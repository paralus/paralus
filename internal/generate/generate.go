//go:build ignore
// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

var fixtureFS http.FileSystem = http.Dir("fixtures/data")

func main() {

	err := vfsgen.Generate(fixtureFS, vfsgen.Options{
		Filename:     "fixtures/fixtures.go",
		PackageName:  "fixtures",
		VariableName: "Fixtures",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
