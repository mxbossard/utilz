package stringz

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	spacesRegexp = regexp.MustCompile(`\s+`)
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
		return string(r[len(r)-n:])
	}
	return s
}

func SummaryRatioEllipsis(s string, length int, leftRatio float32, ellipsis string) string {
	if leftRatio < 0 {
		leftRatio = 0
	} else if leftRatio > 1 {
		leftRatio = 1
	}
	if length < 2*len(ellipsis) {
		ellipsis = ""
	}

	// Clean spaces
	s = strings.TrimSpace(s)
	s = spacesRegexp.ReplaceAllString(s, " ")

	if len(s) <= length {
		return s
	}

	leftCount := int(float32(length-len(ellipsis)) * leftRatio)
	rightCount := length - len(ellipsis) - leftCount

	sb := strings.Builder{}
	sb.WriteString(s[0:leftCount])
	sb.WriteString(ellipsis)
	sb.WriteString(s[len(s)-rightCount:])
	return sb.String()
}

func SummaryRatioEllipsisRune(s string, length int, leftRatio float32, ellipsis string) string {
	// FIXME: don't know if it's better
	if leftRatio < 0 {
		leftRatio = 0
	} else if leftRatio > 1 {
		leftRatio = 1
	}
	elipsisLen := len([]rune(ellipsis))
	if length < 2*elipsisLen {
		ellipsis = ""
	}

	// Clean spaces
	s = strings.TrimSpace(s)
	s = spacesRegexp.ReplaceAllString(s, " ")

	if len([]rune(s)) <= length {
		return s
	}

	leftCount := int(float32(length-elipsisLen) * leftRatio)
	rightCount := length - elipsisLen - leftCount

	sb := strings.Builder{}
	sb.WriteString(string([]rune(s)[0:leftCount]))
	sb.WriteString(ellipsis)
	sb.WriteString(string([]rune(s)[len(s)-rightCount:]))
	return sb.String()
}

func SummaryRatio(s string, length int, leftRatio float32) string {
	return SummaryRatioEllipsis(s, length, leftRatio, "[...]")
}

func Summary(s string, length int) string {
	return SummaryRatioEllipsis(s, length, 0.5, "[...]")
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
