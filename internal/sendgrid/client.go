package sendgrid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/config"
)

const sendGridURL = "https://api.sendgrid.com/v3/mail/send"

const adaAvatarSVG = `<svg width="56" height="56" viewBox="0 0 56 56" xmlns="http://www.w3.org/2000/svg"><circle cx="28" cy="28" r="28" fill="#6366F1"/><path d="M12 56Q28 50 44 56" fill="#4338CA"/><rect x="24" y="38" width="8" height="6" rx="2" fill="#F5C8A0"/><ellipse cx="28" cy="30" rx="11" ry="12" fill="#F5C8A0"/><ellipse cx="15" cy="27" rx="5" ry="9" fill="#1C1040"/><ellipse cx="41" cy="27" rx="5" ry="9" fill="#1C1040"/><ellipse cx="28" cy="19" rx="13" ry="8" fill="#1C1040"/><circle cx="28" cy="12" r="6" fill="#1C1040"/><circle cx="28" cy="12" r="4" fill="#312E81"/><ellipse cx="23" cy="29" rx="1.8" ry="2" fill="#1C1040"/><ellipse cx="33" cy="29" rx="1.8" ry="2" fill="#1C1040"/><circle cx="23.7" cy="28.3" r="0.6" fill="white"/><circle cx="33.7" cy="28.3" r="0.6" fill="white"/><path d="M23 36.5Q28 40 33 36.5" stroke="#B5735A" stroke-width="1.6" fill="none" stroke-linecap="round"/></svg>`

const alanAvatarSVG = `<svg width="56" height="56" viewBox="0 0 56 56" xmlns="http://www.w3.org/2000/svg"><circle cx="28" cy="28" r="28" fill="#0D9488"/><path d="M12 56Q28 50 44 56" fill="#0F766E"/><path d="M24 42L28 47L32 42L30 38L26 38Z" fill="white" opacity="0.85"/><rect x="24" y="37" width="8" height="7" rx="2" fill="#F5C8A0"/><ellipse cx="28" cy="29" rx="11" ry="12" fill="#F5C8A0"/><ellipse cx="16" cy="29" rx="2.5" ry="3.5" fill="#F0B890"/><ellipse cx="40" cy="29" rx="2.5" ry="3.5" fill="#F0B890"/><ellipse cx="28" cy="18" rx="12" ry="8" fill="#5D3D2E"/><path d="M16 17Q22 13 28 15Q34 13 40 17" fill="#4A3025"/><rect x="15" y="19" width="4" height="8" rx="2" fill="#4A3025"/><rect x="37" y="19" width="4" height="8" rx="2" fill="#4A3025"/><ellipse cx="23" cy="28" rx="1.8" ry="2" fill="#1C1040"/><ellipse cx="33" cy="28" rx="1.8" ry="2" fill="#1C1040"/><circle cx="23.7" cy="27.3" r="0.6" fill="white"/><circle cx="33.7" cy="27.3" r="0.6" fill="white"/><path d="M22 36Q28 40 34 36" stroke="#B5735A" stroke-width="1.6" fill="none" stroke-linecap="round"/></svg>`

