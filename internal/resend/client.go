package resend

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

const resendURL = "https://api.resend.com/emails"

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

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

// Send envia o digest por e-mail via Resend.
func (c *Client) Send(subject, digestText string) error {
	from := fmt.Sprintf("%s <%s>", c.cfg.EmailFromName, c.cfg.EmailFrom)

	payload := resendPayload{
		From:    from,
		To:      c.cfg.EmailTo,
		Subject: subject,
		Text:    digestText,
		HTML:    textToHTML(digestText),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serializar payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, resendURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("enviar e-mail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend status %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}

func textToHTML(text string) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><head><meta charset="UTF-8">
<style>
body{font-family:Arial,sans-serif;max-width:680px;margin:0 auto;padding:24px;color:#1a1a1a;line-height:1.7}
h1{font-size:22px;font-weight:600;margin-bottom:4px}
p{margin:0 0 12px}
a{color:#534AB7}
hr{border:none;border-top:1px solid #e5e5e5;margin:20px 0}
.footer{font-size:12px;color:#888;margin-top:32px}
</style></head><body>`)

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			sb.WriteString("<br>")
			continue
		}
		if strings.HasPrefix(line, "---") {
			sb.WriteString("<hr>")
			continue
		}
		sb.WriteString("<p>")
		sb.WriteString(linkify(line))
		sb.WriteString("</p>\n")
	}

	sb.WriteString(`<p class="footer">Gerado automaticamente pelo Agente de Curadoria</p>`)
	sb.WriteString("</body></html>")
	return sb.String()
}

func linkify(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			words[i] = fmt.Sprintf(`<a href="%s">%s</a>`, w, w)
		}
	}
	return strings.Join(words, " ")
}
