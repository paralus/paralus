//go:build ignore
// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

var dataFS http.FileSystem = http.Dir("data")

func main() {

	err := vfsgen.Generate(dataFS, vfsgen.Options{
		Filename:     "defaults.go",
		PackageName:  "authz",
		VariableName: "defaults",
	})
	if err != nil {
		log.Fatalln(err)
	}

}
