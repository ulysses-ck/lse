package main

import (
	"flag"
	"lse/config"
	"lse/util"
)

var (
	configFile  string
	dirsFirst   bool
	showAll     bool
	realDirSize bool
	recurse     bool
	showTree    bool
)

func init() {
	flag.StringVar(&configFile, "cfg", "~/.config/lse.json", "config file")
	flag.BoolVar(&dirsFirst, "d", false, "show directories first")
	flag.BoolVar(&showAll, "a", false, "show hidden files")
	flag.BoolVar(&realDirSize, "s", false, "show real dir size")
	flag.BoolVar(&recurse, "R", false, "display recursively content directories")
	flag.BoolVar(&showTree, "t", false, "shows a tree structure")
	flag.Parse()
}

func main() {
	pattern := "."
	if flag.NArg() > 0 {
		pattern = flag.Arg(0)
	}

	cfg := config.LoadConfig(configFile)

	if showTree {
		util.ShowOutput(cfg, dirsFirst, realDirSize, nil, recurse, showTree, pattern, showAll)
	} else {
		entries := util.CollectEntries(pattern, showAll)
		util.ShowOutput(cfg, dirsFirst, realDirSize, entries, recurse, showTree, pattern, showAll)
	}
}
