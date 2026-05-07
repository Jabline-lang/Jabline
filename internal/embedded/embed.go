// Package embedded makes all standard .jb library files available
// as an embedded filesystem bundled into the Jabline binary at compile time.
package embedded

import "embed"

// Modules contains all bundled standard library .jb files.
//
//go:embed modules
var Modules embed.FS
