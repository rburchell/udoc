package main

type State int

const (
	Plain State = iota
	Argument
	Introduces
)

type string_dict map[estring]bool

func (this string_dict) contains(k estring) bool {
	_, ok := this[k]
	return ok
}
func (this string_dict) insert(k estring) {
	this[k] = true
}

/*! \class DocBlock docblock.h

  The DocBlock class represents a single atom of documentation.

  A documentation block is written as a C multi-line comment and
  documents a single class or a single function. DocBlock knows how
  to generate output for itself.
*/

/*!  Constructs a DocBlock from \a sourceFile, which starts at
  \a sourceLine, has source \a text and documents \a function.
*/

func newDocBlockForFunction(sourceFile File, sourceLine int, text estring, function *Function) *DocBlock {
	f := &DocBlock{
		file:      sourceFile,
		line:      sourceLine,
		f:         function,
		t:         text,
		s:         Plain,
		arguments: make(string_dict),
	}
	f.f.setDocBlock(f)
	return f
}

/*!  Constructs a DocBlock from \a sourceFile, which starts at
  \a sourceLine, has source \a text and documents \a className.
*/

func newDocBlockForClass(sourceFile File, sourceLine int, text estring, className *Class) *DocBlock {
	f := &DocBlock{
		file:      sourceFile,
		line:      sourceLine,
		c:         className,
		t:         text,
		s:         Plain,
		arguments: make(string_dict),
	}
	f.c.setDocBlock(f)
	return f
}

/*!  Constructs a DocBlock from \a sourceFile, which starts at
  \a sourceLine, has source \a text and documents \a intro.
*/

func newDocBlockForIntro(sourceFile File, sourceLine int, text estring, intro *Intro) *DocBlock {
	f := &DocBlock{
		file:      sourceFile,
		line:      sourceLine,
		i:         intro,
		t:         text,
		s:         Plain,
		arguments: make(string_dict),
	}
	f.i.setDocBlock(f)
	return f
}

type DocBlock struct {
	file       File
	line       int
	c          *Class
	f          *Function
	i          *Intro
	t          estring
	s          State
	arguments  string_dict
	isReimp    bool
	introduces bool
}

/*! Parses the text() and calls the Output functions on to generate
  suitable output.
*/

func (this *DocBlock) generate() {
	if this.t.contains("\\internal") {
		return
	}
	if this.f != nil {
		this.generateFunctionPreamble()
	} else if this.c != nil {
		this.generateClassPreamble()
	} else if this.i != nil {
		this.generateIntroPreamble()
	}

	n := 0
	l := this.line
	i := 0
	for i < this.t.length() {
		this.whitespace(&i, &l)
		n++
		this.word(&i, l, n)
	}
	output.endParagraph()
	if this.f != nil {
		super := this.f.super()
		if super != nil {
			output.addText("Reimplements ")
			output.addFunction(super.name()+"().", super)
			output.endParagraph()
		}
	}
	if this.f != nil && !this.isReimp {
		a := this.f.arguments()
		i := 0
		s := 0
		p := " "
		for i < a.length() {
			c := a[i]
			if c == ',' || c == ')' {
				if s > 0 {
					name := a.mid(s, i-s).simplified()
					if name.endsWith("[]") {
						name.truncate(name.length() - 2)
					}
					if name.find(" ") < 0 && !this.arguments.contains(name) {
						docError(this.file, this.line, "Undocumented argument: "+name)
					}
					s = 0
				}
			} else if (p == "(" || p == " " || p == "*" || p == "&") &&
				((c >= 'a' && c <= 'z') ||
					(c >= 'A' && c <= 'Z')) {
				s = i
			}
			p = string(c)
			i++
		}
	}
	if this.i != nil && !this.introduces {
		docError(this.file, this.line, "\\chapter must contain \\introduces")
	}
}

/*! Outputs boilerplante and genetated text to create a suitable
  headline and lead-in text for this DocBlock's function.
*/

func (this *DocBlock) generateFunctionPreamble() {
	output.startHeadlineFunction(this.f)
	addWithClass(this.f.typeStr(), this.f.parent())
	output.addText(" ")
	output.addText(this.f.name())
	a := this.f.arguments()
	if a == "()" {
		output.addText(this.f.arguments())
	} else {
		s := 0
		e := 0
		for e < a.length() {
			for e < a.length() && a[e] != ',' {
				e++
			}
			addWithClass(a.mid(s, e+1-s), this.f.parent())
			s = e + 1
			for a.at(s) == ' ' {
				output.addSpace()
				s++
			}
			e = s
		}
	}
	if this.f.isConst() {
		output.addText(" const")
	}
	output.endParagraph()
}

/*! Steps past whitespace, modifying the character index \a i and the
  line number \a l.
*/

func (this *DocBlock) whitespace(i, l *int) {
	first := (*i == 0)
	ol := *l
	any := false
	for *i < this.t.length() && (this.t[*i] == 32 || this.t[*i] == 9 ||
		this.t[*i] == 13 || this.t[*i] == 10) {
		if this.t[*i] == '\n' {
			*l++
		}
		*i++
		any = true
	}
	if *l > ol+1 {
		if this.s == Introduces {
			this.setState(Plain, "(end of paragraph)", *l)
		}
		this.checkEndState(ol)
		output.endParagraph()
	} else if any && !first && this.s != Introduces {
		output.addSpace()
	}
}

