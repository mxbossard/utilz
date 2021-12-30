package errorz

import (
	"testing"
	//"fmt"
	"errors"

        "github.com/stretchr/testify/assert"
        //"github.com/stretchr/testify/require"

)

type baseError struct {
	content string
}
func (e baseError) Error() string {
	return e.content
}

type error1 struct {
	baseError
}
type error2 struct {
	baseError
}
type error3 struct {
	baseError
}
type error4 struct {
	baseError
}
type error5 struct {
	baseError
}

type errorWithArray struct {
	baseError
	array *[]string
}

var (
	err1a error1 = error1{baseError{"err1a"}}
	err1b error1 = error1{baseError{"err1b"}}
	err2a error2 = error2{baseError{"err2a"}}
	err2b error2 = error2{baseError{"err2b"}}
	err3 error3 = error3{baseError{"err3"}}
	err4 error4 = error4{baseError{"err4"}}
	err5 error5 = error5{baseError{"err5"}}
	err6 error3 = error3{baseError{"err6"}}
	arrayErr errorWithArray = errorWithArray{baseError{"foo"}, &[]string{"bar", "baz"}}
)

func TestIs(t *testing.T) {
	agg := Aggregated{}
	agg.Add(err1a)
	agg.Add(err2a)
	agg.Add(err3)
	agg.Add(arrayErr)

	var err1aBis error = error1{baseError{"err1a"}}
	var arrayErrBis error = errorWithArray{baseError{"foo"}, &[]string{"bar", "baz"}}
	var err1c error = error1{baseError{"err1c"}}

	assert.ErrorIs(t, agg, err1a)
	assert.ErrorIs(t, agg, err1aBis)
	assert.ErrorIs(t, agg, arrayErrBis)

	assert.True(t, agg.Is(err1a))
	assert.True(t, agg.Is(err1aBis))
	assert.True(t, agg.Is(arrayErrBis))
	assert.False(t, agg.Is(err1c))
}

func TestAs(t *testing.T) {
	agg := Aggregated{}
	agg.Add(err1a)
	agg.Add(err2a)
	agg.Add(err2b)
	agg.Add(err3)
	agg.Add(arrayErr)

	assert.ErrorIs(t, &err1a, &err1a)
	assert.ErrorIs(t, &arrayErr, &arrayErr)
	assert.True(t, errors.Is(&err1a, &err1a))
	assert.False(t, errors.Is(&err1a, &err2a))
	assert.True(t, errors.Is(&arrayErr, &arrayErr))
	assert.True(t, errors.Is(err1a, err1a))
	assert.False(t, errors.Is(err1a, err2a))
	assert.True(t, errors.Is(arrayErr, arrayErr))

	e1 := &error1{}
	e2 := &error2{}
	e3 := &error3{}
	e4 := &error4{}
	eWa := &errorWithArray{}

	assert.ErrorAs(t, agg, e1)
	assert.ErrorAs(t, agg, e2)
	assert.ErrorAs(t, agg, e3)
	assert.ErrorAs(t, agg, eWa)

	assert.Equal(t, err1a, *e1)
	assert.Equal(t, err2b, *e2)
	assert.Equal(t, err3, *e3)
	assert.Equal(t, arrayErr, *eWa)

	assert.ErrorIs(t, *e1, err1a)
	assert.ErrorIs(t, *e2, err2b)
	assert.ErrorIs(t, *e3, err3)
	assert.ErrorIs(t, *eWa, arrayErr)

	e1 = &error1{}
	e2 = &error2{}
	e3 = &error3{}
	e4 = &error4{}
	eWa = &errorWithArray{}

	assert.True(t, agg.As(e1))
	assert.True(t, agg.As(e2))
	assert.True(t, agg.As(e3))
	assert.True(t, agg.As(eWa))
	assert.False(t, agg.As(e4))

	assert.Equal(t, err1a, *e1)
	assert.Equal(t, err2b, *e2)
	assert.Equal(t, err3, *e3)
	assert.Equal(t, arrayErr, *eWa)

	assert.ErrorIs(t, *e1, err1a)
	assert.ErrorIs(t, *e2, err2b)
	assert.ErrorIs(t, *e3, err3)
	assert.ErrorIs(t, *eWa, arrayErr)

}

