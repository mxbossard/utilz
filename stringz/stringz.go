package stringz

import (
	"fmt"
	"regexp"
	"strings"
)

func Left(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

func Right(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[len(r)-n : len(r)])
	}
	return s
}

func SplitByRegexp(s, regex string) (parts []string, separators []string) {
	if s == "" {
		return
	}
	if regex == "" {
		parts = []string{s}
		return
	}
	r := regexp.MustCompile(regex)
	var sep string
	for {
		sep = r.FindString(s)
		splitted := r.Split(s, 2)
		if len(splitted) > 0 {
			parts = append(parts, splitted[0])
		}
		if sep == "" {
			break
		}
		separators = append(separators, sep)
		s = splitted[1]
	}
	return
}

func JoinStringers[T fmt.Stringer](stringers []T, separator string) string {
	var b strings.Builder
	if len(stringers) == 0 {
		return ""
	}
	b.WriteString(stringers[0].String())
	for _, s := range stringers[1:] {
		b.WriteString(separator)
		b.WriteString(s.String())
	}
	return b.String()
}
