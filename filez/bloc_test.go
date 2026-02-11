package filez

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlocFile_NewBlocFile(t *testing.T) {
	p := MkTempOrPanic("/tmp/TestBlocFile_NewBlocFile")
	defer os.RemoveAll(p)

	bf, err := NewBlocFile(p, 8)
	require.NoError(t, err)
	require.NotNil(t, bf)
	assert.Equal(t, 8, bf.Cap())
	assert.Equal(t, 0, bf.Len())
}

func TestBlocFile_WriteNewBloc(t *testing.T) {
	// t.Skip()
	p := MkTempOrPanic("/tmp/TestBlocFile_WriteNewBloc")
	defer os.RemoveAll(p)

	bf, err := NewBlocFile(p, 4)
	require.NoError(t, err)
	require.NotNil(t, bf)
	assert.Equal(t, 4, bf.Cap())
	assert.Equal(t, 0, bf.Len())

	expectedBloc0 := "loremIpsum0"
	expectedBloc1 := "loremIpsum1"
	expectedBloc2 := "loremIpsum2"
	expectedBloc3 := "loremIpsum3"
	expectedBloc4 := "loremIpsum4"

	// Test Bloc Writing
	b0, err := bf.WriteNewBloc([]byte(expectedBloc0))
	assert.NoError(t, err)
	assert.NotNil(t, b0)
	assert.Equal(t, p, b0.uid.filepath)
	assert.Equal(t, 0, b0.uid.id)
	assert.Equal(t, expectedBloc0, string(*b0.data))
	assert.Equal(t, 4, bf.Cap())
	assert.Equal(t, 1, bf.Len())

	b1, err := bf.WriteNewBloc([]byte(expectedBloc1))
	assert.NoError(t, err)
	assert.NotNil(t, b1)
	assert.Equal(t, p, b1.uid.filepath)
	assert.Equal(t, 1, b1.uid.id)
	assert.Equal(t, expectedBloc1, string(*b1.data))
	assert.Equal(t, 4, bf.Cap())
	assert.Equal(t, 2, bf.Len())

	b2, err := bf.WriteNewBloc([]byte(expectedBloc2))
	assert.NoError(t, err)
	assert.NotNil(t, b2)
	assert.Equal(t, p, b2.uid.filepath)
	assert.Equal(t, 2, b2.uid.id)
	assert.Equal(t, expectedBloc2, string(*b2.data))

	b3, err := bf.WriteNewBloc([]byte(expectedBloc3))
	assert.NoError(t, err)
	assert.NotNil(t, b3)
	assert.Equal(t, p, b3.uid.filepath)
	assert.Equal(t, 3, b3.uid.id)
	assert.Equal(t, expectedBloc3, string(*b3.data))

	// Test Bloc Retrieval
	read0, err := bf.Get(0)
	assert.NoError(t, err)
	assert.NotNil(t, read0)
	assert.Equal(t, p, read0.uid.filepath)
	assert.Equal(t, 0, read0.uid.id)
	assert.Equal(t, expectedBloc0, string(*read0.data))

	read1, err := bf.Get(1)
	assert.NoError(t, err)
	assert.NotNil(t, read1)
	assert.Equal(t, p, read1.uid.filepath)
	assert.Equal(t, 1, read1.uid.id)
	assert.Equal(t, expectedBloc1, string(*read1.data))

	read2, err := bf.Get(2)
	assert.NoError(t, err)
	assert.NotNil(t, read2)
	assert.Equal(t, p, read2.uid.filepath)
	assert.Equal(t, 2, read2.uid.id)
	assert.Equal(t, expectedBloc2, string(*read2.data))

	read3, err := bf.Get(3)
	assert.NoError(t, err)
	assert.NotNil(t, read3)
	assert.Equal(t, p, read3.uid.filepath)
	assert.Equal(t, 3, read3.uid.id)
	assert.Equal(t, expectedBloc3, string(*read3.data))

	// Test capacity overflow
	b4, err := bf.WriteNewBloc([]byte(expectedBloc4))
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, b4)
	assert.Equal(t, 4, bf.Cap())
	assert.Equal(t, 4, bf.Len())
}

