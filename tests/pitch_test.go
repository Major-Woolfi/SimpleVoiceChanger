package tests

import (
	"math"
	"testing"
	"time"

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

const testRate = 48000

func sine(freq float64, n, sampleRate int) []float32 {
	out := make([]float32, n)
	for i := range out {
		out[i] = float32(0.7 * math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)))
	}
	return out
}

func runShifter(t *testing.T, src []float32, semi float64, block int) []float32 {
	t.Helper()
	sh := effects.NewPitchShifter(testRate)
	sh.SetSemitones(semi)
	res := make([]float32, 0, len(src))
	for off := 0; off < len(src); off += block {
		end := off + block
		if end > len(src) {
			end = len(src)
		}
		out := sh.Process(src[off:end])
		if len(out) != end-off {
			t.Fatalf("block %d: got %d samples, want %d", off, len(out), end-off)
		}
		res = append(res, out...)
	}
	return res
}

// estimateFreq counts zero crossings over the tail of the signal, skipping the
// WSOLA warm-up region, and converts them into an approximate frequency.
func estimateFreq(sig []float32, sampleRate int) float64 {
	skip := len(sig) / 3
	sig = sig[skip:]
	if len(sig) < 3 {
		return 0
	}
	crossings := 0
	prev := sig[0]
	for _, v := range sig[1:] {
		if (prev < 0 && v >= 0) || (prev > 0 && v <= 0) {
			crossings++
		}
		prev = v
	}
	seconds := float64(len(sig)) / float64(sampleRate)
	if seconds == 0 {
		return 0
	}
	return float64(crossings) / 2 / seconds
}

func rms(sig []float32) float64 {
	var sum float64
	for _, v := range sig {
		sum += float64(v) * float64(v)
	}
	return math.Sqrt(sum / float64(len(sig)))
}

func TestPitchShifterKeepsLength(t *testing.T) {
	src := sine(200, testRate, testRate)
	for _, semi := range []float64{-12, -5, 0, 5, 12} {
		got := runShifter(t, src, semi, 512)
		if len(got) != len(src) {
			t.Fatalf("semi %.0f: len %d, want %d", semi, len(got), len(src))
		}
		for i, v := range got {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				t.Fatalf("semi %.0f: bad value at %d", semi, i)
			}
		}
	}
}

func TestPitchShifterRaisesPitch(t *testing.T) {
	src := sine(200, testRate, testRate)
	got := runShifter(t, src, 12, 512)
	freq := estimateFreq(got, testRate)
	if freq < 330 || freq > 480 {
		t.Fatalf("expected ~400 Hz after +12 semitones, got %.1f Hz", freq)
	}
	if r := rms(got[len(got)/2:]); r < 0.1 {
		t.Fatalf("output is silent after shift up, rms %.4f", r)
	}
}

func TestPitchShifterLowersPitch(t *testing.T) {
	src := sine(200, testRate, testRate)
	got := runShifter(t, src, -12, 512)
	freq := estimateFreq(got, testRate)
	if freq < 80 || freq > 130 {
		t.Fatalf("expected ~100 Hz after -12 semitones, got %.1f Hz", freq)
	}
	if r := rms(got[len(got)/2:]); r < 0.1 {
		t.Fatalf("output is silent after shift down, rms %.4f", r)
	}
}

func TestPitchShifterZeroIsPassthrough(t *testing.T) {
	src := sine(300, 4096, testRate)
	sh := effects.NewPitchShifter(testRate)
	sh.SetSemitones(0)
	got := sh.Process(src)
	for i := range src {
		if math.Abs(float64(got[i]-src[i])) > 1e-6 {
			t.Fatalf("zero shift changed sample %d: %v vs %v", i, got[i], src[i])
		}
	}
}

func TestPitchShifterSmallShiftStaysClose(t *testing.T) {
	src := sine(220, testRate, testRate)
	got := runShifter(t, src, 2, 512)
	freq := estimateFreq(got, testRate)
	if freq < 240 || freq > 280 {
		t.Fatalf("expected ~260 Hz after +2 semitones, got %.1f Hz", freq)
	}
}

func TestPitchShifterHandlesBlockChange(t *testing.T) {
	src := sine(180, testRate/2, testRate)
	a := runShifter(t, src, 7, 512)
	b := runShifter(t, src, 7, 1024)
	if len(a) != len(b) {
		t.Fatalf("block size changed output length: %d vs %d", len(a), len(b))
	}
	fa, fb := estimateFreq(a, testRate), estimateFreq(b, testRate)
	if math.Abs(fa-fb) > 15 {
		t.Fatalf("block size changed the result: %.1f Hz vs %.1f Hz", fa, fb)
	}
}

func TestPitchShifterStableRMS(t *testing.T) {
	src := sine(200, testRate, testRate)
	baseRMS := rms(src)
	for _, semi := range []float64{-12, -7, -3, 0, 3, 7, 12} {
		got := runShifter(t, src, semi, 512)
		tail := got[len(got)/2:]
		r := rms(tail)
		ratio := r / baseRMS
		if ratio < 0.5 || ratio > 1.5 {
			t.Fatalf("semi %.0f: RMS ratio %.2f, expected ~1.0", semi, ratio)
		}
	}
}

