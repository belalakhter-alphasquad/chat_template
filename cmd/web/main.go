package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/belalakhter-alphasquad/chat_template/internal/api"
	"github.com/belalakhter-alphasquad/chat_template/internal/buffer"
	"github.com/belalakhter-alphasquad/chat_template/internal/chat"
	"github.com/belalakhter-alphasquad/chat_template/utils"
)

func main() {
	api.NewServer()
	buf := buffer.NewBuffer()
	chat.SetupChat(buf)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGQUIT, syscall.SIGTERM)

	for sig := range quit {
		utils.LogMessage(sig.String(), 1)
		os.Exit(0)

	}
}
