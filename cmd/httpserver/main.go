package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/paokimsiwoong/httpfromtcp/internal/headers"
	"github.com/paokimsiwoong/httpfromtcp/internal/request"
	"github.com/paokimsiwoong/httpfromtcp/internal/response"
	"github.com/paokimsiwoong/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	// server는 request를 go 루틴으로 처리하고 바로 반환되는 함수이므로
	// main 함수가 바로 끝나지 않도록 기다리게 하는 부분이 필요
	// ==> sigChan이 그 역할을 담당
	sigChan := make(chan os.Signal, 1)
	// sigChan은 os.Signal (os이 보내는 시그널을 담는 인터페이스) 구현체를 받는 채널
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// Notify causes package signal to relay incoming signals to c
	// signal.Notify는 sigChan에 뒤이어 주어진 인자들(syscall.SIGINT, syscall.SIGTERM)에서
	// 오는 신호들을 보낸다 (ex: Ctrl+c)
	// @@@@@@@@
	<-sigChan
	// 신호가 올 때까진 여기서 sigChan이 블락해서 main 함수 종료를 막는다
	// @@@@@@@@
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	headers := headers.NewHeaders()

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		err := w.WriteStatusLine(response.StatusBadRequest)
		if err != nil {
			log.Printf("error writing status line: %v", err)
			return
		}

		body := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

		headers.SetOverride("Content-Length", strconv.Itoa(len(body)))
		headers.SetOverride("Connection", "close")
		headers.SetOverride("Content-Type", "text/html")

		err = w.WriteHeaders(headers)
		if err != nil {
			log.Printf("error writing headers: %v", err)
			return
		}

		_, err = w.WriteBody([]byte(body))
		if err != nil {
			log.Printf("error writing body: %v", err)
			return
		}

	case "/myproblem":
		err := w.WriteStatusLine(response.StatusInternalServerError)
		if err != nil {
			log.Printf("error writing status line: %v", err)
			return
		}

		body := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

		headers.SetOverride("Content-Length", strconv.Itoa(len(body)))
		headers.SetOverride("Connection", "close")
		headers.SetOverride("Content-Type", "text/html")

		err = w.WriteHeaders(headers)
		if err != nil {
			log.Printf("error writing headers: %v", err)
			return
		}

		_, err = w.WriteBody([]byte(body))
		if err != nil {
			log.Printf("error writing body: %v", err)
			return
		}
	default:
		err := w.WriteStatusLine(response.StatusOK)
		if err != nil {
			log.Printf("error writing status line: %v", err)
			return
		}

		body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

		headers.SetOverride("Content-Length", strconv.Itoa(len(body)))
		headers.SetOverride("Connection", "close")
		headers.SetOverride("Content-Type", "text/html")

		err = w.WriteHeaders(headers)
		if err != nil {
			log.Printf("error writing headers: %v", err)
			return
		}

		_, err = w.WriteBody([]byte(body))
		if err != nil {
			log.Printf("error writing body: %v", err)
			return
		}
	}
}

// request 종류에 따라 알 맞은 처리를 하는 Handler 타입 함수
// func handler(w io.Writer, req *request.Request) *server.HandlerError {
// 	switch req.RequestLine.RequestTarget {
// 	case "/yourproblem":
// 		return &server.HandlerError{
// 			StatusCode: response.StatusBadRequest,
// 			Message:    []byte("Your problem is not my problem\n"),
// 		}
// 	case "/myproblem":
// 		return &server.HandlerError{
// 			StatusCode: response.StatusInternalServerError,
// 			Message:    []byte("Woopsie, my bad\n"),
// 		}
// 	default:
// 		_, err := w.Write([]byte("All good, frfr\n"))
// 		if err != nil {
// 			log.Fatalf("Error writing response body: %v", err)
// 		}
// 		return nil
// 	}
// }
