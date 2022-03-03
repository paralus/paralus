//go:build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

var fs http.FileSystem = http.Dir("fixtures/data")

func main() {

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "fixtures/fixtures.go",
		PackageName:  "fixtures",
		VariableName: "Fixtures",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
