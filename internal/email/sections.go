package email

import "strings"

// SectionHeaders identifica seções especiais do digest pelo prefixo lowercase da linha.
// Usado pelo renderer HTML (detectSection) e pelo TTS (ParseScript stop condition).
var SectionHeaders = []string{
	"ada's pick",
	"alan's pick",
	"fatos interessantes",
	"interesting facts",
	"hoje na história",
	"hoje na historia",
	"today in history",
	"livro da semana",
	"book of the week",
	"canal/vídeo",
	"canal/video",
	"featured channel",
	"featured video",
	"vídeo em destaque",
}

// IsSectionHeader retorna true se lowerLine contém um dos cabeçalhos de seção especial.
// Recebe a linha já em minúsculas para evitar conversão redundante em loops.
func IsSectionHeader(lowerLine string) bool {
	for _, h := range SectionHeaders {
		if strings.Contains(lowerLine, h) {
			return true
		}
	}
	return false
}
