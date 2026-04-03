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
