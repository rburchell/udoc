package main

type Parser struct {
	t  estring
	i  int
	ln int
	li int
}

/*! \class Parser parser.h

  The Parser class does basic C++ parsing.

  It doesn't actually parse C++: all it does is lend some support to
  the header and source handling, which needs to find certain
  constructs and look at them.
*/

/*! Constructs a Parser for string \a s. The parser's cursor is left
  at the beginning of \a s. */
func newParser(contents estring) *Parser {
	return &Parser{
		t:  contents,
		ln: 1,
	}
}

/*! Returns true if the parser has reached the end of its input, and
  false if not.
*/

func (this Parser) atEnd() bool {
	return this.i >= this.t.length()
}

/*! Returns the parser's current line number.

The line number is that of the first unparsed nonwhitespace
character. This implies that if the parser's cursor is at the end of a
line, then the line number returned is that of the next nonempty line.
*/

func (this *Parser) line() int {
	if this.li > this.i {
		this.ln = 1
		this.li = 0
	}

	for this.li < this.i || this.t.at(this.li) == 32 || this.t.at(this.li) == 9 || this.t.at(this.li) == 13 || this.t.at(this.li) == 10 {
		if this.t.at(this.li) == 10 {
			this.ln++
		}
		this.li++
	}
	return this.ln
}

/*! Scans forward until an instance of \a text is found, and positions
  the cursor at the first character after that string. */

func (this *Parser) scan(text estring) {
	j := 0
	for this.i < this.t.length() && j < text.length() {
		j = 0
		for j < text.length() && this.t.at(this.i+j) == text[j] {
			j++
		}
		if j < text.length() {
			this.i++
		}
	}
	if j == text.length() {
		this.i += j
	}
}

/*! Scans for \a text and returns all the text, without the trailing
  instance of \a text. The cursor is left after \a text. */

func (this *Parser) textUntil(text estring) estring {
	j := this.i
	this.scan(text)
	if this.atEnd() {
		return this.t.mid(j, this.i-j)
	}
	return this.t.mid(j, this.i-j-text.length())
}

/*! Scans past whitespace, leaving the cursor at the end or at a
  nonwhitespace character.
*/

func (this *Parser) whitespace() {
	this.i = this.whitespaceAt(this.i)
}

func spaceless(t estring) estring {
	i := 0
	var r estring
	for i < t.length() {
		if t[i] != 32 && t[i] != 9 && t[i] != 13 && t[i] != 10 {
			r += estring(t[i])
		}
		i++
	}
	return r
}

/*! Returns the C++ identifier at the cursor, or an empty string if
  there isn't any. Steps past the identifier and any trailing whitespace.
*/

func (this *Parser) identifier() estring {
	j := this.complexIdentifier(this.i)
	r := spaceless(this.t.mid(this.i, j-this.i))
	this.i = j
	return r
}

/*! Scans past the simpler identifier starting at \a j, returning the
  first position afte the identifier. If something goes wrong,
  simpleIdentifier() returns \a j.

  A simple identifier is a text label not containing ::, <, >,
  whitespace or the like.
*/

func (this *Parser) simpleIdentifier(j int) int {
	k := this.whitespaceAt(j)
	if this.t.mid(k, 8) == "operator" {
		return this.operatorHack(k)
	}
	if (this.t.at(k) >= 'A' && this.t.at(k) <= 'z') ||
		(this.t.at(k) >= 'a' && this.t.at(k) <= 'z') {
		j = k + 1
		for (this.t.at(j) >= 'A' && this.t.at(j) <= 'z') ||
			(this.t.at(j) >= 'a' && this.t.at(j) <= 'z') ||
			(this.t.at(j) >= '0' && this.t.at(j) <= '9') ||
			(this.t.at(j) == '_') {
			j++
		}
	}
	return j
}

/*! Scans past the complex identifier starting at \a j, returning the
  first position after the identifier. If something goes wrong,
  complexIdentifier() returns \a j.

  A complex identifier is anything that may be used as an identifier
  in C++, even "operator const char *".
*/

