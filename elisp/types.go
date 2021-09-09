package elisp

type lispType int
type vectorLikeType int
type lispInt int64

const (
	lispTypeSymbol lispType = iota
	lispTypeInt
	lispTypeStr
	lispTypeVecLike
	lispTypeCons
	lispTypeFloat
	argsMany      = -1
	argsUnevalled = -2
)

const (
	vectorLikeTypeNormal vectorLikeType = iota
	vectorLikeTypeSubroutine
)

type lispFn0 func() (lispObject, error)
type lispFn1 func(lispObject) (lispObject, error)
type lispFn2 func(lispObject, lispObject) (lispObject, error)
type lispFnM func(...lispObject) (lispObject, error)

type subroutine struct {
	callabe0 lispFn0
	callabe1 lispFn1
	callabe2 lispFn2
	callabem lispFnM
	minArgs  int
	maxArgs  int
}

type lispSymbol struct {
	name     string
	value    lispObject
	function lispObject
}

type lispCons struct {
	car lispObject
	cdr lispObject
}

type lispInteger struct {
	value lispInt
}

type lispFloat struct {
	value float64
}

type lispString struct {
	value string
}

type vectorLikeValue interface{}

type lispVectorLike struct {
	vecType vectorLikeType
	value   vectorLikeValue
}

type lispObject interface {
	getType() lispType
}

func (ls *lispSymbol) getType() lispType {
	return lispTypeSymbol
}

func (lc *lispCons) getType() lispType {
	return lispTypeCons
}

func (li *lispInteger) getType() lispType {
	return lispTypeInt
}

func (lf *lispFloat) getType() lispType {
	return lispTypeFloat
}

func (ls *lispString) getType() lispType {
	return lispTypeStr
}

func (lv *lispVectorLike) getType() lispType {
	return lispTypeVecLike
}
