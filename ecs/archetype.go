package ecs

import (
	"github.com/ujjwalvivek/loom/physics"
)

type Entity uint32

const NullEntity Entity = 0

type ComponentType uint16

const (
	TypePosition ComponentType = 0
	TypeVelocity ComponentType = 1
	TypeSprite   ComponentType = 2
	TypePhysics  ComponentType = 3
	TypeParent   ComponentType = 4
	TypeChildren ComponentType = 5
)

// Concrete component structs
type Position struct {
	X, Y float32
}

type Velocity struct {
	X, Y float32
}

type Sprite struct {
	TextureID uint32
	U0, V0, U1, V1 float32
	FlipX, FlipY   bool
	Angle          float32
}

type PhysicsBody struct {
	Handle physics.BodyHandle
}

type Parent struct {
	Entity Entity
}

type Children struct {
	List []Entity
}

type ComponentMask uint64

func (m ComponentMask) Has(t ComponentType) bool {
	return (m & (1 << t)) != 0
}

func (m ComponentMask) Set(t ComponentType) ComponentMask {
	return m | (1 << t)
}

func (m ComponentMask) Clear(t ComponentType) ComponentMask {
	return m & ^(1 << t)
}

type Archetype struct {
	Mask       ComponentMask
	Types      []ComponentType
	Entities   []Entity
	Components map[ComponentType]interface{} // Maps ComponentType to pointer to type slice, e.g. *[]Position
}

func NewArchetype(mask ComponentMask, types []ComponentType) *Archetype {
	arch := &Archetype{
		Mask:       mask,
		Types:      types,
		Entities:   make([]Entity, 0, 16),
		Components: make(map[ComponentType]interface{}),
	}

	// Initialize component slices dynamically using reflection on setup only
	for _, t := range types {
		switch t {
		case TypePosition:
			slice := make([]Position, 0, 16)
			arch.Components[t] = &slice
		case TypeVelocity:
			slice := make([]Velocity, 0, 16)
			arch.Components[t] = &slice
		case TypeSprite:
			slice := make([]Sprite, 0, 16)
			arch.Components[t] = &slice
		case TypePhysics:
			slice := make([]PhysicsBody, 0, 16)
			arch.Components[t] = &slice
		case TypeParent:
			slice := make([]Parent, 0, 16)
			arch.Components[t] = &slice
		case TypeChildren:
			slice := make([]Children, 0, 16)
			arch.Components[t] = &slice
		}
	}

	return arch
}

// Add appends an entity and default components to the archetype.
// Returns the index of the newly added entity.
func (a *Archetype) Add(entity Entity) int {
	a.Entities = append(a.Entities, entity)
	
	// Grow each component slice by appending a zero value
	for t, slicePtr := range a.Components {
		growSlice(slicePtr, t)
	}

	return len(a.Entities) - 1
}

// Remove deletes an entity at index using O(1) swap-and-pop.
// Returns the entity that was swapped to the deleted index (if any, else NullEntity).
func (a *Archetype) Remove(index int) Entity {
	lastIdx := len(a.Entities) - 1

	var swappedEntity Entity = NullEntity
	if index < lastIdx {
		// Swap entity ID
		swappedEntity = a.Entities[lastIdx]
		a.Entities[index] = swappedEntity

		// Swap components
		for t, slicePtr := range a.Components {
			swapAndPopSlice(slicePtr, index, lastIdx, t)
		}
	} else {
		// Just pop components
		for t, slicePtr := range a.Components {
			popSlice(slicePtr, t)
		}
	}

	a.Entities = a.Entities[:lastIdx]
	return swappedEntity
}

