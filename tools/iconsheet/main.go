// Command iconsheet writes an HTML contact sheet of the hand-drawn icons:
//
//	go run ./tools/iconsheet sheet.html
package main

import (
	"fmt"
	"os"

	"github.com/YutaViriato/Wavemark/internal/sketch"
)

func main() {
	f, _ := os.Create(os.Args[1])
	fmt.Fprint(f, `<!doctype html><style>body{margin:0;padding:24px;background:#faf9f5;color:#141413;font:13px system-ui;display:grid;grid-template-columns:repeat(8,1fr);gap:18px}
div{display:flex;flex-direction:column;align-items:center;gap:6px}.wm-sketch__fill path{fill:#e3dacc}.wm-sketch__ink path{fill:none;stroke:currentColor;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.big .wm-sketch__ink path{stroke-width:1.6}.dark{background:#1f1e1d;color:#faf9f5;padding:8px;border-radius:8px}.dark .wm-sketch__fill path{fill:#3d3d3a}</style>`)
	for _, n := range sketch.Names() {
		fmt.Fprintf(f, `<div>%s%s<span class="dark">%s</span>%s</div>`, sketch.Icon(n, sketch.Options{Size: 48}), `<span class="big">`+sketch.Icon(n, sketch.Options{Size: 96})+`</span>`, sketch.Icon(n, sketch.Options{Size: 40}), n)
	}
	fmt.Fprint(f, `<div style="grid-column:span 4"><span style="font:400 40px Georgia;position:relative">enquanto escreve<span style="position:absolute;left:0;right:0;bottom:-6px;height:12px;color:#d97757">`+sketch.Underline(3, sketch.Options{})+`</span></span></div>`)
	fmt.Fprintf(f, `<div>%s</div>`, sketch.Circle(7, sketch.Options{Size: 48}))
}
