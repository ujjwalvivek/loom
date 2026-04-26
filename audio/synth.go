package audio

import (
	"math"

	loommath "github.com/ujjwalvivek/loom/math"
)

type Waveform byte

const (
	WaveSine     Waveform = 0
	WaveTriangle Waveform = 1
	WaveSquare   Waveform = 2
	WaveSawtooth Waveform = 3
	WaveNoise    Waveform = 4
)

type ADSR struct {
	Attack  float32 // duration in seconds
	Decay   float32 // duration in seconds
	Sustain float32 // level (0.0 to 1.0)
	Release float32 // duration in seconds
}

type SynthPatch struct {
	Wave        Waveform
	Freq        float32
	Env         ADSR
	Volume      float32
	Pan         float32 // -1.0 (left) to 1.0 (right)
	VolumeGroup string  // "sfx" or "music" (defaults to sfx if empty)
}

type Voice struct {
	Active      bool
	Patch       SynthPatch
	Time        float32 // time in seconds since start
	ReleaseTime float32 // time in seconds when release started (-1.0 if not releasing)
	Phase       float32 // oscillator phase accumulator
	Lfsr        *loommath.LFSR
	
	// Spatial properties
	Spatial           bool
	Position          loommath.Vec2
	ReferenceDistance float32
	MaxDistance       float32
}

const MaxVoices = 32

type VoicePool struct {
	Voices [MaxVoices]Voice
}

func NewVoicePool() *VoicePool {
	vp := &VoicePool{}
	for i := 0; i < MaxVoices; i++ {
		vp.Voices[i].Lfsr = loommath.NewLFSR(uint16(i + 1))
	}
	return vp
}

func (vp *VoicePool) Trigger(patch SynthPatch, spatial bool, pos loommath.Vec2, refDist, maxDist float32) int {
	// Find inactive voice
	idx := -1
	for i := 0; i < MaxVoices; i++ {
		if !vp.Voices[i].Active {
			idx = i
			break
		}
	}

	// Voice stealing: if pool is full, steel the oldest voice
	if idx == -1 {
		oldestIdx := 0
		var maxTime float32 = -1.0
		for i := 0; i < MaxVoices; i++ {
			if vp.Voices[i].Time > maxTime {
				maxTime = vp.Voices[i].Time
				oldestIdx = i
			}
		}
		idx = oldestIdx
	}

	// Initialize voice
	v := &vp.Voices[idx]
	v.Active = true
	v.Patch = patch
	v.Time = 0.0
	v.ReleaseTime = -1.0
	v.Phase = 0.0
	v.Spatial = spatial
	v.Position = pos
	v.ReferenceDistance = refDist
	v.MaxDistance = maxDist

	return idx
}

func (vp *VoicePool) Stop(voiceIdx int) {
	if voiceIdx >= 0 && voiceIdx < MaxVoices {
		v := &vp.Voices[voiceIdx]
		if v.Active && v.ReleaseTime < 0 {
			v.ReleaseTime = v.Time
		}
	}
}

// GenerateSample computes the combined left and right stereo audio sample outputs for a single frame increment.
func (vp *VoicePool) GenerateSample(sampleRate float32, listenerPos loommath.Vec2, masterVol, sfxVol, musicVol float32) (float32, float32) {
	var outL, outR float32
	dt := 1.0 / sampleRate

	for i := 0; i < MaxVoices; i++ {
		v := &vp.Voices[i]
		if !v.Active {
			continue
		}

		// Calculate envelope level
		envLvl := calculateADSR(v)
		if v.Time > 0 && envLvl <= 0.0 {
			v.Active = false
			continue
		}

		// Calculate oscillator wave output
		var rawSample float32
		switch v.Patch.Wave {
		case WaveSine:
			rawSample = float32(math.Sin(2.0 * math.Pi * float64(v.Phase)))
		case WaveTriangle:
			if v.Phase < 0.5 {
				rawSample = 4.0*v.Phase - 1.0
			} else {
				rawSample = 3.0 - 4.0*v.Phase
			}
		case WaveSquare:
			if v.Phase < 0.5 {
				rawSample = 1.0
			} else {
				rawSample = -1.0
			}
		case WaveSawtooth:
			rawSample = 2.0*v.Phase - 1.0
		case WaveNoise:
			rawSample = v.Lfsr.Next()*2.0 - 1.0
		}

		// Accumulate phase
		v.Phase += v.Patch.Freq * dt
		if v.Phase >= 1.0 {
			v.Phase -= float32(int(v.Phase)) // keep phase in [0, 1)
		}

		// Apply volume category
		groupVol := sfxVol
		if v.Patch.VolumeGroup == "music" {
			groupVol = musicVol
		}

		// Apply envelope, initial volume, and group volume
		sampleVal := rawSample * envLvl * v.Patch.Volume * groupVol

		// Apply spatial audio volume attenuation and panning
		var pan float32 = v.Patch.Pan
		var distanceVolume float32 = 1.0

		if v.Spatial {
			dx := v.Position.X - listenerPos.X
			dy := v.Position.Y - listenerPos.Y
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

			// Panning based on angle/X difference
			if dist > 0 {
				pan = dx / dist // ranges from -1.0 (full left) to 1.0 (full right)
			} else {
				pan = 0.0
			}

			// Distance attenuation: linear clamping model
			if dist > v.ReferenceDistance {
				if dist >= v.MaxDistance {
					distanceVolume = 0.0
				} else {
					distanceVolume = (v.MaxDistance - dist) / (v.MaxDistance - v.ReferenceDistance)
				}
			}
		}

		sampleVal *= distanceVolume

		// Stereo split panning calculations
		leftGain := 1.0 - pan
		rightGain := 1.0 + pan

		// Normalize gains to maintain constant power
		gainSum := leftGain + rightGain
		if gainSum > 0 {
			leftGain /= gainSum
			rightGain /= gainSum
		}

		outL += sampleVal * leftGain * masterVol
		outR += sampleVal * rightGain * masterVol

		// Advance voice timeline
		v.Time += dt
	}

	return outL, outR
}

func calculateADSR(v *Voice) float32 {
	env := v.Patch.Env
	t := v.Time

	// If release has not started yet
	if v.ReleaseTime < 0 {
		if t < env.Attack {
			if env.Attack > 0 {
				return t / env.Attack
			}
			return 1.0
		}
		
		tDecay := t - env.Attack
		if tDecay < env.Decay {
			if env.Decay > 0 {
				return 1.0 - (1.0-env.Sustain)*(tDecay/env.Decay)
			}
			return env.Sustain
		}
		
		return env.Sustain
	}

	// Release phase
	tRel := t - v.ReleaseTime
	if tRel < env.Release {
		if env.Release > 0 {
			return env.Sustain * (1.0 - (tRel / env.Release))
		}
		return 0.0
	}

	return 0.0
}
