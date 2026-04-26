package audio

import (
	"encoding/binary"
	"log"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
	loommath "github.com/ujjwalvivek/loom/math"
)

type AudioCommandType byte

const (
	AudioPlaySound      AudioCommandType = 0
	AudioPlaySoundAt    AudioCommandType = 1
	AudioStopSound      AudioCommandType = 2
	AudioSetVolume      AudioCommandType = 3
	AudioSequencerStart AudioCommandType = 4
	AudioSequencerStop  AudioCommandType = 5
)

type AudioCommand struct {
	Type        AudioCommandType
	Patch       SynthPatch
	VoiceHandle int
	VolumeGroup string // "master", "sfx", "music"
	VolumeLevel float32
	Spatial     bool
	Position    loommath.Vec2
	RefDistance float32
	MaxDistance float32
	BPM         float32
}

type AudioSystem struct {
	otoContext *oto.Context
	player     *oto.Player
	commands   chan AudioCommand
	voicePool  *VoicePool
	Sequencer  *Sequencer

	MasterVolume float32
	SfxVolume    float32
	MusicVolume  float32

	ListenerPos loommath.Vec2
}

func NewAudioSystem() *AudioSystem {
	op := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		Format:       oto.FormatFloat32LE,
		BufferSize:   time.Millisecond * 5, // Ultra low latency buffer
	}

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		log.Println("Warning: Failed to initialize audio context:", err)
		return &AudioSystem{
			commands: make(chan AudioCommand, 256),
		}
	}

	as := &AudioSystem{
		otoContext:   otoCtx,
		commands:     make(chan AudioCommand, 256),
		voicePool:    NewVoicePool(),
		Sequencer:    NewSequencer(),
		MasterVolume: 1.0,
		SfxVolume:    1.0,
		MusicVolume:  0.8,
	}

	// Wait for audio context readiness in background to avoid blocking main thread
	// and initialize the player only after the context is fully ready
	go func() {
		<-readyChan
		stream := &SynthStream{sys: as}
		player := otoCtx.NewPlayer(stream)
		
		// Constrain internal buffer capacity to 4096 bytes (~11ms at 44.1kHz) 
		// to enforce low-latency execution of the SynthStream.
		player.SetBufferSize(4096)
		
		as.player = player
		player.Play()
	}()

	return as
}

func (as *AudioSystem) PlaySound(patch SynthPatch) {
	select {
	case as.commands <- AudioCommand{
		Type:  AudioPlaySound,
		Patch: patch,
	}:
	default:
		// Drop command if buffer is full to prevent freezing
	}
}

func (as *AudioSystem) PlaySoundAt(patch SynthPatch, pos loommath.Vec2, refDist, maxDist float32) {
	select {
	case as.commands <- AudioCommand{
		Type:        AudioPlaySoundAt,
		Patch:       patch,
		Position:    pos,
		RefDistance: refDist,
		MaxDistance: maxDist,
	}:
	default:
	}
}

func (as *AudioSystem) StartMusic(bpm float32) {
	select {
	case as.commands <- AudioCommand{
		Type: AudioSequencerStart,
		BPM:  bpm,
	}:
	default:
	}
}

func (as *AudioSystem) StopMusic() {
	select {
	case as.commands <- AudioCommand{
		Type: AudioSequencerStop,
	}:
	default:
	}
}

func (as *AudioSystem) StopSound(voiceHandle int) {
	select {
	case as.commands <- AudioCommand{
		Type:        AudioStopSound,
		VoiceHandle: voiceHandle,
	}:
	default:
	}
}

func (as *AudioSystem) SetVolume(group string, volume float32) {
	select {
	case as.commands <- AudioCommand{
		Type:        AudioSetVolume,
		VolumeGroup: group,
		VolumeLevel: volume,
	}:
	default:
	}
}

func (as *AudioSystem) SetListenerPos(pos loommath.Vec2) {
	// Thread-safe assignment
	as.ListenerPos = pos
}

func (as *AudioSystem) processCommands() {
	for {
		select {
		case cmd := <-as.commands:
			switch cmd.Type {
			case AudioPlaySound:
				as.voicePool.Trigger(cmd.Patch, false, loommath.Vec2{}, 0, 0)
			case AudioPlaySoundAt:
				as.voicePool.Trigger(cmd.Patch, true, cmd.Position, cmd.RefDistance, cmd.MaxDistance)
			case AudioStopSound:
				as.voicePool.Stop(cmd.VoiceHandle)
			case AudioSetVolume:
				switch cmd.VolumeGroup {
				case "master":
					as.MasterVolume = cmd.VolumeLevel
				case "sfx":
					as.SfxVolume = cmd.VolumeLevel
				case "music":
					as.MusicVolume = cmd.VolumeLevel
				}
			case AudioSequencerStart:
				as.Sequencer.Start(cmd.BPM)
			case AudioSequencerStop:
				as.Sequencer.Stop()
			}
		default:
			return
		}
	}
}

// SynthStream pulls audio frames from the VoicePool synthesis engine.
type SynthStream struct {
	sys *AudioSystem
}

func (s *SynthStream) Read(buf []byte) (int, error) {
	s.sys.processCommands()

	sampleRate := float32(44100)
	dtPerSample := 1.0 / sampleRate

	// Stereo Float32LE = 8 bytes per frame (4 bytes left, 4 bytes right)
	count := len(buf) / 8

	var maxVal float32
	for i := 0; i < count; i++ {
		// Advance sequencer by single sample time interval
		s.sys.Sequencer.Update(dtPerSample, s.sys.voicePool)

		left, right := s.sys.voicePool.GenerateSample(
			sampleRate,
			s.sys.ListenerPos,
			s.sys.MasterVolume,
			s.sys.SfxVolume,
			s.sys.MusicVolume,
		)

		// Hard clamp output to avoid clipping stutters
		if left < -1.0 {
			left = -1.0
		} else if left > 1.0 {
			left = 1.0
		}
		if right < -1.0 {
			right = -1.0
		} else if right > 1.0 {
			right = 1.0
		}

		absLeft := left
		if absLeft < 0 {
			absLeft = -absLeft
		}
		if absLeft > maxVal {
			maxVal = absLeft
		}
		absRight := right
		if absRight < 0 {
			absRight = -absRight
		}
		if absRight > maxVal {
			maxVal = absRight
		}

		// Convert float values to IEEE-754 bytes directly
		binary.LittleEndian.PutUint32(buf[i*8:i*8+4], math.Float32bits(left))
		binary.LittleEndian.PutUint32(buf[i*8+4:i*8+8], math.Float32bits(right))
	}

	return count * 8, nil
}
