package cli

import (
	hook "github.com/robotn/gohook"
)

// SetupKeyHooks sets up global key hooks for audio control.
func SetupKeyHooks(cmdKeyCode uint16) (chan struct{}, chan struct{}) {
	audioStartCh := make(chan struct{}, 1)
	audioEndCh := make(chan struct{}, 1)

	// Register KeyHold hook for audio start
	hook.Register(hook.KeyHold, []string{""}, func(e hook.Event) {
		if e.Rawcode != cmdKeyCode {
			return
		}
		select {
		case audioStartCh <- struct{}{}:
		default:
		}
	})

	// Register KeyUp hook for audio end
	hook.Register(hook.KeyUp, []string{""}, func(e hook.Event) {
		if e.Rawcode != cmdKeyCode {
			return
		}
		select {
		case audioEndCh <- struct{}{}:
		default:
		}
	})

	// Start the hook event processing
	s := hook.Start()
	go func() {
		<-hook.Process(s)
	}()

	return audioStartCh, audioEndCh
}
