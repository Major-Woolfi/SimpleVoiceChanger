package audio

type Config struct {
	SampleRate int
	BufferSize int
	Channels   int
	InputDevice string
	OutputDevice string
}

func DefaultConfig() Config {
	return Config{
		SampleRate: DefaultSampleRate,
		BufferSize: DefaultBufferSize,
		Channels:   DefaultChannels,
	}
}
