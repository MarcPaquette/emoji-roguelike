package ecs

// EntityID uniquely identifies an entity in the world.
type EntityID uint64

// NilEntity is the zero value — no valid entity has this ID.
const NilEntity EntityID = 0

// ComponentType is a small integer key used to store/retrieve components.
type ComponentType uint8

// Component is implemented by every data struct stored in the world.
type Component interface {
	Type() ComponentType
}
