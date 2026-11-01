package components

import (
	"strconv"
	"strings"
	"time"
)

var ptMonths = []string{"janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
var esMonths = []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}

// FormatDate formats a date for display in the page language.
func FormatDate(t time.Time, lang string) string {
	base, _, _ := strings.Cut(strings.ToLower(lang), "-")
	switch base {
	case "pt":
		return strconv.Itoa(t.Day()) + " de " + ptMonths[t.Month()-1] + " de " + strconv.Itoa(t.Year())
	case "es":
		return strconv.Itoa(t.Day()) + " de " + esMonths[t.Month()-1] + " de " + strconv.Itoa(t.Year())
	}
	return t.Format("Jan 2, 2006")
}
