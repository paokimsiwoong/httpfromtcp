package main

import (
	"fmt"
	"log"
	"net"

	"github.com/paokimsiwoong/httpfromtcp/internal/request"
)

// @@@ 읽어오는 파일 경로 const로 저장하기
// const inputFilePath = "messages.txt"
// @@@ 이젠 파일대신 port 저장
const port = ":42069"

// // @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
// // Go의 net.Listen 함수에서 두 번째 인자(주소)에 :포트번호 형식만 입력하면, 이는 모든 네트워크 인터페이스(0.0.0.0)의 해당 포트에서 연결을 수신(listen)하겠다는 의미입니다.
// // 예를 들어, net.Listen("tcp", ":8000")로 작성하면, 서버는 0.0.0.0:8000(즉, 모든 IP 주소의 8000번 포트)에서 접속을 기다리게 됩니다.
// // 반면, localhost:8000이나 127.0.0.1:8000을 사용하면 오직 루프백 인터페이스(자기 자신)에서만 연결을 수신합니다.
// // 즉, 외부에서 접근하려면 :8000처럼 IP를 생략하는 것이 필요합니다.
// // @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

func main() {

	// 파일 열기
	// file, err := os.Open(inputFilePath)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// @@@ file을 열었으면 반드시 defer로 close 걸기 @@@
	// defer file.Close()
	// @@@ 이제는 getLinesChannel 함수 안의 go 루틴(파일 읽는 루프)가 끝날 때 Close 실행

	// @@@ file 읽기 대신 TCP listener 설정
	// listener, err := net.Listen("tcp", "localhost:42069")
	// net.Listen(네트워크종류, 주소(ip주소:port 형태 => local일때는 localhost:port 사용))
	listener, err := net.Listen("tcp", port)
	// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
	// Perp AI : Go의 net.Listen 함수에서 두 번째 인자(주소)에 :포트번호 형식만 입력하면, 이는 모든 네트워크 인터페이스(0.0.0.0)의 해당 포트에서 연결을 수신(listen)하겠다는 의미입니다.
	// 예를 들어, net.Listen("tcp", ":8000")로 작성하면, 서버는 0.0.0.0:8000(즉, 모든 IP 주소의 8000번 포트)에서 접속을 기다리게 됩니다.
	// 반면, localhost:8000이나 127.0.0.1:8000을 사용하면 오직 루프백 인터페이스(자기 자신)에서만 연결을 수신합니다.
	// 즉, 외부에서 접근하려면 :8000처럼 IP를 생략하는 것이 필요합니다.
	// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
	if err != nil {
		log.Fatalf("error creating listner: %v", err)
	}
	// @@@ net.Listener 인터페이스도 구현 조건에 close 함수 있음
	defer listener.Close()

	// listener 무한 for 루프
	for {
		// net.Conn 인터페이스(generic stream-oriented network connection)는 네트워크 연결을 추상화한 인터페이스
		// TCP, UDP 등 다양한 네트워크 프로토콜의 연결에서 연결된 상대방과 데이터를 송수신할 수 있는
		// Read, Write, Close등의 메소드들을 구현조건으로 가진다
		curConn, err := listener.Accept()
		// net.Conn은 io.ReadCloser를 구현
		if err != nil {
			log.Fatalf("error accepting connection: %v", err)
		}
		fmt.Print("A connection from ", curConn.RemoteAddr(), " has been accepted\n")
		// @@@ solution: net.Conn의 RemoteAddr 메소드 활용하기
		// fmt.Println("Accepted connection from", curConn.RemoteAddr())

		req, err := request.RequestFromReader(curConn)
		if err != nil {
			log.Fatalf("error parsing request line: %v", err)
		}

		// Request line 출력
		fmt.Printf(
			`Request line:
- Method: %s
- Target: %s
- Version: %s
`,
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion,
		)

		// Headers 출력
		fmt.Println("Headers:")

		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		// Body 출력
		fmt.Println("Body:")

		fmt.Println(string(req.Body))

		// lineCh := getLinesChannel(curConn)

		// for line := range lineCh { // range 채널 for 루프는 채널이 close 될 때 종료 된다
		// 	fmt.Printf("read: %s\n", line)
		// 	// @@@ solution처럼 Println을 써도 됨
		// 	// fmt.Println("read:", line)
		// }

		// 사용완료한 종료
		curConn.Close()
		fmt.Print("the connection ", curConn.RemoteAddr(), " has been closed\n")
		// @@@ solution: net.Conn의 RemoteAddr 메소드 활용하기
		// fmt.Println("Connection to ", curConn.RemoteAddr(), "closed")
	}

}

