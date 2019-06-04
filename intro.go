package main

import (
	"log"
)

/*! \class Intro intro.h

  The Intro class introduces any number of classes (except
  zero). The introduction is output before the contained classes,
  and each class is output after its introduction.

  The Intro has a DocBlock, as usual. The DocBlock calls addClass()
  during parsing; output() uses this information to call
  Class::generateOutput() on the right classes afterwards. main()
  calls output() to build all the output.
*/

var intros []*Intro

type Intro struct {
	n        estring
	docBlock *DocBlock
	classes  []*Class // SortedList
}

/*!  Constructs an Intro object to go into file \a name. */
func newIntro(name estring) *Intro {
	i := &Intro{
		n: name,
	}
	log.Printf("New intro: %s", name)
	intros = append(intros, i)
	return i
}

/*! Notifies this Intro that it is documented by \a d. */
func (this *Intro) setDocBlock(d *DocBlock) {
	this.docBlock = d
}

/*! Add \a c to the list of classes being introduced by this object. */

func (this *Intro) addClass(c *Class) {
	this.classes = append(this.classes, c)
}

/*! This static function processes all Intro objects and generates the
  appropriate output, including output for the classes introduced.
*/

func outputIntro() {
	for _, i := range intros {
		i.docBlock.generate()

		for _, c := range i.classes {
			c.generateOutput()
		}
	}
}

/*! Returns the name supplied to the constructor. */

func (this Intro) name() estring {
	return this.n
}
