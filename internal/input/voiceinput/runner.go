package voiceinput

import "context"

// var voiceChan chan string
// Be able to record voice input
// Be able to cancel the ongoing prompt
func Run(ctx context.Context, outCh chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
