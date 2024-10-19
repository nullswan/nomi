package cli

import (
	"time"

	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/logger"
	"github.com/nullswan/nomi/internal/transcription"
)

// InitTranscriptionServer initializes the Transcription Server with predefined buffer settings.
func InitTranscriptionServer(
	oaiKey string,
	audioOpts *audio.AudioOptions,
	log *logger.Logger,
	callback transcription.TranscriptionServerCallbackT,
) (*transcription.TranscriptionServer, error) {
	bufferManagerPrimary := transcription.NewBufferManager(audioOpts)
	bufferManagerPrimary.SetMinBufferDuration(500 * time.Millisecond)
	bufferManagerPrimary.SetOverlapDuration(100 * time.Millisecond)

	bufferManagerSecondary := transcription.NewBufferManager(audioOpts)
	bufferManagerSecondary.SetMinBufferDuration(2 * time.Second)
	bufferManagerSecondary.SetOverlapDuration(400 * time.Millisecond)

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
			EnergyThreshold: 0.005,
			FlushInterval:   310 * time.Millisecond,
			SilenceDuration: 500 * time.Millisecond,
			PauseDuration:   300 * time.Millisecond,
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
