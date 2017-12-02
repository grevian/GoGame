package service

type Platform struct {
	Position
	friction int
}

type Level struct {
	platforms     []*Platform
	width, height float64
}

func NewPlatform(x, y float64) *Platform {
	return &Platform{
		Position{x, y}, 2,
	}
}

func NewLevel() *Level {
	return &Level{
		platforms: []*Platform{
			NewPlatform(20, 150),
			NewPlatform(390, 250),
			NewPlatform(45, 500),
			NewPlatform(120, 360),
		},
		width:  800,
		height: 600,
	}
}
