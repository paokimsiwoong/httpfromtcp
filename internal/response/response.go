package response

import (
	"fmt"
	"io"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	line := ""
	switch statusCode {
	case StatusOK:
		line = "HTTP/1.1 200 OK\r\n"
	case StatusBadRequest:
		line = "HTTP/1.1 400 Bad Request\r\n"
	case StatusInternalServerError:
		line = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		line = fmt.Sprintf("HTTP/1.1 %v \r\n", statusCode)
	}

	_, err := w.Write([]byte(line))
	if err != nil {
		return err
	}

	return nil
}

// func GetDefaultHeaders(contentLen int) headers.Headers {

// }

// func WriteHeaders(w io.Writer, headers headers.Headers) error {

// }
