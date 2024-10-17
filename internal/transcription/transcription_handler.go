package transcription

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nullswan/nomi/internal/audio"
	openai "github.com/sashabaranov/go-openai"
)

type TranscriptionHandler struct {
	client         *openai.Client
	contextTimeout time.Duration
	enableFixing   bool

	sampleRate     int
	channels       int
	bytesPerSample int
	bitsPerSample  int
}

func NewTranscriptionHandler(
	apiKey string,
	audioOpts *audio.AudioOptions,
) *TranscriptionHandler {
	return &TranscriptionHandler{
		client:         openai.NewClient(apiKey),
		contextTimeout: 30 * time.Second,
		sampleRate:     int(audioOpts.SampleRate),
		channels:       audioOpts.Channels,
		bytesPerSample: audioOpts.BytesPerSample,
		bitsPerSample:  audioOpts.BitsPerSample,
	}
}

func (th *TranscriptionHandler) SetContextTimeout(duration time.Duration) {
	th.contextTimeout = duration
}

func (th *TranscriptionHandler) SetEnableFixing(enabled bool) {
	th.enableFixing = enabled
}

func (th *TranscriptionHandler) Transcribe(pcmData []byte) (string, error) {
	// Add WAV header to PCM data
	wavData, err := AddWAVHeader(
		pcmData,
		th.sampleRate,
		th.channels,
		th.bitsPerSample,
	)
	if err != nil {
		return "", fmt.Errorf("failed to add WAV header: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), th.contextTimeout)
	defer cancel()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		Reader:   bytes.NewReader(wavData),
		FilePath: "audio.wav",
	}

	resp, err := th.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("transcription error: %w", err)
	}

	// TODO(nullswan): remove dump audio data to file
	dumpAudioToFile(wavData)

	return resp.Text, nil
}

func dumpAudioToFile(audioData []byte) {
	filename := fmt.Sprintf("audio_%d.wav", time.Now().UnixNano())
	err := os.WriteFile(filename, audioData, 0644)
	if err != nil {
		fmt.Printf("Failed to write audio file: %v\n", err)
	}

	fmt.Printf("Wrote audio data to %s\n", filename)
}
