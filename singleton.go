package main

type singleton_dict map[estring]*Singleton

func (this singleton_dict) find(k estring) *Singleton {
	v, _ := this[k]
	return v
}
func (this singleton_dict) contains(k estring) bool {
	_, ok := this[k]
	return ok
}
func (this singleton_dict) insert(k estring, s *Singleton) {
	this[k] = s
}

var refs singleton_dict

/*! \class Singleton singleton.h

  The Singleton class defines a singleton, ie. a word or phrase
  which may only be mentioned once in the documentation. It is used
  to ensure that only one Intro introduces a given Class or other
  Intro.

  If a Singleton is created for a the same name as an already
  existing Singleton, error messages are omitted for both of them.
*/

type Singleton struct {
	f File
	l int
}

/*! Constructs a Singleton to \a name, which is located at \a file,
  \a line. */

func newSingleton(file File, line int, name estring) {
	s := &Singleton{
		f: file,
		l: line,
	}
	if refs == nil {
		refs = make(singleton_dict)
	}

	other := refs.find(name)
	if other != nil {
		docError(file, line,
			name+" also mentioned at "+
				other.file().Name()+" line "+
				fn(other.line(), 10))
		docError(other.file(), other.line(),
			name+" also mentioned at "+
				file.Name()+" line "+
				fn(line, 10))
	} else {
		refs.insert(name, s)
	}
}

func (this Singleton) file() File {
	return this.f
}

func (this Singleton) line() int {
	return this.l
}
