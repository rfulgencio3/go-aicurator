package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Anthropic
	AnthropicAPIKey string
	AnthropicModel  string

	// SendGrid
	SendGridAPIKey  string
	EmailFrom       string
	EmailFromName   string
	EmailTo         []string

	// Curadoria
	Topics  []string
	Formats []string
	ItemQty int
	Lang    string // "pt", "en", "bilingual"
}

func Load() (*Config, error) {
	c := &Config{}

	c.AnthropicAPIKey = mustEnv("ANTHROPIC_API_KEY")
	c.AnthropicModel = envOr("ANTHROPIC_MODEL", "claude-sonnet-4-20250514")

	c.SendGridAPIKey = mustEnv("SENDGRID_API_KEY")
	c.EmailFrom = mustEnv("EMAIL_FROM")
	c.EmailFromName = envOr("EMAIL_FROM_NAME", "Agente de Curadoria")

	toRaw := mustEnv("EMAIL_TO")
	for _, e := range strings.Split(toRaw, ",") {
		if t := strings.TrimSpace(e); t != "" {
			c.EmailTo = append(c.EmailTo, t)
		}
	}

	topicsRaw := envOr("TOPICS", "Inteligência Artificial,Machine Learning,LLMs e modelos de linguagem,Startups de tecnologia")
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

	qty, err := strconv.Atoi(envOr("ITEM_QTY", "8"))
	if err != nil || qty < 1 {
		return nil, fmt.Errorf("ITEM_QTY inválido")
	}
	c.ItemQty = qty

	c.Lang = envOr("LANG", "bilingual")
	if c.Lang != "pt" && c.Lang != "en" && c.Lang != "bilingual" {
		return nil, fmt.Errorf("LANG deve ser pt, en ou bilingual")
	}

	return c, nil
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
