package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nullswan/ai/internal/input/voiceinput"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	// Start voice input
	channel := make(chan []byte)
	go voiceinput.Run(ctx, channel)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-channel:
			switch string(msg) {
			case "[START]":
				fmt.Println("User started speaking.")
			case "[STOP]":
				fmt.Println("User stopped speaking.")
			default:
				fmt.Println("Chunck size: ", len(msg))
			}
		}
	}
}
