package response

// type testWriter struct {
// 	data string
// }

// // pointer receiver => testWriter가 아니라 *testWriter가 io.Writer 구현
// func (tw *testWriter) Write(p []byte) (n int, err error) {
// 	tw.data += string(p)
// 	return len(p), nil
// }

// func TestStatusLineWrite(t *testing.T) {
// 	//  Test: StatusOK
// 	writer := &testWriter{
// 		data: "",
// 	}
// 	err := WriteStatusLine(writer, StatusOK)
// 	require.NoError(t, err)
// 	assert.Equal(t, "HTTP/1.1 200 OK\r\n", writer.data)

// 	//  Test: StatusBadRequest
// 	writer = &testWriter{
// 		data: "",
// 	}
// 	err = WriteStatusLine(writer, StatusBadRequest)
// 	require.NoError(t, err)
// 	assert.Equal(t, "HTTP/1.1 400 Bad Request\r\n", writer.data)

// 	//  Test: StatusInternalServerError
// 	writer = &testWriter{
// 		data: "",
// 	}
// 	err = WriteStatusLine(writer, StatusInternalServerError)
// 	require.NoError(t, err)
// 	assert.Equal(t, "HTTP/1.1 500 Internal Server Error\r\n", writer.data)

// 	//  Test: Unknown Status Code
// 	writer = &testWriter{
// 		data: "",
// 	}
// 	err = WriteStatusLine(writer, 100)
// 	require.NoError(t, err)
// 	assert.Equal(t, "HTTP/1.1 100 \r\n", writer.data)

// 	// Test: Default Response
// 	writer = &testWriter{
// 		data: "",
// 	}
// 	_ = WriteStatusLine(writer, StatusOK)
// 	headers := GetDefaultHeaders(0)
// 	require.NotNil(t, headers)
// 	err = WriteHeaders(writer, headers)
// 	require.NoError(t, err)
// 	assert.Equal(t, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: close\r\nContent-Type: text/plain\r\n\r\n", writer.data)

// 	// Test: Default Response with non-zero content-length
// 	writer = &testWriter{
// 		data: "",
// 	}
// 	_ = WriteStatusLine(writer, StatusOK)
// 	headers = GetDefaultHeaders(100)
// 	require.NotNil(t, headers)
// 	err = WriteHeaders(writer, headers)
// 	require.NoError(t, err)
// 	assert.Equal(t, "HTTP/1.1 200 OK\r\nContent-Length: 100\r\nConnection: close\r\nContent-Type: text/plain\r\n\r\n", writer.data)

// }
