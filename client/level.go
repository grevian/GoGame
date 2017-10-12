package client

import (
	"github.com/hajimehoshi/ebiten"
)

type Level struct {
	platforms     []*Platform
	width, height float64
}

func NewLevel() *Level {
	return &Level{
		platforms: []*Platform{
			NewPlatform(PLATFORM_DEFAULT_COLOR, 20, 150),
			NewPlatform(PLATFORM_DEFAULT_COLOR, 390, 250),
			NewPlatform(PLATFORM_DEFAULT_COLOR, 45, 500),
			NewPlatform(PLATFORM_DEFAULT_COLOR, 120, 360),
		},
		width:  800,
		height: 600,
	}
}

func (l *Level) Draw(screen *ebiten.Image) {
	for _, p := range mainLevel.platforms {
		p.Draw(screen)
	}
}
