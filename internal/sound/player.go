package sound

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

var speakerOnce sync.Once

func startSpeaker(sr beep.SampleRate) error {
	var err error

	speakerOnce.Do(func() {
		err = speaker.Init(sr, sr.N(time.Second/10))
	})

	if err != nil {
		return fmt.Errorf("error initializing speaker: %w", err)
	}

	return err
}

func PlaySuccess() error {
	f, err := os.Open("success.mp3")
	if err != nil {
		return fmt.Errorf("error opening sound file: %w", err)
	}

	defer f.Close()

	if err := playSound(f); err != nil {
		return fmt.Errorf("error playing sound: %w", err)
	}
	return nil
}

const resampleQuality = 4

func PlayBuffer(buf []byte) error {
	reader := bytes.NewReader(buf)
	r := io.NopCloser(reader)
	defer r.Close()
	if err := playSound(r); err != nil {
		return fmt.Errorf("error playing sound: %w", err)
	}
	return nil
}

func playSound(rd io.ReadCloser) error {
	streamer, format, err := mp3.Decode(rd)
	if err != nil {
		return fmt.Errorf("error decoding sound file: %w", err)
	}

	defer streamer.Close()

	sr := format.SampleRate * 2
	err = startSpeaker(sr)
	if err != nil {
		return fmt.Errorf("error initializing speaker: %w", err)
	}

	resampled := beep.Resample(resampleQuality, format.SampleRate, sr, streamer)
	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
