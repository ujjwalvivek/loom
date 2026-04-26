package audio

import (
	"testing"
	"time"
)

func TestAudioSystemSampleGeneration(t *testing.T) {
	as := NewAudioSystem()
	// Wait a moment for readyChan (if any)
	time.Sleep(100 * time.Millisecond)

	// Define a test patch
	testPatch := SynthPatch{
		Wave:   WaveTriangle,
		Freq:   440.0,
		Volume: 0.5,
		Env: ADSR{
			Attack:  0.01,
			Decay:   0.02,
			Sustain: 0.5,
			Release: 0.05,
		},
	}

	// Trigger the test sound
	as.PlaySound(testPatch)

	// Manually process commands
	as.processCommands()

	// Verify a voice is active
	activeVoices := 0
	for _, v := range as.voicePool.Voices {
		if v.Active {
			activeVoices++
		}
	}
	if activeVoices == 0 {
		t.Fatal("Expected at least one active voice after triggering PlaySound")
	}

	// Read samples
	buf := make([]byte, 800)
	stream := &SynthStream{sys: as}
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Stream Read failed: %v", err)
	}
	if n != 800 {
		t.Fatalf("Expected to read 800 bytes, got %d", n)
	}

	// Verify samples are non-zero
	hasNonZero := false
	for i := 0; i < len(buf); i++ {
		if buf[i] != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("Generated audio buffer is completely silent (all zeros)")
	} else {
		t.Log("Audio generation test passed: buffer contains sound data")
	}
}
