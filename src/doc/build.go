// Copyright 2011 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package doc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type PathInfo interface {
	ImportPath() string
	ProjectPrefix() string
	ProjectName() string
	ProjectURL() string
	Package(*http.Client) (*Package, error)
}

func startsWithUppercase(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(r)
}

// builder holds the state used when building the documentation.
type builder struct {
	fset        *token.FileSet
	examples    []*doc.Example
	buf         bytes.Buffer // scratch space for printNode method.
	importPaths map[string]map[string]string
	pkg         *ast.Package
}

type TypeAnnotation struct {
	Pos, End   int
	ImportPath string
	Name       string
}

type Decl struct {
	Text        string
	Annotations []TypeAnnotation
}

type sortByPos []TypeAnnotation

func (p sortByPos) Len() int           { return len(p) }
func (p sortByPos) Less(i, j int) bool { return p[i].Pos < p[j].Pos }
func (p sortByPos) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// annotationVisitor collects type annotations.
type annotationVisitor struct {
	annotations []TypeAnnotation
	fset        *token.FileSet
	b           *builder
}

func (v *annotationVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.TypeSpec:
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		return nil
	case *ast.FuncDecl:
		if n.Recv != nil {
			ast.Walk(v, n.Recv)
		}
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		return nil
	case *ast.Field:
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		return nil
	case *ast.ValueSpec:
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		return nil
	case *ast.FuncLit:
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		return nil
	case *ast.CompositeLit:
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		return nil
	case *ast.Ident:
		if !ast.IsExported(n.Name) {
			return nil
		}
		v.addAnnoation(n, "", n.Name)
		return nil
	case *ast.SelectorExpr:
		if !ast.IsExported(n.Sel.Name) {
			return nil
		}
		if i, ok := n.X.(*ast.Ident); ok {
			v.addAnnoation(n, i.Name, n.Sel.Name)
			return nil
		}
	}
	return v
}

const packageWrapper = "package p\n"

func (v *annotationVisitor) addAnnoation(n ast.Node, packageName string, name string) {
	pos := v.fset.Position(n.Pos())
	end := v.fset.Position(n.End())
	v.annotations = append(v.annotations, TypeAnnotation{
		pos.Offset - len(packageWrapper),
		end.Offset - len(packageWrapper),
		packageName,
		name})
}

