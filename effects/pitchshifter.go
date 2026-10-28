package effects

import "math"

// PitchShifter moves the pitch of the voice without changing its duration.
//
// Two stages:
//
//  1. time-stretch by the pitch ratio using WSOLA: every splice point is the
//     shift whose input frame correlates best with the tail of the output, so
//     pitch periods stay aligned instead of colliding into phasing mush;
//  2. read the stretched signal back with a step equal to the same ratio.
//     That cancels the duration change and leaves the pitch multiplied.
//
// Frames are summed with a Hann window and divided by the accumulated window
// weight (weighted overlap-add), so the level stays constant for any overlap
// ratio - a plain overlap-add pumps and dips when the ratio moves away from 1.
//
// The shifter is stateful across engine blocks; Process always returns exactly
// len(src) samples so the pipeline stays real-time. The price is a fixed delay
// of roughly one analysis window.
type PitchShifter struct {
	frame  int // analysis/synthesis window in samples
	hop    int // analysis hop (Ta) in samples
	ts     int // synthesis hop (Ts) in samples
	search int // WSOLA search radius in samples
	window []float64

	inbuf  []float32 // input not yet consumed by the analyser
	out    []float32 // stretched samples awaiting the resampler
	weight []float64 // accumulated window weight per output sample

	mOff   int     // analysis cursor, relative to inbuf[0]
	nOff   int     // synthesis cursor, relative to out[0]
	filled int     // how many output samples are finalized
	rsOff  float64 // resampler cursor, relative to out[0]
	step   float64 // resampler step, exactly ts/hop

	semitones float64
	ratio     float64
	primed    bool
}

// NewPitchShifter creates a shifter for the given sample rate. The window is
// about 40 ms: long enough to hold several pitch periods of a deep voice,
// short enough to keep the added latency usable in a voice chat.
func NewPitchShifter(sampleRate int) *PitchShifter {
	frame := sampleRate / 25
	hop := frame / 4
	if hop < 16 {
		frame, hop = 64, 16
	}
	p := &PitchShifter{
		frame:  frame,
		hop:    hop,
		ts:     hop,
		search: hop / 2,
		step:   1,
		window: hannWindow(frame),
	}
	p.reset()
	return p
}

func hannWindow(n int) []float64 {
	w := make([]float64, n)
	for i := range w {
		w[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n)))
	}
	return w
}

// reset drops all state. Called whenever the pitch changes, so a buffer
// stretched with the previous ratio cannot leak artifacts into the voice.
func (p *PitchShifter) reset() {
	p.inbuf = p.inbuf[:0]
	p.out = p.out[:0]
	p.weight = p.weight[:0]
	p.mOff = 0
	p.nOff = 0
	p.filled = 0
	p.rsOff = 0
	p.primed = false
}

// SetSemitones sets the pitch offset in semitones, clamped to [-24, 24].
// Zero makes Process a pure passthrough.
//
// The synthesis hop is forced to the nearest whole sample and the resampler
// step is then derived from that integer hop, not from the ideal ratio. Reading
// with the ideal ratio would consume slightly more stretched samples than the
// analyser produces, the output buffer would starve and the voice would turn
// into silence. Matching the rates exactly keeps the pipeline stable forever.
func (p *PitchShifter) SetSemitones(semi float64) {
	if semi > 24 {
		semi = 24
	}
	if semi < -24 {
		semi = -24
	}
	if semi == p.semitones {
		return
	}
	p.semitones = semi
	if semi == 0 {
		p.ratio = 1
		p.ts = p.hop
		p.step = 1
		p.reset()
		return
	}
	p.ratio = math.Pow(2, semi/12)
	p.ts = int(math.Round(float64(p.hop) * p.ratio))
	if p.ts < 1 {
		p.ts = 1
	}
	p.step = float64(p.ts) / float64(p.hop)
	p.reset()
}

// Semitones returns the current pitch offset.
func (p *PitchShifter) Semitones() float64 {
	return p.semitones
}

// Process shifts the pitch of src and returns exactly len(src) samples.
func (p *PitchShifter) Process(src []float32) []float32 {
	res := make([]float32, len(src))
	p.ProcessInto(res, src)
	return res
}

