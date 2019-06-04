package main

import (
	"os"
	"path/filepath"
	"strings"
)

func main() {
	newWebpage(".")

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".cpp") {
			NewSourceFile(path)
		}
		return nil
	})
	buildHierarchy()
	outputIntro()
	outputClasses()
	webpage.endPage()
}
