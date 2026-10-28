//go:build portaudio

package audio

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

// paRef/paInit/paRefMu track the shared PortAudio library lifecycle: it is
// initialized on the first Open/list and terminated after the last Close, so
// that repeated Start/Stop cycles keep working.
var paRef = 0
var paInit = false
var paRefMu sync.Mutex

// ensurePortAudioInit initializes PortAudio if needed and takes a reference.
func ensurePortAudioInit() error {
	paRefMu.Lock()
	defer paRefMu.Unlock()
	if !paInit {
		if err := portaudio.Initialize(); err != nil {
			return err
		}
		paInit = true
	}
	paRef++
	return nil
}

// hostAPIPreference lists the host APIs to prefer, best first. DirectSound
// exposes full device names and is reliable; WASAPI is preferred when it works
// but can fail to start on some hardware; MME is the last resort (its device
// names are truncated to 31 chars). WDM-KS is never used: it can only open
// devices in exclusive mode and locks the audio driver.
var hostAPIPreference = []portaudio.HostApiType{
	portaudio.DirectSound,
	portaudio.WASAPI,
	portaudio.MME,
}

func matchesDirection(d *portaudio.DeviceInfo, input bool) bool {
	if input {
		return d.MaxInputChannels > 0
	}
	return d.MaxOutputChannels > 0
}

// preferredHostAPIs returns the host APIs that can serve the requested
// direction, in preference order. Only APIs with at least one usable device
// are included.
func preferredHostAPIs(input bool) []*portaudio.HostApiInfo {
	hosts, err := portaudio.HostApis()
	if err != nil {
		return nil
	}
	var out []*portaudio.HostApiInfo
	for _, want := range hostAPIPreference {
		for _, h := range hosts {
			if h.Type != want {
				continue
			}
			for _, d := range h.Devices {
				if matchesDirection(d, input) {
					out = append(out, h)
				}
			}
			break
		}
	}
	return out
}

// candidateDevices returns the devices to try when opening, in order. For a
// named device, the matching device in each preferred host API is tried first,
// then a matching device in any other API. For the system default (empty name)
// the global default device - which on Windows resolves to the real MME device
// behind the OS default - is tried first, because opening the DirectSound
// virtual "Primary Sound Driver" as a default can produce a low-frequency hum
// on some hardware. Duplicates are skipped, so the first candidate that opens
// successfully is used.
func candidateDevices(name string, input bool) []*portaudio.DeviceInfo {
	var out []*portaudio.DeviceInfo
	seen := map[int]bool{}
	add := func(d *portaudio.DeviceInfo) {
		if d == nil || seen[d.Index] {
			return
		}
		seen[d.Index] = true
		out = append(out, d)
	}
	if name != "" {
		for _, h := range preferredHostAPIs(input) {
			for _, d := range h.Devices {
				if d.Name == name && matchesDirection(d, input) {
					add(d)
				}
			}
		}
		if devs, err := portaudio.Devices(); err == nil {
			for _, d := range devs {
				if d.Name == name && matchesDirection(d, input) {
					add(d)
				}
			}
		}
		return out
	}
	if input {
		if d, err := portaudio.DefaultInputDevice(); err == nil {
			add(d)
		}
	} else if d, err := portaudio.DefaultOutputDevice(); err == nil {
		add(d)
	}
	for _, h := range preferredHostAPIs(input) {
		if input {
			add(h.DefaultInputDevice)
		} else {
			add(h.DefaultOutputDevice)
		}
	}
	return out
}

// ListInputDevices returns the names of all available capture devices, or nil
// when the audio backend is unavailable.
func ListInputDevices() []string { return listDevices(true) }

// ListOutputDevices returns the names of all available playback devices, or
// nil when the audio backend is unavailable.
func ListOutputDevices() []string { return listDevices(false) }

