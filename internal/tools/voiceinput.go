package tools

import (
	"fmt"
	"log/slog"

	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/cli"
	"github.com/nullswan/nomi/internal/config"
)

type VoiceInput struct {
	inputStream  *audio.StreamHandler
	audioStartCh <-chan struct{}
	audioEndCh   <-chan struct{}
}

func NewVoiceInput(
	cfg *config.Config,
	logger *slog.Logger,
	voiceInputCh chan string,
) (*VoiceInput, error) {
	inputStream, audioStartCh, audioEndCh, err := cli.InitVoice(
		cfg,
		logger,
		func(text string, isProcessing bool) {
			if !isProcessing {
				fmt.Println(">>>", text)
				voiceInputCh <- text
			}
		},
		cfg.Input.Voice.KeyCode,
		cfg.Input.Voice.Language,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing voice: %w", err)
	}

	return &VoiceInput{
		inputStream:  inputStream,
		audioStartCh: audioStartCh,
		audioEndCh:   audioEndCh,
	}, nil
}

func (v *VoiceInput) GetInputStream() *audio.StreamHandler {
	return v.inputStream
}

func (v *VoiceInput) GetAudioStartCh() <-chan struct{} {
	return v.audioStartCh
}

func (v *VoiceInput) GetAudioEndCh() <-chan struct{} {
	return v.audioEndCh
}

func (v *VoiceInput) Close() {
	if v.inputStream != nil {
		v.inputStream.Stop()
	}
}
