# Metria CuradorIA

**Ada** é uma IA de curadoria com personalidade: humor mordaz britânico, opiniões técnicas firmes e um detector de hype permanentemente ativado — batizada em homenagem a Ada Lovelace.

Toda segunda, quarta e sexta às 07h BRT, Ada busca os conteúdos mais relevantes sobre Tecnologia & IA, comenta com ironia (quando merece) e envia um digest bilíngue por e-mail.

## O que vem em cada digest

- **12 itens** com resumo técnico + comentário pessoal da Ada em PT 🇧🇷 e EN 🇺🇸
- **Ada's Pick da Semana** — o destaque com análise mais profunda
- **Fatos Interessantes** — curiosidades técnicas e históricas da semana
- **Hoje na História** — marcos históricos da data de hoje

## Pré-requisitos

- Go 1.22+
- Chave da [API OpenAI](https://platform.openai.com) (ou Anthropic)
- Chave da [API Resend](https://resend.com) (ou SendGrid)
- Domínio ou e-mail verificado no Resend

## Configuração local

```bash
cp .env.example .env
# Edite .env com suas chaves
```

**Windows (PowerShell):**
```powershell
foreach ($line in Get-Content .env) { if ($line -match '^[^#]') { [System.Environment]::SetEnvironmentVariable(($line -split '=')[0], ($line -split '=',2)[1]) } }
go run .\cmd\main.go
```

**Linux / macOS / Git Bash:**
```bash
make run
```

## Deploy via GitHub Actions (recomendado)

O agendamento roda automaticamente via `.github/workflows/digest.yml` — sem servidor necessário.

### Secrets obrigatórios

No repositório: **Settings → Secrets and variables → Actions**

| Secret | Descrição |
|---|---|
| `OPENAI_API_KEY` | Chave da API OpenAI |
| `RESEND_API_KEY` | Chave da API Resend |
| `EMAIL_FROM` | E-mail remetente verificado no Resend |
| `EMAIL_TO` | Destinatário(s), separados por vírgula |

Para usar Anthropic no lugar da OpenAI, substitua `OPENAI_API_KEY` por `ANTHROPIC_API_KEY` e altere `AI_PROVIDER: openai` → `AI_PROVIDER: anthropic` no workflow.

### Disparo manual

**Actions → Digest Curadoria IA → Run workflow**

## Variáveis de ambiente

| Variável | Obrigatória | Padrão |
|---|---|---|
| `AI_PROVIDER` | Não | `openai` |
| `OPENAI_API_KEY` | Sim (se openai) | — |
| `OPENAI_MODEL` | Não | `gpt-4o` |
| `ANTHROPIC_API_KEY` | Sim (se anthropic) | — |
| `ANTHROPIC_MODEL` | Não | `claude-sonnet-4-20250514` |
| `EMAIL_PROVIDER` | Não | `resend` |
| `RESEND_API_KEY` | Sim (se resend) | — |
| `SENDGRID_API_KEY` | Sim (se sendgrid) | — |
| `EMAIL_FROM` | Sim | — |
| `EMAIL_FROM_NAME` | Não | `Agente de Curadoria` |
| `EMAIL_TO` | Sim | — |
| `TOPICS` | Não | IA, ML, LLMs, Startups |
| `FORMATS` | Não | Artigos, Papers, Vídeos |
| `ITEM_QTY` | Não | `12` |
| `LANG` | Não | `bilingual` |

## Estrutura do projeto

```
go-aicurator/
├── .github/
│   └── workflows/
│       └── digest.yml          # Agendamento via GitHub Actions
├── cmd/
│   └── main.go                 # Entrypoint
├── internal/
│   ├── anthropic/
│   │   └── client.go           # Provider IA: Anthropic + prompt Ada
│   ├── config/
│   │   └── config.go           # Variáveis de ambiente
│   ├── openai/
│   │   └── client.go           # Provider IA: OpenAI + prompt Ada
│   ├── resend/
│   │   └── client.go           # Provider e-mail: Resend
│   └── sendgrid/
│       └── client.go           # Provider e-mail: SendGrid
├── .env.example
├── Makefile
└── go.mod
```

## Makefile

| Target | Descrição |
|---|---|
| `make build` | Compila o binário `./curator` |
| `make run` | Carrega `.env` e executa (Linux/macOS/Git Bash) |
| `make tidy` | Roda `go mod tidy` |
| `make install-cron` | Instala cron local (Linux/macOS) |

## Ada — Persona

Ada ama Go e .NET, tolera JavaScript com ressalvas e se recusa a reconhecer PHP como linguagem de programação. Cita Dijkstra, Shannon e Turing quando pertinente. Detecta hype corporativo com precisão cirúrgica e tem opiniões sobre tudo — sempre embasadas tecnicamente.
