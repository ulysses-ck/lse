package util

import (
	"fmt"
	"lse/ansi"
	"lse/color"
	"lse/config"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Path string
	Info os.FileInfo
}

type DirBlock struct {
	Path    string
	Entries []FileEntry
}

type TreeEntry struct {
	FileEntry
	Prefix string
}

func BuildPrefix(isLastAncestor []bool, isLastCurrent bool) string {
	var prefix string
	for _, isLast := range isLastAncestor {
		if isLast {
			prefix += "    "
		} else {
			prefix += "│   "
		}
	}

	if len(isLastAncestor) > 0 || len(prefix) > 0 {
		if isLastCurrent {
			prefix += "└── "
		} else {
			prefix += "├── "
		}
	}

	return prefix
}

func BuildTreeEntries(cwd string, isLastAncestor []bool, showAll bool, dirsFirst bool) []TreeEntry {
	var treeEntries []TreeEntry

	entries := CollectEntries(cwd, showAll)
	SortEntries(entries, dirsFirst)

	for i, e := range entries {
		isLast := (i == len(entries)-1)

		prefix := BuildPrefix(isLastAncestor, isLast)

		treeEntries = append(treeEntries, TreeEntry{
			FileEntry: e,
			Prefix:    prefix,
		})

		if e.Info.IsDir() {
			nextIsLastAncestor := append(isLastAncestor, isLast)
			subEntries := BuildTreeEntries(e.Path, nextIsLastAncestor, showAll, dirsFirst)
			treeEntries = append(treeEntries, subEntries...)
		}
	}

	return treeEntries
}

func RecurseScan(cwd string, showAll bool) []DirBlock {
	entries := CollectEntries(cwd, showAll)
	var result []DirBlock

	if len(entries) <= 1 {
		return result
	}

	for _, e := range entries {

		if e.Info.IsDir() {
			subEntries := CollectEntries(e.Path, showAll)
			result = append(result, DirBlock{Path: e.Path, Entries: subEntries})
			result = append(result, RecurseScan(e.Path, showAll)...)
		}
	}

	return result
}

func ShowOutput(cfg config.Config, dirsFirst bool, realDirSize bool, entries []FileEntry, recurse bool, showTree bool, cwd string, showAll bool) {
	if showTree {
		// Modo ÁRBOL: Recolecta toda la estructura con prefijos
		// Nota: El modo árbol implica recursión, por lo que 'recurse' es irrelevante aquí.
		treeEntries := BuildTreeEntries(cwd, []bool{}, showAll, dirsFirst)

		if len(treeEntries) == 0 {
			fmt.Printf("No entries found in '%s'\n", cwd)
			return
		}

		fmt.Println(cwd)

		var rows [][]string
		for _, te := range treeEntries {
			rows = append(rows, FormatEntry(te.FileEntry, realDirSize, cfg, true, te.Prefix))
		}
		PrintTable(rows, recurse, true)
		return
	}

	if len(entries) == 0 {
		return
	}

	SortEntries(entries, dirsFirst)

	var rows [][]string
	for _, e := range entries {
		rows = append(rows, FormatEntry(e, realDirSize, cfg, false, ""))
	}

	PrintTable(rows, recurse, false)
}

func CollectEntries(pattern string, showAll bool) []FileEntry {
	paths := CollectPaths(pattern)
	var entries []FileEntry

	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil {
			continue
		}
		name := filepath.Base(path)
		if !showAll && strings.HasPrefix(name, ".") {
			continue
		}
		entries = append(entries, FileEntry{Path: path, Info: info})
	}
	return entries
}

func SortEntries(entries []FileEntry, dirsFirst bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if dirsFirst {
			if a.Info.IsDir() && !b.Info.IsDir() {
				return true
			}
			if !a.Info.IsDir() && b.Info.IsDir() {
				return false
			}
		}
		return strings.ToLower(a.Info.Name()) < strings.ToLower(b.Info.Name())
	})
}

func FormatEntry(e FileEntry, realDirSize bool, cfg config.Config, showTree bool, prefix string) []string {
	fullName := color.Name(e.Info.Name(), e.Info.Mode(), cfg.Icons, cfg.FileTypes)

	if showTree {
		return []string{prefix + fullName}
	}

	perm := color.Permissions(e.Info.Mode().String(), cfg.Permissions)

	var sizeBytes int64
	if e.Info.IsDir() && realDirSize {
		sizeBytes = DirSize(e.Path)
	} else {
		sizeBytes = e.Info.Size()
	}
	size := color.Size(sizeBytes, cfg.Size)

	date := color.Date(e.Info.ModTime(), cfg.Date)

	return []string{perm, size, date, fullName}
}

func PrintTable(rows [][]string, recurse bool, showTree bool) {
	if len(rows) == 0 {
		return
	}

	if showTree && len(rows[0]) == 1 {
		for _, row := range rows {
			fmt.Println(row[0])
		}
		return
	}

	colWidths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, col := range row {
			if w := ansi.VisibleLength(col); w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	for _, row := range rows {
		if recurse {
			fmt.Print(" ")
		}
		for i, col := range row {
			format := ansi.PadString(col, colWidths[i])
			if i < len(row)-1 {
				fmt.Printf("%-*s", colWidths[i], format)
				fmt.Print(" ")
			} else {
				fmt.Print(format)
			}
		}
		fmt.Println()
	}
}

func CollectPaths(pattern string) []string {
	if strings.Contains(pattern, "**") {
		root := strings.Split(pattern, "**")[0]
		if root == "" {
			root = "."
		}
		var paths []string
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			matched, _ := filepath.Match(pattern, path)
			if matched {
				paths = append(paths, path)
			}
			return nil
		})
		return paths
	}

	info, err := os.Stat(pattern)
	if err == nil && info.IsDir() {
		entries, err := os.ReadDir(pattern)
		if err != nil {
			return nil
		}
		var paths []string
		for _, e := range entries {
			paths = append(paths, filepath.Join(pattern, e.Name()))
		}
		return paths
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return matches
}

func DirSize(path string) int64 {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return total
}
