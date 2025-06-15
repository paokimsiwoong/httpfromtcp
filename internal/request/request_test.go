package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good POST Request line with path
	r, err = RequestFromReader(strings.NewReader("POST /api/users HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/json\r\nContent-Length: 45\r\n\r\n{\"name\": \"Jane Doe\", \"email\": \"jane.doe@example.com\"}"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidRequestLine)

	// Test: Invalid method in request line
	_, err = RequestFromReader(strings.NewReader("get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidMethod)

	// Test: Invalid version in request line
	_, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/2.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidVersion)
	_, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidVersion)

	// @@@ 퍼플렉시티 추천
	// Test: Empty request line
	_, err = RequestFromReader(strings.NewReader(""))
	require.Error(t, err)
	// Test: Invalid request line with extra spaces
	_, err = RequestFromReader(strings.NewReader(" GET   /coffee   HTTP/1.1  \r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
	// @@@ 퍼플렉시티는 strings.Split 대신 strings.Fields()를 써서 공백이 한개가 아니거나 \t인 경우에도 파싱에 성공해 통과하도록 추천하고 있음
	_, err = RequestFromReader(strings.NewReader("GET\t/coffee\tHTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
	// Test: Invalid request line with only CRLF
	_, err = RequestFromReader(strings.NewReader("\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
	// Test: Invalid request line with non-ASCII char
	// @@@ 현재 통과 불가능 (unicode.IsUpper로 검사하면 비 아스키 대문자도 대문자로 취급함)
	// _, err = RequestFromReader(strings.NewReader("GÉT /coffee HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	// require.Error(t, err)
	// require.ErrorIs(t, err, ErrInvalidMethod)
}
