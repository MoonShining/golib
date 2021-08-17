package linkedbuffer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkedBufferWrite(t *testing.T) {
	t.Run("short write", func(t *testing.T) {
		data := "ab"
		buffer := New(3)
		buffer.Write([]byte(data))
		assert.Equal(t, data, buffer.String())
	})
	t.Run("equal write", func(t *testing.T) {
		data := "abc"
		buffer := New(3)
		buffer.Write([]byte(data))
		assert.Equal(t, data, buffer.String())
	})
	t.Run("large write", func(t *testing.T) {
		data := "abcdefgh"
		buffer := New(3)
		buffer.Write([]byte(data))
		assert.Equal(t, data, buffer.String())
	})
}

func TestLinkedBufferRead(t *testing.T) {
	t.Run("large read", func(t *testing.T) {
		data := "abc"
		buffer := New(3)
		buffer.Write([]byte(data))

		p := make([]byte, 4)
		n, _ := buffer.Read(p)
		assert.Equal(t, len(data), n)
		assert.Equal(t, string(p[:n]), data)
	})
	t.Run("equal read", func(t *testing.T) {
		data := "abc"
		buffer := New(3)
		buffer.Write([]byte(data))

		p := make([]byte, 3)
		n, _ := buffer.Read(p)
		assert.Equal(t, len(data), n)
		assert.Equal(t, string(p), data)
	})
	t.Run("short read", func(t *testing.T) {
		data := "abcdefgh"
		buffer := New(3)
		buffer.Write([]byte(data))
		assert.Equal(t, data, buffer.String())

		p := make([]byte, 4)
		n, _ := buffer.Read(p)
		assert.Equal(t, len(p), n)
		assert.Equal(t, string(p), data[:n])
	})
	t.Run("exhausted read", func(t *testing.T) {
		data := "abcdefgh"
		buffer := New(3)
		buffer.Write([]byte(data))
		assert.Equal(t, data, buffer.String())

		p := make([]byte, 10)
		n, _ := buffer.Read(p)
		assert.Equal(t, len(data), n)
		assert.Equal(t, string(p[:n]), data)
	})
}

func TestLinkedBufferReadWrite(t *testing.T) {
	buffer := New(4)

	p := make([]byte, 3)
	buffer.Write([]byte("abc"))
	assert.Equal(t, 3, buffer.Len())
	buffer.Read(p)
	assert.Equal(t, string(p), "abc")
	assert.Equal(t, 0, buffer.Len())

	p = make([]byte, 3)
	n, _ := buffer.Read(p)
	assert.Equal(t, 0, n)

	buffer.Write([]byte("defg"))
	assert.Equal(t, 4, buffer.Len())
	buffer.Read(p)
	assert.Equal(t, string(p), "def")
	assert.Equal(t, 1, buffer.Len())

	p = make([]byte, 100)
	buffer.Write([]byte("hijklmn"))
	assert.Equal(t, 8, buffer.Len())
	n, _ = buffer.Read(p)
	assert.Equal(t, string(p[:n]), "ghijklmn")
	assert.Equal(t, 0, buffer.Len())
}
