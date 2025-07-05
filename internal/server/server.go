package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/paokimsiwoong/httpfromtcp/internal/headers"
	"github.com/paokimsiwoong/httpfromtcp/internal/request"
	"github.com/paokimsiwoong/httpfromtcp/internal/response"
)

type Server struct {
	port     int //@@@ 예시의 경우 구조체에 port 저장 안함
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

// request 처리를 하는 함수들의 타입으로 쓰일 Handler 정의
// type Handler func(w io.Writer, req *request.Request) *HandlerError
// @@@ Handler가 header, status code, body를 직접 작성 가능하도록 구조 변경
type Handler func(w *response.Writer, req *request.Request)

// type HandlerError struct {
// 	StatusCode response.StatusCode
// 	Message    []byte
// }

// Server 구조체를 초기화하고 반환하면서 Server.listen 메소드를 고 루틴으로 시작하는 함수
func Serve(port int, handler Handler) (*Server, error) {
	// @@@ 예시의 경우 어차피 *Server를 반환하므로 구조체 선언때도 &Server{}로 바로 포인터 생성
	server := Server{
		port:    port,
		handler: handler,
		closed:  atomic.Bool{},
	}

	// tcp listener 생성 및 Server 구조체에 저장
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, fmt.Errorf("error creating listener: %w", err)
	}
	// @@@ 예시의 경우 listener 생성을 먼저 하고 그 뒤에 Server 구조체 선언
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
				// break
				return
			}
			// 서버가 닫히지 않았는데 에러가 발생했다면 심각한 문제이므로 로그 후 종료
			// log.Fatalf("error accepting connection: %v", err)
			// @@@ 예시처럼 에러 발생시에도 handle함수를 종료하지 않고 다음 for 루프로 넘어가기
			log.Printf("error accepting connection: %v", err)
			continue
		}
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

		// Accept 함수 에러처리 수정전에는 서버가 닫히면서 Accept함수가 에러를 발생하면
		// err != nil && s.closed.Load() 상태가 예외 처리되지 않고
		// 이 go s.handle(curConn) 까지 와서 에러가 발생했기때문에
		// if !s.closed.Load() 가 필요했었지만
		// 이제 서버가 종료되면 위에서 함수를 빠져나가므로 문제없음
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

	writer := &response.Writer{
		State: response.WriterStateInitialized,
	}

	// internal/request의 RequestFromReader를 이용해 conn이 보낸 request 파싱
	req, err := request.RequestFromReader(conn)
	if err != nil {
		// log.Fatalf("error parsing request: %v", err)
		// @@@ 예시를 따라 HandlerError 이용
		// WriteHandlerError(
		// 	&HandlerError{
		// 		StatusCode: response.StatusBadRequest,
		// 		Message:    []byte(err.Error()),
		// 	},
		// 	conn,
		// )
		WriteHandlerError(writer, conn, response.StatusBadRequest, []byte(err.Error()))
		// @@@ log.Fatalf 대신 return
		return
	}

	// @@@ 구조 변경
	// handler가 response body를 임시로 저장할 버퍼 생성
	// buffer := &bytes.Buffer{}
	// @@@ 예시는 bytes.NewBuffer([]byte{}) 사용함
	// io.Writer를 구현하는건 *bytes.Buffer
	// @@@ 구조 변경

	// handler 호출
	s.handler(writer, req)

	// @@@ 구조 변경
	// err = response.WriteStatusLine(conn, response.StatusOK)
	// if err != nil {
	// 	// log.Fatalf("error writing status line to connection: %v", err)
	// 	// @@@ 예시를 따라 HandlerError 이용
	// 	WriteHandlerError(
	// 		&HandlerError{
	// 			StatusCode: response.StatusInternalServerError,
	// 			Message:    []byte(err.Error()),
	// 		},
	// 		conn,
	// 	)
	// 	// @@@ log.Fatalf 대신 return
	// 	return
	// }

	// headers := response.GetDefaultHeaders(buffer.Len())

	// err = response.WriteHeaders(conn, headers)
	// if err != nil {
	// 	// log.Fatalf("error writing headers to connection: %v", err)
	// 	// @@@ 예시를 따라 HandlerError 이용
	// 	WriteHandlerError(
	// 		&HandlerError{
	// 			StatusCode: response.StatusInternalServerError,
	// 			Message:    []byte(err.Error()),
	// 		},
	// 		conn,
	// 	)
	// 	// @@@ log.Fatalf 대신 return
	// 	return
	// }

	// // buffer(여기서는 io.Reader 구현체)에 저장된 response body의 내용을 conn(io.Writer)에 저장
	// _, err = io.Copy(conn, buffer)
	// // @@@ 예시는 b := buffer.Bytes()를 써서
	// // @@@ buffer의 아직 읽히지 않은 []bytes 타입 데이터 부분만 가져온 뒤
	// // @@@ io.Copy 대신 conn.Write(b)를 사용
	// if err != nil {
	// 	// log.Fatalf("error writing response body to connection: %v", err)
	// 	// @@@ 예시를 따라 HandlerError 이용
	// 	WriteHandlerError(
	// 		&HandlerError{
	// 			StatusCode: response.StatusInternalServerError,
	// 			Message:    []byte(err.Error()),
	// 		},
	// 		conn,
	// 	)
	// 	// @@@ log.Fatalf 대신 return
	// 	return
	// }
	// @@@ 구조 변경

	n, err := conn.Write(writer.Data)
	if err != nil {
		log.Printf("conn.Write error: %v", err.Error())
		// @@@ s.handler(writer, req)를 거치고 나면
		// @@@ 앞에 선언된 writer의 State는 WriterStateDone인 상태
		// @@@ 새 writer를 만들어야 한다
		// writer := &response.Writer{
		// 	State: response.WriterStateInitialized,
		// }
		// WriteHandlerError(writer, conn, response.StatusInternalServerError, []byte(err.Error()))
		// @@@@@@ conn.Write가 에러가 난 경우 (ex: write tcp [::1]:42069->[::1]:43908: write: connection reset by peer)
		// @@@@@@ 이미 연결이 닫히거나 해서 쓰기가 불가능하므로 WriteHandlerError 안에서 conn.Write를 또하려해도 불가능
		return
	}

	fmt.Printf("Successfully writes %v to the connection %v\n", n, conn.RemoteAddr())
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

