package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

/*! \class SourceFile sourcefile.h
  The SourceFile class models a C++ source file.

  When a SourceFile object is created, it automatically scans the file
  for documented classes and functions, scans HeaderFile files as directed
  and creates Class and Function objects.

  That's all.
*/

// SourceFile or HeaderFile
type File interface {
	Name() estring
}

type SourceFile struct {
	name     estring
	contents estring
}

func (this *SourceFile) Name() estring {
	return this.name
}

/*!  Constructs a SourceFile named \a f, and parses it if it can be opened. */

func NewSourceFile(fname string) {
	log.Printf("New source file: %s", fname)
	contents, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(fmt.Sprintf("Can't read file %s: %s", fname, err))
	}
	sf := SourceFile{
		name:     estring(fname),
		contents: estring(contents),
	}
	sf.Parse()
}

/*! This happy-happy little function parse (or scans, to be truthful)
  a C++ source file looking for documentation. It's the All of this
  class.
*/

func (this *SourceFile) Parse() {
	any := false
	p := newParser(this.contents)
	pfx := estring("/")
	pfx += "*!" // must not see this as one string
	p.scan(pfx)
	for !p.atEnd() {
		any = true
		p.whitespace()
		var f *Function
		var c *Class
		var i *Intro
		var d estring
		l := p.line()
		if p.lookingAt("\\fn ") {
			p.scan(" ")
			f = this.function(p)
			d = p.textUntil("*/")
		} else if p.lookingAt("\\chapter ") {
			p.scan(" ")
			name := p.word()
			if name.isEmpty() {
				docError(this, p.line(), "\\chapter must be followed by name")
			}
			i = newIntro(name)
			p.whitespace()
			d = p.textUntil("*/")
		} else if p.lookingAt("\\class ") {
			p.scan(" ")
			className := p.identifier()
			if className.isEmpty() {
				docError(this, l, "\\class must be followed by a class name")
			}
			c = findClass(className)
			if c == nil {
				c = newClass(className, nil, 0)
			}
			p.whitespace()
			hn := p.word()
			for p.lookingAt(".") {
				p.step()
				hn += "."
				hn += p.word()
			}
			if hn.length() < 2 || hn.mid(hn.length()-2, hn.length()) != ".h" {
				docError(this, l, "Missing header file name")
			} else {
				if !this.contents.contains("#include \""+hn+"\"") &&
					!this.contents.contains("#include <"+hn+">") {
					docError(this, l, "File does not include "+hn)
				}
				h := findHeaderFile(hn)
				if h == nil {
					if this.Name().contains("/") {
						dir := this.Name()
						i := dir.length() - 1
						for i > 0 && dir[i] != '/' {
							i--
						}
						hn = dir.mid(0, i+1) + hn
					}
					h = newHeaderFile(hn)
					if !h.valid() {
						docError(this, l, "Cannot find header file "+hn+" (for class "+className+")")
					}
				}
				if len(c.members()) == 0 {
					docError(this, l, "Cannot find any "+className+" members in "+hn)
				}
			}
			d = p.textUntil("*/")
		} else if p.lookingAt("\\nodoc") {
			any = true
			d = "hack"
		} else {
			d = p.textUntil("*/")
			f = this.function(p)
		}
		if d.isEmpty() {
			docError(this, l, "Comment contains no documentation")
		} else if f != nil {
			newDocBlockForFunction(this, l, d, f)
		} else if c != nil {
			newDocBlockForClass(this, l, d, c)
		} else if i != nil {
			newDocBlockForIntro(this, l, d, i)
		}

		/* udoc must not see that as one string */
		str := estring("/")
		str += "*!"
		p.scan(str)
	}
	if !any {
		p := newParser(this.contents)
		p.scan("::") // any source in this file at all?
		if !p.atEnd() {
			docError(this, p.line(), "File contains no documentation")
		}
	}
}

/*! This helper parses a function name using \a p or reports an
  error. It returns a pointer to the function, or a null pointer in
  case of error.
*/

func (this *SourceFile) function(p *Parser) *Function {
	var f *Function
	t := p.parseType()
	l := p.line()
	n := p.identifier()
	if n.isEmpty() && p.lookingAt("(") && t.find(":") > 0 {
		// constructor support hack. eeek.
		n = t
		t = ""
	}
	a := p.argumentList()
	p.whitespace()
	cn := false
	if p.lookingAt("const") {
		p.word()
		cn = true
	}
	if !n.isEmpty() && n.find(":") > 0 &&
		!a.isEmpty() {
		f = findFunction(n, a, cn)
		if f != nil {
			f.setArgumentList(a)
		} else {
			f = newFunction(t, n, a, cn, this, l)
		}
	} else {
		docError(this, l, "Unable to parse function name")
	}
	return f
}
