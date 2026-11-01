package sketch

import (
	"math"
	"sort"
)

// Icons are original hand-drawn pictograms on a 48×48 grid. Fill shapes get
// the warm "oat" tone; ink shapes are strokes in the text colour.
var icons = map[string]Drawing{
	"book": {
		Fill: []Shape{G(24, 14, 8, 11, 8, 36, 24, 38), G(24, 14, 40, 11, 40, 36, 24, 38)},
		Ink:  []Shape{G(24, 14, 8, 11, 8, 36, 24, 38), G(24, 14, 40, 11, 40, 36, 24, 38), L(24, 14, 24, 38), L(12, 18, 20, 19), L(12, 24, 20, 25), L(28, 19, 36, 18), L(28, 25, 36, 24)},
	},
	"code": {
		Fill: []Shape{R(7, 10, 34, 28)},
		Ink:  []Shape{R(7, 10, 34, 28), L(7, 17, 41, 17), L(13, 10, 13, 17), P(20, 23, 16, 27.5, 20, 32), P(28, 23, 32, 27.5, 28, 32), L(26, 22, 22, 33)},
	},
	"terminal": {
		Fill: []Shape{R(7, 10, 34, 28)},
		Ink:  []Shape{R(7, 10, 34, 28), P(13, 20, 18, 24, 13, 28), L(21, 29, 30, 29)},
	},
	"rocket": {
		Fill: []Shape{B(24, 6, 30, 13, 31, 27, 24, 32, 17, 27, 18, 13), G(18, 24, 12, 33, 18, 31), G(30, 24, 36, 33, 30, 31)},
		Ink:  []Shape{B(24, 6, 30, 13, 31, 27, 24, 32, 17, 27, 18, 13), P(18, 23, 12, 33, 18, 31), P(30, 23, 36, 33, 30, 31), C(24, 18, 3.2), P(21, 35, 24, 42, 27, 35)},
	},
	"gear": {
		Fill: []Shape{gearShape(24, 24, 15, 11.5, 8)},
		Ink:  []Shape{gearShape(24, 24, 15, 11.5, 8), C(24, 24, 5)},
	},
	"bolt": {
		Fill: []Shape{G(28, 6, 13, 27, 23, 27, 20, 42, 35, 20, 25, 20)},
		Ink:  []Shape{G(28, 6, 13, 27, 23, 27, 20, 42, 35, 20, 25, 20)},
	},
	"shield": {
		Fill: []Shape{B(24, 7, 38, 12, 37, 27, 24, 41, 11, 27, 10, 12)},
		Ink:  []Shape{B(24, 7, 38, 12, 37, 27, 24, 41, 11, 27, 10, 12), P(18, 24, 22, 28, 30, 19)},
	},
	"key": {
		Fill: []Shape{C(16, 31, 7)},
		Ink:  []Shape{C(16, 31, 7), L(21, 26, 39, 8), P(32, 15, 36, 19), P(28, 19, 31, 22)},
	},
	"cloud": {
		Fill: []Shape{B(14, 33, 7, 28, 11, 21, 18, 20, 22, 13, 31, 13, 35, 20, 41, 23, 41, 31, 34, 34)},
		Ink:  []Shape{B(14, 33, 7, 28, 11, 21, 18, 20, 22, 13, 31, 13, 35, 20, 41, 23, 41, 31, 34, 34)},
	},
	"database": {
		Fill: []Shape{G(11, 12, 37, 12, 37, 36, 11, 36)},
		Ink:  []Shape{E(24, 12, 13, 4.5), L(11, 12, 11, 36), L(37, 12, 37, 36), A(24, 24, 13, 4.5, 0, 180), A(24, 36, 13, 4.5, 0, 180)},
	},
	"folder": {
		Fill: []Shape{G(6, 13, 18, 13, 21, 17, 42, 17, 42, 38, 6, 38)},
		Ink:  []Shape{G(6, 13, 18, 13, 21, 17, 42, 17, 42, 38, 6, 38), L(6, 22, 42, 22)},
	},
	"file": {
		Fill: []Shape{G(12, 6, 29, 6, 37, 14, 37, 42, 12, 42)},
		Ink:  []Shape{G(12, 6, 29, 6, 37, 14, 37, 42, 12, 42), P(29, 6, 29, 14, 37, 14), L(17, 22, 31, 22), L(17, 28, 31, 28), L(17, 34, 26, 34)},
	},
	"globe": {
		Fill: []Shape{C(24, 24, 16)},
		Ink:  []Shape{C(24, 24, 16), E(24, 24, 7, 16), L(8, 24, 40, 24), P(11, 16, 24, 14, 37, 16), P(11, 32, 24, 34, 37, 32)},
	},
	"message": {
		Fill: []Shape{G(7, 10, 41, 10, 41, 30, 22, 30, 14, 38, 15, 30, 7, 30)},
		Ink:  []Shape{G(7, 10, 41, 10, 41, 30, 22, 30, 14, 38, 15, 30, 7, 30), L(14, 18, 34, 18), L(14, 23, 27, 23)},
	},
	"user": {
		Fill: []Shape{C(24, 16, 7)},
		Ink:  []Shape{C(24, 16, 7), A(24, 41, 14, 13, 180, 360)},
	},
	"users": {
		Fill: []Shape{C(18, 17, 6), C(32, 19, 5)},
		Ink:  []Shape{C(18, 17, 6), A(18, 41, 12, 12, 180, 360), C(32, 19, 5), A(34, 40, 9, 10, 205, 360)},
	},
	"chart": {
		Fill: []Shape{R(13, 26, 6, 14), R(22, 18, 6, 22), R(31, 11, 6, 29)},
		Ink:  []Shape{P(8, 7, 8, 40, 42, 40), R(13, 26, 6, 14), R(22, 18, 6, 22), R(31, 11, 6, 29)},
	},
	"search": {
		Fill: []Shape{C(20, 20, 11)},
		Ink:  []Shape{C(20, 20, 11), L(28.5, 28.5, 40, 40)},
	},
	"sparkles": {
		Fill: []Shape{star4(22, 25, 15, 4)},
		Ink:  []Shape{star4(22, 25, 15, 4), star4(37, 11, 6, 1.6)},
	},
	"heart": {
		Fill: []Shape{B(24, 40, 9, 27, 8, 15, 16, 10, 24, 16, 32, 10, 40, 15, 39, 27)},
		Ink:  []Shape{B(24, 40, 9, 27, 8, 15, 16, 10, 24, 16, 32, 10, 40, 15, 39, 27)},
	},
	"check": {
		Fill: []Shape{C(24, 24, 16)},
		Ink:  []Shape{C(24, 24, 16), P(16, 24.5, 22, 30, 33, 17)},
	},
	"clock": {
		Fill: []Shape{C(24, 24, 16)},
		Ink:  []Shape{C(24, 24, 16), P(24, 14, 24, 24, 31, 29)},
	},
	"compass": {
		Fill: []Shape{G(24, 11, 28.5, 24, 19.5, 24)},
		Ink:  []Shape{C(24, 24, 16), G(24, 11, 28.5, 24, 24, 37, 19.5, 24)},
	},
	"flask": {
		Fill: []Shape{G(15, 29, 33, 29, 38, 39, 10, 39)},
		Ink:  []Shape{G(19, 6, 29, 6, 28, 8, 28, 19, 38, 39, 10, 39, 20, 19, 20, 8), L(15, 29, 33, 29), C(22, 34, 1.4), C(28, 32, 1)},
	},
	"sprout": {
		Fill: []Shape{B(24, 26, 15, 22, 9, 14, 17, 14), B(24, 20, 30, 12, 39, 9, 37, 18)},
		Ink:  []Shape{P(24, 42, 24, 30, 24, 20), B(24, 26, 15, 22, 9, 14, 17, 14), B(24, 20, 30, 12, 39, 9, 37, 18), L(15, 42, 33, 42)},
	},
	"layers": {
		Fill: []Shape{G(24, 8, 40, 16, 24, 24, 8, 16)},
		Ink:  []Shape{G(24, 8, 40, 16, 24, 24, 8, 16), P(8, 23, 24, 31, 40, 23), P(8, 30, 24, 38, 40, 30)},
	},
	"lock": {
		Fill: []Shape{R(12, 22, 24, 18)},
		Ink:  []Shape{R(12, 22, 24, 18), A(24, 22, 8, 10, 180, 360), L(24, 29, 24, 33)},
	},
	"mail": {
		Fill: []Shape{R(7, 12, 34, 24)},
		Ink:  []Shape{R(7, 12, 34, 24), P(7, 13, 24, 26, 41, 13)},
	},
	"map": {
		Fill: []Shape{G(18, 7, 30, 11, 30, 41, 18, 37)},
		Ink:  []Shape{G(7, 11, 18, 7, 30, 11, 41, 7, 41, 37, 30, 41, 18, 37, 7, 41), L(18, 7, 18, 37), L(30, 11, 30, 41)},
	},
	"pencil": {
		Fill: []Shape{G(32, 7, 41, 16, 17, 40, 8, 31)},
		Ink:  []Shape{G(32, 7, 41, 16, 17, 40, 7, 41, 8, 31), L(28, 11, 37, 20)},
	},
	"play": {
		Fill: []Shape{C(24, 24, 16)},
		Ink:  []Shape{C(24, 24, 16), G(20, 16, 32, 24, 20, 32)},
	},
	"star": {
		Fill: []Shape{star5(24, 25, 17, 7.5)},
		Ink:  []Shape{star5(24, 25, 17, 7.5)},
	},
	"wand": {
		Fill: []Shape{star4(33, 14, 8, 2.2)},
		Ink:  []Shape{L(9, 39, 27, 21), star4(33, 14, 8, 2.2), L(9, 39, 12, 36)},
	},
	"wave": {
		Ink: []Shape{W(6, 42, 16, 3.5, 12), W(6, 42, 25, 3.5, 12), W(6, 42, 34, 3.5, 12)},
	},
	"download": {
		Fill: []Shape{G(9, 32, 39, 32, 39, 40, 9, 40)},
		Ink:  []Shape{L(24, 7, 24, 29), P(16, 21, 24, 29, 32, 21), P(9, 31, 9, 40, 39, 40, 39, 31)},
	},
	"box": {
		Fill: []Shape{G(8, 14, 24, 22, 40, 14, 24, 6)},
		Ink:  []Shape{G(24, 6, 40, 14, 40, 34, 24, 42, 8, 34, 8, 14), P(8, 14, 24, 22, 40, 14), L(24, 22, 24, 42)},
	},
	"cpu": {
		Fill: []Shape{R(19, 19, 10, 10)},
		Ink:  []Shape{R(12, 12, 24, 24), R(19, 19, 10, 10), L(19, 6, 19, 12), L(29, 6, 29, 12), L(19, 36, 19, 42), L(29, 36, 29, 42), L(6, 19, 12, 19), L(6, 29, 12, 29), L(36, 19, 42, 19), L(36, 29, 42, 29)},
	},
	"lightbulb": {
		Fill: []Shape{B(24, 6, 34, 11, 35, 21, 29, 30, 19, 30, 13, 21, 14, 11)},
		Ink:  []Shape{B(24, 6, 34, 11, 35, 21, 29, 30, 19, 30, 13, 21, 14, 11), L(19, 35, 29, 35), L(21, 40, 27, 40)},
	},
	"target": {
		Fill: []Shape{C(24, 24, 9)},
		Ink:  []Shape{C(24, 24, 16), C(24, 24, 9), C(24, 24, 2.5)},
	},
	"flag": {
		Fill: []Shape{G(13, 8, 37, 8, 31, 16, 37, 24, 13, 24)},
		Ink:  []Shape{L(13, 42, 13, 6), G(13, 8, 37, 8, 31, 16, 37, 24, 13, 24)},
	},
	"calendar": {
		Fill: []Shape{R(8, 11, 32, 8)},
		Ink:  []Shape{R(8, 11, 32, 29), L(8, 19, 40, 19), L(16, 7, 16, 14), L(32, 7, 32, 14), C(16, 27, 1.2), C(24, 27, 1.2), C(32, 27, 1.2), C(16, 34, 1.2), C(24, 34, 1.2)},
	},
	"server": {
		Fill: []Shape{R(9, 9, 30, 12), R(9, 27, 30, 12)},
		Ink:  []Shape{R(9, 9, 30, 12), R(9, 27, 30, 12), C(15, 15, 1.4), C(15, 33, 1.4), L(26, 15, 33, 15), L(26, 33, 33, 33)},
	},
	"home": {
		Fill: []Shape{R(12, 22, 24, 18)},
		Ink:  []Shape{P(7, 24, 24, 8, 41, 24), R(12, 22, 24, 18), P(20, 40, 20, 30, 28, 30, 28, 40)},
	},
	"question": {
		Fill: []Shape{C(24, 24, 16)},
		Ink:  []Shape{C(24, 24, 16), P(19, 19, 20.5, 15, 24, 14, 28, 15.5, 29, 19, 24, 23.5, 24, 27.5), C(24, 33, 0.9)},
	},
	"eye": {
		Fill: []Shape{C(24, 24, 6)},
		Ink:  []Shape{A(24, 30, 18, 14, 200, 340), A(24, 18, 18, 14, 20, 160), C(24, 24, 6)},
	},
	"coffee": {
		Fill: []Shape{G(10, 19, 32, 19, 30, 39, 12, 39)},
		Ink:  []Shape{G(10, 19, 32, 19, 30, 39, 12, 39), A(32, 27, 6, 5.5, -90, 90), P(17, 7, 19, 10.5, 17, 14), P(24, 7, 26, 10.5, 24, 14)},
	},
	"plug": {
		Fill: []Shape{R(15, 16, 18, 13)},
		Ink:  []Shape{R(15, 16, 18, 13), L(20, 8, 20, 16), L(28, 8, 28, 16), P(24, 29, 24, 35, 20, 41)},
	},
	"puzzle": {
		Fill: []Shape{R(9, 15, 26, 25)},
		Ink:  []Shape{P(9, 15, 17, 15), A(22, 14, 5, 5, 180, 360), P(27, 15, 35, 15, 35, 23), A(36, 28, 5, 5, -90, 90), P(35, 33, 35, 40, 9, 40, 9, 15)},
	},
	"brush": {
		Fill: []Shape{B(13, 30, 19, 33, 18, 40, 9, 42, 10, 35)},
		Ink:  []Shape{L(18, 31, 39, 9), B(13, 30, 19, 33, 18, 40, 9, 42, 10, 35)},
	},
	"chat": {
		Fill: []Shape{B(8, 22, 12, 12, 24, 9, 36, 12, 40, 22, 35, 31, 22, 32, 13, 38, 14, 30)},
		Ink:  []Shape{B(8, 22, 12, 12, 24, 9, 36, 12, 40, 22, 35, 31, 22, 32, 13, 38, 14, 30), C(17, 21, 1.2), C(24, 21, 1.2), C(31, 21, 1.2)},
	},
	"tag": {
		Fill: []Shape{G(8, 8, 23, 8, 41, 26, 26, 41, 8, 23)},
		Ink:  []Shape{G(8, 8, 23, 8, 41, 26, 26, 41, 8, 23), C(16, 16, 2.4)},
	},
	"pin": {
		Fill: []Shape{B(24, 42, 14, 26, 13, 16, 18, 9, 24, 7, 30, 9, 35, 16, 34, 26)},
		Ink:  []Shape{B(24, 42, 14, 26, 13, 16, 18, 9, 24, 7, 30, 9, 35, 16, 34, 26), C(24, 18, 4)},
	},
}

