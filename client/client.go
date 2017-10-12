package client

import (
	"fmt"
	"image/color"
	_ "image/png"

	log "github.com/Sirupsen/logrus"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

var mainLevel *Level
var usersCharacter *Character

func update(screen *ebiten.Image) error {

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		reset()
	}

	mainLevel.Draw(screen)
	usersCharacter.Update(mainLevel)
	usersCharacter.Draw(screen)

	msg := fmt.Sprintf("FPS: %0.2f, Grounded: %t, Forces: %v", ebiten.CurrentFPS(), usersCharacter.grounded, usersCharacter.forces)

	ebitenutil.DebugPrint(screen, msg)
	return nil
}

func reset() {
	mainLevel = NewLevel()
	usersCharacter = NewCharacter(color.RGBA{0, 255, 0, 128}, 140, 310)
}

func NewClient() {

	reset()

	if err := ebiten.Run(update, 800, 600, 2, "Little Platformer"); err != nil {
		log.WithError(err).Fatal("Ebiten Stopped unexpectedly!")
	}
}
