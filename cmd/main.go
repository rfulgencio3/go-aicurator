package main

import (
	"fmt"
	"log"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/anthropic"
	"github.com/seu-usuario/go-aicurator/internal/config"
	"github.com/seu-usuario/go-aicurator/internal/openai"
	"github.com/seu-usuario/go-aicurator/internal/resend"
	"github.com/seu-usuario/go-aicurator/internal/sendgrid"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	type digester interface {
		GenerateDigest() (string, error)
	}
	var ai digester
	switch cfg.AIProvider {
	case "anthropic":
		ai = anthropic.New(cfg)
	case "openai":
		ai = openai.New(cfg)
	default:
		log.Fatalf("provider de IA não suportado: %s", cfg.AIProvider)
	}

	log.Printf("Gerando digest via %s...", cfg.AIProvider)
	digest, err := ai.GenerateDigest()
	if err != nil {
		log.Fatalf("erro ao gerar digest: %v", err)
	}

	subject := fmt.Sprintf("%s — %s", cfg.EmailFromName, time.Now().Format("02/01/2006"))

	log.Printf("Enviando e-mail para %v via %s...", cfg.EmailTo, cfg.EmailProvider)

	type mailer interface {
		Send(subject, body string) error
	}
	var m mailer
	switch cfg.EmailProvider {
	case "resend":
		m = resend.New(cfg)
	case "sendgrid":
		m = sendgrid.New(cfg)
	default:
		log.Fatalf("provider de e-mail não suportado: %s", cfg.EmailProvider)
	}
	if err := m.Send(subject, digest); err != nil {
		log.Fatalf("erro ao enviar e-mail: %v", err)
	}

	log.Println("Digest enviado com sucesso!")
}
