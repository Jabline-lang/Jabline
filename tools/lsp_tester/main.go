package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type DiagnosticParam struct {
	URI         string        `json:"uri"`
	Diagnostics []interface{} `json:"diagnostics"`
}

func main() {

	fmt.Println("[TEST] Starting jabline lsp server...")
	cmd := exec.Command("./jabline", "lsp")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer cmd.Process.Kill()

	go io.Copy(os.Stderr, stderr)

	reader := bufio.NewReader(stdout)

	idCounter := 0
	send := func(method string, params interface{}) int {
		idCounter++
		msg := Request{
			JSONRPC: "2.0",
			ID:      idCounter,
			Method:  method,
			Params:  params,
		}
		body, _ := json.Marshal(msg)
		header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
		fmt.Fprintf(stdin, "%s%s", header, body)
		return idCounter
	}

	sendNotification := func(method string, params interface{}) {
		msg := Notification{
			JSONRPC: "2.0",
			Method:  method,
			Params:  params,
		}
		body, _ := json.Marshal(msg)
		header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
		fmt.Fprintf(stdin, "%s%s", header, body)
	}

	read := func() (string, json.RawMessage, interface{}) {
		tp := textproto.NewReader(reader)
		headers, err := tp.ReadMIMEHeader()
		if err != nil {
			log.Fatalf("Failed to read header: %v", err)
		}
		lengthStr := headers.Get("Content-Length")
		length, _ := strconv.Atoi(lengthStr)

		body := make([]byte, length)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			log.Fatalf("Failed to read body: %v", err)
		}

		var base struct {
			Method string          `json:"method"`
			Result json.RawMessage `json:"result"`
			Params interface{}     `json:"params"`
			ID     *int            `json:"id"`
		}
		json.Unmarshal(body, &base)
		return base.Method, base.Result, base.Params
	}

	fmt.Println("[TEST] Sending 'initialize'...")
	send("initialize", map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   "file:///tmp/test",
		"capabilities": map[string]interface{}{},
	})

	_, result, _ := read()
	fmt.Printf("[PASS] Server Initialized. Capabilities received: %t\n", len(result) > 0)

	sendNotification("initialized", map[string]interface{}{})

	fmt.Println("[TEST] Sending 'textDocument/didOpen' with invalid code...")
	invalidCode := "let x = ;"
	sendNotification("textDocument/didOpen", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        "file:///tmp/test/main.jb",
			"languageId": "jabline",
			"version":    1,
			"text":       invalidCode,
		},
	})

	fmt.Println("[TEST] Waiting for diagnostics...")
	
	gotDiagnostics := false
	for i := 0; i < 5; i++ {
		method, _, params := read()
		if method == "textDocument/publishDiagnostics" {

			pBytes, _ := json.Marshal(params)
			var diagParams DiagnosticParam
			json.Unmarshal(pBytes, &diagParams)
			
			if len(diagParams.Diagnostics) > 0 {
				fmt.Printf("[PASS] Diagnostics received! Found %d errors.\n", len(diagParams.Diagnostics))
				gotDiagnostics = true
				break
			}
		}
	}
	if !gotDiagnostics {
		fmt.Println("[FAIL] No diagnostics received for invalid code.")
	}

	fmt.Println("[TEST] Testing Autocomplete...")
	validCode := "let myVar = 10;\n"
	sendNotification("textDocument/didChange", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     "file:///tmp/test/main.jb",
			"version": 2,
		},
		"contentChanges": []map[string]interface{}{
			{"text": validCode},
		},
	})
	
	send("textDocument/completion", map[string]interface{}{
		"textDocument": map[string]interface{}{"uri": "file:///tmp/test/main.jb"},
		"position":     map[string]interface{}{"line": 1, "character": 0},
	})
	
	method, result, _ := read()
	if method == "" && len(result) > 0 {
		var items []interface{}
		json.Unmarshal(result, &items)
		fmt.Printf("[PASS] Completion successful. Items: %d\n", len(items))
	} else {
		fmt.Println("[FAIL] Completion failed or unexpected message:", method)
	}

	fmt.Println("[TEST] Testing Hover on 'myVar'...")
	send("textDocument/hover", map[string]interface{}{
		"textDocument": map[string]interface{}{"uri": "file:///tmp/test/main.jb"},
		"position":     map[string]interface{}{"line": 0, "character": 5},
	})
	
	_, hoverResult, _ := read()
	fmt.Printf("[PASS] Hover result received: %s\n", string(hoverResult))

	fmt.Println("[TEST] Shutting down...")
	send("shutdown", nil)
	read()
	sendNotification("exit", nil)
	
	time.Sleep(100 * time.Millisecond)
	fmt.Println("[DONE] All tests completed.")
}
