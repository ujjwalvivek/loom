package ecs

import (
	"sync"

	loommath "github.com/ujjwalvivek/loom/math"
)

type System interface {
	Reads() []ComponentType
	Writes() []ComponentType
	Execute(world *World, cb *CommandBuffer, dt float32)
}

// BuildSystemBatches schedules non-conflicting systems to run in parallel batches
func BuildSystemBatches(systems []System) [][]System {
	var batches [][]System
	assigned := make([]bool, len(systems))
	assignedCount := 0

	for assignedCount < len(systems) {
		var currentBatch []System
		for i, sys := range systems {
			if assigned[i] {
				continue
			}

			// Check conflict with existing systems in current batch
			conflict := false
			for _, other := range currentBatch {
				if hasConflict(sys, other) {
					conflict = true
					break
				}
			}

			if !conflict {
				currentBatch = append(currentBatch, sys)
				assigned[i] = true
				assignedCount++
			}
		}
		batches = append(batches, currentBatch)
	}

	return batches
}

func hasConflict(a, b System) bool {
	for _, wA := range a.Writes() {
		for _, wB := range b.Writes() {
			if wA == wB {
				return true
			}
		}
	}
	for _, wA := range a.Writes() {
		for _, rB := range b.Reads() {
			if wA == rB {
				return true
			}
		}
	}
	for _, rA := range a.Reads() {
		for _, wB := range b.Writes() {
			if rA == wB {
				return true
			}
		}
	}
	return false
}

// RunSystemGraph runs systems grouped by conflict-free batches in parallel goroutines.
func RunSystemGraph(w *World, cb *CommandBuffer, dt float32) {
	w.mu.Lock()
	systemsCopy := make([]System, len(w.systems))
	copy(systemsCopy, w.systems)
	w.mu.Unlock()

	batches := BuildSystemBatches(systemsCopy)

	for _, batch := range batches {
		var wg sync.WaitGroup
		for _, sys := range batch {
			wg.Add(1)
			go func(s System) {
				defer wg.Done()
				s.Execute(w, cb, dt)
			}(sys)
		}
		wg.Wait()
		
		// Flush commands to update archetypes before the next batch runs
		cb.Merge(w)
	}
}

// GetWorldPosition recursively calculates an entity's absolute position by traversing parent chains.
func GetWorldPosition(world *World, entity Entity) loommath.Vec2 {
	posVal := world.Get(entity, TypePosition)
	if posVal == nil {
		return loommath.Vec2{}
	}
	pos := posVal.(*Position)

	parentVal := world.Get(entity, TypeParent)
	if parentVal == nil {
		return loommath.Vec2{X: pos.X, Y: pos.Y}
	}
	parent := parentVal.(*Parent)

	parentPos := GetWorldPosition(world, parent.Entity)
	return loommath.Vec2{
		X: parentPos.X + pos.X,
		Y: parentPos.Y + pos.Y,
	}
}
