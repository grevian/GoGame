package client

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten"
)

type Physics struct {
	grounded      bool
	forces        [2]float64
	width, height float64
}

type Character struct {
	image         *ebiten.Image
	op            *ebiten.DrawImageOptions
	doublejumping bool
	Position
	Physics
}

const GRAVITY_ACCELLERATION = 0.08
const GRAVITY_MAX_FORCE = 9.2
const HORIZONTAL_ACCELLERATION = 0.7
const HORIZONTAL_MAX_FORCE = 6

var KeySpace_debounce = false

func NewCharacter(characterColor color.RGBA, x, y float64) *Character {
	WIDTH, HEIGHT := 20, 20

	image, _ := ebiten.NewImage(WIDTH, HEIGHT, ebiten.FilterLinear)
	image.Fill(characterColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Reset()
	op.GeoM.Translate(x, y)

	return &Character{
		image,
		op,
		false,
		Position{x, y},
		Physics{false, [2]float64{0, 0}, float64(WIDTH), float64(HEIGHT)},
	}
}

func (c *Character) Draw(screen *ebiten.Image) {
	c.op.GeoM.Reset()
	c.op.GeoM.Translate(c.X, c.Y)
	screen.DrawImage(c.image, c.op)
}

func (c *Character) Update(level *Level) {

	// If we're moving/falling, Test if we've landed on any platforms
	if !c.grounded {
		for _, platform := range level.platforms {
			if (c.Y+c.height >= platform.Y && c.Y <= platform.Y+10) && // Character bottom lines up around the platform
				(c.X >= platform.X && c.X <= platform.X+100) { // Character is over top of the platform
				c.grounded = true
				c.doublejumping = false // Reset any jumps
				c.forces[1] = 0
				break
			}
		}
	}

	// If we didn't find ourselves on a platform, we're moving
	if !c.grounded {
		// Apply Gravity to the Y forces
		c.forces[1] = math.Min(c.forces[1]+GRAVITY_ACCELLERATION, GRAVITY_MAX_FORCE)

		// Apply wind resistance to the X axis forces
		if c.forces[0] > 0 {
			c.forces[0] -= HORIZONTAL_ACCELLERATION / 6
		} else if c.forces[0] < 0 {
			c.forces[0] += HORIZONTAL_ACCELLERATION / 6
		}
	} else {
		// We're standing on a platform, so apply friction to X axis forces
		if c.forces[0] > 0 {
			c.forces[0] -= HORIZONTAL_ACCELLERATION / 2
		} else if c.forces[0] < 0 {
			c.forces[0] += HORIZONTAL_ACCELLERATION / 2
		}
	}

	// Debounce our spacebar (Don't register it again after it's been pressed, until it's been released)
	if !ebiten.IsKeyPressed(ebiten.KeySpace) && KeySpace_debounce {
		KeySpace_debounce = false
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) && !KeySpace_debounce {
		KeySpace_debounce = true
		if c.grounded {
			c.forces[1] = -(GRAVITY_MAX_FORCE / 2)
		} else if !c.doublejumping {
			// We get a second jump, slightly reduced in power
			c.doublejumping = true
			c.forces[1] = -GRAVITY_MAX_FORCE / 3
		}
	}

	// If the user is pressing left or right, apply forces to the X axis
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if c.grounded {
			c.forces[0] = math.Max(c.forces[0]-HORIZONTAL_ACCELLERATION, -HORIZONTAL_MAX_FORCE)
		} else {
			c.forces[0] = math.Max(c.forces[0]-HORIZONTAL_ACCELLERATION/2, -HORIZONTAL_MAX_FORCE)
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if c.grounded {
			c.forces[0] = math.Min(c.forces[0]+HORIZONTAL_ACCELLERATION, HORIZONTAL_MAX_FORCE)
		} else {
			c.forces[0] = math.Min(c.forces[0]+HORIZONTAL_ACCELLERATION/2, HORIZONTAL_MAX_FORCE)
		}
	}

	// Clamp small values to zero so floating point errors don't make you slide around forever
	c.forces[0] = clamp(c.forces[0], 0.2)
	c.forces[1] = clamp(c.forces[1], 0.02)

	// Apply the accumulated horizontal forces to the character, if necessary
	if c.forces[0] != 0 {
		c.grounded = false
		c.Position.X += c.forces[0]
		// Correct for level boundaries
		c.Position.X = math.Max(0, c.Position.X)
		c.Position.X = math.Min(level.width-c.width, c.Position.X)
	}

	// Apply accumulated Vertical force if necessary
	if c.forces[1] != 0 {
		c.grounded = false
		c.Position.Y += c.forces[1]
		// Correct for level boundaries
		c.Position.Y = math.Max(0, c.Position.Y)
		c.Position.Y = math.Min(level.height-c.height, c.Position.Y)
	}

}

// Return the input value unless it is below a certain range, in which case return 0
func clamp(value float64, clampRange float64) float64 {
	if math.Abs(value) <= clampRange {
		return 0
	}
	return value
}