// @@@ internal/request/request.go의 RequestFromReader 함수 대신 사용
// // io.ReadCloser 타입을 받아 저장된 line을 읽어 chan에 입력하고 그 채널을 반환하는 함수
// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	// io.ReadCloser는 Reader와 Closer 인터페이스 구현하는 인터페이스
// 	// os.File은 io.ReadCloser를 구현한다

// 	// line string을 한줄 씩 받고 내보내는 채널 생성
// 	lineCh := make(chan string)

// 	// // 파일 열기
// 	// file, err := os.Open(inputFilePath)
// 	// if err != nil {
// 	// 	log.Fatalf("error opening file: %v", err)
// 	// }
// 	// // @@@ file을 열었으면 반드시 defer로 close 걸기 @@@
// 	// defer file.Close()
// 	// 함수 인자로 os.File(io.ReadCloser)를 받으므로 직접 os.Open을 할 필요가 없다
// 	// // @@@ defer f.Close()를 go 루틴 바깥에서 해버리면
// 	// // @@@ go 루틴이 파일을 읽으며 돌아가기도 전에 함수가 반환하면서 파일이 close되어 문제가 생긴다

// 	raw := make([]byte, 8)

// 	// 8 bytes 덩어리들이 한 줄이 될때까지 저장하는 변수
// 	var line string

// 	// 파일을 읽는 for loop는 go routine으로
// 	go func() {
// 		// @@@ solution처럼 defer 사용하기 <= func() {} anonymous 함수이므로 defer 사용 가능
// 		defer f.Close()
// 		defer close(lineCh)
// 		// @@@ defer는 LIFO이므로 채널 close가 먼저

// 		for {
// 			n, err := f.Read(raw)
// 			if err != nil {
// 				if err == io.EOF { // io.EOF는 파일 읽기가 문제 없이 종료 되었을 때 Read의 err 반환값
// 					// if errors.Is(err, io.EOF) { 형태도 가능
// 					break
// 				}
// 				log.Fatalf("error reading file: %v", err)
// 			}

// 			// []byte인 raw를 string으로 변환
// 			str := string(raw[:n])
// 			strSlice := strings.Split(str, "\n")

// 			for _, part := range strSlice[:(len(strSlice) - 1)] {
// 				line += part
// 				// fmt.Printf("read: %s\n", line)
// 				// 채널에 완성된 한줄을 송신
// 				lineCh <- line
// 				line = ""
// 			}

// 			// fmt.Printf("read: %s\n", string(raw[:n]))

// 			line += strSlice[len(strSlice)-1]
// 		}

// 		// @@@ line이 길이가 0인 경우 제외하기 위해 if 추가
// 		if line != "" {
// 			// fmt.Printf("read: %s\n", line)
// 			// solution은 마지막 남은 line을 for 루프 안 Read 함수 에러 처리 부분에서 진행함
// 			// // => 마지막 읽기 실행 후 err == io.EOF로 if err != nil 블록이 실행되기 때문

// 			lineCh <- line
// 			line = ""
// 		}

// 		// @@@ solution처럼 defer 사용하기
// 		// 파일을 다 읽고나면 채널 close
// 		// close(lineCh)
// 		// // @@@ 함수 바깥에서 for line := range lineCh로 for 루프를 도는데
// 		// // @@@ range 채널 for 루프는 채널이 close 될 때 종료 되므로 for 루프가 정상적으로 종료될 수 있도록 반드시 채널을 close 해준다

// 		// 읽기가 끝난 os.File Close
// 		// f.Close()
// 		// @@@ defer f.Close()를 go 루틴 바깥에서 해버리면
// 		// @@@ go 루틴이 파일을 읽으며 돌아가기도 전에 함수가 반환하면서 파일이 close되어 문제가 생긴다

// 		// @@@ solution처럼 defer 사용하기
// 	}()

// 	return lineCh
// }
