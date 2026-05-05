# CLAUDE.md — Guideline para IA (Claude Code / VS Code)

> Este arquivo fornece contexto e regras para assistentes de IA que trabalham neste repositório.
> Leia-o inteiramente antes de sugerir qualquer mudança de código.

---

## Visão geral do projeto

**go-aicurator** é um agente de curadoria de conteúdo escrito em Go, personificado pela **Ada** — uma IA com personalidade, humor mordaz e opiniões técnicas firmes, batizada em homenagem a Ada Lovelace.

O agente busca conteúdos recentes sobre Tecnologia & IA, gera um digest bilíngue com comentários da Ada e envia por e-mail. Roda de forma autônoma via GitHub Actions.

**Fluxo principal:**
```
GitHub Actions (cron) → cmd/main.go → config.Load() → [openai|anthropic].GenerateDigest() → [resend|sendgrid].Send()
```

---

## Stack e dependências

| Item | Detalhe |
|---|---|
| Linguagem | Go 1.22+ |
| Módulo | `github.com/seu-usuario/go-aicurator` |
| Provider de IA | OpenAI (padrão, `gpt-4o` + `web_search_preview`) ou Anthropic (`claude-sonnet-4-20250514` + `web_search_20250305`) |
| Provider de e-mail | Resend (padrão) ou SendGrid |
| Agendamento | GitHub Actions (`.github/workflows/digest.yml`) |
| Dependências externas | Nenhuma — apenas stdlib Go |

O projeto **não usa nenhum pacote externo** (`go.mod` sem `require`). Mantenha assim sempre que possível.

---

## Estrutura de diretórios

```
go-aicurator/
├── .github/
│   └── workflows/
│       └── digest.yml           # Cron: seg, qua, sex às 07h BRT (10h UTC)
├── cmd/
│   └── main.go                  # Entrypoint — orquestra os pacotes internos
├── internal/
│   ├── anthropic/
│   │   └── client.go            # Provider IA Anthropic + prompt da Ada
│   ├── config/
│   │   └── config.go            # Carrega e valida variáveis de ambiente
│   ├── openai/
│   │   └── client.go            # Provider IA OpenAI + prompt da Ada
│   ├── resend/
│   │   └── client.go            # Provider e-mail Resend + renderização HTML
│   └── sendgrid/
│       └── client.go            # Provider e-mail SendGrid + renderização HTML
├── .env.example
├── .gitignore
├── Makefile
├── CLAUDE.md
└── go.mod
```

### Regra de organização

- `cmd/` contém apenas o entrypoint. Lógica de negócio **nunca** vai aqui.
- `internal/` é privado ao módulo. Não crie pacotes públicos sem necessidade.
- Cada pacote em `internal/` tem responsabilidade única. Não misture concerns.
- Novos pacotes devem entrar em `internal/<nome>/`.

---

## Convenções de código Go

### Estilo geral
- Siga `gofmt` e `goimports` — sem exceções.
- Nomes de variáveis e funções em `camelCase`, tipos em `PascalCase`.
- Comentários de exported functions em inglês (`// FunctionName does X`).
- Comentários internos (lógica, decisões) em português é aceitável.
- Erros sempre em minúsculo e sem ponto final: `fmt.Errorf("carregar config: %w", err)`.

### Tratamento de erros
- **Nunca** ignore erros com `_`.
- Use `%w` para wrapping: `fmt.Errorf("contexto: %w", err)`.
- `log.Fatalf` apenas no `cmd/main.go`. Pacotes internos retornam `error`.
- Não use `panic` fora de testes.

### Configuração
- Toda configuração vem de variáveis de ambiente via `internal/config`.
- **Nenhum valor hardcoded** de URL, chave, modelo ou endereço de e-mail fora de `config.go`.
- Novos parâmetros configuráveis devem ser adicionados à struct `Config` com fallback via `envOr()`.

### HTTP e I/O
- Sempre defina `Timeout` em clientes `http.Client`.
- Sempre feche `resp.Body` com `defer`.
- Leia o corpo completo antes de fechar para reusar conexões.

---

## Variáveis de ambiente

Todas documentadas em `.env.example`. **Nunca** commite valores reais.

