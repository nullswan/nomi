package knowledgebase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/tools"
)

func OnStart(
	ctx context.Context,
	selector tools.Selector,
	logger tools.Logger,
	internalLogger *slog.Logger,
	knownledgebase tools.KnowledgeBase,
) error {
	devices, err := audio.GetDevices()
	if err != nil {
		return fmt.Errorf("failed to get audio devices: %w", err)
	}

	if len(devices) == 0 {
		return errors.New("no audio devices found")
	}

	deviceNames := make([]string, len(devices))
	for i, device := range devices {
		deviceNames[i] = device.Name
	}
	selectedDevice := selector.SelectString("Select audio device", deviceNames)
	if selectedDevice == "" {
		return errors.New("no audio device selected")
	}

	var selectedDeviceInfo *portaudio.DeviceInfo
	for _, device := range devices {
		if device.Name == selectedDevice {
			selectedDeviceInfo = device
			break
		}
	}

	audioCh := make(chan []float32)
	defer close(audioCh)

	callback := func(buffer []float32) {
		audioCh <- buffer
	}
	audioStream, err := tools.NewAudioStream(
		internalLogger,
		selectedDeviceInfo,
		callback,
	)
	if err != nil {
		return fmt.Errorf("failed to create audio stream: %w", err)
	}

	if err := audioStream.Start(); err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	defer func() {
		if err := audioStream.Close(); err != nil {
			logger.Error("failed to close audio stream")
		}
	}()

	headers := http.Header{
		"Authorization": []string{"Bearer " + os.Getenv("OPENAI_API_KEY")},
		"OpenAI-Beta":   []string{"realtime=v1"},
	}

	c, cresp, err := websocket.DefaultDialer.Dial(
		"wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview",
		headers,
	)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	defer cresp.Body.Close()
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					internalLogger.With("error", err).
						Error("Failed to read message")
					continue
				}
				msgType := struct {
					Type string `json:"type"`
				}{}

				err = json.Unmarshal(message, &msgType)
				if err != nil {
					internalLogger.With("error", err).Error(
						"Failed to unmarshal message",
					)
					continue
				}

				switch msgType.Type {
				// case "response.done":
				// 	internalLogger.Info(
				// 		fmt.Sprintf("Received %+s\n", message),
				// 	)
				case "response.text.done":
					msg := ResponseTextDoneEvent{}
					err = json.Unmarshal(message, &msg)
					if err != nil {
						internalLogger.With("error", err).
							Error("Failed to unmarshal response text done")
						continue
					}

					fmt.Printf("\nDone.\n")
					// fmt.Printf("\n%s\n", msg.Text)
				case "response.text.delta":
					msg := ResponseTextDeltaEvent{}
					err = json.Unmarshal(message, &msg)
					if err != nil {
						internalLogger.With("error", err).
							Error("Failed to unmarshal response text delta")
						continue
					}

					fmt.Printf("%s", msg.Delta)
				default:
					// internalLogger.With("type", msgType.Type).
					// 	Info("Received")
				}
			}
		}
	}()

	defer func() {
		err := c.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to write close message: %s", err))
		}
	}()

	knownledege, err := knownledgebase.QueryAll()
	if err != nil {
		return fmt.Errorf("failed to get knowledge base: %w", err)
	}

	sessionUpdate := NewSessionUpdateEvent(
		knownledege,
	)
	data, err := sessionUpdate.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal session update: %w", err)
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			err = c.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				return fmt.Errorf("failed to write close message: %w", err)
			}
			return fmt.Errorf("context done: %w", ctx.Err())
		case buffer := <-audioCh:
			encodedBuffer, err := base64EncodeAudio(
				buffer,
				int(audioStream.GetSampleRate()),
				audioStream.GetChannels(),
			)
			if err != nil {
				internalLogger.With("error", err).
					Error("failed to encode audio buffer")
				continue
			}

			message := NewInputAudioBufferAppend(encodedBuffer)
			data, err := json.Marshal(message)
			if err != nil {
				internalLogger.With("error", err).
					Error("json marshal error")
				continue
			}
			err = c.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return fmt.Errorf("failed to write message: %w", err)
			}
		}
	}
}

