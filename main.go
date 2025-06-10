package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// @@@ 읽어오는 파일 경로 const로 저장하기
const inputFilePath = "messages.txt"

func main() {

	// 파일 열기
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	// @@@ file을 열었으면 반드시 defer로 close 걸기 @@@
	// defer file.Close()
	// @@@ 이제는 getLinesChannel 함수 안의 go 루틴(파일 읽는 루프)가 끝날 때 Close 실행

	lineCh := getLinesChannel(file)

	for line := range lineCh { // range 채널 for 루프는 채널이 close 될 때 종료 된다
		fmt.Printf("read: %s\n", line)
		// @@@ solution처럼 Println을 써도 됨
		// fmt.Println("read:", line)
	}
}

// line을 읽는 함수
func getLinesChannel(f io.ReadCloser) <-chan string {
	// io.ReadCloser는 Reader와 Closer 인터페이스 구현하는 인터페이스
	// os.File은 io.ReadCloser를 구현한다

	// line string을 한줄 씩 받고 내보내는 채널 생성
	lineCh := make(chan string)

	// // 파일 열기
	// file, err := os.Open(inputFilePath)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// // @@@ file을 열었으면 반드시 defer로 close 걸기 @@@
	// defer file.Close()
	// 함수 인자로 os.File(io.ReadCloser)를 받으므로 직접 os.Open을 할 필요가 없다
	// // @@@ defer f.Close()를 go 루틴 바깥에서 해버리면
	// // @@@ go 루틴이 파일을 읽으며 돌아가기도 전에 함수가 반환하면서 파일이 close되어 문제가 생긴다

	raw := make([]byte, 8)

	// 8 bytes 덩어리들이 한 줄이 될때까지 저장하는 변수
	var line string

	// 파일을 읽는 for loop는 go routine으로
	go func() {
		// @@@ solution처럼 defer 사용하기 <= func() {} anonymous 함수이므로 defer 사용 가능
		defer f.Close()
		defer close(lineCh)
		// @@@ defer는 LIFO이므로 채널 close가 먼저

		for {
			n, err := f.Read(raw)
			if err != nil {
				if err == io.EOF { // io.EOF는 파일 읽기가 문제 없이 종료 되었을 때 Read의 err 반환값
					// if errors.Is(err, io.EOF) { 형태도 가능
					break
				}
				log.Fatalf("error reading file: %v", err)
			}

			// []byte인 raw를 string으로 변환
			str := string(raw[:n])
			strSlice := strings.Split(str, "\n")

			for _, part := range strSlice[:(len(strSlice) - 1)] {
				line += part
				// fmt.Printf("read: %s\n", line)
				// 채널에 완성된 한줄을 송신
				lineCh <- line
				line = ""
			}

			// fmt.Printf("read: %s\n", string(raw[:n]))

			line += strSlice[len(strSlice)-1]
		}

		// @@@ line이 길이가 0인 경우 제외하기 위해 if 추가
		if line != "" {
			// fmt.Printf("read: %s\n", line)
			// solution은 마지막 남은 line을 for 루프 안 Read 함수 에러 처리 부분에서 진행함
			// // => 마지막 읽기 실행 후 err == io.EOF로 if err != nil 블록이 실행되기 때문

			lineCh <- line
			line = ""
		}

		// @@@ solution처럼 defer 사용하기
		// 파일을 다 읽고나면 채널 close
		// close(lineCh)
		// // @@@ 함수 바깥에서 for line := range lineCh로 for 루프를 도는데
		// // @@@ range 채널 for 루프는 채널이 close 될 때 종료 되므로 for 루프가 정상적으로 종료될 수 있도록 반드시 채널을 close 해준다

		// 읽기가 끝난 os.File Close
		// f.Close()
		// @@@ defer f.Close()를 go 루틴 바깥에서 해버리면
		// @@@ go 루틴이 파일을 읽으며 돌아가기도 전에 함수가 반환하면서 파일이 close되어 문제가 생긴다

		// @@@ solution처럼 defer 사용하기
	}()

	return lineCh
}
