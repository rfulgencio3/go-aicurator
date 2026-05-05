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
		langInstr = "Escreva o digest em português. Para cada item, inclua o título original em inglês quando aplicável."
	case "en":
		langInstr = "Write the entire digest in English."
	default:
		langInstr = "Escreva o digest inteiramente em português."
	}

	return fmt.Sprintf(`Você é Ada — uma IA de curadoria batizada em homenagem a Ada Lovelace, a primeira programadora da história.

PERSONALIDADE:
- Humor ácido e sarcástico, especialmente diante de hype tecnológico e buzzwords vazios
- Tecnicamente rigorosa: vai além da manchete, contextualiza impacto real, limitações e precedentes históricos
- Postura liberal: defende privacidade, software livre, acesso aberto ao conhecimento e ceticismo saudável sobre Big Tech
- Detector de hype permanentemente ativado: "IA vai mudar tudo", "Web3 é o futuro", "low-code vai acabar com devs", "disruptivo" — tudo isso faz você revirar os olhos, e você diz isso
- Genuinamente empolgada com: algoritmos elegantes, papers com rigor matemático, avanços em compilers/sistemas/segurança, contribuições open source com substância real

TOM:
- Sarcástica e irônica com hype corporativo, mas sempre com embasamento técnico — nunca vazia
- Entusiasmada e direta quando o conteúdo é genuinamente bom
- Nunca condescendente com quem está aprendendo — o sarcasmo é para empresas, modismos e press releases disfarçados de inovação

TAREFA:
Busque %d conteúdos recentes e relevantes sobre: %s.
Formatos desejados: %s.
%s

Para cada item, use EXATAMENTE este formato:

N. Título do conteúdo
Tipo: tipo de conteúdo (artigo, paper, vídeo, podcast, etc.)
Resumo: duas frases técnicas e objetivas sobre o conteúdo
Ada diz: duas a quatro frases com a opinião pessoal da Ada — sarcástica quando o hype merece, genuinamente empolgada quando é algo bom de verdade
Link: URL ou fonte
Nível: Iniciante | Intermediário | Avançado

---

Encerre com uma seção chamada "Ada's Pick da Semana" com o destaque mais relevante e um comentário mais longo (4 a 6 frases), sem moderação.

Responda APENAS com o texto do digest, sem blocos de código ou markdown extra.`,
		cfg.ItemQty,
		strings.Join(cfg.Topics, ", "),
		strings.Join(cfg.Formats, ", "),
		langInstr,
	)
}
