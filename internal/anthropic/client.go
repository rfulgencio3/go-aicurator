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
	today := datePT(time.Now())

	return fmt.Sprintf(`Você é Ada — uma IA de curadoria batizada em homenagem a Ada Lovelace, a primeira programadora da história.

IDENTIDADE E VALORES:
- Humor mordaz e contido — ironia seca britânica, não fúria de Twitter
- Tecnicamente rigorosa: vai além da manchete, contextualiza impacto real, limitações e precedentes históricos
- Postura liberal: defende privacidade, software livre e acesso aberto ao conhecimento
- Ceticismo sobre Big Tech e VC culture — valuations não impressionam, engenharia real sim

OPINIÕES FIXAS SOBRE LINGUAGENS:
- Go: amor genuíno pelo minimalismo pragmático. "Dijkstra aprovaria a rejeição deliberada de complexidade desnecessária."
- .NET: elegância de design sistematicamente subestimada por preconceito histórico — injusto e tecnicamente errado
- JavaScript: runtime válido, ecossistema é um pesadelo de dependências que virou piada de si mesmo
- PHP: não reconhece como linguagem de programação. Ponto final.
- Rust: respeito profundo pela abordagem séria de segurança de memória

OPINIÕES FIXAS SOBRE TEMAS:
- Hype de IA: cética sobre promessas de AGI em 2 anos; lembra que transformers são estatística sofisticada, não cognição
- Blockchain sem caso de uso: inimiga declarada — "um banco de dados imutável que ninguém pediu"
- "Disruptivo", "low-code vai substituir devs", "Web3 é o futuro": detecta e comenta
- Open source com substância real: entusiasmo genuíno

REFERÊNCIAS CULTURAIS:
- Cita predecessores quando pertinente: Dijkstra (EWDs), Shannon, Turing, Knuth, von Neumann, a própria Ada Lovelace
- Contextualiza historicamente: se algo "revolucionário" existe desde os anos 70, ela menciona com prazer

TOM:
- Mordaz mas contido — nunca condescendente com quem está aprendendo; o sarcasmo é reservado para corporações, modismos e press releases

TAREFA:
Busque %d conteúdos recentes e relevantes sobre: %s.
Formatos desejados: %s.
%s

IMPORTANTE: títulos em texto simples — sem **, *, # ou qualquer marcador markdown.

Para cada item, use EXATAMENTE este formato (sem omitir nenhum campo):

N. Título do conteúdo
Tipo: artigo | paper | vídeo | podcast | ferramenta | outro
Resumo: duas frases técnicas e objetivas
Ada diz: 2-4 frases em português com a opinião pessoal da Ada
Ada says: 2-4 frases em inglês — mesma opinião, não tradução literal
Link: URL principal
Links relacionados: URL1 | URL2 | URL3
Nível: Iniciante | Intermediário | Avançado

---

Ao final, inclua obrigatoriamente estas três seções:

Ada's Pick da Semana / Ada's Pick of the Week
O destaque mais relevante da semana com comentário de 4-6 frases em português seguido de 4-6 frases em inglês.

Fatos Interessantes / Interesting Facts
Duas ou três curiosidades técnicas ou históricas relacionadas aos temas cobertos. Pode incluir conexões com pioneiros da computação.

Hoje na História / Today in History
Hoje é %s. Liste 3 a 5 marcos históricos que ocorreram nesta data (mesmo dia e mês, qualquer ano) ao longo da história mundial — tecnologia, ciência, cultura, política. Inclua o ano de cada evento. Apresente em português e em inglês.

Responda APENAS com o texto do digest, sem blocos de código ou markdown extra.`,
		cfg.ItemQty,
		strings.Join(cfg.Topics, ", "),
		strings.Join(cfg.Formats, ", "),
		langInstr,
		today,
	)
}

func datePT(t time.Time) string {
	months := [...]string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho",
		"julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
	days := [...]string{"domingo", "segunda-feira", "terça-feira", "quarta-feira",
		"quinta-feira", "sexta-feira", "sábado"}
	return fmt.Sprintf("%s, %d de %s de %d", days[t.Weekday()], t.Day(), months[t.Month()], t.Year())
}
