package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	// State 가 0이면 초기화 상태
	// 1이면 파싱 처리 완료
	State int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// 테스트에서 require.ErrorIs(errors.Is의 wrapper)를 사용할 수 있도록 에러 변수 선언
var ErrEmptyReader = errors.New("reader does not contain any data")
var ErrNotParsed = errors.New("nothing has been parsed")
var ErrInvalidState = errors.New("parsing state is 1: it has already been done")
var ErrUnknownState = errors.New("parsing state is not 0 or 1: unknown state")
var ErrInvalidRequestLine = errors.New("request line must contain exactly three parts: method, request-target, HTTP-version")
var ErrInvalidMethod = errors.New("request method must be capital alphabetic characters")
var ErrInvalidVersion = errors.New("HTTP-version must be HTTP/1.1")

const bufferSize = 8

// @@@ 예시 따라서 Request의 State 필드에 들어갈 값 const 지정
const requestStateInitialized = 0
const requestStateDone = 1

// @@@ 예시 따라서 crlf도 const 지정
const crlf = "\r\n"

// io.Reader를 받아 HTTP request를 파싱하는 함수
func RequestFromReader(reader io.Reader) (*Request, error) {
	// 파싱 완료된 데이터를 담을 구조체 선언
	req := Request{State: requestStateInitialized}

	// 첫단계에서는 전체 데이터를 한꺼번에 메모리에 올려도 문제 없으므로
	// io.ReadAll 사용
	// raw, err := io.ReadAll(reader)
	// if err != nil {
	// 	return nil, fmt.Errorf("error reading io reader: %w", err)
	// }
	// @@@ 이제는 조각조각 데이터가 들어와도 처리 가능하도록 변경
	buffer := make([]byte, 0, bufferSize)
	// make([]byte, 8) 이렇게만 두면 len 8, cap 8로 이미 8개의 0이 들어있는 취급이라
	// 뒤에 buffer = append(buffer, chunk...)를 하면 8개의 0이 대체되는 것이 아니라
	// 그 0 뒤에 chunk의 데이터가 추가된다
	bytesRead := 0
	bytesParsed := 0

	for req.State != requestStateDone { // req.State == 1 즉 파싱 완료가 되기 전까지 루프 반복
		chunk := make([]byte, 8)
		// @@@@@ 과제 tips에서는 chunk를 따로 만들지 않고
		// @@@@@ reader.Read(buffer[bytesRead:])
		// @@@@@ 그대신 Read하기 전에 buffer가 가득찬지 확인하고 가득찬 경우
		// @@@@@ 크키가 2배인 새 버퍼를 만들고 거기에 구 buffer를 복사한다
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
		// 일반적으로는 const bufferSize = 8 와 같이 버퍼 크기를 매우 작은 값으로 두기 보단
		// 1024, 4096과 같은 값을 사용하지만
		// 여기서는 1~3바이트 조각으로 들어오는 테스트 케이스들을 테스트 하기 위해 작은 값으로 둔 상태
		// 일반적인 케이스처럼 버퍼 크기를 키울 경우 매번 chunk를 생성하기보다는
		// 과제 tips처럼 Read에 buffer를 넣는 방식으로 바꾸는게 좋아 보임
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

		n, err := reader.Read(chunk)
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Read가 끝났을 때 행하는 코드들

				// @@@ n!=0, io.EOF이 반환되는 경우가 현재는 없다?
				// n이 0이 아닐 경우
				// if n != 0 {
				// 	// 현재까지 읽은 바이트 길이 기록
				// 	bytesRead += n

				// 	// 새로 읽은 부분을 buffer에 추가
				// 	buffer = append(buffer, chunk[:n]...)
				// 	// chunk 슬라이스 뒤에 ...을 붙여서 unpack한 뒤에 append에 입력
				// 	// @@@ chunk안의 유효 데이터만 buffer에 붙일 수 있도록 슬라이싱 [:n] 필요

				// 	m, err := req.parse(buffer)
				// 	if err != nil {
				// 		if errors.Is(err, ErrInvalidState) {
				// 			return &req, fmt.Errorf("error trying to read data in a done state: %w", err)
				// 		}
				// 		return nil, fmt.Errorf("error parsing buffer: %w", err)
				// 	}
				// 	// 현재까지 파싱한 바이트 길이 기록
				// 	if m != 0 {
				// 		bytesParsed += m
				// 		// 파싱 완료된 부분들은 buffer에서 필요 없음
				// 		oldBuffer := buffer
				// 		buffer = make([]byte, len(oldBuffer)-m)
				// 		_ = copy(buffer, oldBuffer[m:])
				// 	}
				// }

				// reader에 들어있는 데이터가 없는 경우
				if bytesRead == 0 {
					return nil, ErrEmptyReader
				}
				// reader를 다 읽었는데도 파싱된 데이터가 없는 경우
				if bytesParsed == 0 {
					return nil, ErrNotParsed
				}

				req.State = requestStateDone
				break
			}
			return nil, fmt.Errorf("error reading io reader: %w", err)
		}
		// 현재까지 읽은 바이트 길이 기록
		bytesRead += n

		// 새로 읽은 부분을 buffer에 추가
		buffer = append(buffer, chunk[:n]...)
		// chunk 슬라이스 뒤에 ...을 붙여서 unpack한 뒤에 append에 입력
		// @@@ chunk안의 유효 데이터만 buffer에 붙일 수 있도록 슬라이싱 [:n] 필요

		m, err := req.parse(buffer)
		if err != nil {
			if errors.Is(err, ErrInvalidState) {
				return &req, fmt.Errorf("error trying to read data in a done state: %w", err)
			}
			return nil, fmt.Errorf("error parsing buffer: %w", err)
		}
		// 현재까지 파싱한 바이트 길이 기록
		if m != 0 {
			bytesParsed += m
			// 파싱 완료된 부분들은 buffer에서 필요 없음
			// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
			// oldBuffer := buffer
			// buffer = make([]byte, len(oldBuffer)-m)
			// _ = copy(buffer, oldBuffer[m:])
			// 이 방식은 데이터 복사 완료 후에도
			// 구 버퍼를 가리키는 변수가 남아있어
			// 메모리 회수에 불리하므로 변경하기
			// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
			newBuffer := make([]byte, len(buffer)-m)
			_ = copy(newBuffer, buffer[m:])
			buffer = newBuffer
			// @@@@@ 과제 tips 방식을 사용했을 경우에도
			// @@@@@ 위와 같이 파싱된 부분을 제외한 부분을 새로운 buffer를 생성해 복사하고
			// @@@@@ bytesRead에서 파싱된 길이만큼을 빼준다
		}
	}

	return &req, nil
}

