package fixtures

import (
	"io/ioutil"
	"text/template"
)

var (
	// DownloadTemplate is the template for rendering download yaml
	DownloadTemplate *template.Template
)

func init() {
	f, err := Fixtures.Open("download.yaml")
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	DownloadTemplate, err = template.New("download").Parse(string(b))
	if err != nil {
		panic(err)
	}
}
