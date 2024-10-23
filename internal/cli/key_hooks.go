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
	pressedKeys := make(map[uint16]bool)

	// Start the hook event processing
	s := hook.Start()
	go func() {
		for e := range s {
			if !(e.Kind == hook.KeyDown || e.Kind == hook.KeyHold || e.Kind == hook.KeyUp) {
				continue
			}

			switch e.Kind {
			case hook.KeyDown:
				pressedKeys[e.Rawcode] = true
			case hook.KeyHold:
				pressedKeys[e.Rawcode] = true
			case hook.KeyUp:
				delete(pressedKeys, e.Rawcode)
			}

			if e.Kind == hook.KeyUp {
				select {
				case audioEndCh <- struct{}{}:
					ts.FlushBuffers()
				default:
				}
			}

			if e.Rawcode != cmdKeyCode {
				if e.Kind == hook.KeyDown || e.Kind == hook.KeyHold {
					select {
					case audioEndCh <- struct{}{}:
						ts.Reset()
					default:
					}
				}
				continue
			}

			// Check if only the command key is pressed
			if !(pressedKeys[cmdKeyCode] && len(pressedKeys) == 1) {
				continue
			}

			if e.Kind == hook.KeyHold || e.Kind == hook.KeyDown {
				select {
				case audioStartCh <- struct{}{}:
				default:
				}
			}
		}
	}()

	return audioStartCh, audioEndCh
}
