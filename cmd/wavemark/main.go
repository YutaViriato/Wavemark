// Command wavemark serves a folder of Markdown files as a website.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/server"
	"github.com/YutaViriato/Wavemark/internal/site"
	"github.com/YutaViriato/Wavemark/internal/spec"
)

const usage = `wavemark: transforma uma pasta de Markdown em site.

Uso:
  wavemark dev   [pasta] [--addr :4321]   servidor de desenvolvimento (recarrega e mostra erros na página)
  wavemark serve [pasta] [--addr :8080]   servidor de produção (só sobe sem erros; mantém a última versão válida)
  wavemark check [pasta] [--json]         valida o conteúdo e sai com código 1 se houver erros
  wavemark spec  [--json]                 referência de sintaxe e componentes para autores

A pasta padrão é a atual.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	var code int
	switch cmd {
	case "dev":
		code = runServer(args, true, ":4321")
	case "serve":
		code = runServer(args, false, ":8080")
	case "check":
		code = runCheck(args, os.Stdout)
	case "spec":
		code = runSpec(args, os.Stdout)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "comando desconhecido %q\n\n%s", cmd, usage)
		code = 2
	}
	os.Exit(code)
}

// parse handles "[pasta] [flags]" in any order.
func parse(fs *flag.FlagSet, args []string) (string, error) {
	var positional []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return "", err
		}
		args = fs.Args()
		if len(args) > 0 {
			positional = append(positional, args[0])
			args = args[1:]
		}
	}
	switch len(positional) {
	case 0:
		return ".", nil
	case 1:
		return positional[0], nil
	}
	return "", fmt.Errorf("esperava uma pasta, recebi %d argumentos", len(positional))
}

func runServer(args []string, dev bool, defaultAddr string) int {
	name := "serve"
	if dev {
		name = "dev"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	addr := fs.String("addr", defaultAddr, "endereço de escuta")
	root, err := parse(fs, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	logger := log.New(os.Stderr, "wavemark ", log.Ltime)
	srv, err := server.New(root, dev, logger)
	if err != nil {
		var ce *server.ContentError
		if errors.As(err, &ce) {
			printDiags(os.Stderr, ce.Diags)
		}
		logger.Printf("erro: %v", err)
		return 1
	}
	st := srv.Site()
	errs, warns := diag.Count(st.Diags)
	if dev {
		printDiags(os.Stderr, st.Diags)
	}
	logger.Printf("%d páginas · %d erro(s) · %d aviso(s)", len(st.Pages), errs, warns)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.Watch(ctx); err != nil {
			logger.Printf("watcher desativado: %v", err)
		}
	}()
	host := *addr
	if len(host) > 0 && host[0] == ':' {
		host = "localhost" + host
	}
	logger.Printf("servindo %s em http://%s", st.Root, host)
	if err := srv.ListenAndServe(ctx, *addr); err != nil {
		logger.Printf("erro: %v", err)
		return 1
	}
	return 0
}

func runCheck(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "saída em JSON")
	root, err := parse(fs, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	st, err := site.Build(site.Options{Root: root, Dev: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 2
	}
	errs, warns := diag.Count(st.Diags)
	if *asJSON {
		report := struct {
			OK          bool              `json:"ok"`
			Pages       int               `json:"pages"`
			Errors      int               `json:"errors"`
			Warnings    int               `json:"warnings"`
			Diagnostics []diag.Diagnostic `json:"diagnostics"`
		}{errs == 0, len(st.Pages), errs, warns, st.Diags}
		if report.Diagnostics == nil {
			report.Diagnostics = []diag.Diagnostic{}
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
	} else {
		printDiags(out, st.Diags)
		fmt.Fprintf(out, "%d páginas · %d erro(s) · %d aviso(s)\n", len(st.Pages), errs, warns)
	}
	if errs > 0 {
		return 1
	}
	return 0
}

func runSpec(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("spec", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "schemas dos componentes em JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *asJSON {
		b, err := spec.JSON()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		_, _ = out.Write(append(b, '\n'))
		return 0
	}
	_, _ = io.WriteString(out, spec.Markdown())
	return 0
}

func printDiags(w io.Writer, ds []diag.Diagnostic) {
	for _, d := range ds {
		fmt.Fprintln(w, d.String())
	}
}
