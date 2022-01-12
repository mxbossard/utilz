package format

import (
	"testing"
        "github.com/stretchr/testify/assert"
)

func TestPadLeft(t *testing.T) {
        cases := []struct {
                in string
		l int
		want string
        }{
                {"", 3, "   "},
                {"foo", 3, "foo"},
                {"foo", 5, "  foo"},
                {"  bar", 5, "  bar"},
                {"  bar", 7, "    bar"},
                {"bar\n", 5, " bar\n"},
                {"foo\nbar", 7, "foo\nbar"},
                {"foo\nbar", 8, " foo\nbar"},
        }
        for i, c := range cases {
                got := PadLeft(c.in, c.l)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

func TestPadRight(t *testing.T) {
        cases := []struct {
                in string
		l int
		want string
        }{
                {"", 3, "   "},
                {"foo", 3, "foo"},
                {"foo", 5, "foo  "},
                {"  bar", 5, "  bar"},
                {"  bar", 7, "  bar  "},
                {"bar\n", 5, "bar\n "},
                {"foo\nbar", 7, "foo\nbar"},
                {"foo\nbar", 8, "foo\nbar "},
        }
        for i, c := range cases {
                got := PadRight(c.in, c.l)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

func TestTruncateLeft(t *testing.T) {
        cases := []struct {
                in string
		l int
		want string
        }{
                {"", 3, ""},
                {"foo", 3, "foo"},
                {"foo", 2, "oo"},
                {"  bar", 3, "bar"},
                {"  bar", 2, "ar"},
                {"bar\n", 3, "ar\n"},
                {"foo\nbar", 5, "o\nbar"},
        }
        for i, c := range cases {
                got := TruncateLeft(c.in, c.l)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

func TestTruncateRight(t *testing.T) {
        cases := []struct {
                in string
		l int
		want string
        }{
                {"", 3, ""},
                {"foo", 3, "foo"},
                {"foo", 2, "fo"},
                {"  bar", 3, "  b"},
                {"  bar", 2, "  "},
                {"bar\n", 3, "bar"},
                {"foo\nbar", 5, "foo\nb"},
        }
        for i, c := range cases {
                got := TruncateRight(c.in, c.l)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

func TestTruncateLeftPrefix(t *testing.T) {
        cases := []struct {
                in string
		l int
		prefix string
		want string
        }{
                {"", 3, "..", ""},
                {"foobar", 6, "..", "foobar"},
                {"foobar", 5, "..", "..bar"},
                {"foobar", 2, "..", ".."},
                {"foobar", 2, "abc", "ab"},
                {"   foobar", 7, "..", "..oobar"},
                {"foobar   ", 7, "..", "..ar   "},
                {"foobar\n", 5, "..", "..ar\n"},
                {"bazfoo\nbar", 7, "..", "..o\nbar"},
        }
        for i, c := range cases {
                got := TruncateLeftPrefix(c.in, c.l, c.prefix)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

func TestTruncateRightSuffix(t *testing.T) {
        cases := []struct {
                in string
		l int
		prefix string
		want string
        }{
                {"", 3, "..", ""},
                {"foobar", 6, "..", "foobar"},
                {"foobar", 5, "..", "foo.."},
                {"foobar", 2, "..", ".."},
                {"foobar", 2, "abc", "bc"},
                {"   foobar", 7, "..", "   fo.."},
                {"foobar   ", 7, "..", "fooba.."},
                {"\nfoobar", 5, "..", "\nfo.."},
                {"foo\nbarbaz", 7, "..", "foo\nb.."},
        }
        for i, c := range cases {
                got := TruncateRightSuffix(c.in, c.l, c.prefix)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}
