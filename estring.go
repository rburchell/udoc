package main

import (
	"strconv"
	"strings"
)

type estring string

func fn(n int, b int) estring {
	return estring(strconv.FormatInt(int64(n), b))
}

func (this *estring) truncate(l int) {
	if l == 0 {
		*this = ""
	} else if l < len(*this) {
		*this = (*this)[0:l]
	}
}

/*! Returns a copy of this string where all upper-case letters (A-Z -
  this is ASCII only) have been changed to lower case. */

func (this estring) lower() estring {
	var result estring
	i := 0
	for i < this.length() {
		if this[i] >= 'A' && this[i] <= 'Z' {
			result += estring(this[i] + 32)
		} else {
			result += estring(this[i])
		}
		i++
	}
	return result
}

func (this estring) isEmpty() bool {
	return this == ""
}
func (this estring) mid(start, num int) estring {
	if len(this) == 0 {
		num = 0
	} else if num > len(this) || start+num > len(this) {
		num = len(this) - start
	}

	var result estring
	if num == 0 || start >= this.length() {
		return result
	}

	result = this[start : start+num]
	return result
}
func (this estring) at(i int) byte {
	if i < 0 || i >= len(this) {
		return 0
	}
	return this[i]
}
func (this estring) length() int {
	return len(this)
}
func (this *estring) append(o estring) {
	*this += o
}
func (this estring) contains(o estring) bool {
	return this.find(o) != -1
}
func (this estring) find(o estring) int {
	return strings.Index(string(this), string(o))
}
func (this estring) findAt(o estring, pos int) int {
	ret := strings.Index(string(this[pos:]), string(o))
	if ret >= 0 {
		return pos + ret
	}
	return ret
}
func (this estring) startsWith(o estring) bool {
	return strings.HasPrefix(string(this), string(o))
}
func (this estring) endsWith(o estring) bool {
	return strings.HasSuffix(string(this), string(o))
}
func (this estring) simplified() estring {
	// scan for the first nonwhitespace character
	i := 0
	first := 0
	for i < this.length() && first == i {
		c := this[i]
		if c == 9 || c == 10 || c == 13 || c == 32 {
			first++
		}
		i++
	}
	// scan on to find the last nonwhitespace character and detect any
	// sequences of two or more whitespace characters within the
	// string.
	last := first
	spaces := 0
	identity := true
	for identity && i < this.length() {
		c := this[i]
		if c == 9 || c == 10 || c == 13 || c == 32 {
			spaces++
		} else {
			if spaces > 1 {
				identity = false
			}
			spaces = 0
			last = i
		}
		i++
	}
	if identity {
		return this.mid(first, last+1-first)
	}

	var result estring
	i = 0
	spaces = 0
	for i < this.length() {
		c := this[i]
		if c == 9 || c == 10 || c == 13 || c == 32 {
			spaces++
		} else {
			if spaces > 0 && !result.isEmpty() {
				result += " "
			}
			spaces = 0
			result += estring(c)
		}
		i++
	}
	return result
}
func (this *estring) replace(find, replace estring) {
	*this = estring(strings.ReplaceAll(string(*this), string(find), string(replace)))
}

/*! Returns a string representing the number \a n in the \a base
  system, which is 10 (decimal) by default and must be in the range
  2-36.

  For 0, "0" is returned.

  For bases 11-36, lower-case letters are used for the digits beyond
  9.
*/

func fromNumber(n int, base int) estring {
	var r estring
	if 0 == n {
		r += "0" // Short-circuit, 0 is 0 in any base, no need for extra processing
	} else {
		if n < 0 {
			n = -n   // Negate, otherwise negative numbers get messed up
			r += "-" // But we remember we need a - sign
		}
		r.appendNumber(n, base)
	}
	return r
}

/*! Converts \a n to a number in the \a base system and appends the
  result to this string. If \a n is 0, "0" is appended.

  Uses lower-case for digits above 9.
*/

func (this *estring) appendNumber(n int, base int) {
	top := 1
	for top*base <= n {
		top = base * top
	}
	for top > 0 {
		d := (n / top) % base
		c := '0' + d
		if d > 9 {
			c = 'a' + d - 10
		}

		*this += estring(c)
		top = top / base
	}
}

/*! Returns the number encoded by this string, and sets \a *ok to true
  if that number is valid, or to false if the number is invalid. By
  default the number is encoded in base 10, if \a base is specified
  that base is used. \a base must be at least 2 and at most 36.

  If the number is invalid (e.g. negative), number() returns 0.

  If \a ok is a null pointer, it is not modified.
*/

func (this estring) number(ok *bool, base int) int {
	i := 0
	n := 0

	good := !this.isEmpty()
	for good && i < this.length() {
		if this[i] < '0' || this[i] > 'z' {
			good = false
		}

		digit := int(this[i] - '0')

		// hex or something?
		if digit > 9 {
			c := int(this[i])
			if c > 'Z' {
				c = c - 32
			}
			digit = c - 'A' + 10
		}

		// is the digit too large?
		if digit >= base {
			good = false
		}

		// Would n overflow if we multiplied by 10 and added digit?
		// FIXME
		//if n > UINT_MAX/base {
		//	good = false
		//}
		n *= base
		//if n >= (UINT_MAX-UINT_MAX%base) && digit > (UINT_MAX%base) {
		//	good = false
		//}
		n += digit

		i++
	}

	if !good {
		n = 0
	}

	if ok != nil {
		*ok = good
	}

	return n
}