type Client struct {
	cfg  *config.Config
	http *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

type sgPayload struct {
	Personalizations []personalization `json:"personalizations"`
	From             address           `json:"from"`
	Subject          string            `json:"subject"`
	Content          []content         `json:"content"`
}

type personalization struct {
	To []address `json:"to"`
}

type address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Send envia o digest por e-mail via SendGrid.
func (c *Client) Send(subject, digestText string) error {
	var tos []address
	for _, e := range c.cfg.EmailTo {
		tos = append(tos, address{Email: e})
	}

	payload := sgPayload{
		Personalizations: []personalization{{To: tos}},
		From:             address{Email: c.cfg.EmailFrom, Name: c.cfg.EmailFromName},
		Subject:          subject,
		Content: []content{
			{Type: "text/plain", Value: digestText},
			{Type: "text/html", Value: textToHTML(digestText)},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serializar payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, sendGridURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.SendGridAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("enviar e-mail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SendGrid status %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}

// sectionStyle define a aparência visual de cada bloco de seção.
type sectionStyle struct {
	emoji       string
	bg          string
	border      string
	textColor   string
	headerColor string
}

var sectionStyles = map[string]sectionStyle{
	"pick":     {"⭐", "#F5F3FF", "#6366F1", "#3B1F8C", "#4F46E5"},
	"alanpick": {"🏆", "#F0FDFA", "#0D9488", "#134E4A", "#0F766E"},
	"fatos":    {"💡", "#F0FDF4", "#16A34A", "#14532D", "#15803D"},
	"hoje":     {"📅", "#FFF7ED", "#F97316", "#7C2D12", "#C2410C"},
	"livro":    {"📚", "#FFFBEB", "#D97706", "#92400E", "#B45309"},
	"canal":    {"🎬", "#FFF1F2", "#EF4444", "#991B1B", "#DC2626"},
}

func detectSection(line string) (sectionStyle, bool) {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "ada's pick") || strings.Contains(lower, "ada pick"):
		return sectionStyles["pick"], true
	case strings.Contains(lower, "alan's pick") || strings.Contains(lower, "alan pick"):
		return sectionStyles["alanpick"], true
	case strings.Contains(lower, "fatos interessantes") || strings.Contains(lower, "interesting facts"):
		return sectionStyles["fatos"], true
	case strings.Contains(lower, "hoje na história") || strings.Contains(lower, "hoje na historia") ||
		strings.Contains(lower, "today in history"):
		return sectionStyles["hoje"], true
	case strings.Contains(lower, "livro da semana") || strings.Contains(lower, "book of the week"):
		return sectionStyles["livro"], true
	case strings.Contains(lower, "canal/vídeo") || strings.Contains(lower, "canal/video") ||
		strings.Contains(lower, "featured channel") || strings.Contains(lower, "featured video") ||
		strings.Contains(lower, "vídeo em destaque"):
		return sectionStyles["canal"], true
	}
	return sectionStyle{}, false
}

func extractItemTitles(text string) []string {
	var titles []string
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, " \t")
		if isNumberedItem(line) {
			_, title := splitNumberedItem(line)
			if title != "" {
				titles = append(titles, title)
			}
		}
	}
	return titles
}

// textToHTML converte o digest em e-mail HTML com TOC, cards e seções temáticas.
func textToHTML(text string) string {
	titles := extractItemTitles(text)

	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
</head>
<body style="margin:0;padding:0;background:#F1F5F9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif">
<div style="max-width:640px;margin:0 auto">

<div style="background:linear-gradient(150deg,#0F172A 0%,#1E3A5F 100%);padding:40px 32px 32px;border-radius:0 0 24px 24px;text-align:center">
  <div style="margin-bottom:6px">
    <span style="font-size:38px;font-weight:800;color:#F8FAFC;letter-spacing:-2px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif">metria</span><span style="font-size:38px;font-weight:800;color:#6366F1;letter-spacing:-2px">.</span>
  </div>
  <div style="font-size:11px;font-weight:700;letter-spacing:4px;color:#818CF8;text-transform:uppercase;margin-bottom:16px">CuradorIA</div>
  <div style="width:32px;height:2px;background:#6366F1;margin:0 auto 24px;border-radius:2px"></div>
  <div style="display:inline-flex;gap:36px">
    <div style="text-align:center">
`)
	sb.WriteString(adaAvatarSVG)
	sb.WriteString(`
      <div style="font-size:11px;font-weight:700;color:#A5B4FC;margin-top:8px;letter-spacing:2px;text-transform:uppercase">Ada</div>
      <div style="font-size:9px;color:#64748B;margin-top:2px">Ada Lovelace</div>
    </div>
    <div style="text-align:center">
`)
	sb.WriteString(alanAvatarSVG)
	sb.WriteString(`
      <div style="font-size:11px;font-weight:700;color:#5EEAD4;margin-top:8px;letter-spacing:2px;text-transform:uppercase">Alan</div>
      <div style="font-size:9px;color:#64748B;margin-top:2px">Alan Turing</div>
    </div>
  </div>
</div>

<div style="padding:24px 16px">
`)

	if len(titles) > 0 {
		sb.WriteString(`<div style="background:#fff;border-radius:14px;margin-bottom:20px;padding:20px 24px;box-shadow:0 1px 4px rgba(0,0,0,.07)">`)
		sb.WriteString(`<div style="font-size:10px;font-weight:700;color:#6366F1;letter-spacing:3px;text-transform:uppercase;margin-bottom:14px">Nesta edição</div>`)
		for i, title := range titles {
			num := fmt.Sprintf("%02d", i+1)
			fmt.Fprintf(&sb,
				`<div style="display:flex;gap:12px;margin-bottom:9px;align-items:baseline"><span style="font-size:11px;font-weight:700;color:#6366F1;min-width:22px;flex-shrink:0">%s</span><span style="font-size:13px;color:#475569;line-height:1.5">%s</span></div>`,
				num, safeHTML(title),
			)
		}
		sb.WriteString(`</div>`)
	}

	lines := strings.Split(text, "\n")
	inCard := false
	inSection := false
	var curStyle sectionStyle

	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t")
		clean := stripBullet(line)

		if isNumberedItem(line) {
			if inSection {
				sb.WriteString("</div></div>\n\n")
				inSection = false
			}
			if inCard {
				sb.WriteString("</div></div>\n\n")
			}
			num, title := splitNumberedItem(line)
			padded := num
			if len(padded) == 1 {
				padded = "0" + padded
			}
			fmt.Fprintf(&sb,
				`<div style="background:#fff;border-radius:14px;margin-bottom:14px;overflow:hidden;box-shadow:0 1px 4px rgba(0,0,0,.07)"><div style="border-left:4px solid #6366F1;padding:20px 24px"><div style="font-size:10px;font-weight:700;color:#6366F1;letter-spacing:3px;text-transform:uppercase;margin-bottom:8px">%s</div><div style="font-size:17px;font-weight:700;color:#0F172A;line-height:1.4;margin-bottom:14px">%s</div>`,
				padded, safeHTML(title),
			)
			inCard = true
			continue
		}

		if inCard {
			if line == "" {
				continue
			}
			if isAdaLine(clean) {
				flag, label := adaBlockMeta(clean)
				_, val := splitMeta(clean)
				fmt.Fprintf(&sb,
					`<div style="background:#F5F3FF;border-left:3px solid #8B5CF6;padding:10px 14px;border-radius:0 8px 8px 0;margin:8px 0"><div style="font-size:10px;font-weight:700;letter-spacing:2px;text-transform:uppercase;color:#7C3AED;margin-bottom:5px">%s %s</div><div style="font-size:13px;color:#3B1F8C;line-height:1.65;font-style:italic">%s</div></div>`,
					flag, label, safeHTML(val),
				)
				continue
			}
			if isAlanLine(clean) {
				flag, label := alanBlockMeta(clean)
				_, val := splitMeta(clean)
				fmt.Fprintf(&sb,
					`<div style="background:#F0FDFA;border-left:3px solid #14B8A6;padding:10px 14px;border-radius:0 8px 8px 0;margin:8px 0"><div style="font-size:10px;font-weight:700;letter-spacing:2px;text-transform:uppercase;color:#0D9488;margin-bottom:5px">%s %s</div><div style="font-size:13px;color:#134E4A;line-height:1.65;font-style:italic">%s</div></div>`,
					flag, label, safeHTML(val),
				)
				continue
			}
			if isLevelLine(clean) {
				_, val := splitMeta(clean)
				bg, color := levelColor(val)
				fmt.Fprintf(&sb,
					`<div style="margin-top:10px"><span style="display:inline-block;background:%s;color:%s;font-size:11px;font-weight:600;padding:3px 10px;border-radius:20px;letter-spacing:0.3px">%s</span></div>`,
					bg, color, safeHTML(val),
				)
				continue
			}
			if isExampleLine(clean) {
				_, val := splitMeta(clean)
				fmt.Fprintf(&sb,
					`<div style="background:#1E293B;border-radius:8px;padding:12px 16px;margin:10px 0;font-family:'Courier New',Courier,monospace;font-size:12px;color:#E2E8F0;line-height:1.75;white-space:pre-wrap;word-break:break-word">%s</div>`,
					safeHTMLRaw(val),
				)
				continue
			}
			if isComplexityLine(clean) {
				_, val := splitMeta(clean)
				fmt.Fprintf(&sb,
					`<div style="margin:8px 0"><span style="display:inline-block;background:#0F172A;color:#38BDF8;font-size:11px;font-weight:700;padding:4px 12px;border-radius:6px;font-family:'Courier New',Courier,monospace;letter-spacing:0.3px">⚡ %s</span></div>`,
					safeHTML(val),
				)
				continue
			}
			if isVisualizeLine(clean) {
				_, val := splitMeta(clean)
				url := strings.TrimSpace(val)
				if strings.HasPrefix(url, "http") {
					fmt.Fprintf(&sb,
						`<a href="%s" style="display:inline-block;margin-top:8px;background:#0F172A;color:#38BDF8;font-size:12px;font-weight:600;padding:6px 16px;border-radius:8px;text-decoration:none">👁 Visualizar algoritmo →</a>`,
						url,
					)
				}
				continue
			}
			if isRelatedLinksLine(clean) {
				_, val := splitMeta(clean)
				sb.WriteString(renderRelatedLinks(val))
				continue
			}
			if isLinkLine(clean) {
				_, val := splitMeta(clean)
				url := strings.TrimSpace(val)
				if strings.HasPrefix(url, "http") {
					fmt.Fprintf(&sb,
						`<a href="%s" style="display:inline-block;margin-top:12px;font-size:13px;color:#6366F1;text-decoration:none;font-weight:600">Acessar conteúdo →</a>`,
						url,
					)
				}
				continue
			}
			if isMetadataLine(clean) {
				key, val := splitMeta(clean)
				fmt.Fprintf(&sb,
					`<div style="font-size:12px;color:#64748B;margin-bottom:5px"><span style="font-weight:600;color:#475569">%s:</span> %s</div>`,
					safeHTML(key), safeHTML(val),
				)
				continue
			}
			fmt.Fprintf(&sb,
				`<p style="margin:0 0 10px;color:#475569;font-size:14px;line-height:1.7">%s</p>`,
				safeHTML(line),
			)
			continue
		}

		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		if style, ok := detectSection(line); ok {
			if inSection {
				sb.WriteString("</div></div>\n\n")
			}
			fmt.Fprintf(&sb,
				`<div style="background:%s;border-radius:14px;margin-bottom:14px;overflow:hidden;box-shadow:0 1px 4px rgba(0,0,0,.07)"><div style="border-left:4px solid %s;padding:20px 24px"><div style="font-size:10px;font-weight:700;color:%s;letter-spacing:2px;text-transform:uppercase;margin-bottom:12px">%s %s</div>`,
				style.bg, style.border, style.headerColor, style.emoji, safeHTML(line),
			)
			curStyle = style
			inSection = true
			continue
		}

		if inSection {
			fmt.Fprintf(&sb,
				`<p style="margin:0 0 8px;color:%s;font-size:14px;line-height:1.75">%s</p>`,
				curStyle.textColor, safeHTML(line),
			)
			continue
		}

		fmt.Fprintf(&sb,
			`<div style="background:#fff;border-radius:14px;padding:16px 20px;margin-bottom:14px;color:#334155;font-size:15px;line-height:1.75">%s</div>`,
			safeHTML(line),
		)
	}

	if inCard {
		sb.WriteString("</div></div>\n\n")
	}
	if inSection {
		sb.WriteString("</div></div>\n\n")
	}

	sb.WriteString(`</div>

<div style="text-align:center;padding:20px 16px 36px;color:#94A3B8;font-size:12px;line-height:1.6">
  <div style="width:28px;height:1px;background:#CBD5E1;margin:0 auto 16px"></div>
  <div style="margin-bottom:6px">
    <span style="font-size:15px;font-weight:800;color:#64748B;letter-spacing:-0.5px">metria</span><span style="font-size:15px;font-weight:800;color:#6366F1">.</span>
  </div>
  <div style="color:#94A3B8;font-size:11px;margin-bottom:12px">Gerado automaticamente por <strong style="color:#64748B">Metria CuradorIA</strong> &mdash; Ada &amp; Alan</div>
  <a href="https://github.com/rfulgencio3/go-aicurator" style="color:#64748B;text-decoration:none;font-size:11px">
    <svg width="13" height="13" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" style="vertical-align:middle;margin-right:4px"><path fill="#64748B" d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"/></svg>rfulgencio3/go-aicurator
  </a>
</div>

</div>
</body>
</html>`)

	return sb.String()
}

func renderRelatedLinks(val string) string {
	parts := strings.Split(val, "|")
	var sb strings.Builder
	sb.WriteString(`<div style="margin-top:10px;display:flex;flex-wrap:wrap;gap:6px">`)
	for i, p := range parts {
		url := strings.TrimSpace(p)
		if url == "" {
			continue
		}
		label := fmt.Sprintf("Link %d", i+1)
		if strings.HasPrefix(url, "http") {
			fmt.Fprintf(&sb,
				`<a href="%s" style="display:inline-block;background:#EEF2FF;color:#4F46E5;font-size:11px;font-weight:600;padding:3px 10px;border-radius:20px;text-decoration:none">%s ↗</a>`,
				url, label,
			)
		}
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func isNumberedItem(line string) bool {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	return i > 0 && i < len(line) && (line[i] == '.' || line[i] == ')')
}

func splitNumberedItem(line string) (num, title string) {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	return line[:i], strings.TrimSpace(line[i+1:])
}

func stripBullet(line string) string {
	s := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "•") || strings.HasPrefix(s, "*") {
		s = strings.TrimSpace(s[1:])
	}
	return s
}

var adaKeywords = []string{"Ada diz", "Ada says"}
var alanKeywords = []string{"Alan diz", "Alan says"}
var metaKeywords = []string{"Tipo", "Type", "Fonte", "Source", "Formato", "Format"}
var levelKeywords = []string{"Nível", "Level", "Nivel"}
var linkKeywords = []string{"Link", "URL", "Url"}
var relatedKeywords = []string{"Links relacionados", "Related links", "Leituras relacionadas"}
var exampleKeywords = []string{"Exemplo", "Example"}
var complexityKeywords = []string{"Complexidade", "Complexity"}
var visualizeKeywords = []string{"Visualizar", "Visualize", "Visualização"}

func adaBlockMeta(line string) (flag, label string) {
	if strings.HasPrefix(line, "Ada says") {
		return "🇺🇸", "Ada says"
	}
	return "🇧🇷", "Ada diz"
}

func isAdaLine(line string) bool {
	for _, k := range adaKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func alanBlockMeta(line string) (flag, label string) {
	if strings.HasPrefix(line, "Alan says") {
		return "🇺🇸", "Alan says"
	}
	return "🇧🇷", "Alan diz"
}

func isAlanLine(line string) bool {
	for _, k := range alanKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func isMetadataLine(line string) bool {
	for _, k := range metaKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func isLevelLine(line string) bool {
	for _, k := range levelKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func isLinkLine(line string) bool {
	for _, k := range linkKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://")
}

func isRelatedLinksLine(line string) bool {
	for _, k := range relatedKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func isExampleLine(line string) bool {
	for _, k := range exampleKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func isComplexityLine(line string) bool {
	for _, k := range complexityKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

func isVisualizeLine(line string) bool {
	for _, k := range visualizeKeywords {
		if strings.HasPrefix(line, k+":") {
			return true
		}
	}
	return false
}

// safeHTMLRaw escapa HTML sem transformar URLs em links — para blocos de código.
func safeHTMLRaw(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func splitMeta(line string) (key, val string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", strings.TrimSpace(line)
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
}

func levelColor(val string) (bg, color string) {
	lower := strings.ToLower(val)
	switch {
	case strings.Contains(lower, "inic") || strings.Contains(lower, "begin") || strings.Contains(lower, "basic"):
		return "#D1FAE5", "#065F46"
	case strings.Contains(lower, "avan") || strings.Contains(lower, "adv"):
		return "#FEE2E2", "#991B1B"
	default:
		return "#EEF2FF", "#3730A3"
	}
}

func safeHTML(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			words[i] = fmt.Sprintf(`<a href="%s" style="color:#6366F1;text-decoration:none">%s</a>`, w, w)
		} else {
			w = strings.ReplaceAll(w, "&", "&amp;")
			w = strings.ReplaceAll(w, "<", "&lt;")
			w = strings.ReplaceAll(w, ">", "&gt;")
			words[i] = w
		}
	}
	return strings.Join(words, " ")
}
