package audio

type RingBuffer struct {
	data     []float32
	readPos  int
	writePos int
	size     int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]float32, size),
		size: size,
	}
}

func (r *RingBuffer) Write(samples []float32) int {
	for _, s := range samples {
		r.data[r.writePos] = s
		r.writePos = (r.writePos + 1) % r.size
		if r.writePos == r.readPos {
			r.readPos = (r.readPos + 1) % r.size
		}
	}
	return len(samples)
}

func (r *RingBuffer) Read(out []float32) int {
	n := len(out)
	if n > r.size-(r.writePos-r.readPos)%r.size {
		n = r.size - (r.writePos-r.readPos)%r.size
	}
	for i := 0; i < n; i++ {
		out[i] = r.data[r.readPos]
		r.readPos = (r.readPos + 1) % r.size
	}
	return n
}

func (r *RingBuffer) Available() int {
	return (r.writePos - r.readPos + r.size) % r.size
}

func (r *RingBuffer) Free() int {
	return r.size - r.Available() - 1
}

func (r *RingBuffer) Reset() {
	r.readPos = 0
	r.writePos = 0
}
