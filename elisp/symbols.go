package elisp

import (
	"fmt"
	"reflect"
)

type symbolInit struct {
	loc      *lispObject
	name     string
	unintern bool
}

type globals struct {
	nil_                   lispObject
	t                      lispObject
	internalInterpreterEnv lispObject
	unbound                lispObject
	error_                 lispObject
	quit                   lispObject
	userError              lispObject
	wrongLengthArgument    lispObject
	wrongTypeArgument      lispObject
	argsOutOfRange         lispObject
	voidFunction           lispObject
	voidVariable           lispObject
	wrongNumberofArguments lispObject
	endOfFile              lispObject
	errorConditions        lispObject
	errorMessage           lispObject
}

func (ec *execContext) initialDefsSymbols() {
	g := &ec.g

	syms := []symbolInit{
		{loc: &g.unbound, name: "unbound", unintern: true},
		{loc: &g.nil_, name: "nil"},
		{loc: &g.t, name: "t"},
		{
			loc:      &g.internalInterpreterEnv,
			name:     "internal-interpreter-environment",
			unintern: true,
		},
		{loc: &g.error_, name: "error"},
		{loc: &g.quit, name: "quit"},
		{loc: &g.userError, name: "user-error"},
		{loc: &g.wrongLengthArgument, name: "wrong-length-argument"},
		{loc: &g.wrongTypeArgument, name: "wrong-type-argument"},
		{loc: &g.argsOutOfRange, name: "args-out-of-range"},
		{loc: &g.voidFunction, name: "void-function"},
		{loc: &g.voidVariable, name: "void-variable"},
		{loc: &g.wrongNumberofArguments, name: "wrong-number-of-arguments"},
		{loc: &g.endOfFile, name: "end-of-file"},
		{loc: &g.errorConditions, name: "error-conditions"},
		{loc: &g.errorMessage, name: "error-message"},
	}

	ec.initializeSymbols(syms)

	xSymbol(ec.g.nil_).value = ec.g.nil_
	xSymbol(ec.g.t).value = ec.g.t

	ec.nil_ = ec.g.nil_
	ec.t = ec.g.t
}

func (ec *execContext) initializeSymbols(syms []symbolInit) {
	count := ec.countGlobals()
	if len(syms) != count {
		panic(fmt.Sprintf("%v globals exist but got %v initializers", count, len(syms)))
	}

	names := make(map[string]bool)

	for _, sym := range syms {
		if sym.name == "" {
			panic("no symbol name defined")
		}

		_, ok := names[sym.name]
		if ok {
			panic(fmt.Sprintf("repeated symbol name: %v", sym.name))
		}
		names[sym.name] = true

		*sym.loc = ec.makeSymbolBase(sym.name)
		symbol := xSymbol(*sym.loc)

		if ec.g.unbound == nil {
			panic("unbound not initialized yet")
		}

		symbol.value = ec.g.unbound

		if !sym.unintern {
			ec.internSymbol(symbol)
		}
	}

	for _, sym := range syms {
		symbol := xSymbol(*sym.loc)
		symbol.function = ec.g.nil_
		symbol.plist = ec.g.nil_
	}
}

func (ec *execContext) countGlobals() int {
	return reflect.ValueOf(ec.g).NumField()
}