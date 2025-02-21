// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"go/token"
	"os"
	"testing"
)

// TODO(ondbyte) fix this by pointing back to the actual file like, for now using something qaccessible
//var src = readFile("../printer/nodes.go")
var src = readFile("./parser.go")

func readFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func BenchmarkParse(b *testing.B) {
	b.SetBytes(int64(len(src)))
	for i := 0; i < b.N; i++ {
		if _, _, err := ParseFile(token.NewFileSet(), "", src, ParseComments); err != nil {
			b.Fatalf("benchmark failed due to parse error: %s", err)
		}
	}
}

func BenchmarkParseOnly(b *testing.B) {
	b.SetBytes(int64(len(src)))
	for i := 0; i < b.N; i++ {
		if _, _, err := ParseFile(token.NewFileSet(), "", src, ParseComments|SkipObjectResolution); err != nil {
			b.Fatalf("benchmark failed due to parse error: %s", err)
		}
	}
}

func BenchmarkResolve(b *testing.B) {
	b.SetBytes(int64(len(src)))
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fset := token.NewFileSet()
		_, file, err := ParseFile(fset, "", src, SkipObjectResolution)
		if err != nil {
			b.Fatalf("benchmark failed due to parse error: %s", err)
		}
		b.StartTimer()
		handle := fset.File(file.Package)
		resolveFile(file, handle, nil)
	}
}
