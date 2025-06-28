package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/paokimsiwoong/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// Status Line을 주어진 statusCode에 맞게 io.Writer에 작성하는 함수
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

// 기본 response 헤더를 작성하는 함수
func GetDefaultHeaders(contentLen int) headers.Headers {
	headerMap := headers.NewHeaders()
	// headerMap["content-length"] = strconv.Itoa(contentLen)
	// headerMap["connection"] = "close"
	// headerMap["content-type"] = "text/plain"
	headerMap["Content-Length"] = strconv.Itoa(contentLen)
	headerMap["Connection"] = "close"
	headerMap["Content-Type"] = "text/plain"

	return headerMap
}

// headers에 저장되어 있는 헤더들을 w에 쓰는 함수
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		header := ""

		// 소문자로만 되어 있는 헤더 이름 첫글자 대문자로 변경?
		// split := strings.Split(key, "-")

		// for _, k := range split {
		// 	header += strings.ToUpper(string(k[0]))
		// 	header += k[1:]
		// 	header += "-"
		// }

		// header = header[:len(header)-1]

		header += key

		header += ": "

		header += value + "\r\n"

		_, err := w.Write([]byte(header))
		if err != nil {
			return err
		}
	}

	// headers 맵 순회가 끝나면 헤더 블록이 끝났다고 알리는 \r\n를 마지막으로 쓰고 종료
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
