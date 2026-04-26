package audio

import (
	loommath "github.com/ujjwalvivek/loom/math"
)

type StepEvent struct {
	Step  int
	Patch SynthPatch
}

type Sequencer struct {
	BPM          float32
	Steps        int
	Pattern      []StepEvent
	Playing      bool
	
	currentStep  int
	timeAcc      float32
	stepDuration float32
	
	// Generative systems
	EuclideanOn  bool
	MarkovOn     bool
	MarkovSynth  *MarkovMelody
	lfsr         *loommath.LFSR
}

type MarkovMelody struct {
	Scale     []float32
	LastIndex int
}

func NewMarkovMelody() *MarkovMelody {
	return &MarkovMelody{
		Scale: []float32{
			261.63, // C4
			293.66, // D4
			329.63, // E4
			392.00, // G4
			440.00, // A4
			523.25, // C5
			587.33, // D5
			659.25, // E5
		},
		LastIndex: 0,
	}
}

func (m *MarkovMelody) NextFreq(lfsr *loommath.LFSR) float32 {
	r := lfsr.Next()
	// 45% chance to step up, 35% chance to step down, 20% chance to stay
	if r < 0.45 {
		m.LastIndex = (m.LastIndex + 1) % len(m.Scale)
	} else if r < 0.80 {
		m.LastIndex = (m.LastIndex - 1 + len(m.Scale)) % len(m.Scale)
	}
	return m.Scale[m.LastIndex]
}

func NewSequencer() *Sequencer {
	return &Sequencer{
		BPM:          120.0,
		Steps:        16,
		Pattern:      make([]StepEvent, 0, 16),
		stepDuration: 60.0 / (120.0 * 4.0), // default sixteenth notes
		MarkovSynth:  NewMarkovMelody(),
		lfsr:         loommath.NewLFSR(42),
	}
}

// GenerateEuclideanPattern distributes 'pulses' as evenly as possible over 'steps' steps.
func GenerateEuclideanPattern(pulses, steps int) []bool {
	pattern := make([]bool, steps)
	if pulses <= 0 || steps <= 0 {
		return pattern
	}
	if pulses >= steps {
		for i := range pattern {
			pattern[i] = true
		}
		return pattern
	}

	bucket := 0
	for i := 0; i < steps; i++ {
		bucket += pulses
		if bucket >= steps {
			bucket -= steps
			pattern[i] = true
		}
	}
	return pattern
}

// SetupEuclideanMelody configures the sequencer to generate patterns algorithmically on step ticks.
func (s *Sequencer) SetupEuclideanMelody(pulses, steps int, patch SynthPatch) {
	s.Pattern = s.Pattern[:0]
	s.Steps = steps
	
	euclPattern := GenerateEuclideanPattern(pulses, steps)
	for i, hit := range euclPattern {
		if hit {
			// Trigger notes
			p := patch
			s.Pattern = append(s.Pattern, StepEvent{
				Step:  i,
				Patch: p,
			})
		}
	}
	s.EuclideanOn = true
}

func (s *Sequencer) Start(bpm float32) {
	s.BPM = bpm
	s.stepDuration = 60.0 / (bpm * 4.0) // sixteenth notes pacing
	s.Playing = true
	s.currentStep = 0
	s.timeAcc = 0.0
}

func (s *Sequencer) Stop() {
	s.Playing = false
}

// Update ticks the transport timeline and triggers synth notes on the voice pool.
func (s *Sequencer) Update(dt float32, vp *VoicePool) {
	if !s.Playing {
		return
	}

	s.timeAcc += dt
	if s.timeAcc >= s.stepDuration {
		s.timeAcc -= s.stepDuration
		
		// Run current step events
		s.triggerStep(vp)
		
		// Move to next step
		s.currentStep = (s.currentStep + 1) % s.Steps
	}
}

func (s *Sequencer) triggerStep(vp *VoicePool) {
	for _, event := range s.Pattern {
		if event.Step == s.currentStep {
			p := event.Patch
			if s.MarkovOn {
				// Mutate note frequency algorithmically
				p.Freq = s.MarkovSynth.NextFreq(s.lfsr)
			}
			vp.Trigger(p, false, loommath.Vec2{}, 0, 0)
		}
	}
}
