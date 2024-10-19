package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/config"
	"github.com/nullswan/nomi/internal/logger"
	"github.com/nullswan/nomi/internal/transcription"
)

// TODO(nullswan): These values should be configurable + auto-tuned
const (
	primaryBufferMinBufferDuration = 500 * time.Millisecond
	primaryBufferOverlapDuration   = 100 * time.Millisecond

	secondaryBufferMinBufferDuration = 2 * time.Second
	secondaryBufferOverlapDuration   = 400 * time.Millisecond

	vadEnergyThreshold = 0.005
	vadFlushInterval   = 310 * time.Millisecond
	vadSilenceDuration = 500 * time.Millisecond
	vadPauseDuration   = 300 * time.Millisecond
)

// InitTranscriptionServer initializes the Transcription Server with predefined buffer settings.
func InitTranscriptionServer(
	oaiKey string,
	audioOpts *audio.AudioOptions,
	log *logger.Logger,
	callback transcription.TranscriptionServerCallbackT,
) (*transcription.TranscriptionServer, error) {
	bufferManagerPrimary := transcription.NewBufferManager(audioOpts)
	bufferManagerPrimary.SetMinBufferDuration(primaryBufferMinBufferDuration)
	bufferManagerPrimary.SetOverlapDuration(primaryBufferOverlapDuration)

	bufferManagerSecondary := transcription.NewBufferManager(audioOpts)
	bufferManagerSecondary.SetMinBufferDuration(
		secondaryBufferMinBufferDuration,
	)
	bufferManagerSecondary.SetOverlapDuration(secondaryBufferOverlapDuration)

	textReconciler := transcription.NewTextReconciler(log)
	tsHandler := transcription.NewTranscriptionHandler(
		oaiKey,
		audioOpts,
		log,
	)
	tsHandler.SetEnableFixing(true)

	return transcription.NewTranscriptionServer(
		bufferManagerPrimary,
		bufferManagerSecondary,
		tsHandler,
		textReconciler,
		log,
		callback,
	), nil
}

// InitVAD initializes the Voice Activity Detection with predefined configurations.
func InitVAD(
	ts *transcription.TranscriptionServer,
	log *logger.Logger,
) *audio.VAD {
	vad := audio.NewVAD(
		audio.VADConfig{
			EnergyThreshold: vadEnergyThreshold,
			FlushInterval:   vadFlushInterval,
			SilenceDuration: vadSilenceDuration,
			PauseDuration:   vadPauseDuration,
		},
		audio.VADCallbacks{
			OnSpeechStart: func() {
				log.Debug("VAD: Speech started")
			},
			OnSpeechEnd: func() {
				log.Debug("VAD: Speech ended")
				ts.FlushBuffers()
			},
			OnFlush: func(buffer []float32) {
				log.With("buf_sz", len(buffer)).Debug("VAD: Buffer flushed")

				data, err := audio.Float32ToPCM(buffer)
				if err != nil {
					log.With("error", err).
						Error("Failed to convert float32 to PCM")
					return
				}

				ts.AddAudio(data)
			},
			OnPause: func() {
				log.Debug("VAD: Speech paused")
				ts.FlushPrimaryBuffer()
			},
		},
		log,
	)
	return vad
}

func InitVoice(
	cfg *config.Config,
	log *logger.Logger,
	handleTranscription func(text string, isProcessing bool),
	cmdKeyCode uint16,
) (*audio.AudioStream, <-chan struct{}, <-chan struct{}, error) {
	if !cfg.Input.Voice.Enabled {
		return nil, nil, nil, nil
	}

	if err := portaudio.Initialize(); err != nil {
		return nil, nil, nil, fmt.Errorf(
			"failed to initialize PortAudio: %w",
			err,
		)
	}

	audioOpts, err := audio.ComputeAudioOptions(&audio.AudioOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error computing audio options: %w",
			err,
		)
	}

	oaiKey := os.Getenv("OPENAI_API_KEY")
	if oaiKey == "" {
		return nil, nil, nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ts, err := InitTranscriptionServer(
		oaiKey,
		audioOpts,
		log,
		handleTranscription,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error initializing transcription server: %w",
			err,
		)
	}
	ts.Start()

	vad := InitVAD(ts, log)
	vad.Start()

	inputStream, err := audio.NewInputStream(
		log,
		audioOpts,
		func(buffer []float32) {
			vad.Feed(buffer)
		},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"failed to create input stream: %w",
			err,
		)
	}

	audioStartCh, audioEndCh := SetupKeyHooks(cmdKeyCode)
	return inputStream, audioStartCh, audioEndCh, nil
}