func addWithClass(s estring, in *Class) {
	var c *Class
	i := 0
	for c == nil && i < s.length() {
		if s[i] >= 'A' && s[i] <= 'Z' {
			j := i
			for (s.at(j) >= 'A' && s.at(j) <= 'Z') ||
				(s.at(j) >= 'a' && s.at(j) <= 'z') ||
				(s.at(j) >= '0' && s.at(j) <= '9') {
				j++
			}
			c = findClass(s.mid(i, j-i))
			i = j
		}
		i++
	}
	if c != nil && c != in {
		output.addClass(s, c)
	} else {
		output.addText(s)
	}
}

/*! Sets the DocBlock to state \a newState based on directive \a w,
  and gives an error from line \a l if the transition from the old
  state to the new is somehow wrong.
*/

func (this *DocBlock) setState(newState State, w estring, l int) {
	if this.s != Plain && newState != Plain {
		docError(this.file, l,
			"udoc directive "+w+
				" negates preceding directive")
	}
	if this.s == Introduces && this.i == nil {
		docError(this.file, l,
			"udoc directive "+w+
				" is only valid with \\chapter")
	}
	this.s = newState
}

/*! Verifies that all state is appropriate for ending a paragraph or
  documentation block, and emits appropriate errors if not. \a l must
  be the line number at which the paragraph/doc block ends.
*/

func (this DocBlock) checkEndState(l int) {
	if this.s != Plain {
		docError(this.file, l, "udoc directive hanging at end of paragraph")
	}
}

/*! Steps past and processes a word, which in this context is any
  nonwhitespace. \a i is the character index, which is moved, \a l
  is the line number, and \a n is the word number.
*/

func (this *DocBlock) word(i *int, l, n int) {
	j := *i
	for j < this.t.length() && !(this.t[j] == 32 || this.t[j] == 9 ||
		this.t[j] == 13 || this.t[j] == 10) {
		j++
	}
	w := this.t.mid(*i, j-*i)
	*i = j
	if w == "RFC" {
		for this.t[j] == ' ' || this.t[j] == '\n' {
			j++
		}
		start := j

		for this.t[j] <= '9' && this.t[j] >= '0' {
			j++
		}

		ok := false
		n := this.t.mid(start, j-start).number(&ok, 10)
		if ok {
			rfc := estring("http://www.rfc-editor.org/rfc/rfc")
			rfc += fromNumber(n, 10)
			rfc += ".txt"

			output.addLink(rfc, "RFC "+fromNumber(n, 10))

			*i = j
		} else {
			this.plainWord(w, l)
		}
	} else if w.lower().startsWith("http://") {
		output.addLink(w, w)
	} else if w.at(0) != '\\' {
		this.plainWord(w, l)
	} else if w == "\\a" {
		if this.f != nil {
			this.setState(Argument, w, l)
		} else {
			docError(this.file, l, "\\a is only defined function documentation")
		}
	} else if w == "\\introduces" {
		if this.i != nil {
			this.setState(Introduces, w, l)
		} else {
			docError(this.file, l, "\\introduces is only valid after \\chapter")
		}
		this.introduces = true
	} else if w == "\\overload" {
		this.overload(l, n)
	} else if w == "\\code" {
		this.code(i)
	} else if w == "\\warning" {
		this.warning(i)
	} else if w == "\\note" {
		this.note(i)
	} else if w == "\\section1" {
		this.section(i, 1)
	} else if w == "\\section2" {
		this.section(i, 2)
	} else if w == "\\section3" {
		this.section(i, 3)
	} else {
		docError(this.file, l, "udoc directive unknown: "+w)
	}
}

/*! Adds the plain word or link \a w to the documentation, reporting
  an error from line \a l if the link is dangling.
*/