func listDevices(input bool) []string {
	paRefMu.Lock()
	started := false
	if !paInit {
		if err := portaudio.Initialize(); err != nil {
			paRefMu.Unlock()
			return nil
		}
		paInit = true
		started = true
	}
	apis := preferredHostAPIs(input)
	if started {
		portaudio.Terminate()
		paInit = false
	}
	paRefMu.Unlock()

	seen := make(map[string]bool, 8)
	out := make([]string, 0, 8)
	for _, h := range apis {
		for _, d := range h.Devices {
			if !matchesDirection(d, input) || seen[d.Name] {
				continue
			}
			seen[d.Name] = true
			out = append(out, d.Name)
		}
	}

	// MME truncates device names to 31 chars, so the same physical device can
	// appear twice: once with the full DirectSound/WASAPI name and once with a
	// shorter MME prefix (e.g. "Microphone (USB PnP Audio Device" vs
	// "Microphone (USB PnP Audio Device)"). Drop entries that are a strict
	// prefix of a longer entry - they are truncations of that device.
	kept := out[:0]
	for _, name := range out {
		prefix := false
		for _, other := range out {
			if other != name && len(other) > len(name) && strings.HasPrefix(other, name) {
				prefix = true
				break
			}
		}
		if !prefix {
			kept = append(kept, name)
		}
	}
	return kept
}

// PortAudioCapture implements CaptureSource using PortAudio.
type PortAudioCapture struct {
	stream     *portaudio.Stream
	buffer     []float32
	bufferSize int
	sampleRate int
	deviceName string
	inited     bool
	mu         sync.Mutex
	closed     int32
	readPos    int
	writePos   int
	overflow   int32
}

// NewPortAudioCapture creates a new PortAudio capture source. An empty
// deviceName selects the system default input device.
func NewPortAudioCapture(sampleRate, bufferSize int, deviceName string) (*PortAudioCapture, error) {
	return &PortAudioCapture{
		bufferSize: bufferSize,
		sampleRate: sampleRate,
		deviceName: deviceName,
		buffer:     make([]float32, bufferSize*4),
	}, nil
}

// Open opens the capture stream.
func (c *PortAudioCapture) Open(sampleRate, bufferSize int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	atomic.StoreInt32(&c.closed, 0)

	c.sampleRate = sampleRate
	c.bufferSize = bufferSize
	c.buffer = make([]float32, bufferSize*4)
	c.readPos = 0
	c.writePos = 0
	atomic.StoreInt32(&c.overflow, 0)

	if err := ensurePortAudioInit(); err != nil {
		return fmt.Errorf("portaudio init: %w", err)
	}
	c.inited = true

	var lastErr error
	for _, device := range candidateDevices(c.deviceName, true) {
		params := portaudio.LowLatencyParameters(device, nil)
		params.Input.Channels = 1
		params.SampleRate = float64(sampleRate)
		params.FramesPerBuffer = bufferSize

		stream, err := portaudio.OpenStream(params, c.processAudio)
		if err != nil {
			lastErr = err
			continue
		}
		if err := stream.Start(); err != nil {
			stream.Close()
			lastErr = err
			continue
		}
		c.stream = stream
		return nil
	}
	c.releasePortAudio()
	if lastErr != nil {
		return fmt.Errorf("open input stream: %w", lastErr)
	}
	return fmt.Errorf("no input device available")
}

// releasePortAudio drops this instance's PortAudio reference (once).
func (c *PortAudioCapture) releasePortAudio() {
	if c.inited {
		c.inited = false
		terminatePortAudio()
	}
}

// processAudio is the PortAudio callback for capture.
func (c *PortAudioCapture) processAudio(in []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.closed) == 1 {
		return
	}

	for _, sample := range in {
		c.buffer[c.writePos] = sample
		c.writePos = (c.writePos + 1) % len(c.buffer)
		if c.writePos == c.readPos {
			atomic.StoreInt32(&c.overflow, 1)
			c.readPos = (c.readPos + 1) % len(c.buffer)
		}
	}
}

// Read reads audio samples from the capture buffer.
func (c *PortAudioCapture) Read(out []float32) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.closed) == 1 {
		return 0, fmt.Errorf("capture closed")
	}

	n := 0
	for i := 0; i < len(out); i++ {
		if c.readPos == c.writePos {
			break
		}
		out[i] = c.buffer[c.readPos]
		c.readPos = (c.readPos + 1) % len(c.buffer)
		n++
	}

	return n, nil
}

// Close closes the capture stream.
func (c *PortAudioCapture) Close() {
	atomic.StoreInt32(&c.closed, 1)

	c.mu.Lock()
	if c.stream != nil {
		c.stream.Stop()
		c.stream.Close()
		c.stream = nil
	}
	c.releasePortAudio()
	c.mu.Unlock()
}

// PortAudioPlayback implements PlaybackSink using PortAudio.
type PortAudioPlayback struct {
	stream     *portaudio.Stream
	buffer     []float32
	bufferSize int
	sampleRate int
	deviceName string
	channels   int
	inited     bool
	mu         sync.Mutex
	closed     int32
	readPos    int
	writePos   int
	underflow  int32
}

