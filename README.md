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

For a local SIT build, select its GIR directories explicitly:

```bash
./vetro --gir-path /path/to/sit-build/gtk --in example.vetro --out example.ui
```

The same option works with `--lsp`. `VETRO_GIR_PATH` provides the default when
`--gir-path` is omitted. The first directory has priority over later directories
and system metadata. `--gir` has priority over all directories. Missing, empty
or malformed explicit metadata fails rather than silently using another toolkit.
SIT retains the GTK namespaces and GtkBuilder format, so existing Vetro sources
keep their widget names.

GTK's build tree puts `Gtk-4.0.gir`, `Gdk-4.0.gir` and `Gsk-4.0.gir` together
in `gtk/`. For an installed toolkit, use its `share/gir-1.0` directory. Separate
directories with `:` on Linux or `;` on Windows.

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