func (b *builder) printDecl(decl ast.Node) Decl {
	b.buf.Reset()
	b.buf.WriteString(packageWrapper)
	err := (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(&b.buf, b.fset, decl)
	if err != nil {
		return Decl{Text: err.Error()}
	}
	text := string(b.buf.Bytes()[len(packageWrapper):])
	v := &annotationVisitor{
		b:    b,
		fset: token.NewFileSet(),
	}
	f, err := parser.ParseFile(v.fset, "", b.buf.Bytes(), 0)
	if err != nil {
		return Decl{Text: text}
	}
	ast.Walk(v, f)
	sort.Sort(sortByPos(v.annotations))
	return Decl{Text: text, Annotations: v.annotations}
}

func (b *builder) printNode(node interface{}) string {
	b.buf.Reset()
	err := (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(&b.buf, b.fset, node)
	if err != nil {
		b.buf.Reset()
		b.buf.WriteString(err.Error())
	}
	return b.buf.String()
}

type Value struct {
	Decl Decl
	Doc  string
}

func (b *builder) values(vdocs []*doc.Value) []*Value {
	var result []*Value
	for _, d := range vdocs {
		result = append(result, &Value{
			Decl: b.printDecl(d.Decl),
			Doc:  d.Doc,
		})
	}
	return result
}

type Example struct {
	Name   string
	Doc    string
	Code   string
	Output string
}

var exampleOutputRx = regexp.MustCompile(`(?i)//[[:space:]]*output:`)

func (b *builder) getExamples(name string) []Example {
	var docs []Example
	for _, e := range b.examples {
		n := e.Name
		if i := strings.LastIndex(n, "_"); i >= 0 {
			if i < len(n)-1 && !startsWithUppercase(n[i+1:]) {
				n = n[:i]
			}
		}
		if n != name {
			continue
		}

		output := e.Output
		code := b.printNode(&printer.CommentedNode{
			Node:     e.Code,
			Comments: e.Comments,
		})

		// additional formatting if this is a function body
		if i := len(code); i >= 2 && code[0] == '{' && code[i-1] == '}' {
			// remove surrounding braces
			code = code[1 : i-1]
			// unindent
			code = strings.Replace(code, "\n    ", "\n", -1)
			// remove output comment
			if j := exampleOutputRx.FindStringIndex(code); j != nil {
				code = strings.TrimSpace(code[:j[0]])
			}
		} else {
			// drop output, as the output comment will appear in the code
			output = ""
		}
		docs = append(docs, Example{Name: e.Name, Doc: e.Doc, Code: code, Output: output})
	}
	return docs
}

type Func struct {
	Decl     Decl
	Doc      string
	Name     string
	Recv     string
	Examples []Example
}

func (b *builder) funcs(fdocs []*doc.Func) []*Func {
	var result []*Func
	for _, d := range fdocs {
		var exampleName string
		switch {
		case d.Recv == "":
			exampleName = d.Name
		case d.Recv[0] == '*':
			exampleName = d.Recv[1:] + "_" + d.Name
		default:
			exampleName = d.Recv + "_" + d.Name
		}
		result = append(result, &Func{
			Decl:     b.printDecl(d.Decl),
			Doc:      d.Doc,
			Name:     d.Name,
			Recv:     d.Recv,
			Examples: b.getExamples(exampleName),
		})
	}
	return result
}

type Type struct {
	Doc      string
	Name     string
	Decl     Decl
	Consts   []*Value
	Vars     []*Value
	Funcs    []*Func
	Methods  []*Func
	Examples []Example
}

func (b *builder) types(tdocs []*doc.Type) []*Type {
	var result []*Type
	for _, d := range tdocs {
		result = append(result, &Type{
			Doc:      d.Doc,
			Name:     d.Name,
			Decl:     b.printDecl(d.Decl),
			Consts:   b.values(d.Consts),
			Vars:     b.values(d.Vars),
			Funcs:    b.funcs(d.Funcs),
			Methods:  b.funcs(d.Methods),
			Examples: b.getExamples(d.Name),
		})
	}
	return result
}

type File struct {
	Name string
}

func (b *builder) files(filenames []string) []*File {
	var result []*File
	for _, f := range filenames {
		_, name := path.Split(f)
		result = append(result, &File{
			Name: name,
		})
	}
	return result
}

type Package struct {
	// The import path for this package.
	ImportPath string

	// Package name or "" if no package for this import path.
	Name string

	// Synopsis and full documentation for package.
	Doc string

	// The time this object was created.
	Updated time.Time

	// Top-level declarations.
	Consts []*Value
	Funcs  []*Func
	Types  []*Type
	Vars   []*Value

	// Examples
	Examples []Example

	// Non-test files.
	Files []*File
}

func buildDoc(importPath string, files []string) (*Package, error) {

	b := &builder{
		fset:        token.NewFileSet(),
		importPaths: make(map[string]map[string]string),
	}

	pkgs := make(map[string]*ast.Package)
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		if src, err := parser.ParseFile(b.fset, f, nil, parser.ParseComments); err == nil {
			name := src.Name.Name
			pkg, found := pkgs[name]
			if !found {
				pkg = &ast.Package{Name: name, Files: make(map[string]*ast.File)}
				pkgs[name] = pkg
			}
			pkg.Files[f] = src
		}
	}
	score := 0
	for _, pkg := range pkgs {
		switch {
		case score < 3 && strings.HasSuffix(importPath, pkg.Name):
			b.pkg = pkg
			score = 3
		case score < 2 && pkg.Name != "main":
			b.pkg = pkg
			score = 2
		case score < 1:
			b.pkg = pkg
			score = 1
		}
	}

	if b.pkg == nil {
		return nil, fmt.Errorf("Package %s not found", importPath)
	}

	ast.PackageExports(b.pkg)
	pdoc := doc.New(b.pkg, importPath, 0)

	pdoc.Doc = strings.TrimRight(pdoc.Doc, " \t\n\r")

	// Collect examples.
	for _, f := range files {
		if !strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := parser.ParseFile(b.fset, f, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		if src.Name.Name != pdoc.Name && src.Name.Name != pdoc.Name+"_test" {
			continue
		}
		b.examples = append(b.examples, doc.Examples(src)...)
	}

	return &Package{
		Consts:     b.values(pdoc.Consts),
		Doc:        pdoc.Doc,
		Examples:   b.getExamples(""),
		Files:      b.files(pdoc.Filenames),
		Funcs:      b.funcs(pdoc.Funcs),
		ImportPath: pdoc.ImportPath,
		Name:       pdoc.Name,
		Types:      b.types(pdoc.Types),
		Updated:    time.Now(),
		Vars:       b.values(pdoc.Vars),
	}, nil
}
