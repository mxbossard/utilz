package logging

import (
	"fmt"
	"os"
)

func ErrorPrint(a ...interface{}) (n int, err error) {
	s := fmt.Sprint(a...)
	return os.Stderr.WriteString(s)
}

func ErrorPrintf(format string, a ...interface{}) (n int, err error) {
	s := fmt.Sprintf(format, a...)
	return os.Stderr.WriteString(s)
}
