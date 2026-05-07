package ai

import (
	"fmt"
	"strings"
	"time"
)

// StripDisclaimer remove linhas de disclaimer ou desculpas geradas pelo LLM.
func StripDisclaimer(text string) string {
	skipPatterns := []string{
		"desculpe",
		"não posso acessar",
		"não tenho acesso",
		"simulação do digest",
		"i'm unable",
		"i'm sorry",
		"i cannot",
		"i am unable",
		"as of my knowledge",
		"my knowledge cutoff",
		"aqui está uma simulação",
		"com base no meu conhecimento",
		"conhecimento atualizado até",
		"however, i can help",
		"based on available information",
		"i can help generate",
		"here is a sample digest",
		"aqui está um exemplo",
		"aqui está uma lista",
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		lower := strings.ToLower(line)
		skip := false
		for _, pat := range skipPatterns {
			if strings.Contains(lower, pat) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// DatePT formata time.Time como data completa em português.
func DatePT(t time.Time) string {
	months := [...]string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho",
		"julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
	days := [...]string{"domingo", "segunda-feira", "terça-feira", "quarta-feira",
		"quinta-feira", "sexta-feira", "sábado"}
	return fmt.Sprintf("%s, %d de %s de %d", days[t.Weekday()], t.Day(), months[t.Month()], t.Year())
}

// BuildSourcesInstruction gera a instrução de fontes para o prompt.
// Quando articlesCtx está vazio, instrui o LLM a usar conhecimento de treinamento.
// Quando há artigos RSS disponíveis, injeta-os como base obrigatória da curadoria.
func BuildSourcesInstruction(articlesCtx string) string {
	if articlesCtx == "" {
		return "Use seu conhecimento de treinamento mais recente para selecionar conteúdos representativos e relevantes. Priorize conteúdos reconhecidamente importantes e indique a data aproximada de cada publicação."
	}
	return fmt.Sprintf(`ARTIGOS REAIS COLETADOS ESTA SEMANA — use como base obrigatória da curadoria:

%s
Selecione os itens do digest A PARTIR DESTES ARTIGOS REAIS. Todos os itens devem referenciar artigos da lista acima. Complemente com contexto de treinamento apenas na análise de Ada e Alan — nunca para inventar artigos.`, articlesCtx)
}
