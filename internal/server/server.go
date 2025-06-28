package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/paokimsiwoong/httpfromtcp/internal/response"
)

type Server struct {
	port     int
	listener net.Listener
	closed   atomic.Bool
}

// Server 구조체를 초기화하고 반환하면서 Server.listen 메소드를 고 루틴으로 시작하는 함수
func Serve(port int) (*Server, error) {
	server := Server{
		port:   port,
		closed: atomic.Bool{},
	}

	// tcp listener 생성 및 Server 구조체에 저장
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, fmt.Errorf("error creating listener: %w", err)
	}
	server.listener = listener

	// go 루틴으로 listen 루프 시작
	go server.listen()

	return &server, nil
}

// tcp listener로 들어오는 연결들 처리하는 메소드
func (s *Server) listen() {

	// server가 열려 있으면 for 루프
	// for !s.closed.Load() { // @@@ 이 조건 필요 없음 (서버가 닫히면 s.listener.Accept()가 에러를 반환하므로 그 에러 처리를 이용해 for loop 종료 가능)
	for {
		curConn, err := s.listener.Accept()
		// Accept()는 blocking 함수
		// 즉, 연결이 들어오지 않으면 거기서 계속 멈춰 있다
		// // 서버를 종료할 때 s.listener.Close()를 호출하면,
		// // blocking 중이던 Accept()가 에러를 반환하면서 block 종료
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
		// if err != nil && !s.closed.Load() {
		// 	log.Fatalf("error accepting connection: %v", err)
		// }
		// 서버 종료로 인해 Accept가 에러를 낸 경우 err != nil이고 s.closed.Load()이므로
		// 이 조건 처리가 필요
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
		if err != nil {
			if s.closed.Load() {
				// 서버가 닫히면서 s.listener.Accept()가 에러를 반환한 경우 for 루프 종료
				break
			}
			// 서버가 닫히지 않았는데 에러가 발생했다면 심각한 문제이므로 로그 후 종료
			log.Fatalf("error accepting connection: %v", err)
		}
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

		// Accept 함수 에러처리 수정전에는 서버가 닫히면서 Accept함수가 에러를 발생하면
		// err != nil && s.closed.Load() 상태가 예외 처리되지 않고 이 go s.handle(curConn) 까지 와서 에러가 발생했기때문에
		// if !s.closed.Load() 가 필요했었지만
		// 이제 서버가 종료되면 위에서 break로 빠져나가므로 문제없음
		// if !s.closed.Load() {
		// 	go s.handle(curConn)
		// }
		go s.handle(curConn)
	}
}

// net.Conn을 받아서 response를 하는 메소드
func (s *Server) handle(conn net.Conn) {
	// connection 종료 defer
	defer conn.Close()

	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Fatalf("error writing status line to connection: %v", err)
	}

	headers := response.GetDefaultHeaders(0)

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Fatalf("error writing headers to connection: %v", err)
	}

	fmt.Printf("Successfully writes to the connection %v\n", conn.RemoteAddr())
}

// close 함수
func (s *Server) Close() error {
	// 서버 종료 true 저장
	s.closed.Store(true)

	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}
