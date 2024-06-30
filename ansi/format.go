package ansi

import (
	"strings"
)

func AnsiRulesIndex(in string) (int, [][]int) {
	pos := ansiRulePattern.FindAllStringSubmatchIndex(in, -1)
	ansiLength := 0
	for _, p := range pos {
		ansiLength += p[1] - p[0]
	}
	return ansiLength, pos
}

func AnsiRules(in string) (rules []string) {
	_, pos := AnsiRulesIndex(in)
	for _, p := range pos {
		rules = append(rules, in[p[0]:p[1]])
	}
	return
}

func PadLeft(in string, pad int) (out string) {
	ansiLength, _ := AnsiRulesIndex(in)
	spaceCount := pad - len(in) - ansiLength
	if spaceCount > 0 {
		out += strings.Repeat(" ", spaceCount)
	}
	out += in
	return
}

func PadRight(in string, pad int) (out string) {
	ansiLength, _ := AnsiRulesIndex(in)
	out += in
	spaceCount := pad - len(in) - ansiLength
	if spaceCount > 0 {
		out += strings.Repeat(" ", spaceCount)
	}
	return
}

func TruncateLeftPrefix(in string, length int, prefix string) string {
	if len(prefix) >= int(length) {
		panic("prefix must be smaller than length")
	}
	ansiLength, pos := AnsiRulesIndex(in)
	if len(in)-ansiLength > length {
		// Need to truncate
		length -= len(prefix)
		var out string
		if len(pos) > 0 {
			ptr := len(in)
			remains := length
			for i := len(pos) - 1; i >= 0; i-- {
				left := pos[i][0]
				right := pos[i][1]
				if remains > 0 {
					notAnsi := in[max(right, ptr-remains):ptr]
					out = notAnsi + out
					remains = remains - len(notAnsi)
				}
				out = in[left:right] + out
				ptr = left
			}
			out = in[ptr-remains:ptr] + out
		} else {
			out = in[len(in)-length:]
		}
		out = prefix + out
		return out
	}
	return in
}

func TruncateLeft(in string, length int) string {
	return TruncateLeftPrefix(in, length, "")
}

func TruncateRightSuffix(in string, length int, suffix string) string {
	if len(suffix) >= int(length) {
		panic("suffix must be smaller than length")
	}
	ansiLength, pos := AnsiRulesIndex(in)
	if len(in)-ansiLength > length {
		// Need to truncate
		length -= len(suffix)
		var out string
		if len(pos) > 0 {
			ptr := 0
			remains := length
			for i := 0; i < len(pos); i++ {
				left := pos[i][0]
				right := pos[i][1]
				if remains > 0 {
					notAnsi := in[ptr:min(left, ptr+remains)]
					out = out + notAnsi
					remains = remains - len(notAnsi)
				}
				out = out + in[left:right]
				ptr = right
			}
			out = out + in[ptr:ptr+remains]
		} else {
			out = in[:length]
		}
		out = out + suffix
		return out
	}
	return in
}

func TruncateRight(in string, length int) string {
	return TruncateRightSuffix(in, length, "")
}

func TruncateMid(in string, length int, replacement string) string {
	if len(replacement) >= int(length) {
		panic("replacement must be smaller than length")
	}

	ansiLength, _ := AnsiRulesIndex(in)
	if len(in)-ansiLength > length {
		length -= len(replacement)
		left := length / 2
		right := length - left

		out := TruncateRight(in, left) + replacement + TruncateLeft(in, right)
		return out
	}
	return in
}
