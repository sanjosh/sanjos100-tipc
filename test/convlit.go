// errchk $G -e $D/$F.go

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// explicit conversion of constants is work in progress.
// the ERRORs in this block are debatable, but they're what
// the language spec says for now.
var x1 = string(1);
var x2 string = string(1);
var x3 = int(1.5);	// ERROR "convert|truncate"
var x4 int = int(1.5);	// ERROR "convert|truncate"
var x5 = "a" + string(1);
var x6 = int(1e100);	// ERROR "overflow"
var x7 = float(1e1000);	// ERROR "overflow"

// implicit conversions merit scrutiny
var s string;
var bad1 string = 1;	// ERROR "conver|incompatible"
var bad2 = s + 1;		// ERROR "conver|incompatible"
var bad3 = s + 'a';	// ERROR "conver|incompatible"
var bad4 = "a" + 1;	// ERROR "literals|incompatible|convert"
var bad5 = "a" + 'a';	// ERROR "literals|incompatible|convert"

var bad6 int = 1.5;	// ERROR "convert|truncate"
var bad7 int = 1e100;	// ERROR "overflow"
var bad8 float32 = 1e200;	// ERROR "overflow"

// but these implicit conversions are okay
var good1 string = "a";
var good2 int = 1.0;
var good3 int = 1e9;
var good4 float = 1e20;