func TestBlocFile_UpdateLastBloc(t *testing.T) {
	// t.Skip()
	p := MkTempOrPanic("/tmp/TestBlocFile_UpdateLastBloc")
	defer os.RemoveAll(p)

	bf, err := NewBlocFile(p, 4)
	require.NoError(t, err)
	require.NotNil(t, bf)

	expectedBloc0 := "loremIpsum0"
	expectedBloc1 := "loremIpsum1"
	expectedBloc2 := "loremIpsum2"
	expectedBloc2b := "foo-loremIpsum2-bar"

	// Test Bloc Writing
	b0, err := bf.WriteNewBloc([]byte(expectedBloc0))
	assert.NoError(t, err)
	assert.NotNil(t, b0)
	assert.Equal(t, p, b0.uid.filepath)
	assert.Equal(t, 0, b0.uid.id)
	assert.Equal(t, expectedBloc0, string(*b0.data))

	b1, err := bf.WriteNewBloc([]byte(expectedBloc1))
	assert.NoError(t, err)
	assert.NotNil(t, b1)
	assert.Equal(t, p, b1.uid.filepath)
	assert.Equal(t, 1, b1.uid.id)
	assert.Equal(t, expectedBloc1, string(*b1.data))

	b2, err := bf.WriteNewBloc([]byte(expectedBloc2))
	assert.NoError(t, err)
	assert.NotNil(t, b2)
	assert.Equal(t, p, b2.uid.filepath)
	assert.Equal(t, 2, b2.uid.id)
	assert.Equal(t, expectedBloc2, string(*b2.data))

	// Test update last bloc
	b2b, err := bf.UpdateLastBloc([]byte(expectedBloc2b))
	assert.NoError(t, err)
	assert.NotNil(t, b2b)
	assert.Equal(t, p, b2b.uid.filepath)
	assert.Equal(t, 2, b2b.uid.id)
	assert.Equal(t, expectedBloc2b, string(*b2b.data))

	// Test Bloc Retrieval
	read0, err := bf.Get(0)
	assert.NoError(t, err)
	assert.NotNil(t, read0)
	assert.Equal(t, p, read0.uid.filepath)
	assert.Equal(t, 0, read0.uid.id)
	assert.Equal(t, expectedBloc0, string(*read0.data))

	read1, err := bf.Get(1)
	assert.NoError(t, err)
	assert.NotNil(t, read1)
	assert.Equal(t, p, read1.uid.filepath)
	assert.Equal(t, 1, read1.uid.id)
	assert.Equal(t, expectedBloc1, string(*read1.data))

	read2, err := bf.Get(2)
	assert.NoError(t, err)
	assert.NotNil(t, read2)
	assert.Equal(t, p, read2.uid.filepath)
	assert.Equal(t, 2, read2.uid.id)
	assert.Equal(t, expectedBloc2b, string(*read2.data))

	// Test reading not existing bloc
	read3, err := bf.Get(3)
	assert.Equal(t, ErrNotExist, err)
	assert.Nil(t, read3)
	assert.Equal(t, 4, bf.Cap())
	assert.Equal(t, 3, bf.Len())
}

func TestBlocFile_BlocRead(t *testing.T) {
	p := MkTempOrPanic("/tmp/TestBlocFile_BlocRead")
	defer os.RemoveAll(p)

	bf, err := NewBlocFile(p, 4)
	require.NoError(t, err)
	require.NotNil(t, bf)

	expectedBloc0 := "loremIpsum0"

	// Test Bloc Writing
	b0, err := bf.WriteNewBloc([]byte(expectedBloc0))
	assert.NoError(t, err)
	assert.NotNil(t, b0)
	assert.Equal(t, p, b0.uid.filepath)
	assert.Equal(t, 0, b0.uid.id)
	assert.Equal(t, expectedBloc0, string(*b0.data))

	buf := make([]byte, 100)
	n, err := b0.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, len([]byte(expectedBloc0)), n)
	assert.Equal(t, expectedBloc0, string(buf[0:n]))
}

