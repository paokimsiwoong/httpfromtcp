package headers

import (
	"errors"
	"strings"
	"unicode"
)

// 파싱된 HTTP headers 저장하는 맵 타입 정의
type Headers map[string]string

// CRLF
const crlf = "\r\n"

// 에러 변수 선언
var ErrMissingColon = errors.New("header line must contain colon")
var ErrMissingName = errors.New("header line must contain header name")
var ErrInvalidName = errors.New("there must be no spaces betweern colon and header name")

// var ErrMultipleColon = errors.New("there must be one and only one colon")
// @@@ Host: localhost:42069\r\n 와 같이 값에 :가 또 들어갈 수도 있다

// Headers 맵 인스턴스 생성하는 함수
func NewHeaders() Headers {
	// headers := make(Headers, 8)
	// capacity 지정은 예상되는 엔트리 수가 많고 그 수를 대충 예상할 수 있을 때 지정해서
	// 맵 크기를 키울 때마다 발생하는 재할당, 재해싱 비용을 줄여 최적화할때 한다
	headers := make(Headers)
	return headers
}

// HTTP header 파싱 메소드
// @@@ golang에서 구조체뿐 아니라 임의의 사용자 정의 타입에 다 메소드를 붙일 수 있다
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// @@@ h는 pass by value로 원본 맵의 복사본이지만
	// @@@ 고에서 map은 실제 데이터가 위치한 메모리 주소를 저장하는 참조타입
	// // @@@ ==> h를 통해 맵 내부의 데이터를 변경하면 원본 맵이 가리키는 데이터들도 변함
	// // // @@@ slice, map, channel, function, interface : 참조 타입 (referece type)

	strData := string(data)
	// CRLF를 포함하고 있는지 확인
	if !strings.Contains(strData, crlf) {
		// 없으면 데이터를 더 읽은 후 다시 이 함수를 호출하도록 알리는 내용을 담아 반환 (0 바이트 파싱됨, 파싱 미완효, 에러 nil)
		return 0, false, nil
	}

	// CRLF로 시작하는지 확인 (헤더 라인들이 끝날 때 \r\n 두번 반복)
	if strData[:2] == crlf {
		// 헤더 라인 파싱이 끝났다고 알림
		return 0, true, nil
	}

	// 헤더 라인 분리
	line := strings.Split(strData, crlf)[0]
	// :의 인덱스 찾기
	colonIdx := strings.Index(line, ":")
	// :이 없으면 에러
	if colonIdx == -1 {
		return 0, false, ErrMissingColon
	}
	// 헤더 네임이 없으면 에러
	if colonIdx == 0 {
		return 0, false, ErrMissingName
	}
	// 헤더 이름과 : 사이에는 공백이 있으면 안된다
	// if line[colonIdx-1] == " " { @@@ 스트링을 인덱스로 접근해서 반환받은 값은 byte 타입이므로 " "(string)과 ==으로 비교하면 에러 발생
	// if line[colonIdx-1] == ' ' { @@@ ' '로 바이트 타입은 일치하지만 공백에 해당하는 \t와 같은 케이스도 커버하도록 변경
	if unicode.IsSpace(rune(line[colonIdx-1])) {
		// unicode.IsSpace(r)는 유니코드에서 공백으로 정의된 모든 문자를 판별
		return 0, false, ErrInvalidName
	}
	// 공백 제거
	trimmed := strings.ReplaceAll(line, " ", "")
	// @@@ 공백이 \t와 같이 다른 것일 경우 어떻게 처리할지?

	// @@@ colon : 처리 방식 변경
	// @@@ Host: localhost:42069\r\n 와 같이 값에 :가 또 들어갈 수도 있다
	// 헤더 이름과 값으로 분리
	// nameAndValue := strings.Split(trimmed, ":")

	// // :이 복수개 들어간 경우 에러
	// if len(nameAndValue) != 2 {
	// 	return 0, false, ErrMultipleColon
	// }

	// 헤더 맵에 입력
	// h[nameAndValue[0]] = nameAndValue[1]

	// 헤더 이름과 값을 구분하는 첫번쨰 : 인덱스 찾기
	colonIdx = strings.Index(trimmed, ":")
	// 헤더 맵에 입력
	h[trimmed[:colonIdx]] = trimmed[colonIdx+1:]

	// 파싱 완료 후 처리된 바이트 길이 반환
	return len(line) + 2, false, nil
	// @@@ 헤더 부분 + CRLF 2바이트
}
