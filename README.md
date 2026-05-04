# aicurator — AI Content Curator

Busca conteúdos sobre Tecnologia & IA via Claude (com web search) e envia um digest bilíngue por e-mail via SendGrid.

## Pré-requisitos

- Go 1.22+
- Chave da API Anthropic (https://console.anthropic.com)
- Chave da API SendGrid (https://app.sendgrid.com)
- Remetente verificado no SendGrid (Sender Authentication)

## Configuração

```bash
cp .env.example .env
# edite .env com suas chaves e preferências
```

## Uso rápido (teste manual)

```bash
make run
```

## Build e agendamento via cron

```bash
# compila o binário
make build

# instala o cron (roda às 07h de seg, qua e sex)
make install-cron
```

Para ver o crontab instalado:
```bash
crontab -l
```

Para ver os logs:
```bash
tail -f /tmp/curator.log
```

## Personalização

Todas as opções ficam no `.env`:

| Variável | Descrição | Padrão |
|---|---|---|
| `TOPICS` | Tópicos separados por vírgula | IA, ML, LLMs, Startups |
| `FORMATS` | Tipos de conteúdo | Artigos, Papers, Vídeos |
| `ITEM_QTY` | Itens por digest | 8 |
| `LANG` | Idioma: `pt`, `en`, `bilingual` | bilingual |
| `EMAIL_TO` | Destinatários (vírgula) | — |
| `ANTHROPIC_MODEL` | Modelo Claude | claude-sonnet-4-20250514 |

## Estrutura do projeto

```
aicurator/
├── cmd/main.go                   # entrypoint
├── internal/
│   ├── config/config.go          # carrega variáveis de ambiente
│   ├── anthropic/client.go       # chama a API do Claude
│   └── sendgrid/client.go        # envia e-mail via SendGrid
├── .env.example
├── Makefile
└── go.mod
```
