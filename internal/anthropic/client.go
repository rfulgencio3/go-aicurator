package anthropic

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

const apiURL = "https://api.anthropic.com/v1/messages"

type Client struct {
	cfg    *config.Config
	http   *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 120 * time.Second},
	}
}

// requestBody é o payload enviado à API.
type requestBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Tools     []tool    `json:"tools"`
	Messages  []message `json:"messages"`
}

type tool struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// responseBody representa a resposta da API.
type responseBody struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateDigest chama a API e retorna o texto do digest pronto.
func (c *Client) GenerateDigest() (string, error) {
	prompt := c.buildPrompt()

	body := requestBody{
		Model:     c.cfg.AnthropicModel,
		MaxTokens: 2048,
		Tools:     []tool{{Type: "web_search_20250305", Name: "web_search"}},
		Messages:  []message{{Role: "user", Content: prompt}},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("serializar request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cfg.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("chamar API: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ler resposta: %w", err)
	}

	var result responseBody
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("deserializar resposta: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	var parts []string
	for _, block := range result.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			parts = append(parts, block.Text)
		}
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return strings.Join(parts, "\n"), nil
}

func (c *Client) buildPrompt() string {
	cfg := c.cfg

	var langInstr string
	switch cfg.Lang {
	case "bilingual":
		langInstr = "Escreva o digest em português. Para cada item, inclua o título original em inglês quando aplicável. Use 🇧🇷 e 🇺🇸 para indicar o idioma de origem."
	case "en":
		langInstr = "Write the entire digest in English."
	default:
		langInstr = "Escreva o digest inteiramente em português."
	}

	return fmt.Sprintf(`Você é um agente de curadoria de conteúdo especializado em tecnologia e IA.

Busque e cuide de %d conteúdos recentes e relevantes sobre: %s.

Formatos desejados: %s.

%s

Para cada item inclua:
- Título claro e descritivo
- Tipo de conteúdo (artigo, paper, vídeo, etc.)
- Resumo em 2-3 frases explicando por que é relevante
- Link ou fonte
- Nível: Iniciante / Intermediário / Avançado

Organize com uma introdução curta, os itens numerados e encerre com uma tendência ou destaque da semana.

Responda APENAS com o texto do digest, sem blocos de código ou markdown extra.`,
		cfg.ItemQty,
		strings.Join(cfg.Topics, ", "),
		strings.Join(cfg.Formats, ", "),
		langInstr,
	)
}
