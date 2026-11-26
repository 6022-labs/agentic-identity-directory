//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// abigen copies every shared Solidity struct into each contract's generated file,
// so folding multiple contracts into one package double-declares types like KeyValue.
// This strips duplicate type declarations: the first file alphabetically keeps each
// type, later files lose their copy. Drops the associated doc comments too so the
// reformatted output stays clean.
func main() {
	directory := strings.TrimSpace(strings.TrimSuffix(os.Args[1], "/"))
	if len(directory) == 0 {
		directory = "."
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("failed to read directory, error: %v\n", err)
		os.Exit(1)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".go") && !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	seen := map[string]bool{}
	for _, name := range names {
		dedupe_file(directory+"/"+name, seen)
	}
}

func dedupe_file(path string, seen map[string]bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("failed to parse %s, error: %v\n", path, err)
		os.Exit(1)
	}

	dropped := map[*ast.CommentGroup]bool{}
	keptDecls := make([]ast.Decl, 0, len(file.Decls))
	removed := 0

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			keptDecls = append(keptDecls, decl)
			continue
		}

		keptSpecs := make([]ast.Spec, 0, len(genDecl.Specs))
		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			if seen[typeSpec.Name.Name] {
				dropped[typeSpec.Doc] = true
				dropped[typeSpec.Comment] = true
				removed++
				continue
			}
			seen[typeSpec.Name.Name] = true
			keptSpecs = append(keptSpecs, spec)
		}

		if len(keptSpecs) == 0 {
			dropped[genDecl.Doc] = true
			continue
		}
		genDecl.Specs = keptSpecs
		keptDecls = append(keptDecls, genDecl)
	}

	if removed == 0 {
		return
	}

	file.Decls = keptDecls
	keptComments := file.Comments[:0]
	for _, group := range file.Comments {
		if !dropped[group] {
			keptComments = append(keptComments, group)
		}
	}
	file.Comments = keptComments

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		fmt.Printf("failed to format %s, error: %v\n", path, err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Removed %d duplicate type declaration(s) from %s\n", removed, path)
}
