package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// LSP JSON-RPC structures
type Request struct {
	ID     any             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Response struct {
	ID     any    `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Notification struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Server struct {
	reader *bufio.Reader
	writer io.Writer
	files  map[string]string
}

func NewServer() *Server {
	return &Server{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		files:  make(map[string]string),
	}
}

func (s *Server) Start() {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return
		}

		if !strings.HasPrefix(line, "Content-Length: ") {
			continue
		}

		lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length: "))
		length, _ := strconv.Atoi(lengthStr)

		// Skip \r\n
		s.reader.ReadString('\n')

		content := make([]byte, length)
		_, err = io.ReadFull(s.reader, content)
		if err != nil {
			return
		}

		s.handleMessage(content)
	}
}

func (s *Server) handleMessage(data []byte) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return
	}

	switch req.Method {
	case "initialize":
		s.respond(req.ID, map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync": 1, // Full
				"completionProvider": map[string]any{
					"resolveProvider":   false,
					"triggerCharacters": []string{"(", ".", ":"},
				},
			},
		})
	case "textDocument/didOpen", "textDocument/didChange":
		s.handleDidOpenOrChange(req.Params)
	case "textDocument/completion":
		s.handleCompletion(req.ID, req.Params)
	}
}

func (s *Server) respond(id any, result any) {
	res := Response{ID: id, Result: result}
	data, _ := json.Marshal(res)
	fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
}

func (s *Server) notify(method string, params any) {
	notification := Notification{Method: method}
	notification.Params, _ = json.Marshal(params)
	data, _ := json.Marshal(notification)
	fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
}
