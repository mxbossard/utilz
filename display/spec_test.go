package display

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScreen(t *testing.T) {
	s := NewScreen()
	assert.NotNil(t, s)
}

func TestGetAsyncScreen(t *testing.T) {
	s := NewAsyncScreen("/tmp/foo")
	assert.NotNil(t, s)
}

func TestGetSession(t *testing.T) {
	s := NewAsyncScreen("/tmp/foo")
	require.NotNil(t, s)
	session := s.Session("foo", 42)
	assert.NotNil(t, session)
}

func TestGetSessionPrinter(t *testing.T) {
	s := NewAsyncScreen("/tmp/foo")
	require.NotNil(t, s)
	session := s.Session("foo", 42)
	require.NotNil(t, session)
	prtr := session.Printer("bar")
	assert.NotNil(t, prtr)
}
