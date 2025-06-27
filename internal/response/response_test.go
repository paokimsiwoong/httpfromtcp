package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWriter struct {
	data string
}

// pointer receiver => testWriter가 아니라 *testWriter가 io.Writer 구현
func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.data += string(p)
	return len(p), nil
}

func TestStatusLineWrite(t *testing.T) {
	//  Test: StatusOK
	writer := &testWriter{
		data: "",
	}
	err := WriteStatusLine(writer, StatusOK)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 200 OK\r\n", writer.data)

	//  Test: StatusBadRequest
	writer = &testWriter{
		data: "",
	}
	err = WriteStatusLine(writer, StatusBadRequest)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 400 Bad Request\r\n", writer.data)

	//  Test: StatusInternalServerError
	writer = &testWriter{
		data: "",
	}
	err = WriteStatusLine(writer, StatusInternalServerError)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 500 Internal Server Error\r\n", writer.data)

	//  Test: Unknown Status Code
	writer = &testWriter{
		data: "",
	}
	err = WriteStatusLine(writer, 100)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 100 \r\n", writer.data)

}
