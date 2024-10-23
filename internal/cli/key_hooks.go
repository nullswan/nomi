package cli

import (
	"github.com/nullswan/nomi/internal/transcription"
	hook "github.com/robotn/gohook"
)

// SetupKeyHooks sets up global key hooks for audio control.
func SetupKeyHooks(
	cmdKeyCode uint16,
	ts *transcription.TranscriptionServer,
) (chan struct{}, chan struct{}) {
	audioStartCh := make(chan struct{}, 1)
	audioEndCh := make(chan struct{}, 1)

	// Start the hook event processing
	s := hook.Start()
	go func() {
		for e := range s {
			if e.Rawcode != cmdKeyCode {
				continue
			}

			if e.Kind == hook.KeyHold || e.Kind == hook.KeyDown {
				select {
				case audioStartCh <- struct{}{}:
				default:
				}
			}

			if e.Kind == hook.KeyUp {
				select {
				case audioEndCh <- struct{}{}:
					ts.FlushBuffers()
				default:
				}
			}
		}
	}()

	return audioStartCh, audioEndCh
}
