package ecs

import (
	"encoding/json"
	"sync"
)

type EntityRecord struct {
	Archetype *Archetype
	Index     int
}

type World struct {
	mu         sync.Mutex
	entities   []EntityRecord // Index is Entity ID
	freeList   []Entity
	archetypes map[ComponentMask]*Archetype
	systems    []System
}

func NewWorld() *World {
	w := &World{
		entities:   make([]EntityRecord, 1), // Reserve 0 as NullEntity
		freeList:   make([]Entity, 0),
		archetypes: make(map[ComponentMask]*Archetype),
		systems:    make([]System, 0),
	}
	return w
}

func (w *World) NewEntity() Entity {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	var id Entity
	if len(w.freeList) > 0 {
		id = w.freeList[len(w.freeList)-1]
		w.freeList = w.freeList[:len(w.freeList)-1]
		w.entities[id] = EntityRecord{}
	} else {
		id = Entity(len(w.entities))
		w.entities = append(w.entities, EntityRecord{})
	}
	return id
}

func (w *World) Destroy(entity Entity) {
	if entity == NullEntity || int(entity) >= len(w.entities) {
		return
	}
	record := w.entities[entity]
	if record.Archetype != nil {
		swapped := record.Archetype.Remove(record.Index)
		if swapped != NullEntity {
			// Update the index of the swapped entity
			w.entities[swapped].Index = record.Index
		}
	}
	w.entities[entity] = EntityRecord{}
	w.freeList = append(w.freeList, entity)
}

func (w *World) Add(entity Entity, t ComponentType, val interface{}) {
	if entity == NullEntity || int(entity) >= len(w.entities) {
		return
	}

	record := w.entities[entity]
	var oldMask ComponentMask = 0
	var oldArch *Archetype = nil
	if record.Archetype != nil {
		oldMask = record.Archetype.Mask
		oldArch = record.Archetype
	}

	newMask := oldMask.Set(t)
	if newMask == oldMask {
		// Entity already has this component, just update its value
		setComponentValue(record.Archetype, record.Index, t, val)
		return
	}

	// Migrate to new archetype
	newArch := w.getOrCreateArchetype(newMask)
	newIdx := newArch.Add(entity)

	if oldArch != nil {
		// Copy existing components
		for _, compType := range oldArch.Types {
			CopyComponent(oldArch, record.Index, newArch, newIdx, compType)
		}

		// Swap and pop from old archetype
		swapped := oldArch.Remove(record.Index)
		if swapped != NullEntity {
			w.entities[swapped].Index = record.Index
		}
	}

	// Set the new component value
	setComponentValue(newArch, newIdx, t, val)

	// Update entity record
	w.entities[entity] = EntityRecord{
		Archetype: newArch,
		Index:     newIdx,
	}
}

func (w *World) Remove(entity Entity, t ComponentType) {
	if entity == NullEntity || int(entity) >= len(w.entities) {
		return
	}

	record := w.entities[entity]
	if record.Archetype == nil || !record.Archetype.Mask.Has(t) {
		return
	}

	oldArch := record.Archetype
	newMask := oldArch.Mask.Clear(t)

	// Swap and pop from old archetype
	var newArch *Archetype = nil
	var newIdx int

	if newMask != 0 {
		newArch = w.getOrCreateArchetype(newMask)
		newIdx = newArch.Add(entity)

		// Copy components except the removed one
		for _, compType := range newArch.Types {
			CopyComponent(oldArch, record.Index, newArch, newIdx, compType)
		}
	}

	swapped := oldArch.Remove(record.Index)
	if swapped != NullEntity {
		w.entities[swapped].Index = record.Index
	}

	// Update entity record
	w.entities[entity] = EntityRecord{
		Archetype: newArch,
		Index:     newIdx,
	}
}

func (w *World) Get(entity Entity, t ComponentType) interface{} {
	if entity == NullEntity || int(entity) >= len(w.entities) {
		return nil
	}
	record := w.entities[entity]
	if record.Archetype == nil || !record.Archetype.Mask.Has(t) {
		return nil
	}
	slicePtr := record.Archetype.GetSlice(t)
	return getComponentFromSlice(slicePtr, record.Index, t)
}

func (w *World) getOrCreateArchetype(mask ComponentMask) *Archetype {
	if arch, ok := w.archetypes[mask]; ok {
		return arch
	}

	// Collect component types from the mask bits
	var types []ComponentType
	for i := ComponentType(0); i < 64; i++ {
		if mask.Has(i) {
			types = append(types, i)
		}
	}

	arch := NewArchetype(mask, types)
	w.archetypes[mask] = arch
	return arch
}

