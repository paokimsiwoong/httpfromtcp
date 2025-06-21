package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderLineParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:420\r\nHost: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:420", headers["Host"])
	assert.Equal(t, 21, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Perp AI 추가 3개
	// Test: Header line에 colon이 없는 경우 (에러 발생)
	headers = NewHeaders()
	data = []byte("Host localhost42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMissingColon)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header name이 아예 없는 경우 (colon이 첫번째 위치, 에러 발생)
	headers = NewHeaders()
	data = []byte(": localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMissingName)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header name과 colon 사이에 탭(\t) 문자가 있는 경우 (에러 발생)
	headers = NewHeaders()
	data = []byte("Host\t: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
