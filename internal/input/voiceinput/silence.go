package voiceinput

import "bytes"

func compressAudio(input []byte) []byte {
	frameSize := framesPerBuffer * bytesPerSample * channels

	output := make([]byte, len(input))
	writeIndex := 0
	zeroFrame := make([]byte, frameSize)

	for i := 0; i < len(input); i += frameSize {
		frameEnd := i + frameSize
		if frameEnd > len(input) {
			frameEnd = len(input)
		}
		frame := input[i:frameEnd]

		if !bytes.Equal(frame, zeroFrame[:len(frame)]) {
			copy(output[writeIndex:], frame)
			writeIndex += len(frame)
		}
	}

	return output[:writeIndex]
}