func (this *Parser) complexIdentifier(j int) int {
	k := this.whitespaceAt(j)
	if this.t.at(k) == ':' && this.t.at(k+1) == ':' {
		k = this.whitespaceAt(k + 2)
	}
	l := this.simpleIdentifier(k)
	if l == k {
		return j
	}
	j = this.whitespaceAt(l)

	for this.t.at(j) == ':' && this.t.at(j+1) == ':' {
		if this.t.mid(j+2, 8) == "operator" {
			j = this.operatorHack(j + 2)
		} else if this.t.at(j+2) == '~' {
			j = this.simpleIdentifier(j + 3)
		} else {
			j = this.simpleIdentifier(j + 2)
		}
	}

	j = this.whitespaceAt(j)
	if this.t.at(j) == '<' {
		k = this.complexIdentifier(j + 1)
		if k > j+1 && this.t.at(k) == '>' {
			j = k + 1
		}
	}
	return j
}

/*! Parses a type name starting at \a j and returns the first
  character after the type name (and after trailing whitespace). If
  a type name can't be parsed, \a j is returned.
*/

func (this *Parser) parseTypeAt(j int) int {
	// first, we have zero or more of const, static etc.
	l := j
	k := 0
	for {
		k = l
		l = this.whitespaceAt(k)
		for this.t.at(l) >= 'a' && this.t.at(l) <= 'z' {
			l++
		}
		modifier := this.t.mid(k, l-k).simplified()
		if !(modifier == "const" ||
			modifier == "inline" ||
			modifier == "unsigned" ||
			modifier == "signed" ||
			modifier == "class" ||
			modifier == "struct" ||
			modifier == "virtual" ||
			modifier == "static") {
			l = k
		}
		if l <= k {
			break
		}
	}

	l = this.complexIdentifier(k)
	if l == k {
		return j
	}

	k = this.whitespaceAt(l)
	if this.t.at(k) == ':' && this.t.at(k+1) == ':' {
		l = this.whitespaceAt(this.simpleIdentifier(k + 2))
		if l == k {
			return j
		}
		k = l
	}

	if this.t.at(k) == '&' || this.t.at(k) == '*' {
		k = this.whitespaceAt(k + 1)
	}
	return k
}

/*! Parses a type specifier and returns it as a string. If the cursor
  doesn't point to one, type() returns an empty string.

*/

func (this *Parser) parseType() estring {
	j := this.parseTypeAt(this.i)
	r := this.t.mid(this.i, j-this.i).simplified() // simplified() is not quite right
	this.i = j
	for r.startsWith("class ") {
		r = r.mid(6, len(r))
	}
	r.replace(" class ", " ")

	tlen := len(r)

	// ### it might be nice to treat types as a real type rather than a string,
	// and expose reference/pointer as a field. then, the formatters can decide
	// to do whatever they want with this.
	if tlen >= 2 {
		lastc := r[tlen-1]
		switch lastc {
		case '*':
			fallthrough
		case '&':
			s := make([]rune, 0, len(r))
			for idx, c := range r {
				if idx == tlen-2 {
					if c == ' ' {
						continue
					}
				}
				s = append(s, c)
			}
			r = estring(s)
		}
	}
	return r
}

/*! Parses an argument list (for a particularly misleading meaning of
  parse) and returns it. The cursor must be on the leading '(', it
  will be left immediately after the trailing ')'.

  The argument list is returned including parentheses. In case of an
  error, an empty string is returned and the cursor is left near the
  error.
*/

