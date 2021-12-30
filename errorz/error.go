package errorz

import (
	//"fmt"
	"errors"
	"strings"
	"reflect"
)

type Aggregated struct {
	errors []error
}

func (a Aggregated) GotError() bool {
	return len(a.errors) > 0
}

func (a *Aggregated) Add(e error) {
	a.errors = append([]error{e}, a.errors...)
}

func (a *Aggregated) AddAll(errs ...error) {
	a.errors = append(errs, a.errors...)
	//for _, e := range errs {
	//	a.Add(e)
	//}
}

func (a *Aggregated) Concat(agg Aggregated) {
	a.AddAll(agg.errors...)
}

func (a Aggregated) Error() string {
	builder := strings.Builder{}
	for i, e := range a.errors {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(e.Error())
	}
	return builder.String()
}

func (a Aggregated) Errors() []error {
	return a.errors
}

func (a Aggregated) Get(target interface{}) (errs []error) {
	if target == nil {
		return nil
	}
	for _, e := range a.errors {
		if errors.As(e, target) {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		errs = nil
	}

	// Reverse array order
	//for i, j := 0, len(errs)-1; i < j; i, j = i+1, j-1 {
    	//	errs[i], errs[j] = errs[j], errs[i]
	//}

	return errs
}

func (a Aggregated) Is(target error) bool {
	for _, e := range a.errors {
		if errors.Is(e, target) || reflect.DeepEqual(e, target) {
			return true
		}
	}
	return false
}

func (a Aggregated) As(target interface{}) bool {
	for _, e := range a.errors {
		if errors.As(e, target) {
			target = &e
			return true
		}
	}
	return false
}

func (a Aggregated) Unwrap() error {
	if len(a.errors) > 1 {
		unwrapped := Aggregated{}
		unwrapped.errors = a.errors[:len(a.errors) - 1]
		return unwrapped
	}
	return nil
}
