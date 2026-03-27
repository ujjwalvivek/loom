package ecs

import (
	"testing"
)

type MoveSystem struct{}

func (m *MoveSystem) Reads() []ComponentType {
	return []ComponentType{TypeVelocity}
}

func (m *MoveSystem) Writes() []ComponentType {
	return []ComponentType{TypePosition}
}

func (m *MoveSystem) Execute(w *World, cb *CommandBuffer, dt float32) {
	q := w.Query(TypePosition, TypeVelocity)
	for q.Next() {
		positions := *q.GetSlice(TypePosition).(*[]Position)
		velocities := *q.GetSlice(TypeVelocity).(*[]Velocity)
		idx := q.Index()
		positions[idx].X += velocities[idx].X * dt
		positions[idx].Y += velocities[idx].Y * dt
	}
}

func TestECSBasic(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	w.Add(e, TypePosition, Position{X: 10, Y: 20})
	w.Add(e, TypeVelocity, Velocity{X: 1, Y: 2})

	// Run systems
	w.RegisterSystem(&MoveSystem{})
	w.RunSystems(1.0)

	pos := w.Get(e, TypePosition).(*Position)
	if pos.X != 11 || pos.Y != 22 {
		t.Errorf("Expected position (11, 22), got (%f, %f)", pos.X, pos.Y)
	}

	// Test serialization
	data, err := w.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	w2 := NewWorld()
	if err := w2.Deserialize(data); err != nil {
		t.Fatal(err)
	}

	pos2 := w2.Get(e, TypePosition).(*Position)
	if pos2.X != 11 || pos2.Y != 22 {
		t.Errorf("After deserialization: expected position (11, 22), got (%f, %f)", pos2.X, pos2.Y)
	}
}
