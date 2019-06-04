package main

import (
	"log"
)

type Function struct {
	c    *Class
	t    estring
	n    estring
	a    estring
	args estring
	f    File
	l    int
	db   *DocBlock
	ol   bool
	cn   bool
}

func (this Function) parent() *Class {
	return this.c
}
func (this Function) isConst() bool {
	return this.cn
}
func (this Function) name() estring {
	return this.n
}
func (this Function) arguments() estring {
	return this.args
}
func (this *Function) setArgumentList(al estring) {
	this.args = al
}
func (this Function) file() File {
	return this.f
}
func (this Function) line() int {
	return this.l
}
func (this Function) docBlock() *DocBlock {
	return this.db
}
func (this *Function) setDocBlock(d *DocBlock) {
	this.db = d
}
func (this Function) super() *Function {
	return nil
}

/*! Returns a pointer to the Function object that is named (fully
  qualified) \a name, accepts \a arguments, and is/is not const as
  described by \a constness. If there is no such Function object,
  find() returns 0.

  If only \a name is supplied, any \a arguments and \a constness are
  accepted.
*/

func findFunction(name, arguments estring, constness bool) *Function {
	for _, f := range functions {
		if arguments.isEmpty() {
			if f.n != name {
				continue
			}
		} else {
			t := typesOnly(arguments)
			if f.n != name {
				continue
			}
			if f.a != t {
				continue
			}
			if f.cn != constness {
				continue
			}
		}
		return f
	}
	return nil
}

/*! \class Function function.h
  The Function class models a member function.

  Member functions are the only functions in udoc's world.

  Each function has a file() and line() number, which consequently are
  the ones in the class declaration, and it should have a docBlock().
  (The DocBlock's file and line number presumably are in a .cpp file.)
*/

var functions []*Function

/*!  Constructs a function whose return type is \a type, whose full
  name (including class) is \a name, whose arguments are \a
  arguments, and with \a constness. \a originFile and \a originLine
  point to the function's defining source, which will be used in any
  error messages.
*/

func newFunction(typeStr, name, arguments estring, constness bool, originFile File, originLine int) *Function {
	if false {
		log.Printf("New function: %s%s", name, arguments)
	}
	f := &Function{
		t:  typeStr,
		f:  originFile,
		l:  originLine,
		cn: constness,
	}

	functions = append(functions, f)

	i := name.length() - 1
	for i > 0 && name[i] != ':' {
		i--
	}
	if i == 0 || name[i-1] != ':' {
		// bad. error. how?
		return nil
	}
	f.n = name
	f.a = typesOnly(arguments)
	f.args = arguments
	f.c = findClass(name.mid(0, i-1))
	if f.c == nil {
		f.c = newClass(name.mid(0, i-1), nil, 0)
	}
	f.c.insert(f)
	return f
}

func (this Function) typeStr() estring {
	return this.t
}

/*! Returns a version of the argument list \a a which is stripped of
  argument names. For example, "( int a, const EString & b, int )" is
  transformed into "( int a, const EString &, int )".
*/

func typesOnly(a estring) estring {
	if a == "()" {
		return a
	}
	var r estring
	p := newParser(a)
	p.step() // past the '('
	var t estring
	s := estring("( ")
	for {
		t = p.parseType()
		if t.startsWith("class ") {
			t = t.mid(6, len(t))
		}
		if t.startsWith("struct ") {
			t = t.mid(7, len(t))
		}
		if !t.isEmpty() {
			r += s
			r += t
		}
		p.scan(",")
		s = ", "

		if t.isEmpty() {
			break
		}
	}
	r += " )"
	return r
}

/*! Returns true if \a s is the variable name of one of this
  function's arguments as specified in the .cpp file, and false if
  not.
*/

func (this Function) hasArgument(s estring) bool {
	if s.isEmpty() {
		return false
	}
	i := 0
	for i >= 0 && i < this.args.length() {
		i = this.args.findAt(s, i)
		if i >= 0 {
			i += s.length()
			for this.args.at(i) == ' ' {
				i++
			}
			if this.args.at(i) == '[' && this.args.at(i+1) == ']' {
				i += 2
			}
			for this.args.at(i) == ' ' {
				i++
			}
			if this.args.at(i) == ')' || this.args.at(i) == ',' {
				return true
			}
		}
	}
	return false
}

func (this Function) hasOverload() bool {
	return this.ol
}

/*! Notifies this function that it has an "\overload" directive. */

func (this *Function) setOverload() {
	this.ol = true
}
