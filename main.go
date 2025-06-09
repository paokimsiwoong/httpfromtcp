package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	// 파일 열기
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	// @@@ file을 열었으면 반드시 defer로 close 걸기 @@@
	defer file.Close()

	raw := make([]byte, 8)

	// n, err := file.Read(raw)

	// for err != io.EOF {
	// 	fmt.Printf("read: %s\n", string(raw[:n]))
	// 	n, err = file.Read(raw)
	// }

	// 8 bytes 덩어리들이 한 줄이 될때까지 저장하는 변수
	var line string

	// perplexity나 solution의 구조로 바꾸기
	for {
		// file.Read 코드가 한줄만 쓰이도록 for 루프 안으로 이동
		n, err := file.Read(raw)
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
			fmt.Printf("read: %s\n", line)
			line = ""
		}

		// fmt.Printf("read: %s\n", string(raw[:n]))

		line += strSlice[len(strSlice)-1]
	}

	// @@@ line이 길이가 0인 경우 제외하기 위해 if 추가
	if line != "" {
		fmt.Printf("read: %s\n", line)
		// solution은 마지막 남은 line을 for 루프 안 Read 함수 에러 처리 부분에서 진행함
		// // => 마지막 읽기 실행 후 err == io.EOF로 if err != nil 블록이 실행되기 때문
	}
}
