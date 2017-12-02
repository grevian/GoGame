package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"
)

var PLATFORM_DEFAULT_COLOR = color.RGBA{255, 0, 0, 128}

type Platform struct {
	image *ebiten.Image
	op    *ebiten.DrawImageOptions
	Position
	friction int
}

type Level struct {
	platforms     []*Platform
	width, height float64
}

func NewPlatform(platformColor color.RGBA, x, y float64) *Platform {
	image, _ := ebiten.NewImage(100, 10, ebiten.FilterLinear)
	image.Fill(platformColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Reset()
	op.GeoM.Translate(x, y)

	return &Platform{
		image, op, Position{x, y}, 2,
	}
}

func (p *Platform) Draw(screen *ebiten.Image) {
	screen.DrawImage(p.image, p.op)
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
