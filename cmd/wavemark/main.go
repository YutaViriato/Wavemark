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
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/YutaViriato/Wavemark/internal/browse"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/export"
	"github.com/YutaViriato/Wavemark/internal/scaffold"
	"github.com/YutaViriato/Wavemark/internal/server"
	"github.com/YutaViriato/Wavemark/internal/site"
	"github.com/YutaViriato/Wavemark/internal/spec"
	"github.com/YutaViriato/Wavemark/internal/theme"
)

// version is set at release time with -ldflags "-X main.version=v1.2.3".
var version = "dev"

const usage = `wavemark turns a folder of Markdown files into a website.

Usage:
  wavemark init  [folder]                   create a starter site
  wavemark dev   [folder] [--addr :4321]    preview with live reload; errors show on the page
  wavemark browse [folder] [--no-open] [--all] [--addr localhost:4400]
                                            read every Markdown file of a project (READMEs,
                                            agent instructions, plans, specs) in the browser
  wavemark check [folder] [--json]          validate; exits with status 1 if there are errors
  wavemark build [folder] [--out dist] [--base /path]
                                            write static HTML for GitHub Pages or any web server
  wavemark serve [folder] [--addr :8080]    run the site as a production server
  wavemark spec  [--json]                   syntax and component reference for authors
  wavemark version                          print the version

The default folder is the current directory.
Docs: https://github.com/YutaViriato/Wavemark
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	var code int
	switch cmd {
	case "init":
		code = runInit(args, os.Stdout)
	case "dev":
		code = runServer(args, true, ":4321")
	case "serve":
		code = runServer(args, false, ":8080")
	case "browse":
		code = runBrowse(args)
	case "check":
		code = runCheck(args, os.Stdout)
	case "build":
		code = runBuild(args, os.Stdout)
	case "spec":
		code = runSpec(args, os.Stdout)
	case "version", "--version", "-v":
		fmt.Println("wavemark", currentVersion())
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
		code = 2
	}
	os.Exit(code)
}

func currentVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

// parse handles "[folder] [flags]" in any order.
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
	return "", fmt.Errorf("expected one folder, got %d arguments", len(positional))
}

func runInit(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	dir, err := parse(fs, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	files, err := scaffold.Init(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	for _, f := range files {
		fmt.Fprintln(out, "created", filepath.Join(dir, f))
	}
	fmt.Fprintf(out, "\nNext:\n  wavemark dev %s      preview it\n  wavemark spec        read the syntax reference\n  wavemark check %s    validate before publishing\n", dir, dir)
	return 0
}

func runServer(args []string, dev bool, defaultAddr string) int {
	name := "serve"
	if dev {
		name = "dev"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	addr := fs.String("addr", defaultAddr, "listen address")
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
		logger.Printf("error: %v", err)
		return 1
	}
	st := srv.Site()
	errs, warns := diag.Count(st.Diags)
	if dev {
		printDiags(os.Stderr, st.Diags)
	}
	logger.Printf("%d pages · %d error(s) · %d warning(s)", len(st.Pages), errs, warns)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.Watch(ctx); err != nil {
			logger.Printf("watcher disabled: %v", err)
		}
	}()
	host := *addr
	if len(host) > 0 && host[0] == ':' {
		host = "localhost" + host
	}
	logger.Printf("serving %s at http://%s", st.Root, host)
	if err := srv.ListenAndServe(ctx, *addr); err != nil {
		logger.Printf("error: %v", err)
		return 1
	}
	return 0
}

func runBrowse(args []string) int {
	fs := flag.NewFlagSet("browse", flag.ContinueOnError)
	addr := fs.String("addr", "", "listen address (default: the first free port from localhost:4400)")
	noOpen := fs.Bool("no-open", false, "do not open the browser")
	all := fs.Bool("all", false, "include files that .gitignore excludes")
	root, err := parse(fs, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if root, err = filepath.Abs(root); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	logger := log.New(os.Stderr, "wavemark ", log.Ltime)
	th, err := theme.New()
	if err != nil {
		logger.Printf("error: %v", err)
		return 1
	}
	srv, err := server.NewWith(server.Options{
		Root:  root,
		Dev:   true,
		Theme: th,
		Build: func(*site.Site) (*site.Site, error) {
			return browse.Build(browse.Options{Root: root, Theme: th, All: *all, Dev: true})
		},
		SkipDir: browse.SkipDir,
		Ignore:  func(rel string) bool { return !browse.Relevant(rel) },
		NoLLMS:  true,
	}, logger)
	if err != nil {
		logger.Printf("error: %v", err)
		return 1
	}
	ln, err := listenBrowse(*addr)
	if err != nil {
		logger.Printf("error: %v", err)
		return 1
	}
	url := "http://" + displayAddr(ln.Addr())
	logger.Printf("%d Markdown files in %s", len(srv.Site().Pages), root)
	if host, _, _ := net.SplitHostPort(ln.Addr().String()); host != "127.0.0.1" && host != "::1" {
		logger.Printf("warning: listening on every network interface; anyone who can reach this machine can read these files")
	}
	logger.Printf("browsing at %s (Ctrl+C to stop)", url)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.Watch(ctx); err != nil {
			logger.Printf("watcher disabled: %v", err)
		}
	}()
	if !*noOpen {
		if err := openBrowser(url); err != nil {
			logger.Printf("open %s in your browser", url)
		}
	}
	if err := srv.Serve(ctx, ln); err != nil {
		logger.Printf("error: %v", err)
		return 1
	}
	return 0
}

// listenBrowse listens on addr or, by default, on the first free port from
// 4400 on the loopback interface: the files stay private to this machine.
func listenBrowse(addr string) (net.Listener, error) {
	if addr != "" {
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		}
		return net.Listen("tcp", addr)
	}
	var err error
	for port := 4400; port < 4420; port++ {
		var ln net.Listener
		if ln, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port)); err == nil {
			return ln, nil
		}
	}
	return nil, fmt.Errorf("no free port between 4400 and 4419 (%w); choose one with --addr", err)
}

func displayAddr(a net.Addr) string {
	host, port, err := net.SplitHostPort(a.String())
	if err != nil {
		return a.String()
	}
	switch host {
	case "127.0.0.1", "::1", "", "0.0.0.0", "::":
		host = "localhost"
	}
	return net.JoinHostPort(host, port)
}

// openBrowser opens url with $BROWSER or the system's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	if b := strings.Fields(strings.Split(os.Getenv("BROWSER"), string(os.PathListSeparator))[0]); len(b) > 0 {
		cmd = exec.Command(b[0], append(b[1:], url)...)
	} else {
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			// Without a display, xdg-open may start a text browser in this
			// terminal; print the address instead.
			if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
				return errors.New("no display")
			}
			cmd = exec.Command("xdg-open", url)
		}
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func runCheck(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "JSON output")
	root, err := parse(fs, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	st, err := site.Build(site.Options{Root: root, Dev: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
		fmt.Fprintf(out, "%d pages · %d error(s) · %d warning(s)\n", len(st.Pages), errs, warns)
	}
	if errs > 0 {
		return 1
	}
	return 0
}

func runBuild(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	outDir := fs.String("out", "dist", "output folder")
	base := fs.String("base", "", "URL path the site is published under, e.g. /my-repo (default: the path of baseURL in wavemark.yaml)")
	root, err := parse(fs, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	th, err := theme.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	st, err := site.Build(site.Options{Root: root, Render: func(s *site.Site, p *site.Page) ([]byte, error) {
		return th.RenderPage(s, p, theme.Options{})
	}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	if diag.HasErrors(st.Diags) {
		printDiags(os.Stderr, st.Diags)
		n, _ := diag.Count(st.Diags)
		fmt.Fprintf(os.Stderr, "error: the content has %d error(s); nothing was written\n", n)
		return 1
	}
	b := *base
	if b == "" {
		b = export.BaseFromURL(st.Config.BaseURL)
	}
	res, err := export.Write(st, th, export.Options{Out: *outDir, Base: b})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	_, warns := diag.Count(st.Diags)
	printDiags(out, st.Diags)
	where := "/"
	if b != "" {
		where = b + "/"
	}
	fmt.Fprintf(out, "%d pages · %d files · %d warning(s) → %s (published under %s)\n", res.Pages, res.Files, warns, res.Out, where)
	return 0
}

func runSpec(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("spec", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "component schemas as JSON")
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
