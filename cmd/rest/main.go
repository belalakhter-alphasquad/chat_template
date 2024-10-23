package main

import (
	"os"
	"os/signal"
	"syscall"

	"practice.example/internal/api"
	"practice.example/internal/chat"
	"practice.example/utils"
)

func main() {
	api.NewServer()
	chat.SetupChat()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGQUIT, syscall.SIGTERM)

	for sig := range quit {
		utils.LogMessage(sig.String(), 1)
		os.Exit(0)

	}
}
