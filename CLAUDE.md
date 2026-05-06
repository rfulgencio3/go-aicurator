# CLAUDE.md — Guideline para IA (Claude Code / VS Code)

> Este arquivo fornece contexto e regras para assistentes de IA que trabalham neste repositório.
> Leia-o inteiramente antes de sugerir qualquer mudança de código.
> **Mantenha este arquivo e o README.md sempre atualizados após cada conjunto de mudanças.**

---

## Visão geral do projeto

**go-aicurator** é um digest de tecnologia, ciência e cultura narrado por dois co-curadores com personalidades opostas:

- **Ada** — homenagem a Ada Lovelace. Humor mordaz britânico, ceticismo técnico, anti-hype.
- **Alan** — homenagem a Alan Turing. Entusiasta militante, matemático, centro-esquerda, pro-minorias.

O agente busca conteúdos recentes, gera um digest bilíngue com comentários de ambos e envia por e-mail. Roda de forma autônoma via GitHub Actions.

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
| Provider de IA | OpenAI (padrão, `gpt-4o`, Chat Completions API — sem web search) ou Anthropic (`claude-sonnet-4-20250514` + `web_search_20250305`) |
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
│   │   └── client.go            # Provider IA Anthropic + prompt Ada & Alan
│   ├── config/
│   │   └── config.go            # Carrega e valida variáveis de ambiente
│   ├── openai/
│   │   └── client.go            # Provider IA OpenAI + prompt Ada & Alan
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
| `EMAIL_FROM_NAME` | Não | `Metria CuradorIA` |
| `EMAIL_TO` | Sim | — (vírgula para múltiplos) |
| `TOPICS` | Não | 10 tópicos (ver abaixo) |
| `FORMATS` | Não | Artigos, Papers, Vídeos |
| `ITEM_QTY` | Não | `12` |
| `DIGEST_LANG` | Não | `bilingual` |
| `LANG` | Não | legado; use `DIGEST_LANG` |

**Tópicos padrão (TOPICS):**
Estruturas de Dados e Algoritmos, Inteligência Artificial e Machine Learning, LLMs e Modelos de Linguagem, Astronomia e Exploração Espacial, Neurociência e Comportamento Humano, Estoicismo e Filosofia Prática, Desenvolvimento Pessoal e Performance, Geopolítica e Relações Internacionais, Tempo e Clima, Tecnologia e Startups

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

## Prompt — Ada & Alan

O prompt é construído em `buildPrompt()` — presente em `internal/openai/client.go` e `internal/anthropic/client.go`. **Ambos devem ser mantidos em sincronia.**

### Ada
- Voz: frases curtas, sem exclamações, prefere ponto-e-vírgula
- Frases características: "Dijkstra já havia resolvido isso em [ano]. EWD[n].", "Não reconheço esse nome." (PHP)
- Anti-hype: detecta press release disfarçado com precisão cirúrgica
- Referências: Dijkstra (EWDs), Shannon, Turing, Knuth (TAOCP), von Neumann, Ada Lovelace
- Elogio máximo: "Isto é aceitável."

### Alan
- Voz: entusiástico, usa exclamações, conecta tudo à matemática e à história
- Frases características: "Como Turing demonstrou com a máquina universal...", "Isso é o que democratização parece na prática!"
- Pro-minorias: celebra Grace Hopper, Katherine Johnson, Dorothy Vaughan, Mary Jackson
- Ama JavaScript militantemente; defende PHP pragmaticamente; discorda da Ada explicitamente sobre ambos

### Tensão Ada × Alan (use nas opiniões)
- JavaScript, PHP e IA Generativa: pontos de conflito
- Rust, open source com substância, respeito por Turing: raros pontos de concordância

### Formato por item

```
N. Título do conteúdo
Tipo: artigo | paper | vídeo | podcast | ferramenta | outro
Resumo: duas frases técnicas e objetivas
Ada diz: 2-4 frases em português
Ada says: 2-4 frases em inglês
Alan diz: 2-4 frases em português
Alan says: 2-4 frases em inglês
Link: URL real e verificada (omitir se não tiver certeza — nunca inventar)
Links relacionados: URLs reais separadas por | (omitir se não verificadas)
Nível: Iniciante | Intermediário | Avançado

# Apenas para itens de algoritmo/estrutura de dados:
Exemplo: trace passo a passo com entrada pequena
Complexidade: tempo O(?) | espaço O(?)
Visualizar: URL de ferramenta visual (VisuAlgo, Algorithm Visualizer, etc.)
```

### REGRA ABSOLUTA DE FORMATO
A resposta deve começar diretamente com "1. [Título do primeiro item]". Nenhum disclaimer, aviso ou texto introdutório antes do item 1.

