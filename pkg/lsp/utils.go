package lsp

import (
	"strings"
)

func getLine(content string, lineNumber int) string {
	lines := strings.Split(content, "\n")
	if lineNumber >= 1 && lineNumber <= len(lines) {
		return lines[lineNumber-1]
	}
	return ""
}

func getByteOffset(content string, line, col int) int {
	lines := strings.Split(content, "\n")
	byteOffset := 0
	for i := 0; i < line-1; i++ {
		byteOffset += len(lines[i]) + 1
	}
	
	if col-1 > len(lines[line-1]) {
		return -1
	}
	return byteOffset + col - 1
}

func getCharColumn(content string, byteOffset int) int {
	lineStartByteOffset := 0
	lineNumber := 1
	for i := 0; i < len(content) && i < byteOffset; i++ {
		if content[i] == '\n' {
			lineNumber++
			lineStartByteOffset = i + 1
		}
	}
	return (byteOffset - lineStartByteOffset) + 1
}

func getTokenByteOffset(content string, line, col int) int {
	return getByteOffset(content, line, col)
}

func isWhitespaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func ptr[T any](v T) *T {
	return &v
}