func (this *Parser) argumentList() estring {
	var r estring
	j := this.whitespaceAt(this.i)
	if this.t.at(j) != '(' {
		return r
	}
	r = "("
	this.i = this.whitespaceAt(j + 1)
	if this.t.at(this.i) == ')' {
		this.i++
		return "()"
	}
	var s estring
	more := true
	for more {
		tp := this.parseType()
		if tp.isEmpty() {
			return "" // error message here?
		}
		this.whitespace()
		j = this.simpleIdentifier(this.i)
		if j > this.i { // there is a variable name
			tp = tp + " " + this.t.mid(this.i, j-this.i).simplified()
			this.i = j
		}
		r = r + s + tp
		this.whitespace()
		if this.t.at(this.i) == '=' { // there is a default value...
			for this.i < this.t.length() && this.t.at(this.i) != ',' && this.t.at(this.i) != ')' {
				this.i++
			}
			this.whitespace()
		} else if this.t.at(this.i) == '[' && this.t.at(this.i+1) == ']' { // this argument is an array
			this.i = this.i + 2
			r += "[]"
			this.whitespace()
		}
		s = ", "
		if this.t.at(this.i) == ',' {
			more = true
			this.i++
		} else {
			more = false
		}
	}
	if this.t.at(this.i) != ')' {
		return ""
	}
	r += ")"
	this.i++
	return r
}

/*! Steps the Parser past one character. */

func (this *Parser) step() {
	this.i++
}

/*! Returns true if the first unparsed characters of the string are
  the same as \a pattern, and false if not. */

func (this *Parser) lookingAt(pattern estring) bool {
	return this.t.mid(this.i, pattern.length()) == pattern
}

/*! Parses and steps past a single word. If the next nonwhitespace
  character is not a word character, this function returns an empty
  string.
*/

func (this *Parser) word() estring {
	j := this.simpleIdentifier(this.i)
	for this.t.at(j) == '-' {
		k := this.simpleIdentifier(j + 1)
		if k > j+1 {
			j = k
		}
	}
	r := this.t.mid(this.i, j-this.i).simplified()
	if !r.isEmpty() {
		this.i = j
	}
	return r
}

/*! Parses and steps past a single value, which is either a number or
  an identifier.
*/

func (this *Parser) value() estring {
	j := this.whitespaceAt(this.i)
	if this.t.at(j) == '-' ||
		(this.t.at(j) >= '0' && this.t.at(j) <= '9') {
		k := j
		if this.t.at(k) == '-' {
			k++
		}
		for this.t.at(k) >= '0' && this.t.at(k) <= '9' {
			k++
		}
		r := (this.t.mid(j, k-j))
		this.i = k
		return r
	}
	return this.identifier()
}

/*! Steps past the whitespace starting at \a j and return the index of
  the first following nonwhitespace character.
*/

func (this *Parser) whitespaceAt(j int) int {
	k := 0
	for {
		k = j

		for this.t.at(j) == 32 || this.t.at(j) == 9 || this.t.at(j) == 13 || this.t.at(j) == 10 {
			j++
		}

		if this.t.at(j) == '/' && this.t.at(j+1) == '/' {
			for j < this.t.length() && this.t.at(j) != '\n' {
				j++
			}
		}

		if j <= k {
			break
		}
	}

	return j
}

/*! Reads past an operator name starting at \a j and returns the index
  of the following characters. If \a j does not point to an operator
  name, operatorHack() returns \a j.
*/

func (this *Parser) operatorHack(j int) int {
	k := j + 8
	k = this.whitespaceAt(k)

	// Four possible cases: We're looking at a single character, two
	// characters, '()', or "EString".

	chars := 0

	if this.t.at(k) == '(' && this.t.at(k+1) == ')' {
		chars = 2
	} else if ((this.t.at(k) > ' ' && this.t.at(k) < '@') ||
		(this.t.at(k) > 'Z' && this.t.at(k) < 'a')) &&
		!(this.t.at(k) >= '0' && this.t.at(k) <= '9') {
		chars = 1
		if this.t.at(k+1) != '(' &&
			((this.t.at(k+1) > ' ' && this.t.at(k+1) < '@') ||
				(this.t.at(k) > 'Z' && this.t.at(k) < 'a')) &&
			!(this.t.at(k+1) >= '0' && this.t.at(k+1) <= '9') {

			chars = 2
		}
	} else {
		i := this.parseTypeAt(k)
		if i > k {
			chars = i - k
		}
	}

	if chars > 0 {
		k = this.whitespaceAt(k + chars)
		if this.t[k] == '(' {
			return k
		}
	}
	return j
}
