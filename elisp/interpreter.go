package elisp

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	eightBitCodeOffset rune = 0x3fff00
	max5ByteChar            = 0x3fff7f
	eofRune                 = -1
	nbsp                    = '\u00A0'
)

type Interpreter interface {
	Print(obj lispObject) (string, error)
	ReadString(source string) (lispObject, error)
}

type interpreter struct {
	obarray     map[string]*lispSymbol
	nil_        lispObject
}

type readContext struct {
	source string
	runes  []rune
	i      int
}

func (ctx *readContext) read() rune {
	if ctx.i == len(ctx.runes) {
		return eofRune
	}

	r := ctx.runes[ctx.i]
	ctx.i++
	return r
}

func (ctx *readContext) unread(c rune) {
	if c == eofRune {
		return
	}

	if ctx.i == 0 {
		panic("nothing to unread")
	}

	ctx.i--
}

func (inp *interpreter) Print(obj lispObject) (string, error) {
	str, err := inp.prin1(obj)
	if err != nil {
		return "", err
	}

	return str.(*lispString).value, nil
}

func (inp *interpreter) prin1(obj lispObject) (lispObject, error) {
	type_ := obj.getType()
	var s string

	switch type_ {
	case symbol:
		s = obj.(*lispSymbol).name
	case integer:
		s = fmt.Sprint(obj.(*lispInteger).value)
	case string_:
		s = "\"" + obj.(*lispString).value + "\""
	case vector:
		return nil, fmt.Errorf("prin1 unimplemented")
	case cons:
		// TODO: Clean up when string functions are avaiable (don't type-assert)
		lc := obj.(*lispCons)
		current := lc

		carStr, err := inp.prin1(lc.car)
		if err != nil {
			return nil, err
		}
		s = "(" + carStr.(*lispString).value

		for {
			next, ok := current.cdr.(*lispCons)
			if ok {
				nextStr, err := inp.prin1(next.car)
				if err != nil {
					return nil, err
				}

				s += " " + nextStr.(*lispString).value
				current = next
			} else {
				if current.cdr != inp.nil_ {
					cdrStr, err := inp.prin1(current.cdr)
					if err != nil {
						return nil, err
					}
					s += " . " + cdrStr.(*lispString).value
				}

				break
			}
		}

		s += ")"
	case float:
		s = fmt.Sprint(obj.(*lispFloat).value)
	default:
		return nil, fmt.Errorf("prin1 unimplemented for '%v'", type_)
	}

	return inp.makeString(s), nil
}

func (inp *interpreter) setCdr(obj lispObject, newCdr lispObject) lispObject {
	obj.(*lispCons).cdr = newCdr
	return newCdr
}

func (inp *interpreter) stringToNumber(s string) (lispObject, error) {
	nInt, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return inp.makeInteger(lispInt(nInt)), nil
	}

	nFloat, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return inp.makeFloat(nFloat), nil
	}

	return nil, fmt.Errorf("unknown number format")
}

func (inp *interpreter) makeSymbol(name string) *lispSymbol {
	return &lispSymbol{
		name: name,
	}
}

func (inp *interpreter) makeInteger(value lispInt) *lispInteger {
	return &lispInteger{
		value: value,
	}
}

func (inp *interpreter) makeCons(car lispObject, cdr lispObject) *lispCons {
	return &lispCons{
		car: car,
		cdr: cdr,
	}
}

func (inp *interpreter) makeFloat(value float64) *lispFloat {
	return &lispFloat{
		value: value,
	}
}

func (inp *interpreter) makeString(value string) *lispString {
	return &lispString{
		value: value,
	}
}

func (inp *interpreter) makeList(obj lispObject, objs ...lispObject) *lispCons {
	tmp := inp.makeCons(obj, inp.nil_)
	val := tmp

	for _, obj := range objs {
		tmp.cdr = inp.makeCons(obj, inp.nil_)
		tmp = tmp.cdr.(*lispCons)
	}

	return val
}

func (inp *interpreter) readEscape(ctx *readContext, stringp bool) (rune, error) {
	c := ctx.read()
	if c == eofRune {
		return 0, fmt.Errorf("eof")
	}

	switch c {
	case 'a':
		return '\a', nil
	case 'b':
		return '\b', nil
	case 'd':
		return 0177, nil
	case 'e':
		return 033, nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'v':
		return '\v', nil
	case '\n':
		return -1, nil
	case ' ':
		if stringp {
			return -1, nil
		}
		return ' ', nil
	case 'M':
	case 'S':
	case 'H':
	case 'A':
	case 's':
	case 'C':
	case '^':
	case '0', '1', '2', '3', '4', '5', '6', '7':
		num := c - '0'

		for count := 0; count < 2; count++ {
			c = ctx.read()

			if c >= '0' && c <= '7' {
				num *= 8
				num += c - '0'
			} else {
				ctx.unread(c)
				break
			}
		}

		if num >= 0x80 && num < 0x100 {
			num = byte8toChar(num)
		}

		return num, nil
	case 'x':
	case 'U':
	case 'u':
	case 'N':
	default:
		return c, nil
	}

	return 0, fmt.Errorf("unimplemented escape code: '%v", string(c))
}

func (inp *interpreter) initialDefinitions() {
	symNil := inp.intern("nil")
	symNil.value = symNil

	symT := inp.intern("t")
	symT.value = symT

	// TODO:
	// quote
	// backquote (`)
}

func NewInterpreter() Interpreter {
	inp := interpreter{}
	inp.obarray = make(map[string]*lispSymbol)

	inp.initialDefinitions()

	inp.nil_ = inp.intern("nil")

	return &inp
}

