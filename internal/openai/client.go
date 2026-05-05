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