// Aliases map names from Lucide, Font Awesome and Mintlify to our icons.
var aliases = map[string]string{
	"book-open": "book", "books": "book", "docs": "book", "documentation": "book", "guide": "book",
	"code-simple": "code", "brackets": "code", "brackets-curly": "code", "window": "code",
	"terminal-window": "terminal", "square-terminal": "terminal", "console": "terminal", "cli": "terminal", "command": "terminal",
	"rocket-launch": "rocket", "launch": "rocket", "start": "rocket",
	"settings": "gear", "cog": "gear", "gears": "gear", "sliders": "gear", "config": "gear", "wrench": "gear", "tools": "gear",
	"zap": "bolt", "lightning": "bolt", "lightning-bolt": "bolt", "flash": "bolt", "bolt-lightning": "bolt", "fast": "bolt",
	"shield-check": "shield", "shield-halved": "shield", "security": "shield", "secure": "shield",
	"cloud-arrow-up": "cloud", "cloud-upload": "cloud", "upload": "cloud",
	"db": "database", "storage": "database", "hard-drive": "database",
	"folder-open": "folder", "directory": "folder",
	"file-lines": "file", "file-text": "file", "document": "file", "page": "file", "file-code": "file",
	"earth": "globe", "world": "globe", "language": "globe", "languages": "globe", "web": "globe", "earth-americas": "globe",
	"comment": "message", "comments": "message", "message-square": "message",
	"messages": "chat", "message-circle": "chat", "conversation": "chat", "chat-bubble": "chat",
	"person": "user", "circle-user": "user", "account": "user", "profile": "user",
	"people": "users", "team": "users", "user-group": "users", "group": "users", "community": "users",
	"chart-bar": "chart", "chart-line": "chart", "bar-chart": "chart", "analytics": "chart", "graph": "chart", "chart-simple": "chart", "stats": "chart",
	"magnifying-glass": "search", "magnifier": "search", "find": "search",
	"sparkle": "sparkles", "wand-magic-sparkles": "wand", "magic": "wand", "ai": "sparkles", "stars": "sparkles",
	"love": "heart", "like": "heart",
	"circle-check": "check", "check-circle": "check", "badge-check": "check", "done": "check", "success": "check",
	"time": "clock", "hourglass": "clock", "history": "clock", "stopwatch": "clock",
	"navigation": "compass", "explore": "compass", "discover": "compass",
	"beaker": "flask", "lab": "flask", "flask-conical": "flask", "experiment": "flask", "research": "flask", "vial": "flask",
	"leaf": "sprout", "seedling": "sprout", "plant": "sprout", "growth": "sprout", "tree": "sprout",
	"layer-group": "layers", "stack": "layers", "stacks": "layers",
	"lock-keyhole": "lock", "locked": "lock", "padlock": "lock", "privacy": "lock",
	"envelope": "mail", "email": "mail", "inbox": "mail", "newsletter": "mail",
	"map-pin": "pin", "location": "pin", "location-dot": "pin", "marker": "pin",
	"pen": "pencil", "edit": "pencil", "pen-to-square": "pencil", "write": "pencil", "pen-nib": "pencil",
	"circle-play": "play", "video": "play", "player": "play",
	"favorite": "star", "featured": "star",
	"water": "wave", "waves": "wave", "ocean": "wave", "sea": "wave",
	"arrow-down-to-line": "download", "install": "download",
	"package": "box", "cube": "box", "cubes": "box", "boxes": "box", "archive": "box",
	"microchip": "cpu", "chip": "cpu", "processor": "cpu", "model": "cpu",
	"bulb": "lightbulb", "idea": "lightbulb", "light-bulb": "lightbulb",
	"bullseye": "target", "goal": "target", "crosshairs": "target", "focus": "target",
	"flag-checkered": "flag", "milestone": "flag",
	"calendar-days": "calendar", "date": "calendar", "event": "calendar", "schedule": "calendar",
	"servers": "server", "hosting": "server", "infrastructure": "server", "network": "server",
	"house":           "home",
	"circle-question": "question", "help": "question", "faq": "question", "question-mark": "question", "circle-help": "question",
	"view": "eye", "visible": "eye", "preview": "eye",
	"mug": "coffee", "cup": "coffee", "mug-hot": "coffee",
	"api": "plug", "integration": "plug", "integrations": "plug", "connect": "plug", "plugin": "plug", "webhook": "plug",
	"puzzle-piece": "puzzle", "extension": "puzzle", "extensions": "puzzle", "addon": "puzzle",
	"paintbrush": "brush", "palette": "brush", "design": "brush", "theme": "brush",
	"tags": "tag", "label": "tag", "price": "tag", "pricing": "tag",
}

