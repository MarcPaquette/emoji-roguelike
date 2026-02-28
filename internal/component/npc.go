package component

import "emoji-roguelike/internal/ecs"

const CNPC ecs.ComponentType = 16

// NPCKind classifies what an NPC does when interacted with.
type NPCKind uint8

const (
	NPCKindDialogue NPCKind = 0 // townsfolk — speech marks on dialogue
	NPCKindHealer   NPCKind = 1 // heals player to full, repeatable
	NPCKindShop     NPCKind = 2 // opens shop modal
	NPCKindAnimal   NPCKind = 3 // flavor only — no speech marks
)

// NPC is a non-hostile, interactable entity with dialogue.
type NPC struct {
	Name  string
	Kind  NPCKind
	Lines []string // dialogue pool; a random line is shown on each interact
}

func (NPC) Type() ecs.ComponentType { return CNPC }
