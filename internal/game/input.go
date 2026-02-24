package game

import "github.com/gdamore/tcell/v2"

// Action represents a player-requested game action.
type Action uint8

const (
	ActionNone Action = iota
	ActionMoveN
	ActionMoveS
	ActionMoveE
	ActionMoveW
	ActionMoveNE
	ActionMoveNW
	ActionMoveSE
	ActionMoveSW
	ActionWait
	ActionPickup
	ActionInventory
	ActionDescend
	ActionAscend
	ActionQuit
	ActionSpecialAbility
)

// keyToAction maps a tcell key event to a game action.
func keyToAction(ev *tcell.EventKey) Action {
	// Named keys.
	switch ev.Key() {
	case tcell.KeyUp:
		return ActionMoveN
	case tcell.KeyDown:
		return ActionMoveS
	case tcell.KeyRight:
		return ActionMoveE
	case tcell.KeyLeft:
		return ActionMoveW
	case tcell.KeyEscape:
		return ActionQuit
	}

	// Rune keys.
	switch ev.Rune() {
	case 'k', 'K':
		return ActionMoveN
	case 'j', 'J':
		return ActionMoveS
	case 'l', 'L':
		return ActionMoveE
	case 'h', 'H':
		return ActionMoveW
	case 'y', 'Y':
		return ActionMoveNW
	case 'u', 'U':
		return ActionMoveNE
	case 'b', 'B':
		return ActionMoveSW
	case 'n', 'N':
		return ActionMoveSE
	case '.':
		return ActionWait
	case ',':
		return ActionPickup
	case 'i', 'I':
		return ActionInventory
	case '>':
		return ActionDescend
	case '<':
		return ActionAscend
	case 'q', 'Q':
		return ActionQuit
	case 'z', 'Z':
		return ActionSpecialAbility
	}
	return ActionNone
}

// actionToDelta converts a movement action to (dx, dy).
func actionToDelta(a Action) (int, int) {
	switch a {
	case ActionMoveN:
		return 0, -1
	case ActionMoveS:
		return 0, 1
	case ActionMoveE:
		return 1, 0
	case ActionMoveW:
		return -1, 0
	case ActionMoveNE:
		return 1, -1
	case ActionMoveNW:
		return -1, -1
	case ActionMoveSE:
		return 1, 1
	case ActionMoveSW:
		return -1, 1
	}
	return 0, 0
}
