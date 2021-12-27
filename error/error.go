package error

import (
	//"fmt"
	"strings"
)

type Aggregated struct {
	errors []error
}

func (a *Aggregated) Add(e error) {
	a.errors = append(a.errors, e)
}

func (a Aggregated) Error() string {
	builder := strings.Builder{}
	for _, e := range a.errors {
		builder.WriteString(e.Error())
		builder.WriteString("\n")
	}
	return builder.String()
}

func (a Aggregated) Is(target error) bool {
	for _, e := range a.errors {
		if e == target {
			return true
		}
	}
	return false
}

func (a Aggregated) Get(target error) (errors []error) {
	for _, e := range a.errors {
		if e == target {
			errors = append(errors, e)
		}
	}
	if len(errors) == 0 {
		errors = nil
	}
	return errors
}