// 주어진 에러 정보를 response.Writer로 쓰는 함수
// @@@ 예시의 경우 일반 함수대신 HandlerError의 메소드로 작성
func WriteHandlerError(r *response.Writer, w io.Writer, statusCode response.StatusCode, message []byte) {
	err := r.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("error writing handler error status line to connection: %v", err)
		return
	}

	headers := headers.NewHeaders()

	headers.SetOverride("Content-Length", strconv.Itoa(len(message)))
	headers.SetOverride("Connection", "close")
	headers.SetOverride("Content-Type", "text/plain")

	err = r.WriteHeaders(headers)
	if err != nil {
		log.Printf("error writing handler error headers to connection: %v", err)
		return
	}

	_, err = r.WriteBody(message)
	if err != nil {
		log.Printf("error writing handler error message to connection: %v", err)
		return
	}

	_, err = w.Write(r.Data)
	if err != nil {
		log.Printf("error writing handler response to connection: %v", err)
		return
	}
}

// HandlerError 구조체에 적힌 에러 정보를 io.Writer로 쓰는 함수
// @@@ 예시의 경우 일반 함수대신 HandlerError의 메소드로 작성
// func WriteHandlerError(handlerError *HandlerError, w io.Writer) {
// 	err := response.WriteStatusLine(w, handlerError.StatusCode)
// 	if err != nil {
// 		log.Fatalf("error writing status line to connection: %v", err)
// 	}

// 	headers := response.GetDefaultHeaders(len(handlerError.Message))

// 	err = response.WriteHeaders(w, headers)
// 	if err != nil {
// 		log.Fatalf("error writing headers to connection: %v", err)
// 	}

// 	_, err = w.Write(handlerError.Message)
// 	if err != nil {
// 		log.Fatalf("error writing error message to connection: %v", err)
// 	}
// }