// Reflection helpers for dynamic slice allocations (run ONLY during structural mutations, never on hot paths)
func growSlice(slicePtr interface{}, t ComponentType) {
	switch t {
	case TypePosition:
		p := slicePtr.(*[]Position)
		*p = append(*p, Position{})
	case TypeVelocity:
		p := slicePtr.(*[]Velocity)
		*p = append(*p, Velocity{})
	case TypeSprite:
		p := slicePtr.(*[]Sprite)
		*p = append(*p, Sprite{})
	case TypePhysics:
		p := slicePtr.(*[]PhysicsBody)
		*p = append(*p, PhysicsBody{})
	case TypeParent:
		p := slicePtr.(*[]Parent)
		*p = append(*p, Parent{})
	case TypeChildren:
		p := slicePtr.(*[]Children)
		*p = append(*p, Children{List: make([]Entity, 0, 4)})
	}
}

func swapAndPopSlice(slicePtr interface{}, index, lastIdx int, t ComponentType) {
	switch t {
	case TypePosition:
		s := slicePtr.(*[]Position)
		(*s)[index] = (*s)[lastIdx]
		*s = (*s)[:lastIdx]
	case TypeVelocity:
		s := slicePtr.(*[]Velocity)
		(*s)[index] = (*s)[lastIdx]
		*s = (*s)[:lastIdx]
	case TypeSprite:
		s := slicePtr.(*[]Sprite)
		(*s)[index] = (*s)[lastIdx]
		*s = (*s)[:lastIdx]
	case TypePhysics:
		s := slicePtr.(*[]PhysicsBody)
		(*s)[index] = (*s)[lastIdx]
		*s = (*s)[:lastIdx]
	case TypeParent:
		s := slicePtr.(*[]Parent)
		(*s)[index] = (*s)[lastIdx]
		*s = (*s)[:lastIdx]
	case TypeChildren:
		s := slicePtr.(*[]Children)
		(*s)[index] = (*s)[lastIdx]
		*s = (*s)[:lastIdx]
	}
}

func popSlice(slicePtr interface{}, t ComponentType) {
	switch t {
	case TypePosition:
		s := slicePtr.(*[]Position)
		*s = (*s)[:len(*s)-1]
	case TypeVelocity:
		s := slicePtr.(*[]Velocity)
		*s = (*s)[:len(*s)-1]
	case TypeSprite:
		s := slicePtr.(*[]Sprite)
		*s = (*s)[:len(*s)-1]
	case TypePhysics:
		s := slicePtr.(*[]PhysicsBody)
		*s = (*s)[:len(*s)-1]
	case TypeParent:
		s := slicePtr.(*[]Parent)
		*s = (*s)[:len(*s)-1]
	case TypeChildren:
		s := slicePtr.(*[]Children)
		*s = (*s)[:len(*s)-1]
	}
}

func (a *Archetype) GetSlice(t ComponentType) interface{} {
	return a.Components[t]
}

// Helper to copy component values between different archetypes during structural migration
func CopyComponent(srcArch *Archetype, srcIdx int, dstArch *Archetype, dstIdx int, t ComponentType) {
	switch t {
	case TypePosition:
		src := srcArch.Components[t].(*[]Position)
		dst := dstArch.Components[t].(*[]Position)
		(*dst)[dstIdx] = (*src)[srcIdx]
	case TypeVelocity:
		src := srcArch.Components[t].(*[]Velocity)
		dst := dstArch.Components[t].(*[]Velocity)
		(*dst)[dstIdx] = (*src)[srcIdx]
	case TypeSprite:
		src := srcArch.Components[t].(*[]Sprite)
		dst := dstArch.Components[t].(*[]Sprite)
		(*dst)[dstIdx] = (*src)[srcIdx]
	case TypePhysics:
		src := srcArch.Components[t].(*[]PhysicsBody)
		dst := dstArch.Components[t].(*[]PhysicsBody)
		(*dst)[dstIdx] = (*src)[srcIdx]
	case TypeParent:
		src := srcArch.Components[t].(*[]Parent)
		dst := dstArch.Components[t].(*[]Parent)
		(*dst)[dstIdx] = (*src)[srcIdx]
	case TypeChildren:
		src := srcArch.Components[t].(*[]Children)
		dst := dstArch.Components[t].(*[]Children)
		(*dst)[dstIdx] = (*src)[srcIdx]
	}
}
