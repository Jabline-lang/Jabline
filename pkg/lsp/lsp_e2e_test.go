package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// LSPClient is a minimal LSP client for testing.
type LSPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	msgID  int
}

func newLSPClient() (*LSPClient, error) {
	cmd := exec.Command("go", "run", ".", "lsp")
	cmd.Dir = "..\\.."

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	return &LSPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		msgID:  0,
	}, nil
}

// sendJSON sends a raw JSON message to the server.
func (c *LSPClient) sendJSON(msg map[string]interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := io.WriteString(c.stdin, header); err != nil {
		return err
	}
	if _, err := c.stdin.Write(data); err != nil {
		return err
	}
	return nil
}

// request sends an RPC request and returns the id.
func (c *LSPClient) request(method string, params interface{}) (int, error) {
	c.msgID++
	return c.msgID, c.sendJSON(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      c.msgID,
		"method":  method,
		"params":  params,
	})
}

// notify sends a notification (no id).
func (c *LSPClient) notify(method string, params interface{}) error {
	return c.sendJSON(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
}

// readRaw reads one LSP message and returns the raw JSON.
func (c *LSPClient) readRaw() ([]byte, error) {
	contentLength := 0
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			fmt.Sscanf(line, "Content-Length: %d", &contentLength)
		}
	}
	if contentLength == 0 {
		return nil, fmt.Errorf("no Content-Length header")
	}

	content := make([]byte, contentLength)
	if _, err := io.ReadFull(c.stdout, content); err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}
	return content, nil
}

// readMsg reads one LSP message as a generic map.
func (c *LSPClient) readMsg() (map[string]interface{}, error) {
	content, err := c.readRaw()
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w (content: %s)", err, string(content))
	}
	return result, nil
}

func (c *LSPClient) close() {
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
}

func TestLSP_E2E_Initialize(t *testing.T) {
	client, err := newLSPClient()
	if err != nil {
		t.Fatalf("newLSPClient: %v", err)
	}
	defer client.close()

	// Send initialize
	pid := os.Getpid()
	id, err := client.request("initialize", map[string]interface{}{
		"processId": pid,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"completion": map[string]interface{}{
					"completionItem": map[string]interface{}{
						"snippetSupport": true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("send initialize: %v", err)
	}

	// Read response
	resp, err := client.readMsg()
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}

	if resp["id"] != float64(id) {
		t.Errorf("expected id %d, got %v", id, resp["id"])
	}
	if _, ok := resp["result"]; !ok {
		t.Fatalf("expected 'result' in response, got keys: %v", resp)
	}

	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map, got: %T", resp["result"])
	}
	caps, ok := result["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatalf("capabilities not found in result: %v", result)
	}
	if caps["hoverProvider"] != true {
		t.Errorf("expected hoverProvider true, got %v", caps["hoverProvider"])
	}
	if caps["definitionProvider"] != true {
		t.Errorf("expected definitionProvider true, got %v", caps["definitionProvider"])
	}
	t.Logf("Server capabilities: %d keys", len(caps))
	for k := range caps {
		t.Logf("  capability: %s", k)
	}

	// Send initialized notification
	if err := client.notify("initialized", map[string]interface{}{}); err != nil {
		t.Fatalf("send initialized: %v", err)
	}

	// Send didOpen
	if err := client.notify("textDocument/didOpen", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        "file:///test.jb",
			"languageId": "jabline",
			"version":    1,
			"text":       "let x = 42;\nfn add(a, b) { return a + b; }\nlet r = add(x, 7);",
		},
	}); err != nil {
		t.Fatalf("send didOpen: %v", err)
	}

	// Discard any pending diagnostics notification
	client.readRaw()

	// Send hover request for 'x' at line 0
	id, err = client.request("textDocument/hover", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": "file:///test.jb",
		},
		"position": map[string]interface{}{
			"line":      0,
			"character": 4,
		},
	})
	if err != nil {
		t.Fatalf("send hover: %v", err)
	}

	resp, err = client.readMsg()
	if err != nil {
		t.Errorf("hover readMsg error: %v", err)
	} else {
		if hoverResult, ok := resp["result"].(map[string]interface{}); ok {
			contents := hoverResult["contents"]
			t.Logf("hover: %v", contents)
		}
	}

	// Send completion request (at top of file)
	id, err = client.request("textDocument/completion", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": "file:///test.jb",
		},
		"position": map[string]interface{}{
			"line":      0,
			"character": 0,
		},
	})
	if err != nil {
		t.Fatalf("send completion: %v", err)
	}

	resp, err = client.readMsg()
	if err != nil {
		t.Errorf("completion readMsg error: %v", err)
	} else {
		if items, ok := resp["result"].([]interface{}); ok {
			t.Logf("completion returned %d items", len(items))
		} else {
			t.Logf("completion result type: %T", resp["result"])
		}
	}

	// Send signatureHelp for add( at position of 'x' argument (line 2, col 12)
	id, err = client.request("textDocument/signatureHelp", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": "file:///test.jb",
		},
		"position": map[string]interface{}{
			"line":      2,
			"character": 12,
		},
	})
	if err != nil {
		t.Fatalf("send signatureHelp: %v", err)
	}

	resp, err = client.readMsg()
	if err != nil {
		t.Errorf("signatureHelp readMsg error: %v", err)
	} else {
		t.Logf("signatureHelp result: %v", resp["result"])
	}

	// Send shutdown
	id, err = client.request("shutdown", nil)
	if err != nil {
		t.Fatalf("send shutdown: %v", err)
	}

	resp, err = client.readMsg()
	if err != nil {
		t.Errorf("shutdown readMsg error: %v", err)
	}

	// Send exit notification
	client.notify("exit", nil)

	// Wait for process to exit
	done := make(chan bool, 1)
	go func() {
		client.cmd.Wait()
		done <- true
	}()
	select {
	case <-done:
		t.Log("LSP server exited cleanly")
	case <-time.After(3 * time.Second):
		t.Error("LSP server did not exit after shutdown")
	}
}
