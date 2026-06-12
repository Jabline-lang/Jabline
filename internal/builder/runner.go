//go:build runner

package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"

	"jabline/internal/embedded"
	"jabline/pkg/compiler"
	"jabline/pkg/vm"
)

var MagicMarker = []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}

func main() {
	selfPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error obtaining executable path: %s\n", err)
		os.Exit(1)
	}

	selfBytes, err := ioutil.ReadFile(selfPath)
	if err != nil {
		fmt.Printf("Error reading executable: %s\n", err)
		os.Exit(1)
	}

	// Read backwards from selfBytes
	fsize := len(selfBytes)
	markerSize := len(MagicMarker)
	
	if fsize < markerSize+8 {
		fmt.Printf("Invalid standalone binary (too small)\n")
		os.Exit(1)
	}

	// Verify Marker
	for i := 0; i < markerSize; i++ {
		if selfBytes[fsize-markerSize+i] != MagicMarker[i] {
			fmt.Printf("No Jabline bytecode found in binary.\n")
			os.Exit(1)
		}
	}

	// Read bytecode length
	sizeBuf := selfBytes[fsize-markerSize-8 : fsize-markerSize]
	bcSize := binary.LittleEndian.Uint64(sizeBuf)

	if uint64(fsize) < bcSize+uint64(markerSize)+8 {
		fmt.Printf("Invalid standalone binary: corrupted bytecode payload\n")
		os.Exit(1)
	}

	// Read bytecode
	bytecodeData := selfBytes[uint64(fsize)-uint64(markerSize)-8-bcSize : uint64(fsize)-uint64(markerSize)-8]

	bytecode, err := compiler.Deserialize(bytecodeData)
	if err != nil {
		fmt.Printf("Error deserializing bytecode: %s\n", err)
		os.Exit(1)
	}

	vm.EmbeddedModules = embedded.Modules
	loader := vm.NewModuleLoaderWithEmbed(embedded.Modules)
	machine := vm.NewWithLoader(bytecode.Instructions, bytecode.Constants, "main", loader)
	vm.GlobalVM = machine

	if bytecode.Sandbox != "" {
		if err := machine.SetSandboxLevel(bytecode.Sandbox); err != nil {
			fmt.Printf("Error: invalid embedded sandbox level %q: %s\n", bytecode.Sandbox, err)
			os.Exit(1)
		}
	}

	err = machine.Run()
	if err != nil {
		fmt.Printf("Jabline Runtime Error: %s\n", err)
		os.Exit(1)
	}
}
