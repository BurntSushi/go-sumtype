package main

import (
	"github.com/BurntSushi/go-sumtype/pkg/sumtype"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(sumtype.Analyzer)
}
