package component

import (
	"emoji-rougelike/internal/ecs"

	"github.com/gdamore/tcell/v2"
)

const CRenderable ecs.ComponentType = 3

type Renderable struct {
	Glyph       string
	FGColor     tcell.Color
	BGColor     tcell.Color
	RenderOrder int
}

func (Renderable) Type() ecs.ComponentType { return CRenderable }
