package main

import (
	"github.com/ujjwalvivek/loom/audio"
	"github.com/ujjwalvivek/loom/engine"
)

// Audio asset configurations
var (
	JumpPatch = audio.SynthPatch{
		Wave:   audio.WaveTriangle,
		Freq:   330.0,
		Volume: 0.30,
		Env: audio.ADSR{
			Attack:  0.01,
			Decay:   0.22,
			Sustain: 0.0,
			Release: 0.0,
		},
	}

	CoinPatch = audio.SynthPatch{
		Wave:   audio.WaveSquare,
		Freq:   987.77, // B5 note
		Volume: 0.25,
		Env: audio.ADSR{
			Attack:  0.01,
			Decay:   0.35,
			Sustain: 0.0,
			Release: 0.0,
		},
	}

	BumpPatch = audio.SynthPatch{
		Wave:   audio.WaveNoise,
		Freq:   120.0,
		Volume: 0.35,
		Env: audio.ADSR{
			Attack:  0.005,
			Decay:   0.12,
			Sustain: 0.0,
			Release: 0.0,
		},
	}

	StompPatch = audio.SynthPatch{
		Wave:   audio.WaveSawtooth,
		Freq:   100.0,
		Volume: 0.30,
		Env: audio.ADSR{
			Attack:  0.005,
			Decay:   0.12,
			Sustain: 0.0,
			Release: 0.0,
		},
	}

	// Level Win Fanfare
	WinPatch = audio.SynthPatch{
		Wave:   audio.WaveSquare,
		Freq:   523.25, // C5
		Volume: 0.35,
		Env: audio.ADSR{
			Attack:  0.02,
			Decay:   2.0,
			Sustain: 0.0,
			Release: 0.0,
		},
	}
)

// Initialize the background procedural music and sequencer transport.
func SetupMarioAudio(ctx *engine.Context) {
	// Configure background music playing on a square synth patch for classic 8-bit chiptune
	bgmPatch := audio.SynthPatch{
		Wave:        audio.WaveSquare,
		Freq:        261.63, // C4 base
		Volume:      0.08,   // keep it balanced so it doesn't overpower SFX
		VolumeGroup: "music",
		Env: audio.ADSR{
			Attack:  0.01,
			Decay:   0.12, // Staccato decay
			Sustain: 0.0,  // deactivates when finished to prevent overlap screeches
			Release: 0.05,
		},
	}

	seq := ctx.Audio.Sequencer
	seq.Pattern = seq.Pattern[:0]
	seq.Steps = 64
	seq.MarkovOn = false
	seq.EuclideanOn = false

	addNote := func(step int, freq float32) {
		p := bgmPatch
		p.Freq = freq
		seq.Pattern = append(seq.Pattern, audio.StepEvent{
			Step:  step,
			Patch: p,
		})
	}

	// Mario Intro (Steps 0-31)
	addNote(0, 659.25)  // E5
	addNote(2, 659.25)  // E5
	addNote(6, 659.25)  // E5
	addNote(10, 523.25) // C5
	addNote(12, 659.25) // E5
	addNote(16, 783.99) // G5
	addNote(24, 392.00) // G4

	// Main Melody Loop (Steps 32-63)
	addNote(32, 523.25) // C5
	addNote(38, 392.00) // G4
	addNote(44, 329.63) // E4
	addNote(50, 440.00) // A4
	addNote(54, 493.88) // B4
	addNote(58, 466.16) // Bb4
	addNote(60, 440.00) // A4

	seq.Start(150.0) // 150 BPM for authentic theme tempo
}