func TestBlocFile_Cursor(t *testing.T) {
	// t.Skip()
	p := MkTempOrPanic("/tmp/TestBlocFile_Cursor")
	defer os.RemoveAll(p)

	bf, err := NewBlocFile(p, 4)
	require.NoError(t, err)
	require.NotNil(t, bf)

	expectedBloc0 := "loremIpsum0"
	expectedBloc1 := "loremIpsum1"
	expectedBloc2 := "loremIpsum2"
	expectedBloc3 := "loremIpsum3"

	c3 := bf.Cursor(TopToBottom)
	require.NotNil(t, c3)
	assert.False(t, c3.HasNext())

	// Test Bloc Writing
	b0, err := bf.WriteNewBloc([]byte(expectedBloc0))
	assert.NoError(t, err)
	assert.NotNil(t, b0)
	assert.Equal(t, p, b0.uid.filepath)
	assert.Equal(t, 0, b0.uid.id)
	assert.Equal(t, expectedBloc0, string(*b0.data))

	b1, err := bf.WriteNewBloc([]byte(expectedBloc1))
	assert.NoError(t, err)
	assert.NotNil(t, b1)
	assert.Equal(t, p, b1.uid.filepath)
	assert.Equal(t, 1, b1.uid.id)
	assert.Equal(t, expectedBloc1, string(*b1.data))

	b2, err := bf.WriteNewBloc([]byte(expectedBloc2))
	assert.NoError(t, err)
	assert.NotNil(t, b2)
	assert.Equal(t, p, b2.uid.filepath)
	assert.Equal(t, 2, b2.uid.id)
	assert.Equal(t, expectedBloc2, string(*b2.data))

	b3, err := bf.WriteNewBloc([]byte(expectedBloc3))
	assert.NoError(t, err)
	assert.NotNil(t, b3)
	assert.Equal(t, p, b3.uid.filepath)
	assert.Equal(t, 3, b3.uid.id)
	assert.Equal(t, expectedBloc3, string(*b3.data))

	// Test cursor as Reader
	c1 := bf.Cursor(TopToBottom)
	require.NotNil(t, c1)

	buf := new(strings.Builder)
	n, err := io.Copy(buf, c1)
	assert.NoError(t, err)
	assert.Equal(t, len([]byte(expectedBloc0))+len([]byte(expectedBloc1))+len([]byte(expectedBloc2))+len([]byte(expectedBloc3)), int(n))
	assert.Equal(t, expectedBloc0+expectedBloc1+expectedBloc2+expectedBloc3, buf.String())

	// Test cursor as Bloc Iterator
	c2 := bf.Cursor(TopToBottom)
	require.NotNil(t, c2)
	assert.True(t, c2.HasNext())
	c2b0, err := c2.Next()
	assert.NoError(t, err)
	require.NotNil(t, c2b0)
	assert.Equal(t, expectedBloc0, string(*c2b0.data))

	assert.True(t, c2.HasNext())
	c2b1, err := c2.Next()
	assert.NoError(t, err)
	require.NotNil(t, c2b1)
	assert.Equal(t, expectedBloc1, string(*c2b1.data))

	assert.True(t, c2.HasNext())
	c2b2, err := c2.Next()
	assert.NoError(t, err)
	require.NotNil(t, c2b2)
	assert.Equal(t, expectedBloc2, string(*c2b2.data))

	assert.True(t, c2.HasNext())
	c2b3, err := c2.Next()
	assert.NoError(t, err)
	require.NotNil(t, c2b3)
	assert.Equal(t, expectedBloc3, string(*c2b3.data))

	assert.False(t, c2.HasNext())
	c2b4, err := c2.Next()
	assert.Equal(t, ErrNotExist, err)
	assert.Nil(t, c2b4)

	// Test a new cursor iteration
	c2 = bf.Cursor(TopToBottom)
	require.NotNil(t, c2)
	assert.True(t, c2.HasNext())
	c2b0, err = c2.Next()
	assert.NoError(t, err)
	require.NotNil(t, c2b0)
	assert.Equal(t, expectedBloc0, string(*c2b0.data))

	assert.True(t, c2.HasNext())
	c2b1, err = c2.Next()
	assert.NoError(t, err)
	require.NotNil(t, c2b1)
	assert.Equal(t, expectedBloc1, string(*c2b1.data))

	// Test old cursor as Bloc Iterator
	assert.True(t, c3.HasNext())
	c3b0, err := c3.Next()
	assert.NoError(t, err)
	require.NotNil(t, c3b0)
	assert.Equal(t, expectedBloc0, string(*c3b0.data))

	assert.True(t, c3.HasNext())
	c3b1, err := c3.Next()
	assert.NoError(t, err)
	require.NotNil(t, c3b1)
	assert.Equal(t, expectedBloc1, string(*c3b1.data))

	assert.True(t, c3.HasNext())
	c3b2, err := c3.Next()
	assert.NoError(t, err)
	require.NotNil(t, c3b2)
	assert.Equal(t, expectedBloc2, string(*c3b2.data))

	assert.True(t, c3.HasNext())
	c3b3, err := c3.Next()
	assert.NoError(t, err)
	require.NotNil(t, c3b3)
	assert.Equal(t, expectedBloc3, string(*c3b3.data))

	assert.False(t, c3.HasNext())
	c3b4, err := c3.Next()
	assert.Equal(t, ErrNotExist, err)
	assert.Nil(t, c3b4)

	// Test cursor BlocOrdering

}