// Lookup returns the drawing for a name or alias.
func Lookup(name string) (Drawing, string, bool) {
	if d, ok := icons[name]; ok {
		return d, name, true
	}
	if a, ok := aliases[name]; ok {
		return icons[a], a, true
	}
	return Drawing{}, "", false
}

// Names lists the canonical icon names.
func Names() []string {
	out := make([]string, 0, len(icons))
	for n := range icons {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// AllNames lists canonical names and aliases, for suggestions.
func AllNames() []string {
	out := Names()
	for a := range aliases {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

// Icon renders a named icon. The seed is derived from the name so every use
// of an icon looks the same.
func Icon(name string, o Options) string {
	d, canonical, ok := Lookup(name)
	if !ok {
		return ""
	}
	return d.SVG(Seed(canonical), o)
}

func gearShape(cx, cy, ro, ri float64, teeth int) Shape {
	var xy []float64
	n := teeth * 4
	for i := range n {
		a := float64(i) / float64(n) * 2 * math.Pi
		r := ri
		if i%4 == 1 || i%4 == 2 {
			r = ro
		}
		xy = append(xy, cx+math.Cos(a)*r, cy+math.Sin(a)*r)
	}
	return G(xy...)
}

func star4(cx, cy, r, w float64) Shape {
	return G(cx, cy-r, cx+w, cy-w, cx+r, cy, cx+w, cy+w, cx, cy+r, cx-w, cy+w, cx-r, cy, cx-w, cy-w)
}

func star5(cx, cy, ro, ri float64) Shape {
	var xy []float64
	for i := range 10 {
		a := -math.Pi/2 + float64(i)*math.Pi/5
		r := ro
		if i%2 == 1 {
			r = ri
		}
		xy = append(xy, cx+math.Cos(a)*r, cy+math.Sin(a)*r)
	}
	return G(xy...)
}

// Circle draws a single hand-drawn circle, for step markers.
func Circle(seed uint64, o Options) string {
	return Drawing{Ink: []Shape{C(24, 24, 20)}}.SVG(seed, o)
}

// Underline draws a hand-drawn marker stroke for highlighted words. The
// viewBox is wide so it stretches under a phrase.
func Underline(seed uint64, o Options) string {
	o.ViewBox = "0 0 200 12"
	d := Drawing{Ink: []Shape{P(2, 8, 50, 5.5, 110, 6.5, 160, 4.5, 198, 6)}}
	return d.SVG(seed, Options{ViewBox: o.ViewBox, Class: o.Class, Wobble: 1.6, Stretch: true})
}

// Arrow draws a hand-drawn arrow pointing right, for <Flow>.
func Arrow(seed uint64, o Options) string {
	o.ViewBox = "0 0 48 24"
	d := Drawing{Ink: []Shape{P(4, 12.5, 18, 11.2, 30, 12.6, 42, 11.8), P(34, 5.5, 42.5, 11.8, 34.5, 18.5)}}
	return d.SVG(seed, Options{ViewBox: o.ViewBox, Class: o.Class, Wobble: 1.1})
}
