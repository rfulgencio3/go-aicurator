package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/config"
	"github.com/seu-usuario/go-aicurator/internal/email"
)

const ttsURL = "https://api.openai.com/v1/audio/speech"

// Client calls the OpenAI TTS API to generate MP3 audio.
type Client struct {
	cfg  *config.Config
	http *http.Client
}

// Segment is one audio chunk: a speaker voice and the text to read aloud.
type Segment struct {
	Voice string
	Text  string
}

// New returns a TTS client.
func New(cfg *config.Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: 120 * time.Second}}
}

// ParseScript extracts narrator/Ada/Alan segments from the plain-text digest.
// It limits output to the first itemLimit numbered items (0 = no limit).
func ParseScript(digest, narratorVoice, adaVoice, alanVoice string, itemLimit int) []Segment {
	var segments []Segment

	segments = append(segments, Segment{
		Voice: narratorVoice,
		Text:  "Bem-vindos ao Ada e Alan News. A seguir, Ada e Alan comentam as principais notícias curadas para você.",
	})

	type item struct {
		title   string
		resumo  string
		adaDiz  string
		alanDiz string
	}

	var items []item
	var cur *item

	for _, raw := range strings.Split(digest, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Stop at special sections that don't belong in the podcast narrative.
		if email.IsSectionHeader(strings.ToLower(line)) {
			break
		}

		if title, ok := parseItemTitle(line); ok {
			if cur != nil {
				items = append(items, *cur)
			}
			cur = &item{title: title}
			continue
		}

		if cur == nil {
			continue
		}

		if after, ok := cutPrefix(line, "Resumo:"); ok {
			cur.resumo = strings.TrimSpace(after)
		} else if after, ok := cutPrefix(line, "Ada diz:"); ok {
			cur.adaDiz = strings.TrimSpace(after)
		} else if after, ok := cutPrefix(line, "Alan diz:"); ok {
			cur.alanDiz = strings.TrimSpace(after)
		}
	}
	if cur != nil {
		items = append(items, *cur)
	}

	if itemLimit > 0 && len(items) > itemLimit {
		items = items[:itemLimit]
	}

	for i, it := range items {
		intro := fmt.Sprintf("Notícia %d: %s.", i+1, it.title)
		if it.resumo != "" {
			intro += " " + it.resumo
		}
		segments = append(segments, Segment{Voice: narratorVoice, Text: intro})
		if it.adaDiz != "" {
			segments = append(segments, Segment{Voice: adaVoice, Text: it.adaDiz})
		}
		if it.alanDiz != "" {
			segments = append(segments, Segment{Voice: alanVoice, Text: it.alanDiz})
		}
	}

	segments = append(segments, Segment{
		Voice: narratorVoice,
		Text:  "É isso por hoje. Até a próxima edição do Ada e Alan News!",
	})

	return segments
}

// GenerateMP3 synthesizes all segments sequentially and concatenates the MP3 bytes.
func (c *Client) GenerateMP3(segments []Segment) ([]byte, error) {
	var buf bytes.Buffer
	for i, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		mp3, err := c.synthesize(seg.Text, seg.Voice)
		if err != nil {
			return nil, fmt.Errorf("segmento %d (%s): %w", i, seg.Voice, err)
		}
		buf.Write(mp3)
	}
	return buf.Bytes(), nil
}

func (c *Client) synthesize(text, voice string) ([]byte, error) {
	// OpenAI TTS accepts up to 4096 characters per request.
	const maxChars = 4096
	if len(text) > maxChars {
		text = text[:maxChars]
	}

	payload, err := json.Marshal(map[string]any{
		"model": c.cfg.TTSModel,
		"input": text,
		"voice": voice,
	})
	if err != nil {
		return nil, fmt.Errorf("serializar: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, ttsURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.OpenAIAPIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chamar API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS status %d: %s", resp.StatusCode, string(raw))
	}

	return io.ReadAll(resp.Body)
}

// parseItemTitle detects lines like "12. Title text" and returns the title.
func parseItemTitle(line string) (title string, ok bool) {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(line) || (line[i] != '.' && line[i] != ')') {
		return "", false
	}
	if _, err := strconv.Atoi(line[:i]); err != nil {
		return "", false
	}
	t := strings.TrimSpace(line[i+1:])
	if t == "" {
		return "", false
	}
	return t, true
}

func cutPrefix(s, prefix string) (string, bool) {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):], true
	}
	return "", false
}