| Variável | Obrigatória | Padrão |
|---|---|---|
| `AI_PROVIDER` | Não | `openai` |
| `OPENAI_API_KEY` | Sim (se `AI_PROVIDER=openai`) | — |
| `OPENAI_MODEL` | Não | `gpt-4o` |
| `ANTHROPIC_API_KEY` | Sim (se `AI_PROVIDER=anthropic`) | — |
| `ANTHROPIC_MODEL` | Não | `claude-sonnet-4-20250514` |
| `EMAIL_PROVIDER` | Não | `resend` |
| `RESEND_API_KEY` | Sim (se `EMAIL_PROVIDER=resend`) | — |
| `SENDGRID_API_KEY` | Sim (se `EMAIL_PROVIDER=sendgrid`) | — |
| `EMAIL_FROM` | Sim | — |
| `EMAIL_FROM_NAME` | Não | `Agente de Curadoria` |
| `EMAIL_TO` | Sim | — (vírgula para múltiplos) |
| `TOPICS` | Não | IA, ML, LLMs, Startups |
| `FORMATS` | Não | Artigos, Papers, Vídeos |
| `ITEM_QTY` | Não | `12` |
| `LANG` | Não | `bilingual` |

---

## Segurança — regras inegociáveis

- `.env` está no `.gitignore`. **Jamais remova essa entrada.**
- Nunca logue valores de API keys ou qualquer secret.
- Nunca inclua secrets em mensagens de erro retornadas ao caller.
- Se adicionar novos campos sensíveis à `Config`, certifique-se de que não aparecem em `fmt.Sprintf("%+v", cfg)`.

---

## Como rodar localmente

```bash
# 1. Copiar e preencher variáveis
cp .env.example .env

# 2. Testar sem compilar (Linux/macOS/Git Bash)
make run

# 3. Compilar
make build
```

**No Windows (PowerShell), carregue as variáveis manualmente antes de rodar:**
```powershell
foreach ($line in Get-Content .env) {
  if ($line -match '^[^#].+=') {
    $parts = $line -split '=', 2
    [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1])
  }
}
go run .\cmd\main.go
```

---

## Prompt da Ada

O prompt é construído em `buildPrompt()` — presente em `internal/openai/client.go` e `internal/anthropic/client.go`. Ambos devem ser mantidos em sincronia.

**Ada é:**
- Uma IA com humor mordaz britânico e opiniões técnicas firmes
- Ama Go e .NET; desconfia do ecossistema JavaScript; não reconhece PHP
- Cita Dijkstra, Shannon, Turing, Knuth e von Neumann quando pertinente
- Anti-hype: sinaliza buzzwords e promessas infundadas
- Liberal: pro-privacidade, pro-open source

**Ao modificar o prompt:**
- Mantenha o formato estruturado por item (Tipo, Resumo, Ada diz, Ada says, Link, Nível)
- Mantenha as seções fixas ao final: Ada's Pick, Fatos Interessantes, Hoje na História
- Não remova a instrução de retornar apenas texto plano (sem markdown extra)
- A data atual é injetada via `datePT(time.Now())` — não hardcode datas
- Teste com `make run` antes de commitar

---

## Renderização HTML do e-mail

A função `textToHTML()` fica em cada cliente de e-mail (`resend/client.go`, `sendgrid/client.go`) e deve ser mantida em sincronia entre os dois.

O parser detecta:
- Itens numerados → cards com borda esquerda roxa
- `Ada diz:` / `Ada says:` → bloco roxo com bandeira PT/EN
- `Nível:` → badge colorido (verde/roxo/vermelho)
- `Link:` → botão "Acessar conteúdo →"
- Cabeçalhos de seção (Ada's Pick, Fatos Interessantes, Hoje na História) → bloco temático

---

## Adicionando novas funcionalidades

### Novo provider de IA
1. Crie `internal/<provider>/client.go` com método `GenerateDigest() (string, error)`.
2. Adicione as variáveis em `internal/config/config.go` no switch de `AI_PROVIDER`.
3. Instancie no `cmd/main.go` no switch da interface `digester`.

### Novo provider de e-mail
1. Crie `internal/<provider>/client.go` com método `Send(subject, body string) error`.
2. Adicione as variáveis em `internal/config/config.go` no switch de `EMAIL_PROVIDER`.
3. Instancie no `cmd/main.go` no switch da interface `mailer`.
4. Implemente `textToHTML()` com o mesmo comportamento de `resend/client.go`.

### Novo tópico ou formato
Apenas atualize `.env` — nenhuma mudança de código necessária.

---

## O que a IA **não deve** fazer neste projeto

- Adicionar dependências externas sem necessidade clara.
- Mover lógica de negócio para `cmd/main.go`.
- Usar `os.Getenv` diretamente fora de `internal/config`.
- Criar arquivos `.env` com valores reais.
- Sugerir `panic` como tratamento de erro em produção.
- Alterar o `.gitignore` de forma que exponha secrets.
- Desincronizar `buildPrompt()` entre `openai/client.go` e `anthropic/client.go`.
- Desincronizar `textToHTML()` entre `resend/client.go` e `sendgrid/client.go`.