### Seções fixas ao final

1. Ada's Pick da Semana / Ada's Pick of the Week
2. Alan's Pick da Semana / Alan's Pick of the Week
3. Fatos Interessantes / Interesting Facts
4. Hoje na História / Today in History
5. Livro da Semana / Book of the Week
6. Canal/Vídeo em Destaque / Featured Channel or Video

### Fontes e criadores preferenciais

- Geopolítica: Clovis de Barros Filho
- Neurociência: Miguel Nicolelis
- Astronomia/Ciência: Sergio Sacani, Ciência Todo Dia, Ciência sem Fim
- Tech/Dev: Fabio Akita, Professor HOK, Alura, Codecon, Hipsters.tech, Café Debug
- Vídeos/Podcasts: Mano Deyvin, Lucas Montano, Flow podcast

**Ao modificar o prompt:**
- Mantenha sincronia entre `openai/client.go` e `anthropic/client.go`
- Não remova a instrução de retornar apenas texto plano (sem markdown extra)
- A data atual é injetada via `datePT(time.Now())` — não hardcode datas
- Teste com `make run` antes de commitar

---

## Renderização HTML do e-mail

A função `textToHTML()` fica em cada cliente de e-mail e deve ser mantida **em sincronia** entre `resend/client.go` e `sendgrid/client.go`.

### Elementos detectados e renderizados

| Campo no texto | Renderização HTML |
|---|---|
| Itens numerados (`N.`) | Card branco com borda esquerda roxa + contador `01`, `02`... |
| TOC ("Nesta edição") | Painel de sumário antes dos cards |
| `Ada diz:` / `Ada says:` | Bloco roxo (`#F5F3FF`, borda `#8B5CF6`) com bandeira PT/EN |
| `Alan diz:` / `Alan says:` | Bloco teal (`#F0FDFA`, borda `#14B8A6`) com bandeira PT/EN |
| `Exemplo:` | Bloco monospace escuro (`#1E293B`, texto `#E2E8F0`) |
| `Complexidade:` | Badge dark com texto cyan (`⚡ O(n log n)`) |
| `Visualizar:` | Botão dark `👁 Visualizar algoritmo →` |
| `Nível:` | Badge colorido (verde/roxo/vermelho) |
| `Link:` | Botão `Acessar conteúdo →` (suprimido se URL falsa via `isFakeURL()`) |
| `Links relacionados:` | Chips pill-shaped por URL (chips falsos suprimidos via `isFakeURL()`) |
| Ada's Pick | Bloco roxo `⭐` |
| Alan's Pick | Bloco teal `🏆` |
| Fatos Interessantes | Bloco verde `💡` |
| Hoje na História | Bloco laranja `📅` |
| Livro da Semana | Bloco âmbar `📚` |
| Canal/Vídeo em Destaque | Bloco vermelho `🎬` |

### Header e footer

- Header: wordmark `metria.` + avatares SVG minimalistas de Ada (fundo roxo, coque vitoriano) e Alan (fundo teal, cabelo curto)
- Footer: `metria.` wordmark + "Ada & Alan" + link GitHub

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

### Nova seção no digest
1. Adicione a instrução no prompt (ambos os providers).
2. Adicione entrada em `sectionStyles` e caso em `detectSection` (ambos os clientes de e-mail).

### Novo campo por item
1. Adicione keyword slice (ex: `var fooKeywords`), funções `isFooLine` e renderização no loop de cards.
2. Replique em ambos os clientes de e-mail.
3. Adicione instrução no prompt (ambos os providers).

### Novo tópico ou formato
Apenas atualize `.env` ou o padrão em `config.go` — nenhuma mudança de código necessária.

---

## O que a IA **não deve** fazer neste projeto

- Adicionar dependências externas sem necessidade clara.
- Mover lógica de negócio para `cmd/main.go`.
- Usar `os.Getenv` diretamente fora de `internal/config`.
- Criar arquivos `.env` com valores reais.
- Sugerir `panic` como tratamento de erro em produção.
- Alterar o `.gitignore` de forma que exponha secrets.
- Desincronizar `buildPrompt()` entre `openai/client.go` e `anthropic/client.go`.
- Desincronizar `textToHTML()`, `isFakeURL()`, `stripDisclaimer()` e helpers entre os clientes de e-mail e IA.
- Remover `stripDisclaimer()` de um cliente sem remover do outro.
- Remover `isFakeURL()` de `resend/client.go` sem remover de `sendgrid/client.go`.
- Commitar sem atualizar `README.md` e `CLAUDE.md` quando houver mudanças que afetem comportamento, configuração ou estrutura do projeto.
