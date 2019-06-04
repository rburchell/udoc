package main

import (
	"fmt"
	"os"
)

/*! \class WebPage webpage.h
  The WebPage class provides documentation output to a web page.

  It implements the same functions as Output, but they're not static,
  and is called when Output's static functions are called.
*/

/*! Constructs a web page generator that'll write to files in
  directory \a dir. */

type estringlist []estring

func (this estringlist) contains(k estring) bool {
	for _, v := range this {
		if v == k {
			return true
		}
	}

	return false
}

func (this *estringlist) clear() {
	*this = nil
}

type webpageT struct {
	fd        *os.File
	directory estring
	pstart    bool
	para      estring
	names     estringlist
	fn        estring
}

var webpage *webpageT

func newWebpage(dir estring) {
	webpage = &webpageT{
		directory: dir,
		pstart:    false,
	}
}

/*! As Output::startHeadline(). \a i is used to derive a file name. */

func (this *webpageT) startHeadlineIntro(i *Intro) {
	this.endPage()
	this.startPage(i.name().lower(), i.name())
}

/*! As Output::startHeadline(). \a c is used to derive a file name. */

func (this *webpageT) startHeadlineClass(c *Class) {
	this.endPage()
	this.startPage(c.name().lower(), c.name()+" documentation")
	this.output("<h1 class=\"classh\">")
	this.para = "</h1>\n"
	this.pstart = true
}

/*! As Output::startHeadline(). \a f is used to create an anchor. */

func (this *webpageT) startHeadlineFunction(f *Function) {
	a := this.anchor(f)
	o := estring("<h2 class=\"functionh\">")
	if !this.names.contains(a) {
		o += "<a name=\"" + this.anchor(f) + "\"></a>"
		this.names = append(this.names, a)
	}
	this.output(o)
	this.para = "</h2>\n"
	this.pstart = true
}

/*! As Output::endParagraph(). */

func (this *webpageT) endParagraph() {
	if this.para.isEmpty() {
		return
	}
	this.output(this.para)
	this.para = ""
}

/*! As Output::addText(). \a text is used escaped (&amp; etc). */

func (this *webpageT) addText(text estring) {
	if this.para.isEmpty() {
		this.output("<p class=\"text\">")
		this.para = "\n"
		this.pstart = true
	}

	i := 0
	if this.pstart {
		for text.at(i) == ' ' {
			i++
		}
		if i >= text.length() {
			return
		}
		this.pstart = false
	}

	var s estring
	for i < text.length() {
		if text[i] == '<' {
			s.append("&lt;")
		} else if text[i] == '>' {
			s.append("&gt;")
		} else if text[i] == '&' {
			s.append("&amp;")
		} else {
			s.append(estring(text[i]))
		}
		i++
	}
	this.output(s)
}

/*! Adds a link to \a url with the given \a title. */

func (this *webpageT) addLink(url, title estring) {
	this.addText("")
	var s estring = ("<a href=\"")
	s.append(url)
	s.append("\">")
	s.append(title)
	s.append("</a>")
	this.output(s)
}

/*! As Output::addArgument(). \a text is output in italics. */

func (this *webpageT) addArgument(text estring) {
	this.addText("")
	this.output("<i>")
	this.addText(text)
	this.output("</i>")
}

/*! As Output::addFunction(). If part of \a text corresponds to the
  name of \a f, then only that part is made into a link, otherwise
  all of \a text is made into a link.
*/

