package stringz

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
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

func TestSummaryRatioEllipsis(t *testing.T) {
	ellipsis := "[...]"
	shortMsg := "bar"
	shortMultilineMsg := `foofoofoo
		barbarbar
		bazbazbaz`
	longMsg := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	multilineMsg := `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
		Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
		Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 
		Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	
	assert.Equal(t, "", SummaryRatioEllipsis("", 10, 0.5, ellipsis))
	assert.Equal(t, "Lorem ipsum dolor sit amet", SummaryRatioEllipsis("  Lorem   ipsum  dolor  sit  amet ", 40, 0.5, ellipsis))
	
	assert.Equal(t, shortMsg, SummaryRatioEllipsis(shortMsg, 20, 0.5, ellipsis))
	assert.Equal(t, shortMsg, SummaryRatioEllipsis(shortMsg, 4, 0.5, ellipsis))
	assert.Equal(t, "br", SummaryRatioEllipsis(shortMsg, 2, 0.5, ellipsis))

	assert.Equal(t, "foofoof[...]azbazbaz", SummaryRatioEllipsis(shortMultilineMsg, 20, 0.5, ellipsis))
	assert.Equal(t, "foofo[...] bazbazbaz", SummaryRatioEllipsis(shortMultilineMsg, 20, 0.34, ellipsis))
	assert.Equal(t, "foofoofoo barbarbar bazbazbaz", SummaryRatioEllipsis(shortMultilineMsg, 41, 0.34, ellipsis))
	assert.Equal(t, "foofoofoo barbarbar bazbazbaz", SummaryRatioEllipsis(shortMultilineMsg, 41, 0.25, ellipsis))
	assert.Equal(t, "foaz", SummaryRatioEllipsis(shortMultilineMsg, 4, 0.5, ellipsis))

	assert.Equal(t, "Lorem i[...]laborum.", SummaryRatioEllipsis(longMsg, 20, 0.5, ellipsis))
	assert.Equal(t, "Lorem[...]t laborum.", SummaryRatioEllipsis(longMsg, 20, 0.34, ellipsis))
	assert.Equal(t, "Lorem ipsum dolor [...]im id est laborum.", SummaryRatioEllipsis(longMsg, 41, 0.5, ellipsis))
	assert.Equal(t, "Lorem ips[...]mollit anim id est laborum.", SummaryRatioEllipsis(longMsg, 41, 0.25, ellipsis))
	assert.Equal(t, "Lom.", SummaryRatioEllipsis(longMsg, 4, 0.5, ellipsis))

	assert.Equal(t, "Lorem i[...]laborum.", SummaryRatioEllipsis(multilineMsg, 20, 0.5, ellipsis))
	assert.Equal(t, "Lorem[...]t laborum.", SummaryRatioEllipsis(multilineMsg, 20, 0.34, ellipsis))
	assert.Equal(t, "Lorem ipsum dolor [...]im id est laborum.", SummaryRatioEllipsis(multilineMsg, 41, 0.5, ellipsis))
	assert.Equal(t, "Lorem ips[...]mollit anim id est laborum.", SummaryRatioEllipsis(multilineMsg, 41, 0.25, ellipsis))
	assert.Equal(t, "Lom.", SummaryRatioEllipsis(multilineMsg, 4, 0.5, ellipsis))
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

func TestJoinStringers(t *testing.T) {
	s1 := strings.Builder{}
	s1.WriteString("foo")
	s2 := strings.Builder{}
	s2.WriteString("bar")
	s3 := strings.Builder{}
	s3.WriteString("baz")

	t0 := JoinStringers([]fmt.Stringer{&s1}, "-")
	assert.Equal(t, "foo", t0)

	t1 := JoinStringers([]fmt.Stringer{&s1, &s2}, "")
	assert.Equal(t, "foobar", t1)

	t2 := JoinStringers([]fmt.Stringer{&s1, &s2}, " ")
	assert.Equal(t, "foo bar", t2)

	t3 := JoinStringers([]fmt.Stringer{&s1, &s2}, "\n")
	assert.Equal(t, "foo\nbar", t3)

	t4 := JoinStringers([]fmt.Stringer{&s1, &s2, &s3}, " ")
	assert.Equal(t, "foo bar baz", t4)

	t5 := JoinStringers([]fmt.Stringer{}, "-")
	assert.Equal(t, "", t5)
}
