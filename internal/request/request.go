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
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// 테스트에서 require.ErrorIs(errors.Is의 wrapper)를 사용할 수 있도록 에러 변수 선언
var ErrInvalidRequestLine = errors.New("request line must contain exactly three parts: method, request-target, HTTP-version")
var ErrInvalidMethod = errors.New("request method must be capital alphabetic characters")
var ErrInvalidVersion = errors.New("HTTP-version must be HTTP/1.1")

// io.Reader를 받아 HTTP request를 파싱하는 함수
func RequestFromReader(reader io.Reader) (*Request, error) {
	// 파싱 완료된 데이터를 담을 구조체 선언
	req := Request{}

	// 첫단계에서는 전체 데이터를 한꺼번에 메모리에 올려도 문제 없으므로
	// io.ReadAll 사용
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading io reader: %w", err)
	}

	err = parseRequestLine(string(raw), &req)
	if err != nil {
		return nil, fmt.Errorf("error parsing string: %w", err)
	}

	return &req, nil
}

func parseRequestLine(raw string, req *Request) error {
	// HTTP는 각 줄 구분을 CRLF(\r\n)으로 한다
	lines := strings.Split(raw, "\r\n")

	// 첫단계에서는 request-line만 처리
	// request-line은 공백 한칸으로 3 파트 분리
	reqLineParts := strings.Split(lines[0], " ")
	// @@@ 퍼플렉시티 추천은 reqLineParts := strings.Fields(lines[0]) ==> 이러면 공백이 복수개거나 \t인 경우도 3개 파트로 나눠진다
	if len(reqLineParts) != 3 {
		return ErrInvalidRequestLine
	}

	if !isUpper(reqLineParts[0]) {
		return ErrInvalidMethod
	}
	// req 구조체에 method 입력
	req.RequestLine.Method = reqLineParts[0]

	// 일단 HTTP/1.1만 지원
	if reqLineParts[2] != "HTTP/1.1" {
		return ErrInvalidVersion
	}
	// req.RequestLine.HttpVersion 에는 HTTP/ 부분 없이 숫자 버전만 입력
	version := strings.Split(reqLineParts[2], "/")[1]
	// req 구조체에 http 버전 입력
	req.RequestLine.HttpVersion = version
	// req 구조체에 request-target 입력
	req.RequestLine.RequestTarget = reqLineParts[1]

	return nil
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
