package format

import (
	"strings"
)

func PadLeft(in string, pad int) (out string) {
	spaceCount := pad - len(in)
	if spaceCount > 0 {
		out += strings.Repeat(" ", spaceCount)
	}
	out += in
	return
}

func PadRight(in string, pad int) (out string) {
	out += in
	spaceCount := pad - len(in)
	if spaceCount > 0 {
		out += strings.Repeat(" ", spaceCount)
	}
	return
}

func TruncateLeft(in string, length int)  (out string) {
	if len(in) > length {
		s:= len(in) - length
	       return in[s:len(in)]
	}
	return in
}

func TruncateRight(in string, length int)  (out string) {
	if len(in) > length {
		return in[:length]
	}
	return in
}

func TruncateLeftPrefix(in string, length int, prefix string)  (out string) {
	prefix = TruncateRight(prefix, length)
	if len(in) > length {
		return prefix + TruncateLeft(in, length - len(prefix))
	}
	return in
}

func TruncateRightSuffix(in string, length int, suffix string)  (out string) {
	suffix = TruncateLeft(suffix, length)
	if len(in) > length {
	       return TruncateRight(in, length - len(suffix)) + suffix
	}
	return in
}