// request의 state에 따라 파싱을 진행할지 안할지 결정하는 함수
func (r *Request) parse(data []byte) (int, error) {
	if r.State != requestStateInitialized {
		if r.State == requestStateDone {
			return 0, ErrInvalidState
		}
		return 0, ErrUnknownState
	}

	n, err := parseRequestLine(string(data), r)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		// data에 아직 더 추가되어야 파싱 가능하다고
		// 0, nil로 알리기
		return 0, nil
	}

	// 파싱 완료 state로 변경
	// @@@ 지금은 request line만 입력되면 1로 변경해 뒷부분이 읽히지 않고 버려진다
	r.State = requestStateDone
	return n, nil
}

// raw 스트링을 받아서 그 안의 request line을 찾아내는 함수
func parseRequestLine(raw string, req *Request) (int, error) {
	// crlf("\r\n")이 포함되어 있지않으면 chunk를 더 읽어서 raw에 붙인 후 다시 이 함수를 실행하도록 일단 반환
	if !strings.Contains(raw, crlf) {
		return 0, nil
		// 단순히 chunk를 더 읽고 다시 이 함수를 호출해야한다고 알 수 있도록
		// 0, nil 반환
		// 그 외의 경우는 raw 스트링의 길이를 반환
	}

	// HTTP는 각 줄 구분을 CRLF("\r\n")으로 한다
	lines := strings.Split(raw, crlf)

	// 이 함수는 request-line만 처리
	// request-line은 공백 한칸으로 3 파트 분리
	reqLineParts := strings.Split(lines[0], " ")
	// @@@ 퍼플렉시티 추천은 reqLineParts := strings.Fields(lines[0]) ==> 이러면 공백이 복수개거나 \t인 경우도 3개 파트로 나눠진다
	if len(reqLineParts) != 3 {
		return 0, ErrInvalidRequestLine
	}

	if !isUpper(reqLineParts[0]) {
		return 0, ErrInvalidMethod
	}
	// req 구조체에 method 입력
	req.RequestLine.Method = reqLineParts[0]

	// 일단 HTTP/1.1만 지원
	if reqLineParts[2] != "HTTP/1.1" {
		return 0, ErrInvalidVersion
	}
	// req.RequestLine.HttpVersion 에는 HTTP/ 부분 없이 숫자 버전만 입력
	version := strings.Split(reqLineParts[2], "/")[1]
	// req 구조체에 http 버전 입력
	req.RequestLine.HttpVersion = version
	// req 구조체에 request-target 입력
	req.RequestLine.RequestTarget = reqLineParts[1]

	return len(raw), nil
}

// unicode.IsUpper를 이용해 입력된 string이 대문자로만 이루어져있는지 확인하는 함수
//
// @@@ unicode.IsUpper는 rune(한글자)만 확인하는 함수
func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
