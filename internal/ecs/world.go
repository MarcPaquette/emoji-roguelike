package ecs

// World is the central entity registry and component store.
type World struct {
	nextID     EntityID
	alive      map[EntityID]bool
	components map[ComponentType]map[EntityID]Component
}

// NewWorld creates an empty World.
func NewWorld() *World {
	return &World{
		nextID:     1,
		alive:      make(map[EntityID]bool),
		components: make(map[ComponentType]map[EntityID]Component),
	}
}

// CreateEntity mints a new entity ID and marks it alive.
func (w *World) CreateEntity() EntityID {
	id := w.nextID
	w.nextID++
	w.alive[id] = true
	return id
}

// DestroyEntity marks the entity dead and removes all its components.
func (w *World) DestroyEntity(id EntityID) {
	if !w.alive[id] {
		return
	}
	w.alive[id] = false
	for _, store := range w.components {
		delete(store, id)
	}
}

// Alive reports whether the entity is alive.
func (w *World) Alive(id EntityID) bool {
	return w.alive[id]
}

// Add attaches a component to an entity.
func (w *World) Add(id EntityID, c Component) {
	t := c.Type()
	if w.components[t] == nil {
		w.components[t] = make(map[EntityID]Component)
	}
	w.components[t][id] = c
}

// Get returns the component of the given type for entity id, or nil.
func (w *World) Get(id EntityID, t ComponentType) Component {
	store := w.components[t]
	if store == nil {
		return nil
	}
	return store[id]
}

// Remove detaches a component from an entity.
func (w *World) Remove(id EntityID, t ComponentType) {
	if store := w.components[t]; store != nil {
		delete(store, id)
	}
}

// Has reports whether entity id has a component of the given type.
func (w *World) Has(id EntityID, t ComponentType) bool {
	return w.Get(id, t) != nil
}

// Query returns all alive entities that have every listed component type.
func (w *World) Query(types ...ComponentType) []EntityID {
	if len(types) == 0 {
		return nil
	}
	// Use the smallest store as the candidate set.
	smallest := types[0]
	for _, t := range types[1:] {
		if len(w.components[t]) < len(w.components[smallest]) {
			smallest = t
		}
	}
	store := w.components[smallest]
	if store == nil {
		return nil
	}
	var result []EntityID
	for id := range store {
		if !w.alive[id] {
			continue
		}
		match := true
		for _, t := range types {
			if t == smallest {
				continue
			}
			if !w.Has(id, t) {
				match = false
				break
			}
		}
		if match {
			result = append(result, id)
		}
	}
	return result
}
