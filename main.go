package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"jabline/cmd"
	"jabline/pkg/compiler"
	"jabline/pkg/vm"
)

func main() {
	if tryRunStandalone() {
		return
	}
	cmd.Execute()
}

func tryRunStandalone() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	f, err := os.Open(exePath)
	if err != nil {
		return false
	}
	defer f.Close()

	markerLen := int64(len(cmd.MagicMarker))
	fInfo, err := f.Stat()
	if err != nil {
		return false
	}
	fileSize := fInfo.Size()

	if fileSize < markerLen+8 {
		return false
	}

	_, err = f.Seek(-markerLen, 2)
	if err != nil { return false }
	
	markerBuf := make([]byte, markerLen)
	_, err = f.Read(markerBuf)
	if err != nil { return false }

	if !bytes.Equal(markerBuf, cmd.MagicMarker) {
		return false
	}

	_, err = f.Seek(-(markerLen + 8), 2)
	if err != nil { return false }
	
	sizeBuf := make([]byte, 8)
	_, err = f.Read(sizeBuf)
	if err != nil { return false }
	
	bytecodeSize := int64(binary.LittleEndian.Uint64(sizeBuf))

	bytecodeStart := fileSize - markerLen - 8 - bytecodeSize
	if bytecodeStart < 0 {
		return false
	}

	_, err = f.Seek(bytecodeStart, 0)
	if err != nil { return false }
	
	bytecodeData := make([]byte, bytecodeSize)
	_, err = f.Read(bytecodeData)
	if err != nil { return false }

	bytecode, err := compiler.Deserialize(bytecodeData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load embedded bytecode: %s\n", err)
		os.Exit(1)
	}

	machine := vm.New(bytecode.Instructions, bytecode.Constants, "<embedded>")
	err = machine.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %s\n", err)
		os.Exit(1)
	}

	return true
}
