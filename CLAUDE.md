# CLAUDE.md — Guideline para IA (Claude Code / VS Code)

> Este arquivo fornece contexto e regras para assistentes de IA que trabalham neste repositório.
> Leia-o inteiramente antes de sugerir qualquer mudança de código.

---

## Visão geral do projeto

**go-aicurator** é um agente de curadoria de conteúdo escrito em Go.

Ele usa a API do Claude (com web search ativo) para buscar conteúdos recentes sobre Tecnologia & IA e envia um digest bilíngue por e-mail via SendGrid. É projetado para rodar de forma autônoma via cron job.

**Fluxo principal:**
```
cron → cmd/main.go → config.Load() → anthropic.GenerateDigest() → sendgrid.Send()
```

---

## Stack e dependências

| Item | Detalhe |
|---|---|
| Linguagem | Go 1.22+ |
| Módulo | `github.com/seu-usuario/go-aicurator` |
| API de IA | Anthropic (`claude-sonnet-4-20250514`) com tool `web_search_20250305` |
| E-mail | SendGrid v3 REST API |
| Agendamento | Cron do sistema operacional |
| Dependências externas | Nenhuma — apenas stdlib Go |

O projeto **não usa nenhum pacote externo** (`go.mod` sem `require`). Mantenha assim sempre que possível. Se precisar adicionar uma dependência, justifique no PR.

---

## Estrutura de diretórios

```
go-aicurator/
├── cmd/
│   └── main.go                  # Entrypoint — orquestra os pacotes internos
├── internal/
│   ├── anthropic/
│   │   └── client.go            # Cliente da API Anthropic + construção do prompt
│   ├── config/
│   │   └── config.go            # Carrega e valida variáveis de ambiente
│   └── sendgrid/
│       └── client.go            # Cliente SendGrid + conversão texto → HTML
├── .env.example                 # Variáveis necessárias (sem valores reais)
├── .gitignore                   # Bloqueia .env e binários
├── Makefile                     # Targets: build, run, tidy, install-cron
├── CLAUDE.md                    # Este arquivo
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

Todas as variáveis estão documentadas em `.env.example`. **Nunca** commite valores reais.

| Variável | Obrigatória | Padrão |
|---|---|---|
| `ANTHROPIC_API_KEY` | Sim | — |
| `ANTHROPIC_MODEL` | Não | `claude-sonnet-4-20250514` |
| `SENDGRID_API_KEY` | Sim | — |
| `EMAIL_FROM` | Sim | — |
| `EMAIL_FROM_NAME` | Não | `Agente de Curadoria` |
| `EMAIL_TO` | Sim | — (vírgula para múltiplos) |
| `TOPICS` | Não | IA, ML, LLMs, Startups |
| `FORMATS` | Não | Artigos, Papers, Vídeos |
| `ITEM_QTY` | Não | `8` |
| `LANG` | Não | `bilingual` |

---

## Segurança — regras inegociáveis

- `.env` está no `.gitignore`. **Jamais remova essa entrada.**
- Nunca logue o valor de `APIKey`, `SendGridAPIKey` ou qualquer secret.
- Nunca inclua secrets em mensagens de erro retornadas ao caller.
- Se adicionar novos campos sensíveis à `Config`, certifique-se de que não aparecem em `fmt.Sprintf("%+v", cfg)` — use um método `String()` customizado se necessário.

---

## Como rodar localmente

```bash
# 1. Copiar e preencher variáveis
cp .env.example .env

# 2. Testar sem compilar
make run

# 3. Compilar
make build

# 4. Instalar cron (07h seg, qua, sex)
make install-cron

# 5. Ver logs
tail -f /tmp/aicurator.log
```

---

## Makefile targets

| Target | O que faz |
|---|---|
| `make build` | Compila o binário `./aicurator` |
| `make run` | Carrega `.env` e executa via `go run` |
| `make tidy` | Roda `go mod tidy` |
| `make install-cron` | Compila e instala entrada no crontab |

---

## Prompt do agente (anthropic/client.go)

O prompt é construído dinamicamente em `buildPrompt()` a partir da config. Ao modificá-lo:

- Mantenha a instrução de retornar **apenas o texto do digest** (sem markdown extra, sem blocos de código).
- Não remova a instrução de nível de profundidade (Iniciante / Intermediário / Avançado).
- Teste com `make run` antes de commitar — o resultado vai direto para o e-mail.
- O modelo usa a tool `web_search_20250305`. Não remova ela do payload — sem ela o digest não terá conteúdo atualizado.

---

## Adicionando novas funcionalidades

### Novo provedor de e-mail
1. Crie `internal/<provedor>/client.go` com a mesma assinatura: `Send(subject, body string) error`.
2. Adicione as variáveis necessárias em `internal/config/config.go`.
3. Instancie no `cmd/main.go` baseado em uma variável `EMAIL_PROVIDER`.

### Novo tópico ou formato
Apenas atualize `.env` — nenhuma mudança de código necessária.

### Suporte a múltiplos idiomas por e-mail separado
Crie uma função em `internal/anthropic/client.go` que aceite `lang string` e itere no `cmd/main.go`.

---

## O que a IA **não deve** fazer neste projeto

- Adicionar dependências externas sem necessidade clara.
- Mover lógica de negócio para `cmd/main.go`.
- Usar `os.Getenv` diretamente fora de `internal/config`.
- Criar arquivos `.env` com valores reais.
- Sugerir `panic` como tratamento de erro em produção.
- Alterar o `.gitignore` de forma que exponha secrets.
- Refatorar pacotes para fora de `internal/` sem justificativa.
