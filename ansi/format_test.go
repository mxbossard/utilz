package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateLeft(t *testing.T) {
	cases := []struct {
		in   string
		l    int
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

func TestTruncateLeftPrefix(t *testing.T) {
	cases := []struct {
		in     string
		l      int
		prefix string
		want   string
	}{
		{"", 3, "..", ""},
		{"foobar", 6, "..", "foobar"},
		{"foobar", 5, "..", "..bar"},
		{"foobar", 3, "..", "..r"},
		{"foobar", 4, "abc", "abcr"},
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

func TestTruncateLeft_Ansi(t *testing.T) {
	cases := []struct {
		in   string
		l    int
		want string
	}{
		{string(Red) + "" + string(Reset), 3, string(Red) + "" + string(Reset)},
		{string(Red) + "foo" + string(Reset), 3, string(Red) + "foo" + string(Reset)},
		{string(Red) + "foo" + string(Reset), 2, string(Red) + "oo" + string(Reset)},
		{"bla" + string(Red) + "foo" + string(Reset), 2, "" + string(Red) + "oo" + string(Reset)},
		{string(Red) + "foo" + string(Reset) + "bla", 2, string(Red) + "" + string(Reset) + "la"},
		{string(Red) + "  bar" + string(Reset), 3, string(Red) + "bar" + string(Reset)},
		{string(Red) + "  bar" + string(Reset), 2, string(Red) + "ar" + string(Reset)},
		{string(Red) + "bar\n" + string(Reset), 3, string(Red) + "ar\n" + string(Reset)},
		{string(Red) + "foo\nbar" + string(Reset), 5, string(Red) + "o\nbar" + string(Reset)},
		{"foo" + string(Reset) + string(Red) + "bar" + string(Reset) + "baz", 7, "o" + string(Reset) + string(Red) + "bar" + string(Reset) + "baz"},
	}
	for i, c := range cases {
		got := TruncateLeft(c.in, c.l)
		assert.Equal(t, c.want, got, "case #%d should be equal", i)
	}
}

func TestTruncateRight(t *testing.T) {
	cases := []struct {
		in   string
		l    int
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

func TestTruncateRightSuffix(t *testing.T) {
	cases := []struct {
		in     string
		l      int
		prefix string
		want   string
	}{
		{"", 3, "..", ""},
		{"foobar", 6, "..", "foobar"},
		{"foobar", 5, "..", "foo.."},
		{"foobar", 3, "..", "f.."},
		{"foobar", 4, "abc", "fabc"},
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

func TestTruncateRight_Ansi(t *testing.T) {
	cases := []struct {
		in   string
		l    int
		want string
	}{
		{string(Red) + "" + string(Reset), 3, string(Red) + "" + string(Reset)},
		{string(Red) + "foobar" + string(Reset), 6, string(Red) + "foobar" + string(Reset)},
		{string(Red) + "foobar" + string(Reset), 5, string(Red) + "fooba" + string(Reset)},
		{"bla" + string(Red) + "foo" + string(Reset), 2, "bl" + string(Red) + "" + string(Reset)},
		{string(Red) + "foo" + string(Reset) + "bla", 2, string(Red) + "fo" + string(Reset) + ""},
		{string(Red) + "foobar" + string(Reset), 3, string(Red) + "foo" + string(Reset)},
		{string(Red) + "foobar" + string(Reset), 4, string(Red) + "foob" + string(Reset)},
		{string(Red) + "   foobar" + string(Reset), 7, string(Red) + "   foob" + string(Reset)},
		{string(Red) + "foobar   " + string(Reset), 7, string(Red) + "foobar " + string(Reset)},
		{string(Red) + "\nfoobar" + string(Reset), 5, string(Red) + "\nfoob" + string(Reset)},
		{string(Red) + "foo\nbarbaz" + string(Reset), 7, string(Red) + "foo\nbar" + string(Reset)},
		{string(Red) + "foo" + string(Reset) + string(Red) + "bar" + string(Reset) + "baz", 7, string(Red) + "foo" + string(Reset) + string(Red) + "bar" + string(Reset) + "b"},
	}
	for i, c := range cases {
		got := TruncateRight(c.in, c.l)
		assert.Equal(t, c.want, got, "case #%d should be equal", i)
	}
}
