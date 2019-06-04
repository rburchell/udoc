package main

var output *outputT = &outputT{}

/*! \class Output output.h
  The Output class coordinates documentation output.

  It provides a number of static functions, each of which calls
  eponymous functions in each of the concrete output classes. The only
  output class currently is WebPage. PostScript and ManPage may be written
  when arnt is bored or they seem useful.
*/
type outputT struct {
	needSpace bool
	o         estring
	u         estring
}

/*! Starts a headline for \a i, with appropriate fonts etc. The
  headline runs until endParagraph() is called.
*/
func (this *outputT) startHeadlineIntro(i *Intro) {
	this.endParagraph()
	webpage.startHeadlineIntro(i)
}

/*! Starts a headline for \a c, with appropriate fonts etc. The
  headline runs until endParagraph() is called.
*/
func (this *outputT) startHeadlineClass(c *Class) {
	this.endParagraph()
	webpage.startHeadlineClass(c)
}

/*! Starts a headline for \a f, with appropriate fonts etc. The
  headline runs until endParagraph() is called.
*/
func (this *outputT) startHeadlineFunction(f *Function) {
	this.endParagraph()
	webpage.startHeadlineFunction(f)
}

/*! Ends the current paragraph on all output devices. */
func (this *outputT) endParagraph() {
	this.needSpace = false
	webpage.endParagraph()
}

/*! Adds \a text as ordinary text to all output devices. */
func (this *outputT) addText(text estring) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addText(text)
}

/*! Adds \a url and \a title as a link to all capable output devices. */
func (this *outputT) addLink(url, title estring) {

	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addLink(url, title)
}

/*! Adds \a text as an argument name to all output devices. */
func (this *outputT) addArgument(text estring) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addArgument(text)
}

/*! Adds a link to \a f titled \a text on all output devices. Each
  device may express the link differently.
*/
func (this *outputT) addFunction(text estring, f *Function) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addFunction(text, f)
}

/*! Adds a link to \a c titled \a text to all output devices. Each
  device may express the link differently.
*/
func (this *outputT) addClass(text estring, c *Class) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addClass(text, c)
}

/*! Adds a code snippet \a text to all output devices. Each
  device may express the snippet differently.
*/
func (this *outputT) addCodeBlock(text estring) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addCodeBlock(text)
}

/*! Adds an emphasized warning note \a text to all output devices. Each
  device may express the link differently.
*/
func (this *outputT) addWarning(text estring) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addWarning(text)
}

/*! Adds an emphasized note \a text to all output devices. Each
  device may express the link differently.
*/
func (this *outputT) addNote(text estring) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addNote(text)
}

/*! Adds a section header of emphasis level \a prio with a given \a text
* to all output devices. Each device may express the link differently.
 */
func (this *outputT) addSection(prio int, text estring) {
	if this.needSpace {
		this.needSpace = false
		this.addText(" ")
	}
	webpage.addSection(prio, text)
}

/*! Adds a single space to all output devices, prettily optimizing so
  there aren't lots of spaces where none are needed.
*/
func (this *outputT) addSpace() {
	this.needSpace = true
}
func (this *outputT) setOwner(o estring) {
	this.o = o
}
func (this *outputT) owner() estring {
	return this.o
}
func (this *outputT) setOwnerHome(u estring) {
	this.u = u
}
func (this *outputT) ownerHome() estring {
	return this.u
}
