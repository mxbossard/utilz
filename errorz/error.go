package errorz

import (
	//"fmt"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Aggregated struct {
	errors []error
}

// Return nil if no errors
func (a Aggregated) Return() error {
	if !a.GotError() {
		return nil
	}
	return a
}

func (a Aggregated) GotError() bool {
	return len(a.errors) > 0
}

func (a *Aggregated) Add(e error) {
	if e != nil {
		a.errors = append([]error{e}, a.errors...)
	}
}

func (a *Aggregated) AddAll(errs ...error) {
	//a.errors = append(errs, a.errors...)
	for _, e := range errs {
		a.Add(e)
	}
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
		unwrapped.errors = a.errors[:len(a.errors)-1]
		return unwrapped
	}
	return nil
}

// Aggregate all errors in a chan error
func ConsumedAggregated(errorsChan chan error) Aggregated {
	var errors Aggregated
	for {
		var err error
		// Use select to not block if no error in channel
		select {
		case err = <-errorsChan:
			errors.Add(err)
		default:
		}
		if err == nil {
			break
		}
	}
	return errors
}

func NewAggregated(errors ...error) Aggregated {
	agg := Aggregated{}
	agg.AddAll(errors...)
	return agg
}

func Fatal(a ...any) {
	fmt.Fprint(os.Stderr, a...)
	os.Exit(1)
}

func Fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func Panic(a ...any) {
	msg := fmt.Sprint(a...)
	panic(msg)
}

func Panicf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	panic(msg)
}
