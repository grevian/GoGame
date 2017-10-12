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
