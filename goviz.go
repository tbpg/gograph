package main

import (
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
	"os"
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
	g := simple.NewDirectedGraph()
	fset := token.NewFileSet()

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	split := strings.Split(os.Args[1], ".")
	path := filepath.Join(gopath, "src", strings.Join(split[0:len(split)-1], "."))
	varName := split[len(split)-1]

	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err) // type error
		}
		tPkgs = append(tPkgs, pkg)
	}

	var rootType types.Object
	for _, pkg := range tPkgs {
		fmt.Println(pkg.Name())
		rootType = pkg.Scope().Lookup(varName)
		if rootType != nil {
			break
		}
	}
	if rootType == nil {
		log.Fatalf("Not found: %s\n", os.Args[1])
	}

	root := newNode(g, fmt.Sprintf("%q", rootType.Name()))
	g.AddNode(root)
	s, ok := rootType.Type().Underlying().(*types.Struct)
	if !ok {
		return
	}
	handleStruct(os.Stdout, "  ", g, root, s)
	writeDOT("out.dot", g)
}

func handleStruct(w io.Writer, p string, g *simple.DirectedGraph, parent node, s *types.Struct) {
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		t := f.Type()
		fmt.Fprintf(w, "%s%v\n", p, f.Type())
		n := newNode(g, fmt.Sprintf("%q", f.Type()))
		g.AddNode(n)
		e := g.NewEdge(parent, n)
		g.SetEdge(e)
		if ss, ok := t.Underlying().(*types.Struct); ok {
			handleStruct(w, p+"  ", g, n, ss)
		}
	}
}

func writeDOT(filename string, g graph.Graph) error {
	b, err := dot.Marshal(g, "goviz", "", "", false)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0644)
}
