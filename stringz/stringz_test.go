package stringz

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLeft(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"", 3, ""},
		{"foo", 3, "foo"},
		{"foo", 2, "fo"},
		{"  bar", 5, "  bar"},
		{"  bar", 4, "  ba"},
		{"bar\n", 4, "bar\n"},
		{"foo\nbar", 6, "foo\nba"},
		{"foo\nbar", 10, "foo\nbar"},
	}
	for i, c := range cases {
		got := Left(c.in, c.n)
		assert.Equal(t, c.want, got, "case #%d should be equal", i)
	}
}

func TestRight(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"", 3, ""},
		{"foo", 3, "foo"},
		{"foobar", 5, "oobar"},
		{"  bar", 5, "  bar"},
		{"  bar", 4, " bar"},
		{"bar\n", 2, "r\n"},
		{"foo\nbar", 4, "\nbar"},
		{"foo\nbar", 10, "foo\nbar"},
	}
	for i, c := range cases {
		got := Right(c.in, c.n)
		assert.Equal(t, c.want, got, "case #%d should be equal", i)
	}
}

func TestSplitByRegexp(t *testing.T) {
	cases := []struct {
		in    string
		regex string
		want1 []string
		want2 []string
	}{
		{"", "[,:]", []string(nil), []string(nil)},
		{"", "", []string(nil), []string(nil)},
		{"foo,bar,baz", "", []string{"foo,bar,baz"}, []string(nil)},
		{"foo,bar,baz", "[,:]", []string{"foo", "bar", "baz"}, []string{",", ","}},
		{"foo,bar:baz", "[,:]", []string{"foo", "bar", "baz"}, []string{",", ":"}},
	}
	for i, c := range cases {
		got1, got2 := SplitByRegexp(c.in, c.regex)
		assert.Equal(t, c.want1, got1, "case #%d 1 should be equal", i)
		assert.Equal(t, c.want2, got2, "case #%d 2 should be equal", i)
	}
}
