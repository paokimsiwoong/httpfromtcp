package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paokimsiwoong/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
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
