package main

import (
	"fmt"
	"os"

	"github.com/American-Certified-Brands/config-sample/ReadConfig"
	"github.com/pschlump/godebug"
)

// GlobalConfigData is the gloal configuration data.
// It holds all the data from the cfg.json file.
type GlobalConfigData struct {
	ExampeWithDefault string `default:"dflt-1"`
	SomePassword      string `default:"dflt-2"`
}

var gCfg GlobalConfigData // global configuration data.

func main() {
	err := ReadConfig.ReadFile("./testdata/a.json", &gCfg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("SUCCESS: read %s\n", godebug.SVarI(gCfg))
}
