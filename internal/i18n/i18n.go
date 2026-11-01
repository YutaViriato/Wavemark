// Package i18n holds the few interface strings the theme renders itself.
package i18n

import "strings"

var tables = map[string]map[string]string{
	"en": {
		"copy":       "Copy",
		"copied":     "Copied",
		"copyCode":   "Copy code",
		"onThisPage": "On this page",
		"previous":   "Previous",
		"next":       "Next",
		"menu":       "Menu",
		"theme":      "Theme",
		"light":      "Light",
		"dark":       "Dark",
		"system":     "System",
		"note":       "Note",
		"tip":        "Tip",
		"info":       "Info",
		"warning":    "Warning",
		"danger":     "Danger",
		"skip":       "Skip to content",
		"notFound":   "Page not found",
		"notFoundP":  "This page does not exist or was moved.",
		"home":       "Back to home",
		"viewSource": "View as Markdown",
		"navigation": "Navigation",
		"published":  "Published",
		"author":     "Author",
		"section":    "Section",
	},
	"pt": {
		"copy":       "Copiar",
		"copied":     "Copiado",
		"copyCode":   "Copiar código",
		"onThisPage": "Nesta página",
		"previous":   "Anterior",
		"next":       "Próxima",
		"menu":       "Menu",
		"theme":      "Tema",
		"light":      "Claro",
		"dark":       "Escuro",
		"system":     "Sistema",
		"note":       "Nota",
		"tip":        "Dica",
		"info":       "Informação",
		"warning":    "Atenção",
		"danger":     "Perigo",
		"skip":       "Pular para o conteúdo",
		"notFound":   "Página não encontrada",
		"notFoundP":  "Esta página não existe ou foi movida.",
		"home":       "Voltar ao início",
		"viewSource": "Ver como Markdown",
		"navigation": "Navegação",
		"published":  "Publicado",
		"author":     "Autor",
		"section":    "Seção",
	},
}

// For returns a lookup function for a BCP 47 language tag ("pt-BR" → pt).
// Unknown languages and keys fall back to English.
func For(lang string) func(string) string {
	base, _, _ := strings.Cut(strings.ToLower(lang), "-")
	t := tables[base]
	return func(key string) string {
		if s, ok := t[key]; ok {
			return s
		}
		if s, ok := tables["en"][key]; ok {
			return s
		}
		return key
	}
}
