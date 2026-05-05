package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// IA
	AIProvider      string // "anthropic" | "openai"
	AnthropicAPIKey string
	AnthropicModel  string
	OpenAIAPIKey    string
	OpenAIModel     string

	// E-mail
	EmailProvider  string // "resend" | "sendgrid"
	ResendAPIKey   string
	SendGridAPIKey string
	EmailFrom      string
	EmailFromName  string
	EmailTo        []string

	// Curadoria
	Topics  []string
	Formats []string
	ItemQty int
	Lang    string // "pt", "en", "bilingual"
}

func Load() (*Config, error) {
	c := &Config{}

	c.AIProvider = envOr("AI_PROVIDER", "openai")
	switch c.AIProvider {
	case "anthropic":
		c.AnthropicAPIKey = mustEnv("ANTHROPIC_API_KEY")
		c.AnthropicModel = envOr("ANTHROPIC_MODEL", "claude-sonnet-4-20250514")
	case "openai":
		c.OpenAIAPIKey = mustEnv("OPENAI_API_KEY")
		c.OpenAIModel = envOr("OPENAI_MODEL", "gpt-4o")
	default:
		return nil, fmt.Errorf("AI_PROVIDER deve ser anthropic ou openai")
	}

	c.EmailProvider = envOr("EMAIL_PROVIDER", "resend")
	switch c.EmailProvider {
	case "resend":
		c.ResendAPIKey = mustEnv("RESEND_API_KEY")
	case "sendgrid":
		c.SendGridAPIKey = mustEnv("SENDGRID_API_KEY")
	default:
		return nil, fmt.Errorf("EMAIL_PROVIDER deve ser resend ou sendgrid")
	}
	c.EmailFrom = mustEnv("EMAIL_FROM")
	c.EmailFromName = envOr("EMAIL_FROM_NAME", "Ada & Alan News")

	toRaw := mustEnv("EMAIL_TO")
	for _, e := range strings.Split(toRaw, ",") {
		if t := strings.TrimSpace(e); t != "" {
			c.EmailTo = append(c.EmailTo, t)
		}
	}

	topicsRaw := envOr("TOPICS", "Estruturas de Dados e Algoritmos,Inteligência Artificial e Machine Learning,LLMs e Modelos de Linguagem,Astronomia e Exploração Espacial,Neurociência e Comportamento Humano,Estoicismo e Filosofia Prática,Desenvolvimento Pessoal e Performance,Geopolítica e Relações Internacionais,Tempo e Clima,Tecnologia e Startups")
	for _, t := range strings.Split(topicsRaw, ",") {
		if v := strings.TrimSpace(t); v != "" {
			c.Topics = append(c.Topics, v)
		}
	}

	formatsRaw := envOr("FORMATS", "Artigos e notícias,Papers acadêmicos,Vídeos e podcasts")
	for _, f := range strings.Split(formatsRaw, ",") {
		if v := strings.TrimSpace(f); v != "" {
			c.Formats = append(c.Formats, v)
		}
	}

	qty, err := strconv.Atoi(envOr("ITEM_QTY", "12"))
	if err != nil || qty < 1 {
		return nil, fmt.Errorf("ITEM_QTY inválido")
	}
	c.ItemQty = qty

	lang, err := digestLang()
	if err != nil {
		return nil, err
	}
	c.Lang = lang

	return c, nil
}

func digestLang() (string, error) {
	if v := strings.TrimSpace(os.Getenv("DIGEST_LANG")); v != "" {
		if isDigestLang(v) {
			return v, nil
		}
		return "", fmt.Errorf("DIGEST_LANG deve ser pt, en ou bilingual")
	}

	if v := strings.TrimSpace(os.Getenv("LANG")); v != "" {
		if isDigestLang(v) {
			return v, nil
		}
		if isLocaleLang(v) {
			return "bilingual", nil
		}
		return "", fmt.Errorf("LANG deve ser pt, en ou bilingual")
	}

	return "bilingual", nil
}

func isDigestLang(v string) bool {
	return v == "pt" || v == "en" || v == "bilingual"
}

func isLocaleLang(v string) bool {
	upper := strings.ToUpper(v)
	return upper == "C" || upper == "POSIX" || strings.ContainsAny(v, "._")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "variável de ambiente obrigatória não definida: %s\n", key)
		os.Exit(1)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
