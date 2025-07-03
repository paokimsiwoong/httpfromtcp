package response

import (
	"errors"
	"fmt"

	"github.com/paokimsiwoong/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	WriterStateInitialized    writerState = 0
	WriterStateStatusLineDone writerState = 1
	WriterStateHeadersDone    writerState = 2
	WriterStateDone           writerState = 3
)

var ErrWriterInvalidState = errors.New("you must call the struct's methods in the correct order")

type Writer struct {
	Data  []byte
	State writerState
}

// Status Line을 주어진 statusCode에 맞게 Writer 구조체에 저장하는 메소드
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != WriterStateInitialized {
		return ErrWriterInvalidState
	}

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

	w.Data = append(w.Data, []byte(line)...)

	w.State = WriterStateStatusLineDone

	return nil
}

// headers에 저장되어 있는 헤더들을 Writer 구조체에 저장하는 메소드
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != WriterStateStatusLineDone {
		return ErrWriterInvalidState
	}

	for key, value := range headers {
		header := ""

		header += key

		header += ": "

		header += value + "\r\n"

		w.Data = append(w.Data, []byte(header)...)
	}

	// headers 맵 순회가 끝나면 헤더 블록이 끝났다고 알리는 \r\n를 마지막으로 쓰고 종료
	w.Data = append(w.Data, []byte("\r\n")...)

	w.State = WriterStateHeadersDone

	return nil
}

// 주어진 body 데이터를 Writer 구조체에 저장하는 메소드
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != WriterStateHeadersDone {
		return 0, ErrWriterInvalidState
	}

	w.Data = append(w.Data, p...)

	w.State = WriterStateDone

	return len(p), nil
}

// chunk 데이터 길이와 데이터 자체를 Writer에 저장하는 함수
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.State != WriterStateHeadersDone {
		return 0, ErrWriterInvalidState
	}

	// chunk 길이는 16진법으로 표현 (%x 이용)
	chunkLen := []byte(fmt.Sprintf("%x", len(p)) + "\r\n")

	// <n>\r\n 부분 입력
	w.Data = append(w.Data, chunkLen...)
	// <data of length n>\r\n 부분 입력
	w.Data = append(w.Data, p...)
	w.Data = append(w.Data, []byte("\r\n")...)

	return len(chunkLen) + len(p) + 2, nil
	// w.Data에 추가되는 바이트 길이는 len(chunkLen) + len(p) + len("\r\n")
}

// chunked encoding이 끝났음을 알리는 마지막줄을 Writer에 저장하는 함수
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.State != WriterStateHeadersDone {
		return 0, ErrWriterInvalidState
	}

	lastChunk := []byte(fmt.Sprintf("%x", 0) + "\r\n\r\n")

	w.Data = append(w.Data, lastChunk...)

	w.State = WriterStateDone

	return len(lastChunk), nil
}

// @@@ 구조 변경 @@@
// Status Line을 주어진 statusCode에 맞게 io.Writer에 작성하는 함수
// func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
// 	line := ""
// 	switch statusCode {
// 	case StatusOK:
// 		line = "HTTP/1.1 200 OK\r\n"
// 	case StatusBadRequest:
// 		line = "HTTP/1.1 400 Bad Request\r\n"
// 	case StatusInternalServerError:
// 		line = "HTTP/1.1 500 Internal Server Error\r\n"
// 	default:
// 		line = fmt.Sprintf("HTTP/1.1 %v \r\n", statusCode)
// 	}

// 	_, err := w.Write([]byte(line))
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // 기본 response 헤더를 작성하는 함수
// func GetDefaultHeaders(contentLen int) headers.Headers {
// 	headerMap := headers.NewHeaders()
// 	// headerMap["content-length"] = strconv.Itoa(contentLen)
// 	// headerMap["connection"] = "close"
// 	// headerMap["content-type"] = "text/plain"
// 	headerMap["Content-Length"] = strconv.Itoa(contentLen)
// 	headerMap["Connection"] = "close"
// 	headerMap["Content-Type"] = "text/plain"

// 	return headerMap
// }

// // headers에 저장되어 있는 헤더들을 w에 쓰는 함수
// func WriteHeaders(w io.Writer, headers headers.Headers) error {
// 	for key, value := range headers {
// 		header := ""

// 		// 소문자로만 되어 있는 헤더 이름 첫글자 대문자로 변경?
// 		// split := strings.Split(key, "-")

// 		// for _, k := range split {
// 		// 	header += strings.ToUpper(string(k[0]))
// 		// 	header += k[1:]
// 		// 	header += "-"
// 		// }

// 		// header = header[:len(header)-1]

// 		header += key

// 		header += ": "

// 		header += value + "\r\n"

// 		_, err := w.Write([]byte(header))
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// headers 맵 순회가 끝나면 헤더 블록이 끝났다고 알리는 \r\n를 마지막으로 쓰고 종료
// 	_, err := w.Write([]byte("\r\n"))
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