func TestUnwrap(t *testing.T) {
	agg := Aggregated{}
	agg.Add(err1a)
	agg.Add(err2a)
	agg.Add(err3)

	assert.Len(t, agg.Errors(), 3)
	assert.ErrorIs(t, agg, err1a)
	assert.ErrorIs(t, agg, err2a)
	assert.ErrorIs(t, agg, err3)

	unwrapped := agg.Unwrap().(Aggregated)
	assert.Len(t, unwrapped.Errors(), 2)
	assert.ErrorIs(t, agg, err1a)
	assert.ErrorIs(t, agg, err2a)
}

func TestAggregated(t *testing.T) {
	agg := Aggregated{}

	assert.False(t, agg.GotError())
	assert.Len(t, agg.Errors(), 0)

	agg.Add(err1a)
	agg.Add(err2a)
	agg.Add(err3)

	assert.True(t, agg.GotError())

	assert.Equal(t, "err3\nerr2a\nerr1a", agg.Error())

	assert.True(t, agg.Is(err1a))
	assert.True(t, agg.Is(err2a))
	assert.True(t, agg.Is(err3))
	assert.False(t, agg.Is(err4))
	assert.False(t, agg.Is(err5))
	assert.False(t, agg.Is(err6))

	assert.True(t, errors.Is(agg, err1a))
	assert.True(t, errors.Is(agg, err2a))
	assert.True(t, errors.Is(agg, err3))
	assert.False(t, errors.Is(agg, err4))
	assert.False(t, errors.Is(agg, err5))
	assert.False(t, errors.Is(agg, err6))

	assert.True(t, agg.As(&error1{}))
	assert.True(t, agg.As(&error2{}))
	assert.True(t, agg.As(&error3{}))
	assert.False(t, agg.As(&error4{}))
	assert.False(t, agg.As(&error5{}))

	assert.True(t, errors.As(agg, &error1{}))
	assert.True(t, errors.As(agg, &error2{}))
	assert.True(t, errors.As(agg, &error3{}))
	assert.False(t, errors.As(agg, &error4{}))
	assert.False(t, errors.As(agg, &error5{}))

	assert.Equal(t, []error{err1a}, agg.Get(&error1{}))
	assert.Equal(t, []error{err2a}, agg.Get(&error2{}))
	assert.Equal(t, []error{err3}, agg.Get(&error3{}))
	assert.Nil(t, agg.Get(&error4{}))
	assert.Nil(t, agg.Get(&error5{}))

	assert.Len(t, agg.Errors(), 3)

	agg2 := Aggregated{}
	agg2.Add(err4)
	agg2.Add(err6)

	agg.Concat(agg2)

	assert.Len(t, agg.Errors(), 5)

	assert.True(t, errors.Is(agg, err1a))
	assert.True(t, errors.Is(agg, err2a))
	assert.True(t, errors.Is(agg, err3))
	assert.True(t, errors.Is(agg, err4))
	assert.False(t, errors.Is(agg, err5))
	assert.True(t, errors.Is(agg, err6))
	assert.True(t, errors.As(agg, &error1{}))
	assert.True(t, errors.As(agg, &error2{}))
	assert.True(t, errors.As(agg, &error3{}))
	assert.True(t, errors.As(agg, &error4{}))
	assert.False(t, errors.As(agg, &error5{}))

	assert.Equal(t, []error{err1a}, agg.Get(&error1{}))
	assert.Equal(t, []error{err2a}, agg.Get(&error2{}))
	assert.Equal(t, []error{err6, err3}, agg.Get(&error3{}))
	assert.Equal(t, []error{err4}, agg.Get(&error4{}))
	assert.Nil(t, agg.Get(&error5{}))
}