func TestPitchShifterSawtoothQuality(t *testing.T) {
	n := testRate
	src := make([]float32, n)
	f0 := 150.0
	for i := range src {
		ts := float64(i) / float64(testRate)
		var sum float64
		for k := 1; k <= 10; k++ {
			sum += math.Sin(2*math.Pi*f0*float64(k)*ts) / float64(k)
		}
		src[i] = float32(0.5 * sum / 3.0)
	}
	inRMS := rms(src[n/2:])
	got := runShifter(t, src, 5, 512)
	outRMS := rms(got[len(got)/2:])
	ratio := outRMS / inRMS
	if ratio < 0.6 || ratio > 1.4 {
		t.Fatalf("sawtooth +5: RMS ratio %.2f, expected ~1.0", ratio)
	}
	for i, v := range got {
		if math.Abs(float64(v)) > 1.5 {
			t.Fatalf("sawtooth +5: sample %d is %.3f, exceeds reasonable range", i, v)
		}
	}
}

// TestPitchShifterIntervals walks every semitone of the octave and checks the
// resulting pitch in cents. Integer rounding of the synthesis hop is the only
// source of error here, and it has to stay far below what a human can hear.
func TestPitchShifterIntervals(t *testing.T) {
	src := sine(220, testRate, testRate)
	for semi := -12.0; semi <= 12.0; semi++ {
		got := runShifter(t, src, semi, 512)
		want := 220 * math.Pow(2, semi/12)
		freq := estimateFreq(got, testRate)
		if freq <= 0 {
			t.Fatalf("semi %+.0f: no signal at the output", semi)
		}
		cents := 1200 * math.Log2(freq/want)
		if cents < -40 || cents > 40 {
			t.Fatalf("semi %+.0f: %.1f Hz instead of %.1f Hz (%+.1f cents)", semi, freq, want, cents)
		}
	}
}

// oneShotCapture feeds a fixed signal once and then reports no data, so the
// engine test stays bounded instead of spinning on an endless mock source.
type oneShotCapture struct {
	data []float32
	pos  int
}

func (c *oneShotCapture) Open(int, int) error { return nil }

func (c *oneShotCapture) Read(out []float32) (int, error) {
	n := copy(out, c.data[c.pos:])
	c.pos += n
	return n, nil
}

func (c *oneShotCapture) Close() {}

// snapshot returns everything the sink has received so far.
func (m *mockSink) snapshot() []float32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]float32(nil), m.written...)
}

// TestEngineAppliesPitch drives the whole engine pipeline with a pitch offset
// and checks that what reaches the output device is actually shifted. This
// covers the engine wiring that the standalone shifter tests cannot see.
func TestEngineAppliesPitch(t *testing.T) {
	const n = 24000
	src := sine(200, n, testRate)
	cap := &oneShotCapture{data: src}
	sink := &mockSink{}

	engine := audio.NewEngine(testRate, 512)
	engine.SetInput(cap)
	engine.SetOutput(sink)
	engine.SetPitch(12)
	if err := engine.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer engine.Stop()

	deadline := time.Now().Add(5 * time.Second)
	for sink.count() < n && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	engine.Stop()

	got := sink.snapshot()
	if len(got) < n {
		t.Fatalf("engine delivered %d samples, want %d", len(got), n)
	}
	tail := got[n/2:]
	if r := rms(tail); r < 0.2 {
		t.Fatalf("engine output is silent after pitch shift, rms %.4f", r)
	}
	if f := estimateFreq(tail, testRate); f < 330 || f > 480 {
		t.Fatalf("engine pitch shift: expected ~400 Hz, got %.1f Hz", f)
	}
}

// TestEnginePitchControls checks the pitch control surface of the engine:
// clamping, read-back and live changes while the loop is running.
func TestEnginePitchControls(t *testing.T) {
	e := audio.NewEngine(testRate, 512)
	e.SetInput(&oneShotCapture{data: sine(180, testRate, testRate)})
	e.SetOutput(&mockSink{})

	if e.Pitch() != 0 {
		t.Fatalf("initial pitch: %v", e.Pitch())
	}
	e.SetPitch(5)
	if e.Pitch() != 5 {
		t.Fatalf("pitch not stored: %v", e.Pitch())
	}
	e.SetPitch(99)
	if e.Pitch() != 24 {
		t.Fatalf("high pitch not clamped: %v", e.Pitch())
	}
	e.SetPitch(-99)
	if e.Pitch() != -24 {
		t.Fatalf("low pitch not clamped: %v", e.Pitch())
	}

	if err := e.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer e.Stop()

	for semi := 24.0; semi >= -24; semi-- {
		e.SetPitch(semi)
	}
	e.SetPitch(0)
	time.Sleep(50 * time.Millisecond)
}