func (this *webpageT) addFunction(text estring, f *Function) {
	name := f.name()
	ll := text.length()
	ls := text.find(name)
	// if we don't find the complete function name, try just the member part
	if ls < 0 {
		i := name.length()
		for i > 0 && name[i] != ':' {
			i--
		}
		if i > 0 {
			name = name.mid(i+1, len(name)-i+1)
			ls = text.find(name)
		}
	}
	if ls >= 0 {
		ll = name.length()
	} else {
		ls = 0
	}
	if ll < text.length() && text.mid(ls+ll, 2) == "()" {
		ll = ll + 2
	}
	this.addText("")
	space := false
	i := 0
	for i < text.length() && !space {
		if text[i] == ' ' {
			space = true
		}
		i++
	}
	if space {
		this.output("<span class=nobr>")
	}
	this.addText(text.mid(0, ls))
	this.output("<a href=\"")
	target := f.parent().name().lower()
	if this.fn != target {
		this.output(target)
	}
	this.output("#" + this.anchor(f) + "\">")
	this.addText(text.mid(ls, ll))
	this.output("</a>")
	this.addText(text.mid(ls+ll, len(text)-ls+ll))
	if space {
		this.output("</span>")
	}
}

/*! As Output::addClass(). If part of \a text corresponds to the
  name of \a c, then only that part is made into a link, otherwise
  all of \a text is made into a link.
*/

func (this *webpageT) addClass(text estring, c *Class) {
	ll := text.length()
	ls := text.find(c.name())
	if ls >= 0 {
		ll = c.name().length()
	} else {
		ls = 0
	}
	this.addText("")
	space := false
	i := 0
	for i < text.length() && !space {
		if text[i] == ' ' {
			space = true
		}
		i++
	}
	if space {
		this.output("<span class=nobr>")
	}
	this.addText(text.mid(0, ls))
	link := true
	target := c.name().lower()
	if target == this.fn {
		link = false
	}
	if link {
		this.output("<a href=\"" + target + "\">")
	}
	this.addText(text.mid(ls, ll))
	if link {
		this.output("</a>")
	}
	this.addText(text.mid(ls+ll, len(text)-ls+ll))
	if space {
		this.output("</span>")
	}
}

func (this *webpageT) addCodeBlock(text estring) {
	this.output("<pre>")
	this.output(text)
	this.output("</pre>")
}

func (this *webpageT) addWarning(text estring) {
	this.output("<p><b>Warning:</b></p>")
	this.addText(text)
}

func (this *webpageT) addNote(text estring) {
	this.output("<p><b>Note:</b></p>")
	this.addText(text)
}

/*! Write \a s to the output file. */

func (this *webpageT) output(s estring) {
	if this.fd == nil || s.isEmpty() {
		return
	}

	this.fd.Write([]byte(s))
}

/*! This private helper returns the anchor (sans '#') corresponding to
  \a f.
*/

func (this *webpageT) anchor(f *Function) estring {
	fn := f.name()
	i := fn.length()
	for i > 0 && fn.at(i) != ':' {
		i--
	}
	if i > 0 {
		fn = fn.mid(i+1, len(fn)-i+1)
	}
	if fn.startsWith("~") {
		fn = "destructor"
	}
	return fn
}

/*! Emits any boilerplate to be emitted at the end of each page. */

func (this *webpageT) endPage() {
	if this.fd == nil {
		return
	}

	this.endParagraph()

	this.para = "\n"
	this.output("<p class=\"rights\">This web page based on source code belonging to ")
	if !output.ownerHome().isEmpty() {
		this.output("<a href=\"" + output.ownerHome() + "\">")
		this.addText(output.owner())
		this.output("</a>. All rights reserved.")
	} else {
		this.addText(output.owner())
		this.output(". All rights reserved.")
	}
	this.output("</body></html>\n")
	this.fd.Close()
}

/*! Starts a new web page with base name \a name and title tag \a
  title. The \a title must not be empty per the HTML standard.
*/

func (this *webpageT) startPage(name, title estring) {
	this.names.clear()
	filename := this.directory + "/" + name
	var err error
	this.fd, err = os.Create(string(filename))
	if err != nil {
		panic(fmt.Sprintf("Can't write %s: %s", filename, err))
	}
	this.output("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.0//EN\">\n<html lang=en><head>")
	this.output("<title>")
	this.para = "\n"
	this.pstart = true
	this.addText(title)
	this.output("</title>\n")
	this.output("<link rel=stylesheet href=\"udoc.css\" type=\"text/css\">\n<link rel=generator href=\"http://archiveopteryx.org/udoc/\">\n</head><body>\n")
}
