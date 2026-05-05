package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/anthropic"
	"github.com/seu-usuario/go-aicurator/internal/config"
	"github.com/seu-usuario/go-aicurator/internal/resend"
	"github.com/seu-usuario/go-aicurator/internal/sendgrid"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	log.Println("Gerando digest via Claude...")
	claude := anthropic.New(cfg)
	digest, err := claude.GenerateDigest()
	if err != nil {
		log.Fatalf("erro ao gerar digest: %v", err)
	}

	subject := fmt.Sprintf("Curadoria Tecnologia & IA — %s", time.Now().Format("02/01/2006"))

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
	}
	if err := m.Send(subject, digest); err != nil {
		log.Fatalf("erro ao enviar e-mail: %v", err)
	}

	log.Println("Digest enviado com sucesso!")
	os.Exit(0)
}
