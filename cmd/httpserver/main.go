package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
		ErrorHandler(w, req, 400)
	case "/myproblem":
		ErrorHandler(w, req, 500)
	default:
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			proxyHandler(w, req, headers)
			return
		}
		err := w.WriteStatusLine(response.StatusOK)
		if err != nil {
			log.Printf("error writing status line: %v", err)
			ErrorHandler(w, req, 500)
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
			ErrorHandler(w, req, 500)
			return
		}

		_, err = w.WriteBody([]byte(body))
		if err != nil {
			log.Printf("error writing body: %v", err)
			ErrorHandler(w, req, 500)
			return
		}
	}
}

func proxyHandler(w *response.Writer, req *request.Request, headers headers.Headers) {
	route := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")

	resp, err := http.Get("https://httpbin.org" + route)
	if err != nil {
		log.Printf("error making HTTP request: %v", err)
		ErrorHandler(w, req, 500)
		return
	}

	// 메모리에 올라가는 데이터 Close defer
	defer resp.Body.Close()
	// @@@ Go의 http.Response.Body는 chunked 인코딩을 자동으로 해제해서, 읽으면 바로 payload body만 나온다

	// status code가 200이 아닌 경우 처리
	if resp.StatusCode > 499 {
		ErrorHandler(w, req, 500)
		return
	}
	if resp.StatusCode > 399 {
		ErrorHandler(w, req, 400)
		return
	}

	err = w.WriteStatusLine(response.StatusOK)
	if err != nil {
		log.Printf("error writing status line: %v", err)
		ErrorHandler(w, req, 500)
		return
	}

	cotentType := resp.Header.Get("Content-Type")
	headers.SetOverride("Content-Type", cotentType)

	// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
	// @@@ https://httpbin.org/stream/{n} 자체는 Transfer-Encoding 헤더 없음
	// @@@ 단순히 n개의 JSON response를 보낸다
	// log.Printf("response headers key Transfer-Encoding value: %v", resp.Header.Get("Transfer-Encoding"))
	// if resp.Header.Get("Transfer-Encoding") == "chunked" {
	// 	headers.SetOverride("Transfer-Encoding", "chunked")
	// } else {
	// 	// ???
	// 	log.Printf("error not chunks?: %v", err)
	// 	return
	// }
	// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

	// @@@ route 상관없이 다 무조건 chunk로 쪼개기 @@@
	headers.SetOverride("Transfer-Encoding", "chunked")

	headers.SetOverride("Trailer", "X-Content-SHA256, X-Content-Length")

	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("error writing headers: %v", err)
		ErrorHandler(w, req, 500)
		return
	}

	rawBodyLength := 0
	// rawBody := []byte{}
	rawBody := make([]byte, 0)
	// 위 두 방식은 길이 0, 용량 0
	// var rawBody []byte
	// nil 슬라이스, 길이 0, 용량 0

	// resp.Body.Read 루프
	for {
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
		// buffer := make([]byte, 0, 32)
		// @@@ Read 메소드는 주어진 []byte 슬라이스의 길이만큼 읽는데
		// @@@ cap만 32로 하고 길이를 0으로 둬버리면 매번 0 바이트만 읽는다
		// @@@ ==> buffer := make([]byte, 32) 와 같이 생성
		// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
		buffer := make([]byte, 1024)
		n, err := resp.Body.Read(buffer)
		log.Printf("%v bytes read from the response body", n)
		if err == io.EOF {
			if n != 0 {
				rawBodyLength += n
				rawBody = append(rawBody, buffer[:n]...)

				_, err = w.WriteChunkedBody(buffer[:n]) // 실제 데이터가 들어있는 부분까지만 입력
				if err != nil {
					log.Printf("error writing a chunk: %v", err)
					ErrorHandler(w, req, 500)
					return
				}

			}

			// resp.Body안의 모든 chunk가 w에 저장되었으므로 w.WriteChunkedBodyDone 실행
			_, err = w.WriteChunkedBodyDone()
			if err != nil {
				log.Printf("error writing a chunk end: %v", err)
				ErrorHandler(w, req, 500)
				return
			}

			break
		}

		// 읽어낸 조각 길이와 조각을 따로 저장
		rawBodyLength += n
		rawBody = append(rawBody, buffer[:n]...)

		_, err = w.WriteChunkedBody(buffer[:n]) // 실제 데이터가 들어있는 부분까지만 입력
		if err != nil {
			log.Printf("error writing a chunk: %v", err)
			ErrorHandler(w, req, 500)
			return
		}
	}

	hash := sha256.Sum256(rawBody)
	hashString := hex.EncodeToString(hash[:])
	// hash는 길이가 정해져있으므로 [:]를 붙여서 []byte 슬라이스로 변환
	log.Printf("hash string: %s", hashString)
	log.Printf("raw body length: %d", rawBodyLength)

	headers["X-Content-SHA256"] = hashString
	headers["X-Content-Length"] = strconv.Itoa(rawBodyLength)

	err = w.WriteTrailers(headers)
	if err != nil {
		log.Printf("error writing trailers: %v", err)
		ErrorHandler(w, req, 500)
		return
	}

	// if strings.HasPrefix(route, "/stream") {
	// } else {
	// 	headers.SetOverride("Trailer", "X-Content-SHA256, X-Content-Length")

	// 	err = w.WriteHeaders(headers)
	// 	if err != nil {
	// 		log.Printf("error writing headers: %v", err)
	// 		ErrorHandler(w, req, 500)
	// 		return
	// 	}

	// 	body, err := io.ReadAll(resp.Body)
	// 	if err != nil {
	// 		log.Printf("error reading response body: %v", err)
	// 		ErrorHandler(w, req, 500)
	// 		return
	// 	}

	// 	_, err = w.WriteBody([]byte(body))
	// 	if err != nil {
	// 		log.Printf("error writing body: %v", err)
	// 		ErrorHandler(w, req, 500)
	// 		return
	// 	}

	// 	hash := sha256.Sum256(body)
	// 	hashString := hex.EncodeToString(hash[:])
	// 	// hash는 길이가 정해져있으므로 [:]를 붙여서 []byte 슬라이스로 변환

	// 	headers["X-Content-SHA256"] = hashString
	// 	headers["X-Content-Length"] = strconv.Itoa(len(body))

	// 	err = w.WriteTrailers(headers)
	// 	if err != nil {
	// 		log.Printf("error writing trailers: %v", err)
	// 		ErrorHandler(w, req, 500)
	// 		return
	// 	}
	// }

}

func ErrorHandler(w *response.Writer, req *request.Request, statusCode int) {
	w.Data = []byte{}

	headers := headers.NewHeaders()

	body := ""

	// status code가 200이 아닌 경우 처리
	if statusCode > 499 {
		err := w.WriteStatusLine(response.StatusInternalServerError)
		if err != nil {
			log.Fatalf("error writing status line: %v", err)
		}

		body = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
	} else if statusCode > 399 {
		err := w.WriteStatusLine(response.StatusBadRequest)
		if err != nil {
			log.Fatalf("error writing status line: %v", err)
		}

		body = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
	} else {
		log.Fatal("error unknwon error code")
	}

	headers.SetOverride("Content-Length", strconv.Itoa(len(body)))
	headers.SetOverride("Connection", "close")
	headers.SetOverride("Content-Type", "text/html")

	err := w.WriteHeaders(headers)
	if err != nil {
		log.Fatalf("error writing headers: %v", err)
	}

	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Fatalf("error writing body: %v", err)
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
