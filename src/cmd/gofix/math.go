// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"go/ast"
)

var _ fmt.Stringer
var _ os.Error

var mathFix = fix{
	"math",
	math,
	`Remove the leading F from math functions such as Fabs.

http://codereview.appspot.com/5158043
`,
}

func init() {
	register(mathFix)
}

var mathRenames = []struct{ in, out string }{
	{"Fabs", "Abs"},
	{"Fdim", "Dim"},
	{"Fmax", "Max"},
	{"Fmin", "Min"},
	{"Fmod", "Mod"},
}

func math(f *ast.File) bool {
	if !imports(f, "math") {
		return false
	}

	fixed := false

	walk(f, func(n interface{}) {
		// Rename functions.
		if expr, ok := n.(ast.Expr); ok {
			for _, s := range mathRenames {
				if isPkgDot(expr, "math", s.in) {
					expr.(*ast.SelectorExpr).Sel.Name = s.out
					fixed = true
					return
				}
			}
		}
	})
	return fixed
}
