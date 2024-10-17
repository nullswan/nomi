package transcription

import (
	"log/slog"
	"sync"

	"github.com/nullswan/nomi/internal/logger"
)

type TranscriptionServer struct {
	logger *slog.Logger

	transcriptionHandler   *TranscriptionHandler
	bufferManagerPrimary   *BufferManager
	bufferManagerSecondary *BufferManager
	textReconciler         *TextReconciler
	callback               func(string)
	wg                     sync.WaitGroup
}

// NewTranscriptionServer initializes the TranscriptionServer with configurable options.
func NewTranscriptionServer(
	bufferManagerPrimary *BufferManager,
	bufferManagerSecondary *BufferManager,
	handler *TranscriptionHandler,
	reconciler *TextReconciler,
	logger *logger.Logger,
	callback func(string),
) *TranscriptionServer {
	return &TranscriptionServer{
		bufferManagerPrimary:   bufferManagerPrimary,
		bufferManagerSecondary: bufferManagerSecondary,
		transcriptionHandler:   handler,
		textReconciler:         reconciler,
		logger: logger.With(
			"component",
			"transcription_server",
		),
		callback: callback,
	}
}

func (ts *TranscriptionServer) Start() {
	ts.wg.Add(2)
	go ts.processLoop(ts.bufferManagerPrimary)
	go ts.processLoop(ts.bufferManagerSecondary)
}

// AddAudio adds incoming audio data to the buffer manager.
func (ts *TranscriptionServer) AddAudio(audio []byte) {
	ts.bufferManagerPrimary.AddAudio(audio)
}

// processLoop continuously listens for buffer flush signals to initiate transcription.
func (ts *TranscriptionServer) processLoop(bm *BufferManager) {
	defer ts.wg.Done()
	for {
		audioChunck, ok := bm.GetAudio()
		if !ok {
			return
		}
		if len(audioChunck.Data) == 0 {
			continue
		}

		transcribed, err := ts.transcriptionHandler.Transcribe(audioChunck.Data)
		if err != nil {
			ts.logger.
				With("error", err).
				Error("Failed to transcribe audio")

			continue
		}

		ts.textReconciler.AddSegment(
			audioChunck.Timestamp,
			transcribed,
			audioChunck.BufferSource,
		)
		ts.callback(ts.textReconciler.GetCombinedText())
	}
}

// GetText retrieves the current transcribed text.
func (ts *TranscriptionServer) GetText() string {
	return ts.textReconciler.GetExistingText()
}

// Close gracefully shuts down the TranscriptionServer.
func (ts *TranscriptionServer) Close() {
	ts.bufferManagerPrimary.Close()
	ts.bufferManagerSecondary.Close()
	ts.wg.Wait()
}

// Reset clears the current transcribed text.
func (ts *TranscriptionServer) Reset() {
	ts.textReconciler.UpdateText("")
}

// Clear clears the buffer manager.
func (ts *TranscriptionServer) Flush() {
	ts.bufferManagerPrimary.Flush()
}
