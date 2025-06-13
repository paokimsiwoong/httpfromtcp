package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// @@@ net.ResolveUDPAddr에 바로 문자열 입력하는 대신 따로 변수로 만들어 두면 뒤에서 또 사용 가능
	serverAddr := "localhost:42069"

	// net.ResolveUDPAddr 함수는
	// 문자열로 된 주소(예: "localhost:6000" 또는 "192.168.1.10:53")를 UDP 네트워크에서 사용할 수 있는 *net.UDPAddr 타입으로 변환
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	// 함수 첫번째 인자 network에는 "udp"(IPv4, IPv6 모두 허용), "udp4"(IPv4 전용), "udp6"(IPv6)전용 3 중 하나 선택해 입력
	// 두번쨰 인자 address에는 *net.UDPAddr로 변환할 문자열 주소 입력
	if err != nil {
		log.Fatalf("error creating udp address: %v", err)
		// @@@ solution 예시
		// fmt.Fprintf는 인자로 주어진 io.Writer에 fomatted string을 쓰는 함수
		// fmt.Fprintf(os.Stderr, "Error resolving UDP address: %v\n", err)

		// exit code 1과 함께 프로그램 종료
		// os.Exit(1)
		// @@@ solution 예시
	}

	// net.DialUDP는 지정한 로컬 주소(laddr)와 원격 주소(raddr)로 UDP 소켓을 생성해,
	// 해당 주소와만 데이터를 주고받을 수 있도록 설정
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	// network인자는 ResolveUDPAddr와 동일
	// laddr은 nil로 두면 자동할당
	// raddr은 통신할 상대방 주소 입력
	if err != nil {
		log.Fatalf("error creating udp connection: %v", err)
	}
	// close defer하기
	defer udpConn.Close()

	// @@@ solution처럼 간단 사용법 프린트하기
	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", serverAddr)

	// bufio는 버퍼가 추가된 io 패키지
	// bufio.NewReader는 io.Reader를 입력받아 버퍼가 추가된 bufio.Reader로 래핑하는 함수
	// ==> 내부적으로 데이터를 직접 읽지 않고 버퍼에 모아두었다가 한번에 여러 바이트를 읽을 수 있다
	reader := bufio.NewReader(os.Stdin)

	for {
		// 사용자 입력을 기다리는걸 알리는 표시는 >
		fmt.Print(">")

		line, err := reader.ReadString('\n')
		// '\n'은 단일 문자 rune, "\n"은 string => ReadString은 단일문자 입력을 기대하므로
		// '' 사용 (rune(int32)은 byte로 변환 가능)
		if err != nil {
			log.Fatalf("error reading stdin: %v", err)
		}

		_, err = udpConn.Write([]byte(line))
		if err != nil {
			log.Fatalf("error writing line to the connection: %v", err)
		}

		// @@@ solution처럼 결과 프린트
		fmt.Printf("Message sent: %s", line)
	}
}