func (inp *interpreter) intern(name string) *lispSymbol {
	sym, ok := inp.obarray[name]
	if !ok {
		sym = inp.makeSymbol(name)
		inp.obarray[name] = sym
	}

	return sym
}

func (inp *interpreter) readList(ctx *readContext) (lispObject, error) {
	var val lispObject = inp.nil_
	var tail lispObject = inp.nil_

	for {
		elt, c, err := inp.read1(ctx)
		if err != nil {
			return nil, err
		}

		if c != 0 {
			switch c {
			case ')':
				return val, nil
			case '.':
				if tail != inp.nil_ {
					obj, err := inp.read0(ctx)
					if err != nil {
						return nil, err
					}
					inp.setCdr(tail, obj)
				} else {
					val, err = inp.read0(ctx)
				}

				_, c, _ = inp.read1(ctx)
				if c == ')' {
					return val, nil
				}

				return nil, fmt.Errorf("'.' in wrong context")
			default:
				return nil, fmt.Errorf("invalid list ending: '%v'", string(c))
			}
		}

		tmp := inp.makeCons(elt, inp.nil_)

		if tail != inp.nil_ {
			inp.setCdr(tail, tmp)
		} else {
			val = tmp
		}

		tail = tmp
	}
}

func (inp *interpreter) readAtom(c rune, ctx *readContext) (lispObject, error) {
	quoted := false
	builder := strings.Builder{}

	for {
		if c == '\\' {
			c = ctx.read()
			if c == eofRune {
				return nil, fmt.Errorf("eof reading atom")
			}

			quoted = true
		}

		builder.WriteRune(c)
		c = ctx.read()

		special := strings.Contains("\"';()[]#`,", string(c))
		if c <= 040 || c == nbsp || special {
			break
		}
	}

	ctx.unread(c)
	s := builder.String()

	if !quoted {
		num, err := inp.stringToNumber(s)
		if err == nil {
			return num, nil
		}
	}

	return inp.intern(s), nil
}

func (inp *interpreter) read1Result(obj lispObject, err error) (lispObject, rune, error) {
	return obj, 0, err
}

func (inp *interpreter) read1(ctx *readContext) (lispObject, rune, error) {
	var err error
	var c rune

	for {
		c = ctx.read()
		if c == eofRune {
			return nil, 0, fmt.Errorf("eof on read1")
		}

		switch {
		case c == '(':
			return inp.read1Result(inp.readList(ctx))
		case c == '[':
			return nil, 0, fmt.Errorf("unimplented read: '%v'", string(c))
		case c == ')' || c == ']':
			return nil, c, nil
		case c == '#':
			return nil, 0, fmt.Errorf("unimplented read: '%v'", string(c))
		case c == ';':
			for {
				c = ctx.read()
				if c == eofRune || c == '\n' {
					break
				}
			}
		case c == '\'' || c == '`':
			obj, err := inp.read0(ctx)
			if err != nil {
				return nil, 0, err
			}

			name := "quote"
			if c == '`' {
				name = "`"
			}

			list := inp.makeList(inp.intern(name), obj)
			return list, 0, nil
		case c == ',':
			return nil, 0, fmt.Errorf("unimplented read: '%v'", string(c))
		case c == '?':
			c = ctx.read()
			if c == eofRune {
				return nil, 0, fmt.Errorf("eof reading character")
			}

			if c == ' ' || c == '\t' {
				return inp.makeInteger(lispInt(c)), 0, nil
			}

			if c == '\\' {
				c, err = inp.readEscape(ctx, false)
				if err != nil {
					return nil, 0, err
				}
			}

			if charByte8(c) {
				c = charToByte8(c)
			}

			next := ctx.read()
			ctx.unread(next)

			special := strings.Contains("\"';()[]#?`,.", string(next))
			if next <= 040 || special {
				return inp.makeInteger(lispInt(c)), 0, nil
			}

			return nil, 0, fmt.Errorf("invalid syntax for '?'")
		case c == '"':
			builder := strings.Builder{}

			for {
				c = ctx.read()
				if c == eofRune || c == '"' {
					break
				}

				if c == '\\' {
					c, err = inp.readEscape(ctx, true)
					if err != nil {
						return nil, 0, err
					}

					if c == -1 {
						continue
					}
				}

				builder.WriteRune(c)
			}

			if c == eofRune {
				return nil, c, fmt.Errorf("eof reading string")
			}

			return inp.makeString(builder.String()), 0, nil
		case c == '.':
			next := ctx.read()
			ctx.unread(next)

			special := strings.Contains("\"';([#?`,", string(next))
			if next <= 040 || special {
				return nil, c, nil
			}
			return inp.read1Result(inp.readAtom(c, ctx))
		case c <= 040 || c == nbsp:
		default:
			return inp.read1Result(inp.readAtom(c, ctx))
		}
	}
}

func (inp *interpreter) read0(ctx *readContext) (lispObject, error) {
	obj, c, err := inp.read1(ctx)
	if err != nil {
		return nil, err
	}

	if c != 0 {
		return nil, fmt.Errorf("unexpected character: '%v'", string(c))
	}

	return obj, nil
}

func (inp *interpreter) ReadString(source string) (lispObject, error) {
	ctx := readContext{
		source: source,
		runes:  []rune(source),
		i:      0,
	}

	return inp.read0(&ctx)
}

// func (inp *interpreter) eval(form lispObject) (lispObject, error) {
// 	if form.getType() == symbol {
// 		// return value in env
// 	} else if form.getType() != cons {
// 		return form, nil
// 	}

// }

func (inp *interpreter) apply(fn lispObject, args ...lispObject) (lispObject, error) {
	return nil, nil
}
