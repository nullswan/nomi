package sound

import (
	"fmt"
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

func PlaySuccess() error {
	f, err := os.Open("success.mp3")
	if err != nil {
		return fmt.Errorf("error opening sound file: %w", err)
	}

	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return fmt.Errorf("error decoding sound file: %w", err)
	}

	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second/10))

	resampled := beep.Resample(4, format.SampleRate, sr, streamer)
	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
