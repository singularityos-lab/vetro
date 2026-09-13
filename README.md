# Vetro

Vetro is a declarative GTK4 UI transpiler with a built-in Language Server Protocol (LSP) server.

You can:
- transpile `.vetro` to GTK `.ui`
- transpile `.ui` back to `.vetro`

## Requirements

- Go 1.24+
- Gtk-4.0.gir (optional)

## Build

```bash
go build -o vetro .
```

## CLI usage

Transpile Vetro to UI:

```bash
./vetro --in example.vetro --out example.ui
```

Transpile UI to Vetro:

```bash
./vetro --in example.ui --out example.vetro
```

Watch mode:

```bash
./vetro --in example.vetro --out example.ui --watch
```

## Run LSP server

```bash
./vetro --lsp
```

## License

MIT — see [LICENSE](LICENSE).

## Use of Generative AI

Maintainers may use generative AI tools as assistants while working on vetro. Non-trivial assisted commits disclose the tool, model, and scope of the work.

AI tools may assist with code comments, documentation, repetitive code, and issue triage. Maintainers make project decisions and review every assisted change before it is merged.

Use these trailers for non-trivial assisted commits:

```plain
Assisted-by: <tool>:<model-version>
AI-Scope: <what the tool generated and the prompt or a short prompt summary>
```

Single-line completions, renames, and formatting changes do not need trailers.

Coding agents must also follow [AGENTS.md](AGENTS.md) before changing files,
creating commits, or opening pull requests.
