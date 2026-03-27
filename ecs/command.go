package ecs

import (
	"sync"
)

type CmdType byte

const (
	CmdDestroy          CmdType = 0
	CmdAddComponent     CmdType = 1
	CmdRemoveComponent  CmdType = 2
)

type EcsCommand struct {
	Type          CmdType
	Entity        Entity
	ComponentType ComponentType
	Value         interface{}
}

type CommandBuffer struct {
	mu       sync.Mutex
	commands []EcsCommand
}

func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]EcsCommand, 0, 32),
	}
}

// Add schedules a component addition.
func (cb *CommandBuffer) Add(entity Entity, t ComponentType, val interface{}) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.commands = append(cb.commands, EcsCommand{
		Type:          CmdAddComponent,
		Entity:        entity,
		ComponentType: t,
		Value:         val,
	})
}

// Remove schedules a component removal.
func (cb *CommandBuffer) Remove(entity Entity, t ComponentType) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.commands = append(cb.commands, EcsCommand{
		Type:          CmdRemoveComponent,
		Entity:        entity,
		ComponentType: t,
	})
}

// Destroy schedules an entity destruction.
func (cb *CommandBuffer) Destroy(entity Entity) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.commands = append(cb.commands, EcsCommand{
		Type:   CmdDestroy,
		Entity: entity,
	})
}

// Merge flushes and executes all recorded structural operations sequentially on the World.
func (cb *CommandBuffer) Merge(world *World) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for _, cmd := range cb.commands {
		switch cmd.Type {
		case CmdDestroy:
			world.Destroy(cmd.Entity)
		case CmdAddComponent:
			world.Add(cmd.Entity, cmd.ComponentType, cmd.Value)
		case CmdRemoveComponent:
			world.Remove(cmd.Entity, cmd.ComponentType)
		}
	}

	// Reset list to reuse capacity
	cb.commands = cb.commands[:0]
}
