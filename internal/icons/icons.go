// Package icons holds the inline SVG icons used by the theme and components.
//
// Paths follow the Lucide icon set (ISC license, https://lucide.dev): 24×24
// viewBox, 2px round strokes, currentColor.
package icons

import "html/template"

const open = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">`
const close = `</svg>`

var paths = map[string]string{
	"info":          `<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>`,
	"lightbulb":     `<path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.7 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/>`,
	"notepad":       `<path d="M8 2v4"/><path d="M12 2v4"/><path d="M16 2v4"/><rect width="16" height="18" x="4" y="4" rx="2"/><path d="M8 10h6"/><path d="M8 14h8"/><path d="M8 18h5"/>`,
	"alert":         `<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>`,
	"octagon":       `<path d="M2.586 16.726A2 2 0 0 1 2 15.312V8.688a2 2 0 0 1 .586-1.414l4.688-4.688A2 2 0 0 1 8.688 2h6.624a2 2 0 0 1 1.414.586l4.688 4.688A2 2 0 0 1 22 8.688v6.624a2 2 0 0 1-.586 1.414l-4.688 4.688a2 2 0 0 1-1.414.586H8.688a2 2 0 0 1-1.414-.586z"/><path d="M12 8v4"/><path d="M12 16h.01"/>`,
	"copy":          `<rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>`,
	"check":         `<path d="M20 6 9 17l-5-5"/>`,
	"check-circle":  `<circle cx="12" cy="12" r="10"/><path d="m9 12 2 2 4-4"/>`,
	"chevron-down":  `<path d="m6 9 6 6 6-6"/>`,
	"chevron-right": `<path d="m9 18 6-6-6-6"/>`,
	"arrow-right":   `<path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>`,
	"arrow-left":    `<path d="m12 19-7-7 7-7"/><path d="M19 12H5"/>`,
	"sun":           `<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>`,
	"moon":          `<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>`,
	"monitor":       `<rect width="20" height="14" x="2" y="3" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/>`,
	"menu":          `<path d="M4 6h16"/><path d="M4 12h16"/><path d="M4 18h16"/>`,
	"x":             `<path d="M18 6 6 18"/><path d="m6 6 12 12"/>`,
	"link":          `<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>`,
	"file-text":     `<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/><path d="M10 9H8"/><path d="M16 13H8"/><path d="M16 17H8"/>`,
}

// SVG returns the markup of an icon, or "" if the name is unknown.
func SVG(name string) string {
	p, ok := paths[name]
	if !ok {
		return ""
	}
	return open + p + close
}

// HTML returns the icon as trusted template HTML.
func HTML(name string) template.HTML {
	return template.HTML(SVG(name)) // #nosec: static markup
}
