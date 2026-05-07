package crawler

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/config"
)

// Article representa um artigo coletado de um feed RSS/Atom.
type Article struct {
	Title   string    `json:"title"`
	URL     string    `json:"url"`
	Source  string    `json:"source"`
	PubDate time.Time `json:"pub_date"`
	Summary string    `json:"summary"`
}

// Feeds padrão usados quando RSS_FEEDS não está configurado.
var defaultFeeds = []string{
	"https://news.ycombinator.com/rss",
	"https://arxiv.org/rss/cs.AI",
	"https://arxiv.org/rss/cs.LG",
	"https://www.nasa.gov/rss/dyn/breaking_news.rss",
	"https://feeds.arstechnica.com/arstechnica/index",
	"https://www.theverge.com/rss/index.xml",
	"https://www.technologyreview.com/feed/",
	"https://techcrunch.com/feed/",
	"https://rss.tecmundo.com.br/feed",
	"https://canaltech.com.br/rss/",
}

// ── RSS 2.0 ──────────────────────────────────────────────────────────────────

type rssFeed struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	GUID    string `xml:"guid"`
	PubDate string `xml:"pubDate"`
	Desc    string `xml:"description"`
}

// ── Atom ─────────────────────────────────────────────────────────────────────

type atomFeed struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	ID        string     `xml:"id"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// ── Date parsing ──────────────────────────────────────────────────────────────

var dateFmts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05+00:00",
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, f := range dateFmts {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	if len(out) > 220 {
		out = out[:220] + "…"
	}
	return out
}

// normalizeXML strips XML namespace declarations so encoding/xml can parse
// Atom feeds with or without the Atom namespace uniformly.
func normalizeXML(raw []byte) []byte {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			t.Name.Space = ""
			var attrs []xml.Attr
			for _, a := range t.Attr {
				if a.Name.Local == "xmlns" || strings.HasPrefix(a.Name.Local, "xmlns:") ||
					a.Name.Space == "xmlns" {
					continue
				}
				a.Name.Space = ""
				attrs = append(attrs, a)
			}
			t.Attr = attrs
			_ = enc.EncodeToken(t)
		case xml.EndElement:
			t.Name.Space = ""
			_ = enc.EncodeToken(t)
		default:
			_ = enc.EncodeToken(tok)
		}
	}
	_ = enc.Flush()
	return buf.Bytes()
}

// ── Feed fetching ─────────────────────────────────────────────────────────────

func fetchOneFeed(feedURL string, maxAge time.Duration, client *http.Client) ([]Article, error) {
	req, err := http.NewRequest(http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-aicurator/1.0; +https://github.com/rfulgencio3/go-aicurator)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2 MB cap
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-maxAge)
	norm := normalizeXML(raw)

	// Try RSS 2.0
	var rss rssFeed
	if xml.Unmarshal(norm, &rss) == nil && len(rss.Channel.Items) > 0 {
		source := strings.TrimSpace(rss.Channel.Title)
		if source == "" {
			source = feedURL
		}
		var out []Article
		for _, item := range rss.Channel.Items {
			pub := parseDate(item.PubDate)
			if !pub.IsZero() && pub.Before(cutoff) {
				continue
			}
			u := strings.TrimSpace(item.Link)
			if u == "" {
				u = strings.TrimSpace(item.GUID)
			}
			if !strings.HasPrefix(u, "http") {
				continue
			}
			out = append(out, Article{
				Title:   strings.TrimSpace(item.Title),
				URL:     u,
				Source:  source,
				PubDate: pub,
				Summary: stripHTML(item.Desc),
			})
		}
		return out, nil
	}

	// Try Atom
	var atom atomFeed
	if xml.Unmarshal(norm, &atom) == nil && len(atom.Entries) > 0 {
		source := strings.TrimSpace(atom.Title)
		if source == "" {
			source = feedURL
		}
		var out []Article
		for _, entry := range atom.Entries {
			dateStr := entry.Published
			if dateStr == "" {
				dateStr = entry.Updated
			}
			pub := parseDate(dateStr)
			if !pub.IsZero() && pub.Before(cutoff) {
				continue
			}
			var u string
			for _, l := range entry.Links {
				if l.Rel == "alternate" || l.Rel == "" {
					u = l.Href
					break
				}
			}
			if u == "" && len(entry.Links) > 0 {
				u = entry.Links[0].Href
			}
			if u == "" && strings.HasPrefix(entry.ID, "http") {
				u = entry.ID
			}
			if !strings.HasPrefix(u, "http") {
				continue
			}
			summary := entry.Summary
			if summary == "" {
				summary = entry.Content
			}
			out = append(out, Article{
				Title:   strings.TrimSpace(entry.Title),
				URL:     u,
				Source:  source,
				PubDate: pub,
				Summary: stripHTML(summary),
			})
		}
		return out, nil
	}

	return nil, fmt.Errorf("formato não reconhecido")
}

// ── Public API ────────────────────────────────────────────────────────────────

// Fetch busca todos os feeds configurados em paralelo e retorna artigos recentes,
// deduplicados e ordenados do mais novo ao mais antigo.
func Fetch(cfg *config.Config) []Article {
	feeds := cfg.RSSFeeds
	if len(feeds) == 0 {
		feeds = defaultFeeds
	}
	maxAge := time.Duration(cfg.CrawlMaxAgeDays) * 24 * time.Hour
	client := &http.Client{Timeout: 15 * time.Second}

	var mu sync.Mutex
	var wg sync.WaitGroup
	seen := make(map[string]bool)
	var all []Article

	for _, feedURL := range feeds {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			articles, err := fetchOneFeed(u, maxAge, client)
			if err != nil {
				log.Printf("crawler: %s: %v", u, err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, a := range articles {
				if !seen[a.URL] {
					seen[a.URL] = true
					all = append(all, a)
				}
			}
		}(feedURL)
	}
	wg.Wait()

	sort.Slice(all, func(i, j int) bool {
		return all[i].PubDate.After(all[j].PubDate)
	})
	if cfg.CrawlMaxItems > 0 && len(all) > cfg.CrawlMaxItems {
		all = all[:cfg.CrawlMaxItems]
	}
	return all
}

// LoadCache carrega artigos do arquivo JSON. Retorna nil se o arquivo não existir.
func LoadCache(path string) []Article {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var articles []Article
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil
	}
	return articles
}

// SaveCache persiste artigos no arquivo JSON.
func SaveCache(path string, articles []Article) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(articles, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar cache: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Merge combina artigos anteriores com os recém-buscados, deduplicando por URL.
// Artigos frescos têm prioridade; o resultado é ordenado do mais novo ao mais antigo.
func Merge(old, fresh []Article) []Article {
	seen := make(map[string]bool)
	var result []Article
	for _, a := range fresh {
		if !seen[a.URL] {
			seen[a.URL] = true
			result = append(result, a)
		}
	}
	for _, a := range old {
		if !seen[a.URL] {
			seen[a.URL] = true
			result = append(result, a)
		}
	}
	return result
}

// FormatContext formata artigos como bloco de texto para injeção no prompt do LLM.
func FormatContext(articles []Article) string {
	if len(articles) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, a := range articles {
		date := "data desconhecida"
		if !a.PubDate.IsZero() {
			date = a.PubDate.Format("02/01/2006")
		}
		fmt.Fprintf(&sb, "%d. %s\n   Fonte: %s | Data: %s | URL: %s\n",
			i+1, a.Title, a.Source, date, a.URL,
		)
		if a.Summary != "" {
			fmt.Fprintf(&sb, "   Resumo: %s\n", a.Summary)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
