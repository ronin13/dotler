package main

import (
	dotler "github.com/ronin13/dotler/dotler"
	"os"
)

func main() {
	dotler.ParseFlags()
	os.Exit(dotler.StartCrawl(dotler.RootURL))
}
