package ecs

type QueryIterator struct {
	types      []ComponentType
	world      *World
	archetypes []*Archetype
	archIdx    int
	entityIdx  int
}

func (w *World) Query(types ...ComponentType) *QueryIterator {
	var mask ComponentMask
	for _, t := range types {
		mask = mask.Set(t)
	}

	var matched []*Archetype
	for _, arch := range w.archetypes {
		if (arch.Mask & mask) == mask {
			// Only include archetypes that contain entities
			if len(arch.Entities) > 0 {
				matched = append(matched, arch)
			}
		}
	}

	return &QueryIterator{
		types:      types,
		world:      w,
		archetypes: matched,
		archIdx:    0,
		entityIdx:  -1,
	}
}

func (qi *QueryIterator) Next() bool {
	if len(qi.archetypes) == 0 {
		return false
	}

	qi.entityIdx++
	for qi.archIdx < len(qi.archetypes) {
		arch := qi.archetypes[qi.archIdx]
		if qi.entityIdx < len(arch.Entities) {
			return true
		}
		// Move to next archetype
		qi.archIdx++
		qi.entityIdx = 0
	}

	return false
}

func (qi *QueryIterator) Entity() Entity {
	arch := qi.archetypes[qi.archIdx]
	return arch.Entities[qi.entityIdx]
}

// GetSlice retrieves the component slice pointer for the active archetype
func (qi *QueryIterator) GetSlice(t ComponentType) interface{} {
	arch := qi.archetypes[qi.archIdx]
	return arch.GetSlice(t)
}

// Index returns the entity index in the current archetype's slices
func (qi *QueryIterator) Index() int {
	return qi.entityIdx
}

// ArchetypeChanged returns true if Next() just transitioned to a new archetype
func (qi *QueryIterator) ArchetypeChanged() bool {
	return qi.entityIdx == 0
}