func setComponentValue(a *Archetype, idx int, t ComponentType, val interface{}) {
	if val == nil {
		return
	}
	switch t {
	case TypePosition:
		s := a.Components[t].(*[]Position)
		(*s)[idx] = val.(Position)
	case TypeVelocity:
		s := a.Components[t].(*[]Velocity)
		(*s)[idx] = val.(Velocity)
	case TypeSprite:
		s := a.Components[t].(*[]Sprite)
		(*s)[idx] = val.(Sprite)
	case TypePhysics:
		s := a.Components[t].(*[]PhysicsBody)
		(*s)[idx] = val.(PhysicsBody)
	case TypeParent:
		s := a.Components[t].(*[]Parent)
		(*s)[idx] = val.(Parent)
	case TypeChildren:
		s := a.Components[t].(*[]Children)
		(*s)[idx] = val.(Children)
	}
}

func getComponentFromSlice(slicePtr interface{}, idx int, t ComponentType) interface{} {
	switch t {
	case TypePosition:
		s := slicePtr.(*[]Position)
		return &(*s)[idx]
	case TypeVelocity:
		s := slicePtr.(*[]Velocity)
		return &(*s)[idx]
	case TypeSprite:
		s := slicePtr.(*[]Sprite)
		return &(*s)[idx]
	case TypePhysics:
		s := slicePtr.(*[]PhysicsBody)
		return &(*s)[idx]
	case TypeParent:
		s := slicePtr.(*[]Parent)
		return &(*s)[idx]
	case TypeChildren:
		s := slicePtr.(*[]Children)
		return &(*s)[idx]
	}
	return nil
}

// Registry Serialization Support
type SerializedEntity struct {
	ID       Entity          `json:"id"`
	Position *Position       `json:"position,omitempty"`
	Velocity *Velocity       `json:"velocity,omitempty"`
	Sprite   *Sprite         `json:"sprite,omitempty"`
	Physics  *PhysicsBody    `json:"physics,omitempty"`
	Parent   *Parent         `json:"parent,omitempty"`
	Children *Children       `json:"children,omitempty"`
}

func (w *World) Serialize() ([]byte, error) {
	var list []SerializedEntity

	for id := Entity(1); id < Entity(len(w.entities)); id++ {
		record := w.entities[id]
		if record.Archetype == nil {
			continue
		}

		se := SerializedEntity{ID: id}
		if record.Archetype.Mask.Has(TypePosition) {
			se.Position = w.Get(id, TypePosition).(*Position)
		}
		if record.Archetype.Mask.Has(TypeVelocity) {
			se.Velocity = w.Get(id, TypeVelocity).(*Velocity)
		}
		if record.Archetype.Mask.Has(TypeSprite) {
			se.Sprite = w.Get(id, TypeSprite).(*Sprite)
		}
		if record.Archetype.Mask.Has(TypePhysics) {
			se.Physics = w.Get(id, TypePhysics).(*PhysicsBody)
		}
		if record.Archetype.Mask.Has(TypeParent) {
			se.Parent = w.Get(id, TypeParent).(*Parent)
		}
		if record.Archetype.Mask.Has(TypeChildren) {
			se.Children = w.Get(id, TypeChildren).(*Children)
		}
		list = append(list, se)
	}

	return json.Marshal(list)
}

func (w *World) Deserialize(data []byte) error {
	var list []SerializedEntity
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	// Reset current World state
	w.entities = make([]EntityRecord, 1)
	w.freeList = make([]Entity, 0)
	w.archetypes = make(map[ComponentMask]*Archetype)

	for _, se := range list {
		// Ensure entities capacity
		for len(w.entities) <= int(se.ID) {
			w.entities = append(w.entities, EntityRecord{})
		}

		if se.Position != nil {
			w.Add(se.ID, TypePosition, *se.Position)
		}
		if se.Velocity != nil {
			w.Add(se.ID, TypeVelocity, *se.Velocity)
		}
		if se.Sprite != nil {
			w.Add(se.ID, TypeSprite, *se.Sprite)
		}
		if se.Physics != nil {
			w.Add(se.ID, TypePhysics, *se.Physics)
		}
		if se.Parent != nil {
			w.Add(se.ID, TypeParent, *se.Parent)
		}
		if se.Children != nil {
			w.Add(se.ID, TypeChildren, *se.Children)
		}
	}
	return nil
}

func (w *World) RegisterSystem(sys System) {
	w.systems = append(w.systems, sys)
}

func (w *World) RunSystems(dt float32) {
	// Schedule and run the scheduled concurrency batches
	cb := NewCommandBuffer()
	
	// Create dependency groups (for this simple execution, we run batch scheduled)
	// We'll execute them batched or sequentially, merging command buffers at the end of each batch
	// Let's implement SystemGraph scheduling inside system.go!
	RunSystemGraph(w, cb, dt)
}
