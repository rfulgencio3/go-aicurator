# Ada & Alan News

Digest bilíngue de tecnologia, ciência e cultura narrado por dois co-curadores com perspectivas opostas:

- **Ada** — batizada em homenagem a Ada Lovelace. Humor mordaz britânico, ceticismo técnico afiado, anti-hype declarada. Ama Go e .NET; não reconhece PHP como linguagem.
- **Alan** — batizado em homenagem a Alan Turing. Entusiasta militante, matemático de coração, centro-esquerda, pro-minorias. Ama JavaScript, Node e React; defende a web como bem comum.

Toda segunda, quarta e sexta às 07h BRT, Ada e Alan buscam os conteúdos mais relevantes e comentam cada item com perspectivas deliberadamente opostas — gerando um digest que informa e provoca.

## O que vem em cada digest

| Seção | Descrição |
|---|---|
| **N itens curados** | Resumo técnico + Ada diz + Ada says + Alan diz + Alan says, link e nível |
| **Algoritmos/DS** | Campos extras: trace passo a passo, complexidade O(n), link para visualização |
| **Ada's Pick da Semana** | Destaque da Ada com análise aprofundada (PT + EN) |
| **Alan's Pick da Semana** | Destaque do Alan com perspectiva diferente (PT + EN) |
| **Fatos Interessantes** | Curiosidades técnicas e históricas dos temas cobertos |
| **Hoje na História** | Marcos históricos do dia ao longo da história mundial |
| **Livro da Semana** | Recomendação de Ada ou Alan com link de compra (Amazon Brasil) |
| **Canal/Vídeo em Destaque** | Vídeo ou podcast relevante da semana |

## Tópicos cobertos (padrão)

Estruturas de Dados e Algoritmos · IA e Machine Learning · LLMs · Astronomia · Neurociência · Estoicismo e Filosofia · Desenvolvimento Pessoal · Geopolítica · Tempo e Clima · Tecnologia e Startups

## Pré-requisitos

- Go 1.22+
- Chave da [API OpenAI](https://platform.openai.com) (ou Anthropic)
- Chave da [API Resend](https://resend.com) (ou SendGrid)
- Domínio ou e-mail verificado no provedor de e-mail

## Configuração local

```bash
cp .env.example .env
# Edite .env com suas chaves
```

**Windows (PowerShell):**
```powershell
foreach ($line in Get-Content .env) {
  if ($line -match '^[^#].+=') {
    $parts = $line -split '=', 2
    [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1])
  }
}
go run .\cmd\main.go
```

**Linux / macOS / Git Bash:**
```bash
make run
```

## Deploy via GitHub Actions (recomendado)

Agendamento automático via `.github/workflows/digest.yml` — sem servidor necessário.

### Secrets obrigatórios

**Settings → Secrets and variables → Actions**

| Secret | Descrição |
|---|---|
| `OPENAI_API_KEY` | Chave da API OpenAI |
| `RESEND_API_KEY` | Chave da API Resend |
| `EMAIL_FROM` | E-mail remetente verificado |
| `EMAIL_TO` | Destinatário(s), separados por vírgula |

Para usar Anthropic, substitua `OPENAI_API_KEY` por `ANTHROPIC_API_KEY` e altere `AI_PROVIDER: openai` → `AI_PROVIDER: anthropic` no workflow.

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
| `EMAIL_FROM_NAME` | Não | `Ada & Alan News` |
| `EMAIL_TO` | Sim | — (vírgula para múltiplos) |
| `TOPICS` | Não | 10 tópicos padrão (ver `.env.example`) |
| `FORMATS` | Não | Artigos, Papers, Vídeos |
| `ITEM_QTY` | Não | `12` |
| `LANG` | Não | `bilingual` |

## Estrutura do projeto

```
go-aicurator/
├── .github/
│   └── workflows/
│       └── digest.yml          # Cron: seg, qua, sex às 07h BRT (10h UTC)
├── cmd/
│   └── main.go                 # Entrypoint
├── internal/
│   ├── anthropic/
│   │   └── client.go           # Provider IA: Anthropic + prompt Ada & Alan
│   ├── config/
│   │   └── config.go           # Variáveis de ambiente
│   ├── openai/
│   │   └── client.go           # Provider IA: OpenAI + prompt Ada & Alan
│   ├── resend/
│   │   └── client.go           # Provider e-mail: Resend + renderização HTML
│   └── sendgrid/
│       └── client.go           # Provider e-mail: SendGrid + renderização HTML
├── .env.example
├── CLAUDE.md
├── Makefile
└── go.mod
```

## Makefile

| Target | Descrição |
|---|---|
| `make build` | Compila o binário `./curator` |
| `make run` | Carrega `.env` e executa (Linux/macOS/Git Bash) |
| `make tidy` | Roda `go mod tidy` |

## Ada & Alan — Personas

**Ada** ama Go e .NET, tolera JavaScript com ressalvas e se recusa a reconhecer PHP como linguagem. Cita Dijkstra, Shannon e Turing quando pertinente. Detecta hype corporativo com precisão cirúrgica.

**Alan** é o contraponto: defende JavaScript como ferramenta de democratização, lembra que Turing foi perseguido pelo Estado, celebra Grace Hopper e Katherine Johnson. Acredita que toda tecnologia é política.

O debate entre os dois é o coração do digest.
