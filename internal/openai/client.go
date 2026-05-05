package openai

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

const apiURL = "https://api.openai.com/v1/responses"

type Client struct {
	cfg  *config.Config
	http *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 120 * time.Second},
	}
}

type requestBody struct {
	Model string `json:"model"`
	Tools []tool `json:"tools"`
	Input string `json:"input"`
}

type tool struct {
	Type string `json:"type"`
}

type responseBody struct {
	Output []outputItem `json:"output"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type outputItem struct {
	Type    string          `json:"type"`
	Content []contentBlock  `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// GenerateDigest chama a Responses API da OpenAI e retorna o texto do digest.
func (c *Client) GenerateDigest() (string, error) {
	body := requestBody{
		Model: c.cfg.OpenAIModel,
		Tools: []tool{{Type: "web_search_preview"}},
		Input: buildPrompt(c.cfg),
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
	req.Header.Set("Authorization", "Bearer "+c.cfg.OpenAIAPIKey)

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
	for _, item := range result.Output {
		if item.Type != "message" {
			continue
		}
		for _, block := range item.Content {
			if block.Type == "output_text" && strings.TrimSpace(block.Text) != "" {
				parts = append(parts, block.Text)
			}
		}
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return strings.Join(parts, "\n"), nil
}

func buildPrompt(cfg *config.Config) string {
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