// https://platform.openai.com/docs/api-reference/realtime-server-events/response/done
type ResponseDoneEvent struct {
	Type     string `json:"type"`
	ID       string `json:"event_id"`
	Response struct {
		Object string `json:"object"`
		ID     string `json:"id"`
		Status string `json:"status"`
		Output struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Type    string `json:"type"`
			Status  string `json:"status"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	} `json:"response"`
}

// https://platform.openai.com/docs/api-reference/realtime-server-events/response/text/delta
type ResponseTextDeltaEvent struct {
	Type         string `json:"type"`
	ID           string `json:"event_id"`
	ResponseID   string `json:"response_id"`
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Delta        string `json:"delta"`
}

// https://platform.openai.com/docs/api-reference/realtime-server-events/response/text/done
type ResponseTextDoneEvent struct {
	Type         string `json:"type"`
	ID           string `json:"event_id"`
	ResponseID   string `json:"response_id"`
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Text         string `json:"text"`
}

type TurnDetection struct {
	Type              string  `json:"type"`
	Threshold         float64 `json:"threshold"`
	PrefixPaddingMs   int     `json:"prefix_padding_ms"`
	SilenceDurationMs int     `json:"silence_duration_ms"`
}

type InputAudioTranscription struct {
	Model string `json:"model"`
}

type Session struct {
	Instructions            string                  `json:"instructions"`
	TurnDetection           TurnDetection           `json:"turn_detection"`
	Voice                   string                  `json:"voice"`
	Temperature             float64                 `json:"temperature"`
	MaxResponseOutputTokens int                     `json:"max_response_output_tokens"`
	Tools                   []string                `json:"tools"`
	Modalities              []string                `json:"modalities"`
	InputAudioFormat        string                  `json:"input_audio_format"`
	OutputAudioFormat       string                  `json:"output_audio_format"`
	InputAudioTranscription InputAudioTranscription `json:"input_audio_transcription"`
	ToolChoice              string                  `json:"tool_choice"`
}

type SessionUpdateEvent struct {
	Type    string  `json:"type"`
	Session Session `json:"session"`
}

func NewSessionUpdateEvent(instructions string) *SessionUpdateEvent {
	return &SessionUpdateEvent{
		Type: "session.update",
		Session: Session{
			Instructions: instructions,
			TurnDetection: TurnDetection{
				Type:              "server_vad",
				Threshold:         0.05,
				PrefixPaddingMs:   300,
				SilenceDurationMs: 500,
			},
			Voice:                   "alloy",
			Temperature:             0.8,
			MaxResponseOutputTokens: 4096,
			Tools:                   []string{},
			Modalities:              []string{"text"},
			InputAudioFormat:        "pcm16",
			OutputAudioFormat:       "pcm16",
			InputAudioTranscription: InputAudioTranscription{
				Model: "whisper-1",
			},
			ToolChoice: "auto",
		},
	}
}

func (s *SessionUpdateEvent) Marshal() ([]byte, error) {
	ret, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal SessionUpdateEvent: %w",
			err,
		)
	}

	return ret, nil
}

type InputAudioBufferAppendEvent struct {
	Type  string `json:"type"`
	Audio string `json:"audio"`
}

func NewInputAudioBufferAppend(buffer string) *InputAudioBufferAppendEvent {
	return &InputAudioBufferAppendEvent{
		Type:  "input_audio_buffer.append",
		Audio: buffer,
	}
}

func (i *InputAudioBufferAppendEvent) Marshal() ([]byte, error) {
	ret, err := json.Marshal(i)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal InputAudioBufferAppend: %w",
			err,
		)
	}

	return ret, nil
}

func lowPassFilter(input []float32, cutoffFreq, sampleRate float32) []float32 {
	output := make([]float32, len(input))
	rc := 1.0 / (2 * math.Pi * float64(cutoffFreq))
	dt := 1.0 / float64(sampleRate)
	alpha := float32(dt / (rc + dt))

	output[0] = input[0]
	for i := 1; i < len(input); i++ {
		output[i] = output[i-1] + alpha*(input[i]-output[i-1])
	}
	return output
}

func downsample(input []float32) []float32 {
	output := make([]float32, len(input)/2)
	for i := range output {
		output[i] = input[i*2]
	}
	return output
}

func convertTo16BitPCM(input []float32) ([]byte, error) {
	buffer := new(bytes.Buffer)
	for _, v := range input {
		s := math.Max(-1, math.Min(1, float64(v)))
		var intSample int16
		if s < 0 {
			intSample = int16(s * 0x8000)
		} else {
			intSample = int16(s * 0x7fff)
		}
		err := binary.Write(buffer, binary.LittleEndian, intSample)
		if err != nil {
			return nil, fmt.Errorf("failed to write PCM data: %w",
				err)
		}
	}
	return buffer.Bytes(), nil
}

// https://platform.openai.com/docs/guides/realtime#audio-formats
const targetSampleRate = 24000

func downsampleAndConvert(
	input []float32,
	sampleRate int,
	channels int,
) ([]byte, error) {
	// Apply a low-pass filter with a cutoff frequency of 12 kHz
	filtered := lowPassFilter(
		input,
		float32(targetSampleRate/channels),
		float32(sampleRate),
	)

	// Downsample the filtered signal to 24 kHz
	downsampled := downsample(filtered)

	// Convert to 16-bit PCM format
	pcmData, err := convertTo16BitPCM(downsampled)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to PCM: %w", err)
	}
	return pcmData, nil
}

// Converts a slice of float32 to base64-encoded PCM16 data
func base64EncodeAudio(
	float32Array []float32,
	sampleRate int,
	channels int,
) (string, error) {
	pcmBytes, err := downsampleAndConvert(float32Array, sampleRate, channels)
	if err != nil {
		return "", fmt.Errorf("failed to convert to PCM: %w", err)
	}
	return base64.StdEncoding.EncodeToString(pcmBytes), nil
}
