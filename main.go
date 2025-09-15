package main

import (
	"flag"
	"fmt"
	"lse/ansi"
	"lse/config"
	"lse/util"
)

var (
	configFile  string
	dirsFirst   bool
	showAll     bool
	realDirSize bool
	recurse     bool
)

func init() {
	flag.StringVar(&configFile, "cfg", "~/.config/lse.json", "config file")
	flag.BoolVar(&dirsFirst, "d", false, "show directories first")
	flag.BoolVar(&showAll, "a", false, "show hidden files")
	flag.BoolVar(&realDirSize, "s", false, "show real dir size")
	flag.BoolVar(&recurse, "R", false, "display recursively content directories")
	flag.Parse()
}

func main() {
	pattern := "."
	if flag.NArg() > 0 {
		pattern = flag.Arg(0)
	}

	cfg := config.LoadConfig(configFile)

	entries := util.CollectEntries(pattern, showAll)
	util.ShowOutput(cfg, dirsFirst, realDirSize, entries, false)

	if recurse {
		fmt.Println()

		subEntries := util.RecurseScan(pattern, showAll)

		for _, sl := range subEntries {
			path := fmt.Sprintf(" %s%s %s/%s", cfg.FileTypes.Directory, cfg.Icons.Directory, sl.Path, ansi.Reset)
			fmt.Println(path)
			util.ShowOutput(cfg, dirsFirst, realDirSize, sl.Entries, recurse)
			fmt.Println()
		}
	}
}
