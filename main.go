package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mirkobrombin/go-cli-builder/v2/pkg/cli"

	"vetro/internal/action"
	"vetro/internal/domain/vetro"
	"vetro/internal/domain/vetro/metadata"
	"vetro/internal/lsp"
)

type VetroCLI struct {
	In       string `cli:"in" help:"Input file path (.vetro or .ui)"`
	Out      string `cli:"out" help:"Output file path (.ui or .vetro)"`
	Watch    bool   `cli:"watch" help:"Watch input file for changes and auto-rebuild"`
	ForceGir bool   `cli:"force-gir" help:"Force regeneration of GTK metadata cache"`
	Gir      string `cli:"gir" help:"Path to a custom .gir file to load (e.g. LibSingularity-1.0.gir)"`
	LSP      bool   `cli:"lsp" help:"Start Vetro Language Server (LSP)"`

	cli.Base
}

func main() {
	root := &VetroCLI{}
	app, err := cli.New(root)
	if err != nil {
		printError(err)
		os.Exit(1)
	}
	app.SetName("vetro")

	if err := app.Run(); err != nil {
		printError(err)
		os.Exit(1)
	}
}

func (c *VetroCLI) Run() error {

	// Initialize Metadata
	if mm, err := metadata.NewManager(); err == nil {
		mm.LoadAllGIRs(c.ForceGir)
		// Load user-specified custom GIR (e.g. a library not in system paths)
		if c.Gir != "" {
			if err := mm.LoadGtkMetadata(c.Gir, c.ForceGir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not load GIR %s: %v\n", c.Gir, err)
			}
		}
		vetro.MetadataManager = mm
	}

	if c.LSP {
		server := lsp.NewServer()
		server.Start()
		return nil
	}

	if c.In == "" || c.Out == "" {
		fmt.Println("Vetro - Declarative GTK4 UI Transpiler")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  vetro --in app.vetro --out app.ui                        # Vetro to GTK UI")
		fmt.Println("  vetro --in app.ui --out app.vetro                        # GTK UI to Vetro")
		fmt.Println("  vetro --in app.vetro --out app.ui --watch                # Watch mode")
		fmt.Println("  vetro --in app.vetro --out app.ui --gir MyLib-1.0.gir   # With custom GIR")
		fmt.Println("  vetro --lsp                                               # Start LSP server")
		return fmt.Errorf("missing required flags --in and --out unless --lsp is used")
	}

	act := &action.TranspileAction{}

	// Initial build
	if err := act.Transpile(c.In, c.Out); err != nil {
		return err
	}
	fmt.Printf("✓ Generated %s\n", c.Out)

	// Check for CSS output
	cssPath := strings.TrimSuffix(c.Out, filepath.Ext(c.Out)) + ".css"
	if _, err := os.Stat(cssPath); err == nil {
		fmt.Printf("✓ Generated %s\n", cssPath)
	}

	if !c.Watch {
		return nil
	}

	// Watch mode
	fmt.Printf("Watching %s for changes...\n", c.In)
	var lastModTime time.Time
	for {
		time.Sleep(500 * time.Millisecond)

		info, err := os.Stat(c.In)
		if err != nil {
			continue
		}

		if info.ModTime().After(lastModTime) {
			lastModTime = info.ModTime()

			fmt.Printf("\n[%s] Rebuilding...\n", time.Now().Format("15:04:05"))
			if err := act.Transpile(c.In, c.Out); err != nil {
				printError(err)
				continue
			}
			fmt.Printf("✓ Generated %s\n", c.Out)
			if _, err := os.Stat(cssPath); err == nil {
				fmt.Printf("✓ Generated %s\n", cssPath)
			}
		}
	}
}

// printError prints an error with nice formatting.
func printError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}