// NewPortAudioPlayback creates a new PortAudio playback sink. An empty
// deviceName selects the system default output device.
func NewPortAudioPlayback(sampleRate, bufferSize int, deviceName string) (*PortAudioPlayback, error) {
	return &PortAudioPlayback{
		bufferSize: bufferSize,
		sampleRate: sampleRate,
		deviceName: deviceName,
		channels:   1,
		buffer:     make([]float32, bufferSize*4),
	}, nil
}

// Open opens the playback stream.
func (p *PortAudioPlayback) Open(sampleRate, bufferSize int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sampleRate = sampleRate
	p.bufferSize = bufferSize
	p.buffer = make([]float32, bufferSize*4)
	p.readPos = 0
	p.writePos = 0
	atomic.StoreInt32(&p.underflow, 0)
	atomic.StoreInt32(&p.closed, 0)

	if err := ensurePortAudioInit(); err != nil {
		return fmt.Errorf("portaudio init: %w", err)
	}
	p.inited = true

	var lastErr error
	for _, device := range candidateDevices(p.deviceName, false) {
		channels := 1
		if device.MaxOutputChannels >= 2 {
			channels = 2
		}
		p.channels = channels

		params := portaudio.LowLatencyParameters(nil, device)
		params.Output.Channels = channels
		params.SampleRate = float64(sampleRate)
		params.FramesPerBuffer = bufferSize

		stream, err := portaudio.OpenStream(params, p.processAudio)
		if err != nil {
			lastErr = err
			continue
		}
		if err := stream.Start(); err != nil {
			stream.Close()
			lastErr = err
			continue
		}
		p.stream = stream
		return nil
	}
	p.releasePortAudio()
	if lastErr != nil {
		return fmt.Errorf("open output stream: %w", lastErr)
	}
	return fmt.Errorf("no output device available")
}

// releasePortAudio drops this instance's PortAudio reference (once).
func (p *PortAudioPlayback) releasePortAudio() {
	if p.inited {
		p.inited = false
		terminatePortAudio()
	}
}

// processAudio is the PortAudio callback for playback. The engine feeds mono
// samples; they are duplicated across all output channels (interleaved).
func (p *PortAudioPlayback) processAudio(out []float32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := p.channels
	if ch < 1 {
		ch = 1
	}

	if atomic.LoadInt32(&p.closed) == 1 {
		for i := range out {
			out[i] = 0
		}
		return
	}

	frames := len(out) / ch
	for f := 0; f < frames; f++ {
		var s float32
		if p.readPos == p.writePos {
			atomic.StoreInt32(&p.underflow, 1)
			s = 0
		} else {
			s = p.buffer[p.readPos]
			p.readPos = (p.readPos + 1) % len(p.buffer)
		}
		for c := 0; c < ch; c++ {
			out[f*ch+c] = s
		}
	}
}

// Buffered returns the number of mono samples currently queued for playback.
// It lets callers (e.g. Engine.PlayTest) pace their writes so the device never
// starves and the output stays glitch-free.
func (p *PortAudioPlayback) Buffered() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	occ := p.writePos - p.readPos
	if occ < 0 {
		occ += len(p.buffer)
	}
	return occ
}

// Write writes mono audio samples to the playback buffer.
func (p *PortAudioPlayback) Write(samples []float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.closed) == 1 {
		return fmt.Errorf("playback closed")
	}

	for _, s := range samples {
		p.buffer[p.writePos] = s
		p.writePos = (p.writePos + 1) % len(p.buffer)
		if p.writePos == p.readPos {
			p.readPos = (p.readPos + 1) % len(p.buffer)
		}
	}

	return nil
}

// Close closes the playback stream.
func (p *PortAudioPlayback) Close() {
	atomic.StoreInt32(&p.closed, 1)

	p.mu.Lock()
	if p.stream != nil {
		p.stream.Stop()
		p.stream.Close()
		p.stream = nil
	}
	p.releasePortAudio()
	p.mu.Unlock()
}

// terminatePortAudio decrements the shared reference count and, once the
// last device is closed, calls portaudio.Terminate exactly once.
func terminatePortAudio() {
	paRefMu.Lock()
	defer paRefMu.Unlock()
	if paRef > 0 {
		paRef--
	}
	if paRef == 0 && paInit {
		portaudio.Terminate()
		paInit = false
	}
}
