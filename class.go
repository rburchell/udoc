package main

import (
	"log"
)

type Class struct {
	n              estring
	f              File
	l              int
	super          *Class
	sub            []*Class // SortedList
	superclassName estring
	m              []*Function // SortedList
	db             *DocBlock
	done           bool
}

/*! \class Class class.h

  The Class class models a C++ class and its documentation.

  A Class has zero or one parent classes, any number of member
  functions and one documentation block.

  The file has an origin file and line.
*/

/*! Constructs a Class object for the class named \a s, which is
  defined on \a sourceLine of \a sourceFile. Initially the class is
  considered to have no member functions, superclasses or
  subclasses.
*/

var classes []*Class

func newClass(s estring, sourceFile File, sourceLine int) *Class {
	if false {
		log.Printf("New class: %s", s)
	}
	c := &Class{
		n: s,
		f: sourceFile,
		l: sourceLine,
	}
	classes = append(classes, c)
	return c
}

func (this *Class) setSource(file File, line int) {
	if false {
		log.Printf("Setting source for %s to %s:%d", this.name(), file.Name(), line)
	}
	this.f = file
	this.l = line
}

/*! \fn EString Class::name() const

  Returns the class name, as specified to the constructor.
*/

func (this Class) name() estring {
	return this.n
}

/*! Returns a pointer to the Class object whose name() is \a s, or a
  null pointer of there is no such object.
*/

func findClass(s estring) *Class {
	for _, c := range classes {
		if c.name() == s {
			return c
		}
	}
	return nil
}

/*! Notifies this Class that \a cn is its parent class. The initial
  value is an empty string, corresponding to a class that inherits
  nothing.

  Note that udoc does not support multiple or non-public inheritance.
*/

func (this *Class) setParent(cn estring) {
	this.superclassName = cn
}

/*! Returns the line number where this class was first seen. Should
  this be the line of the "\class", or of the header file
  class definition? Not sure. */

func (this Class) line() int {
	return this.l
}

/*! Returns the file name where this class was seen. Should this be
  the .cpp containing the "\class", or the header file
  containing class definition? Not sure. */

func (this Class) file() File {
	return this.f
}

/*! This static function processes all classes and generates the
  appropriate output.
*/

func outputClasses() {
	log.Printf("Generating for %d classes", len(classes))
	for _, c := range classes {
		if !c.done {
			c.generateOutput()
		}
	}
}

func (this *Class) setDocBlock(d *DocBlock) {
	this.db = d
}

func (this *Class) insert(memb *Function) {
	this.m = append(this.m, memb)
}

/*! Does everything necessary to generate output for this class and
  all of its member functions.
*/

func (this *Class) generateOutput() {
	if this.db == nil {
		if this.f == nil {
			// if we don't have a file for this class, see if we can
			// get one from a function.
			if len(this.m) > 0 {
				member := this.m[0]
				this.f = member.file()
				this.l = member.line()
			}
		}
		if this.f != nil {
			// if we now have a file, we can complain
			docError(this.file(), this.line(), "Undocumented class: "+this.n)
		}
		return
	} else if this.f != nil {
		this.db.generate()
	}

	for _, f := range this.m {
		if f.docBlock() != nil {
			f.docBlock().generate()
		} else if f.super() == nil {
			docError(f.file(), f.line(),
				"Undocumented function: "+
					f.name()+f.arguments())
		}
	}
	this.done = true
}

/*! Builds a hierarchy tree of documented classes, and emits errors if
  any the inheritance tree isn't fully documented.

  This function must be called before Function::super() can be.
*/

func buildHierarchy() {
	for _, c := range classes {
		n := c.superclassName
		i := n.find("<")
		if i >= 0 {
			n = n.mid(0, i)
		}
		if !n.isEmpty() {
			p := findClass(n)
			c.super = p
			if p != nil {
				p.sub = append(p.sub, c)
			}
			if c.super == nil {
				docError(c.f, c.l, "Class "+c.n+
					" inherits undocumented class "+
					c.superclassName)
			}
		}
	}
}

/*! Returns a pointer to a list of all classes that directly inherit
  this class. The returned list must neither be deleted nor
  changed. If no classes inherit this one, subclasses() returns a
  null pointer.*/

func (this Class) subclasses() []*Class {
	return this.sub
}

/*! Returns a pointer to the list of all the member functions in this
  class. The Class remains owner of the list; the caller should not
  delete or modify the list in any way.
*/

func (this Class) members() []*Function {
	return this.m
}

/*! \fn Class * Class::parent() const

  Returns a pointer to the superclass of this class, or a null
  pointer if this class doesn't inherit anything.
*/

func (this Class) parent() *Class {
	return this.super
}
