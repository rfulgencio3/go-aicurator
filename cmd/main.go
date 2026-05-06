package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/anthropic"
	"github.com/seu-usuario/go-aicurator/internal/config"
	"github.com/seu-usuario/go-aicurator/internal/crawler"
	"github.com/seu-usuario/go-aicurator/internal/ghrelease"
	"github.com/seu-usuario/go-aicurator/internal/openai"
	"github.com/seu-usuario/go-aicurator/internal/resend"
	"github.com/seu-usuario/go-aicurator/internal/sendgrid"
	"github.com/seu-usuario/go-aicurator/internal/tts"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	// ── Crawler RSS ──────────────────────────────────────────────────────────
	var articlesCtx string
	if cfg.CrawlEnabled {
		old := crawler.LoadCache(cfg.ArticleCacheFile)
		fresh := crawler.Fetch(cfg)
		all := crawler.Merge(old, fresh)
		if err := crawler.SaveCache(cfg.ArticleCacheFile, all); err != nil {
			log.Printf("crawler: aviso — não foi possível salvar cache: %v", err)
		}
		articlesCtx = crawler.FormatContext(all)
		log.Printf("Crawler: %d artigos no pool (cache: %s)", len(all), cfg.ArticleCacheFile)
	}

	// ── Geração do digest ────────────────────────────────────────────────────
	type digester interface {
		GenerateDigest(articlesCtx string) (string, error)
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
	digest, err := ai.GenerateDigest(articlesCtx)
	if err != nil {
		log.Fatalf("erro ao gerar digest: %v", err)
	}

	// ── Geração de áudio (TTS) ───────────────────────────────────────────────
	if cfg.TTSEnabled {
		ttsClient := tts.New(cfg)
		segments := tts.ParseScript(digest, cfg.TTSNarratorVoice, cfg.TTSAdaVoice, cfg.TTSAlanVoice, cfg.TTSItemLimit)
		log.Printf("TTS: gerando áudio para %d segmentos...", len(segments))

		mp3, err := ttsClient.GenerateMP3(segments)
		if err != nil {
			log.Printf("TTS: aviso — falha na geração de áudio: %v", err)
		} else {
			if err := os.WriteFile(cfg.TTSOutputFile, mp3, 0644); err != nil {
				log.Printf("TTS: aviso — não foi possível salvar %s: %v", cfg.TTSOutputFile, err)
			} else {
				log.Printf("TTS: áudio salvo em %s (%d bytes)", cfg.TTSOutputFile, len(mp3))
			}

			if cfg.GithubToken != "" && cfg.GithubRepository != "" {
				ghClient := ghrelease.New(cfg.GithubToken, cfg.GithubRepository)
				podcastURL, err := ghClient.UploadPodcast(mp3, time.Now())
				if err != nil {
					log.Printf("TTS: aviso — falha no upload do podcast: %v", err)
				} else {
					log.Printf("TTS: podcast publicado em %s", podcastURL)
					digest += "\nPODCAST: " + podcastURL
				}
			}
		}
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
