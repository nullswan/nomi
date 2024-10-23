package transcription

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// AddWAVHeader adds a WAV header to PCM audio data.
func AddWAVHeader(
	pcmData []byte,
	sampleRate int,
	channels int,
	bitsPerSample int,
) ([]byte, error) {
	var buf bytes.Buffer

	// RIFF header
	if err := binary.Write(&buf, binary.LittleEndian, []byte("RIFF")); err != nil {
		return nil, fmt.Errorf("error writing RIFF header: %w", err)
	}

	// ChunkSize: 36 + Subchunk2Size
	chunkSize := 36 + len(pcmData)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(chunkSize)); err != nil {
		return nil, fmt.Errorf("error writing ChunkSize: %w", err)
	}

	// Format
	if err := binary.Write(&buf, binary.LittleEndian, []byte("WAVE")); err != nil {
		return nil, fmt.Errorf("error writing Format: %w", err)
	}

	// Subchunk1 ID
	if err := binary.Write(&buf, binary.LittleEndian, []byte("fmt ")); err != nil {
		return nil, fmt.Errorf("error writing Subchunk1ID: %w", err)
	}

	// Subchunk1 Size (16 for PCM)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(16)); err != nil {
		return nil, fmt.Errorf("error writing Subchunk1Size: %w", err)
	}

	// AudioFormat (1 for PCM)
	if err := binary.Write(&buf, binary.LittleEndian, uint16(1)); err != nil {
		return nil, fmt.Errorf("error writing AudioFormat: %w", err)
	}

	// NumChannels
	if err := binary.Write(&buf, binary.LittleEndian, uint16(channels)); err != nil {
		return nil, fmt.Errorf("error writing NumChannels: %w", err)
	}

	// SampleRate
	if err := binary.Write(&buf, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return nil, fmt.Errorf("error writing SampleRate: %w", err)
	}

	// ByteRate = SampleRate * NumChannels * BitsPerSample/8
	byteRate := sampleRate * channels * bitsPerSample / 8 // nolint:gomnd
	if err := binary.Write(&buf, binary.LittleEndian, uint32(byteRate)); err != nil {
		return nil, fmt.Errorf("error writing ByteRate: %w", err)
	}

	// BlockAlign = NumChannels * BitsPerSample/8
	blockAlign := channels * bitsPerSample / 8 // nolint:gomnd
	if err := binary.Write(&buf, binary.LittleEndian, uint16(blockAlign)); err != nil {
		return nil, fmt.Errorf("error writing BlockAlign: %w", err)
	}

	// BitsPerSample
	if err := binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
		return nil, fmt.Errorf("error writing BitsPerSample: %w", err)
	}

	// Subchunk2 ID
	if err := binary.Write(&buf, binary.LittleEndian, []byte("data")); err != nil {
		return nil, fmt.Errorf("error writing Subchunk2ID: %w", err)
	}

	// Subchunk2 Size
	subchunk2Size := len(pcmData)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(subchunk2Size)); err != nil {
		return nil, fmt.Errorf("error writing Subchunk2Size: %w", err)
	}

	// PCM Data
	if _, err := buf.Write(pcmData); err != nil {
		return nil, fmt.Errorf("error writing PCM data: %w", err)
	}

	return buf.Bytes(), nil
}
