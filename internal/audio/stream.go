package audio

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	maxFramesPerBuffer = 4096
	minLatency         = 50 * time.Millisecond
)

type StreamHandler struct {
	stream *portaudio.Stream
	logger *slog.Logger
}

type StreamParameters struct {
	SampleRate      float64
	Latency         time.Duration
	FramesPerBuffer int
	Channels        int
	BytesPerSample  int
	BitsPerSample   int
}

func ComputeDefaultAdudioOptions() (*StreamParameters, error) {
	device, err := portaudio.DefaultInputDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get default input device: %w", err)
	}

	opts, err := ComputeAudioOptions(device, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compute audio options: %w", err)
	}

	return opts, nil
}

func ComputeAudioOptions(
	device *portaudio.DeviceInfo,
	opts *StreamParameters,
) (*StreamParameters, error) {
	if opts == nil {
		return nil, errors.New("AudioOptions cannot be nil")
	}

	if device == nil {
		return nil, errors.New("DeviceInfo cannot be nil")
	}

	opts.SampleRate = device.DefaultSampleRate

	if opts.Latency == 0 {
		opts.Latency = minLatency
	}

	opts.FramesPerBuffer = int(
		opts.SampleRate * float64(opts.Latency) / float64(time.Second),
	)

	opts.Channels = device.MaxInputChannels
	opts.BytesPerSample = 2
	opts.BitsPerSample = 16

	return opts, nil
}

func NewInputStream(
	logger *slog.Logger,
	opts *StreamParameters,
	callback func([]float32),
) (*StreamHandler, error) {
	// Get the default input device
	inputDevice, err := portaudio.DefaultInputDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get default input device: %w", err)
	}

	// Compute and validate options
	opts, err = ComputeAudioOptions(inputDevice, opts)
	if err != nil {
		return nil, err
	}

	if opts.FramesPerBuffer > maxFramesPerBuffer {
		logger.With("frames_per_buffer", opts.FramesPerBuffer).Warn(
			fmt.Sprintf("FramesPerBuffer seems too high (> %d)", maxFramesPerBuffer),
		)
	}

	logger = logger.With("component", "audio_stream").
		With("device_name", inputDevice.Name)

	logger.
		With("sample_rate", opts.SampleRate).
		With("frames_per_buffer", opts.FramesPerBuffer).
		With("latency", opts.Latency).
		Info("Using default input device")

	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   inputDevice,
			Channels: opts.Channels,
			Latency:  opts.Latency,
		},
		SampleRate:      opts.SampleRate,
		FramesPerBuffer: opts.FramesPerBuffer,
	}

	stream, err := portaudio.OpenStream(
		streamParams,
		func(in []float32, _ []float32) {
			callback(in)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &StreamHandler{
		stream: stream,
		logger: logger,
	}, nil
}

func (a *StreamHandler) Start() error {
	a.logger.Info("Starting audio stream")
	err := a.stream.Start()
	if err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	return nil
}

func (a *StreamHandler) Stop() error {
	a.logger.Info("Stopping audio stream")
	err := a.stream.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop audio stream: %w", err)
	}

	return nil
}

func (a *StreamHandler) Close() error {
	a.logger.Info("Closing audio stream")
	err := a.stream.Close()
	if err != nil {
		return fmt.Errorf("failed to close audio stream: %w", err)
	}
	err = portaudio.Terminate()
	if err != nil {
		return fmt.Errorf("failed to terminate PortAudio: %w", err)
	}
	return nil
}

func GetDevices() ([]*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("error listing devices: %v", err)
	}

	return devices, nil
}