func (this *DocBlock) plainWord(w estring, l int) {
	if this.s == Introduces {
		newSingleton(this.file, l, w)
		c := findClass(w)
		if c != nil {
			this.i.addClass(c)
		} else {
			docError(this.file, l, "Cannot find class: "+w)
		}
		return
	}
	// find the last character of the word proper
	last := w.length() - 1
	for last > 0 && (w[last] == ',' || w[last] == '.' ||
		w[last] == ':' || w[last] == ')') {
		last--
	}

	if this.s == Argument {
		name := w.mid(0, last+1)
		if name[0] == '*' {
			name = name.mid(1, len(name)-1) // yuck, what an evil hack
		}

		if this.arguments.contains(name) {
			// fine, nothing more to do
		} else if this.f.hasArgument(name) {
			this.arguments.insert(name)
		} else {
			docError(this.file, l, "No such argument: "+name)
		}
		output.addArgument(w)
		this.setState(Plain, "(after argument name)", l)
		return
	} else if w.at(last) == '(' {
		// is the word a plausible function name?
		i := 0
		for i < last && w[i] != '(' {
			i++
		}
		if i > 0 && ((w[0] >= 'a' && w[0] <= 'z') ||
			(w[0] >= 'A' && w[0] <= 'Z')) {
			name := w.mid(0, i)
			var link *Function
			scope := this.c
			if this.f != nil && scope == nil {
				scope = this.f.parent()
			}
			if name.contains(":") {
				link = findFunction(name, "", false)
			} else {
				parent := scope
				for parent != nil && link == nil {
					tmp := parent.name() + "::" + name
					link = findFunction(tmp, "", false)
					if link != nil {
						name = tmp
					} else {
						parent = parent.parent()
					}
				}
			}
			if scope != nil && link == nil && name != "main" {
				docError(this.file, l,
					"No link target for "+name+
						"() (in class "+scope.name()+")")
			} else if link != nil && link != this.f {
				output.addFunction(w, link)
				return
			}
		}
	} else if w.at(0) >= 'A' && w.at(0) <= 'Z' &&
		(this.c == nil || w.mid(0, last+1) != this.c.name()) {
		// is it a plausible class name? or enum, or enum value?
		link := findClass(w.mid(0, last+1))
		thisClass := this.c
		if this.f != nil && this.c == nil {
			thisClass = this.f.parent()
		}
		if link != nil && link != thisClass {
			output.addClass(w, link)
			return
		}
		// here, we could look to see if that looks _very_ much like a
		// class name, e.g. contains all alphanumerics and at least
		// one "::", and give an error about undocumented classes if
		// not.
	}

	// nothing doing. just add it as text.
	output.addText(w)
}

func (this *DocBlock) readUntilEndOfBlock(i *int) (estring, int) {
	p := newParser(this.t[*i:])

	var t estring
	for !p.atEnd() && !p.lookingAt("\n\n") {
		c := p.t.at(p.i)
		p.step()
		if c == '\t' {
			c = ' '
		}
		if c == '\n' {
			continue
		}
		t += estring(c)
	}

	return t.simplified(), p.i
}

/*! Handles the "\note" directive. \a i is the current cursor position.
 */
func (this *DocBlock) note(i *int) {
	text, advance := this.readUntilEndOfBlock(i)
	output.addNote(text)
	*i += advance
}

/*! Handles the "\warning" directive. \a i is the current cursor position.
 */
func (this *DocBlock) warning(i *int) {
	text, advance := this.readUntilEndOfBlock(i)
	output.addWarning(text)
	*i += advance
}

/*! Handles the "\sectionN" directive. \a i is the current cursor position.
* \a sect is the section number.
 */
func (this *DocBlock) section(i *int, sect int) {
	text, advance := this.readUntilEndOfBlock(i)
	output.addSection(sect, text)
	*i += advance
}

/*! Handles the "\code" directive. \a i is the current cursor position.
 */
func (this *DocBlock) code(i *int) {
	p := newParser(this.t[*i:])
	code := p.textUntil("\\endcode")
	output.addCodeBlock(code)
	*i += p.i
}

/*! Handles the "\overload" directive. \a l is the line number where
  directive was seen and \a n is the word number (0 for the first
  word in a documentation block).
*/

func (this *DocBlock) overload(l, n int) {
	if this.f == nil {
		docError(this.file, l,
			"\\overload is only meaningful for functions")
	} else if this.f.hasOverload() {
		docError(this.file, l,
			"\\overload repeated")
	} else {
		this.f.setOverload()
	}

	if n > 0 {
		docError(this.file, l, "\\overload must be the first directive")
	}
}

/*! Generates the routine text that introduces the documentation for
  each class, e.g. what the class inherits.
*/

func (this *DocBlock) generateClassPreamble() {
	output.startHeadlineClass(this.c)
	output.addText("Class ")
	output.addText(this.c.name())
	output.addText(".")
	output.endParagraph()
	p := false
	if this.c.parent() != nil {
		output.addText("Inherits ")
		output.addClass(this.c.parent().name(), this.c.parent())
		p = true
	}

	subclasses := this.c.subclasses()
	if len(subclasses) > 0 {
		if p {
			output.addText(". ")
		}
		output.addText("Inherited by ")
		p = true

		for _, sub := range subclasses {
			// ### FIXME
			//if ( !it ) {
			//    output.addClass( sub.name() + ".", sub );
			//}
			//else if ( it == subclasses.last() ) {
			//    output.addClass( sub.name(), sub );
			//    output.addText( " and " );
			//}
			//else {
			output.addClass(sub.name()+",", sub)
			output.addText(" ")
			//}
		}
	}
	if p {
		output.endParagraph()
	}

	members := this.c.members()
	if len(members) == 0 {
		docError(this.file, this.line,
			"Class "+this.c.name()+" has no member functions")
		return
	} else {
		// huh?
		//List<Function>::Iterator it( members );
		//while ( it )
		//    ++it;
	}
}

/*! Generates routine text to introduce an introduction. Yay! */

func (this *DocBlock) generateIntroPreamble() {
	output.startHeadlineIntro(this.i)
}
