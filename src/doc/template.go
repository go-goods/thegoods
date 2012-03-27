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
	godoc "go/doc"
	"path"
	"text/template"
)

var Funcs = template.FuncMap{
	"comment": commentFmt,
	"cmdName": cmdNameFmt,
	"decl":    declFmt,
}

// commentFmt formats a source code control comment as HTML.
func commentFmt(v string) string {
	var buf bytes.Buffer
	godoc.ToHTML(&buf, v, nil)
	return buf.String()
}

// declFmt formats a Decl as HTML.
func declFmt(decl Decl) string {
	return template.HTMLEscapeString(decl.Text)
}

// cmdNameFmt formats a doc.PathInfo as a command name.
func cmdNameFmt(pi PathInfo) string {
	_, name := path.Split(pi.ImportPath())
	return template.HTMLEscapeString(name)
}
