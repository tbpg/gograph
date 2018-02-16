/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// gograph generates a DOT graph of the given type.
// Usage:
//     gograph [path/to/package.type]
//
// For example,
//     gograph github.com/tbpg/gograph.node
//
// Debug info is printed to Stdout and the .dot file is written to out.dot.
// You can gerate a png with:
//     dot -Tpng out.dot -o out.png
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
)

type node struct {
	graph.Node
	name string
}

func newNode(g *simple.DirectedGraph, name string) node {
	return node{g.NewNode(), name}
}

func (n node) Attributes() []encoding.Attribute {
	return []encoding.Attribute{{Key: "label", Value: n.name}}
}

func main() {
	g, err := typeGraph(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	writeDOT("out.dot", g)
	http.HandleFunc("/dot", handleDOT)
	http.HandleFunc("/rawdot", handleRawDOT)
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.ListenAndServe(":8080", nil)
}

func handleRawDOT(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query()["type"]
	if len(t) != 1 {
		return
	}

	g, err := typeGraph(t[0])
	if err != nil {
		log.Printf("typeGraph error: %v\n", err)
		return
	}
	b, err := marshalDOT(g)
	if err != nil {
		log.Printf("marshalDOT error: %v\n", err)
		return
	}
	w.Write(b)
}

func handleDOT(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query()["type"]
	if len(t) != 1 {
		return
	}

	g, err := typeGraph(t[0])
	if err != nil {
		return
	}
	b, err := marshalDOT(g)
	if err != nil {
		return
	}
	buf := &bytes.Buffer{}
	buf.Write(b)

	cmd := exec.Command("dot", "-Tpng")
	cmd.Stdout = w
	cmd.Stdin = buf
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(w, "failed to generate")
	}
}

func typeGraph(typeString string) (graph.Graph, error) {
	g := simple.NewDirectedGraph()

	rootType, err := findType(typeString)
	if err != nil {
		return nil, err
	}
	s, ok := rootType.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("not a struct")
	}
	fmt.Printf("%s\n", rootType.Type())
	root := newNode(g, fmt.Sprintf("%q", rootType.Type()))
	g.AddNode(root)

	processStruct(os.Stdout, "  ", g, root, s)

	return g, nil
}

func findType(typeString string) (types.Object, error) {
	fset := token.NewFileSet()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	split := strings.Split(typeString, ".")
	path := filepath.Join(gopath, "src", strings.Join(split[0:len(split)-1], "."))
	varName := split[len(split)-1]

	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	conf := types.Config{Importer: importer.Default()}

	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	var tPkgs []*types.Package
	for _, pkg := range pkgs {
		var files []*ast.File
		for _, f := range pkg.Files {
			files = append(files, f)
		}
		pkg, err := conf.Check(pkg.Name, fset, files, info)
		if err != nil {
			return nil, err
		}
		tPkgs = append(tPkgs, pkg)
	}

	for _, pkg := range tPkgs {
		rootType := pkg.Scope().Lookup(varName)
		if rootType != nil {
			return rootType, nil
		}
	}
	return nil, fmt.Errorf("type not found: %q", typeString)
}

func processStruct(w io.Writer, p string, g *simple.DirectedGraph, parent node, s *types.Struct) {
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		t := f.Type()
		fmt.Fprintf(w, "%s%v\n", p, f.Type())
		n := newNode(g, fmt.Sprintf("%q", f.Type()))
		g.AddNode(n)
		e := g.NewEdge(parent, n)
		g.SetEdge(e)
		if ss, ok := t.Underlying().(*types.Struct); ok {
			processStruct(w, p+"  ", g, n, ss)
		}
	}
}

func writeDOT(filename string, g graph.Graph) error {
	b, err := marshalDOT(g)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0644)
}

func marshalDOT(g graph.Graph) ([]byte, error) {
	return dot.Marshal(g, "goviz", "", "", false)
}
