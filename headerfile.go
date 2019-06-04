package main

import (
	"io/ioutil"
	"log"
)

/*! \class HeaderFile headerfile.h
  The HeaderFile class models a header file.

  The HeaderFile file is viewed as a collection of class { ... }
  statements, each of which is scanned for member functions and
  superclass names. Other content is ignored (for now - enums may
  one day be handled).
*/

type HeaderFile struct {
	name     estring
	contents estring
	v        bool
}

func (this *HeaderFile) Name() estring {
	return this.name
}

func (this HeaderFile) valid() bool {
	return this.v
}

var headers []*HeaderFile

/*! Constructs a HeaderFile for \a file, which is presumed to be in the
  current directory.

  The file is parsed immediately.
*/

func newHeaderFile(fn estring) *HeaderFile {
	if false {
		log.Printf("New header file: %s", fn)
	}
	hf := &HeaderFile{
		name: fn,
	}

	c, err := ioutil.ReadFile(string(fn))
	if err != nil {
		log.Printf("Error reading header file: %s: %s", fn, err)
		return hf
	}

	hf.contents = estring(c)
	hf.v = true
	headers = append(headers, hf)
	hf.parse()
	return hf
}

/*! Returns a pointer to the HeaderFile whose unqualified file name is \a
  s, or a null pointer if there is no such HeaderFile.
*/

func findHeaderFile(s estring) *HeaderFile {
	hack := "/" + s
	for _, h := range headers {
		if h.Name() != s && !h.Name().endsWith(hack) {
			continue
		}

		return h
	}
	return nil
}

/*! Parses this header file and creates Class and Function objects as
  appropriate.

  The parsing is minimalistic: All it does is look for a useful
  subset of class declarations, and process those.
*/

func (this *HeaderFile) parse() {
	p := newParser(this.contents)
	p.scan("\nclass ")
	for !p.atEnd() {
		className := p.identifier()
		var superclass estring
		p.whitespace()
		if p.lookingAt(":") {
			p.step()
			inheritance := p.word()
			if inheritance != "public" {
				docError(this, p.line(), "Non-public inheritance for class "+className)
				return
			}
			parent := p.identifier()
			if parent.isEmpty() {
				docError(this, p.line(),
					"Cannot parse superclass name for class "+
						className)
				return
			}
			superclass = parent

			{
				// rudimentary handling of multiple inheritance (by ignoring it)
				p.whitespace()

				if p.lookingAt(",") {
					docError(this, p.line(), "Skipping multiple inheritance on class "+className)
					for !p.lookingAt("{") {
						p.step()
						p.whitespace()
					}
				}
			}
		}
		p.whitespace()
		if p.lookingAt("{") {
			c := findClass(className)
			if c == nil {
				c = newClass(className, nil, 0)
			}
			c.setParent(superclass)
			if c != nil && c.file() != nil {
				docError(this, p.line(),
					"Class "+className+
						" conflicts with "+className+" at "+
						c.file().Name()+":"+
						fn(c.line(), 10))
				docError(c.file(), c.line(),
					"Class "+className+
						" conflicts with "+className+" at "+
						this.Name()+":"+
						fn(p.line(), 10))
			} else {
				if false {
					log.Printf("Found class definition")
				}
				c.setSource(this, p.line())
			}
			p.step()
			ok := false
			for {
				ok = false
				p.whitespace()
				for p.lookingAt("public:") ||
					p.lookingAt("private:") ||
					p.lookingAt("protected:") {
					p.scan(":")
					p.step()
					p.whitespace()
				}
				if p.lookingAt("virtual ") {
					p.scan(" ")
				}
				p.whitespace()
				var t estring
				var n estring
				l := p.line()
				if p.lookingAt("operator ") {
					n = p.identifier()
				} else if p.lookingAt("enum ") {
					p.scan(" ")
					//e := newEnum(c, p.word(), this, l)
					p.whitespace()
					if p.lookingAt("{") {
						again := true
						for again {
							p.step()
							p.whitespace()
							v := p.word()
							if v.isEmpty() {
								docError(this, p.line(),
									"Could not parse enum value")
							} else {
								//e.addValue(v)
							}
							p.whitespace()
							if p.lookingAt("=") {
								p.step()
								p.whitespace()
								p.value()
								p.whitespace()
							}
							again = p.lookingAt(",")
						}
						if p.lookingAt("}") {
							p.step()
							ok = true
						} else {
							docError(this, p.line(),
								"Enum definition for "+
									className+"::"+n+
									" does not end with '}'")
						}
					} else if p.lookingAt(";") {
						// senseless crap
						ok = true
					} else {
						docError(this, l,
							"Cannot parse enum "+
								className+"::"+n)
					}
				} else if p.lookingAt("typedef ") {
					ok = true
				} else {
					t = p.parseType()
					n = p.identifier()
					if n.isEmpty() {
						// constructor/destructor?
						if t == className || t == "~"+className {
							n = t
							t = ""
						} else if t.isEmpty() && p.lookingAt("~") {
							p.step()
							n = "~" + p.identifier()
						}
					}
				}
				if !n.isEmpty() {
					p.whitespace()
					if p.lookingAt(";") {
						ok = true
					}
					a := p.argumentList()
					p.whitespace()
					fc := false
					if p.lookingAt("const") {
						fc = true
						p.word()
					}
					if !n.isEmpty() && n.find(":") < 0 &&
						!a.isEmpty() {
						n = className + "::" + n
						f := findFunction(n, a, fc)
						if f == nil {
							f = newFunction(t, n, a, fc, this, l)
						}
						ok = true
					}
				}
				if ok {
					p.whitespace()
					if p.lookingAt("{") {
						level := 0
						for level > 0 || p.lookingAt("{") {
							if p.lookingAt("{") {
								level++
							} else if p.lookingAt("}") {
								level--
							}
							p.step()
							p.whitespace()
						}
					} else {
						p.scan(";")
					}
				}

				if !ok {
					break
				}
			}
		}
		p.scan("\nclass ")
	}
}