// ProcessInto is Process without the allocation: it writes exactly len(src)
// shifted samples into dst, which must be at least len(src) long. The engine
// calls this on every audio block, so it must not allocate.
func (p *PitchShifter) ProcessInto(dst, src []float32) {
	if p.semitones == 0 || len(src) == 0 {
		copy(dst, src)
		return
	}

	p.inbuf = append(p.inbuf, src...)
	p.stretch()

	for i := range src {
		dst[i] = p.readSample()
	}
	p.trim()
}

// stretch runs the WSOLA loop for every analysis frame currently available.
//
// The analysis cursor moves on a fixed grid of hop samples: only the read
// position inside the frame is shifted by the best correlation offset. Moving
// the cursor itself would make input consumption depend on the signal, and the
// real-time balance with the resampler would break.
func (p *PitchShifter) stretch() {
	for len(p.inbuf)-p.mOff >= p.frame+p.search {
		off := 0
		if p.primed {
			off = p.bestOffset()
		}
		start := p.mOff + off
		if start < 0 {
			start = 0
		}
		if start+p.frame > len(p.inbuf) {
			break
		}
		p.overlay(start)
		p.mOff += p.hop
		p.nOff += p.ts
		p.primed = true
	}
}

// bestOffset returns the shift in [-search, search] whose input frame best
// matches the output region the new frame will overlap.
func (p *PitchShifter) bestOffset() int {
	overlap := p.filled - p.nOff
	if overlap < p.frame/4 {
		return 0
	}
	if overlap > p.frame {
		overlap = p.frame
	}

	step := 1
	if overlap > 384 {
		step = 3
	}

	best, bestScore := 0, math.Inf(-1)
	for off := -p.search; off <= p.search; off++ {
		start := p.mOff + off
		if start < 0 || start+overlap > len(p.inbuf) {
			continue
		}
		var score float64
		for i := 0; i < overlap; i += step {
			score += float64(p.out[p.nOff+i]) / math.Max(p.weight[p.nOff+i], 1e-3) *
				float64(p.inbuf[start+i])
		}
		if score > bestScore {
			bestScore = score
			best = off
		}
	}
	return best
}

// overlay adds the Hann-windowed input frame at inStart into the output at the
// synthesis cursor and accumulates the window weight for normalization.
func (p *PitchShifter) overlay(inStart int) {
	end := p.nOff + p.frame
	if end > len(p.out) {
		grow := make([]float32, end)
		copy(grow, p.out)
		p.out = grow
		growW := make([]float64, end)
		copy(growW, p.weight)
		p.weight = growW
	}
	for i := 0; i < p.frame; i++ {
		w := p.window[i]
		p.out[p.nOff+i] += float32(float64(p.inbuf[inStart+i]) * w)
		p.weight[p.nOff+i] += w
	}
	if end > p.filled {
		p.filled = end
	}
}

// readSample returns one normalized output sample and advances the resampler.
// While the analyser has not produced anything yet the cursor stands still, so
// the warm-up turns into a fixed delay instead of a permanent buffer underrun
// that would mute the voice for the rest of the session.
func (p *PitchShifter) readSample() float32 {
	pos := int(p.rsOff)
	if pos >= p.filled {
		return 0
	}
	v := p.normalized(pos)
	if pos+1 < p.filled {
		f := p.rsOff - math.Floor(p.rsOff)
		v = v*(1-f) + p.normalized(pos+1)*f
	}
	p.rsOff += p.step
	return float32(v)
}

// normalized divides the accumulated overlap-add sum by the accumulated window
// weight, which keeps the output level flat for any overlap density.
func (p *PitchShifter) normalized(i int) float64 {
	return float64(p.out[i]) / math.Max(p.weight[i], 1e-3)
}

// trim drops samples that are already consumed so buffers stay bounded.
// The drop is capped at the synthesis cursor: dropping past it would move the
// region the analyser still reads behind the buffer start.
func (p *PitchShifter) trim() {
	keep := p.frame
	used := int(p.rsOff) - keep
	if limit := p.nOff - p.frame/2; used > limit {
		used = limit
	}
	if used > 0 && used <= len(p.out) {
		p.out = append(p.out[:0], p.out[used:]...)
		p.weight = append(p.weight[:0], p.weight[used:]...)
		p.nOff -= used
		p.filled -= used
		p.rsOff -= float64(used)
	}
	if drop := p.mOff - p.search; drop > 0 && drop < len(p.inbuf) {
		p.inbuf = append(p.inbuf[:0], p.inbuf[drop:]...)
		p.mOff -= drop
	}
}
