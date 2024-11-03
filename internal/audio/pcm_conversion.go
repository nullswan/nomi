package audio

const (
	bytesPerSample = 2
	maxInt16       = 32767
	pcmMask        = 0xff
	pcmByteShift   = 8
)

// Float32ToPCM converts a slice of float32 samples to PCM byte data (int16)
func Float32ToPCM(float32Buffer []float32) ([]byte, error) {
	pcmBuffer := make(
		[]byte,
		0,
		len(float32Buffer)*bytesPerSample,
	)
	for _, sample := range float32Buffer {
		// Clipping
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		// Convert float32 to int16
		intSample := int16(sample * maxInt16)

		// note: little-endian
		pcmBuffer = append(
			pcmBuffer,
			byte(intSample&pcmMask),
			byte((intSample>>pcmByteShift)&pcmMask),
		)
	}
	return pcmBuffer, nil
}
