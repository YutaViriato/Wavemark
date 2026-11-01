// Command icons exports Wavemark's hand-drawn artwork (the same code the
// product renders with) as JSON for the showreel stage:
//
//	go run ./tools/showreel/icons > tools/showreel/stage/gen/art.json
package main

import (
	"encoding/json"
	"os"

	"github.com/YutaViriato/Wavemark/internal/sketch"
)

func main() {
	art := struct {
		Icons      map[string]string `json:"icons"`
		Underlines []string          `json:"underlines"`
		Arrows     []string          `json:"arrows"`
		Circles    []string          `json:"circles"`
	}{Icons: map[string]string{}}
	for _, n := range sketch.Names() {
		art.Icons[n] = sketch.Icon(n, sketch.Options{})
	}
	for i := uint64(1); i <= 8; i++ {
		art.Underlines = append(art.Underlines, sketch.Underline(i, sketch.Options{}))
		art.Arrows = append(art.Arrows, sketch.Arrow(i, sketch.Options{}))
		art.Circles = append(art.Circles, sketch.Circle(i, sketch.Options{}))
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(art); err != nil {
		os.Exit(1)
	}
}
