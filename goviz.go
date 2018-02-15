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

	rootType, err := findType(os.Args[1])
	if err != nil {
		log.Fatalf("error getting type: %v", err)
	}
	s, ok := rootType.Type().Underlying().(*types.Struct)
	if !ok {
		return
	}
	fmt.Printf("%s\n", rootType.Type())
	root := newNode(g, fmt.Sprintf("%q", rootType.Type()))
	g.AddNode(root)

	handleStruct(os.Stdout, "  ", g, root, s)
	writeDOT("out.dot", g)
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

	for _, pkg := range tPkgs {
		rootType := pkg.Scope().Lookup(varName)
		if rootType != nil {
			return rootType, nil
		}
	}
	return nil, fmt.Errorf("type not found: %q", typeString)
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
