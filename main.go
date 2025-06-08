package main

import (
	"fmt"
	"io"
	"log"
	"os"
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
		fmt.Printf("read: %s\n", string(raw[:n]))
	}
}
