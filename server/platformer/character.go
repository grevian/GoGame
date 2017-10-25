package platformer

import (
	"math"

	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/grevian/GoGame/common/platformer"
)

type User struct {
	name string
}

type Position struct {
	X, Y float64
}

type Physics struct {
	grounded      bool
	forces        [2]float64
	width, height float64
}

type Character struct {
	updateLock    sync.RWMutex
	doublejumping bool
	clientUpdates pb.GameServer_ConnectServer
	user          *User
	position      Position
	physics       Physics
}

const GRAVITY_ACCELLERATION = 0.08
const GRAVITY_MAX_FORCE = 9.2
const HORIZONTAL_ACCELLERATION = 0.7
const HORIZONTAL_MAX_FORCE = 6

func NewNetworkCharacter(x, y float64, user *User, server pb.GameServer_ConnectServer) *Character {
	// TODO Attach name/command stream/etc.
	WIDTH, HEIGHT := 20, 20

	c := &Character{
		doublejumping: false,
		clientUpdates: server,
		user:          user,
		position:      Position{x, y},
		physics:       Physics{false, [2]float64{0, 0}, float64(WIDTH), float64(HEIGHT)},
	}

	// Start processing character updates from the position & command streams immediately
	c.Start()

	return c
}

func (c *Character) Start() {
	entry := log.WithField("user", c.user.name)

	// Process Position updates
	go func() {
		for {
			clientUpdate, err := c.clientUpdates.Recv()
			if err != nil {
				if err == io.EOF {
					entry.Info("Client update stream closed, stopping processing")
					return
				}
				entry.WithError(err).Error("Error processing client update")
				return
			} else {
				c.updateLock.Lock()
				switch update := clientUpdate.UpdatePayload.(type) {
				case *pb.ClientUpdate_C:
					c.commandUpdate(update.C)
				case *pb.ClientUpdate_P:
					c.positionUpdate(update.P)
				default:
					entry.WithField("type", clientUpdate.String()).Error("Unknown update type encountered")
				}
				c.updateLock.Unlock()
			}
		}
	}()
}

func (c *Character) positionUpdate(position *pb.Position) {
	ACCEPTABLE_SLEW := 150.0
	skew_x := math.Abs(math.Abs(c.position.X) - math.Abs(float64(position.X)))
	skew_y := math.Abs(math.Abs(c.position.Y) - math.Abs(float64(position.Y)))
	if skew_x > ACCEPTABLE_SLEW || skew_y > ACCEPTABLE_SLEW {
		log.WithFields(log.Fields{
			"allowable_skew": ACCEPTABLE_SLEW,
			"actual_skew_x":  skew_x,
			"actual_skew_y":  skew_y,
			"user":           c.user.name,
		}).Error("Position Update skewed too far, rejecting update and correcting client")
		return
	}

	// Apply position update force vectors
	// TODO Probably worth sanity checking these too
	c.physics.forces[0] = float64(position.VelX)
	c.physics.forces[1] = float64(position.VelY)
}

func (c *Character) commandUpdate(command *pb.Command) {
	log.WithField("Command", command.Command.String()).Info("Processing Command")
	switch command.Command {
	case pb.Command_QUIT:
		break
	case pb.Command_JOINED:
		// Reset the character position
		c.position.X = 140
		c.position.Y = 310
		c.physics.forces[0] = 0
		c.physics.forces[1] = 0
		break
	case pb.Command_JUMP:
		if c.physics.grounded {
			c.physics.forces[1] = -(GRAVITY_MAX_FORCE / 2)
		} else if !c.doublejumping {
			// We get a second jump, slightly reduced in power
			c.doublejumping = true
			c.physics.forces[1] = -GRAVITY_MAX_FORCE / 3
		}
	default:
		log.WithField("Command", command.Command.String()).Error("Unknown command")
	}
}

func (c *Character) ServerUpdate(update *pb.ServerUpdate) {
	// Send updates from the server to a player

	// TODO Should we preemptively apply them to users on the server too? or Queue them to apply in Tick? probably
	err := c.clientUpdates.Send(update)
	if err != nil {
		log.WithFields(log.Fields{
			"originatingUser": update.UserIdentifier, // TODO Look up the username
			"destinationUser": c.user.name,
			"update":          update.String(),
		}).WithError(err).Error("Could not send update to player")
	}
}

func (c *Character) Tick(level *Level, tickDuration time.Duration) []*pb.ServerUpdate {
	updates := []*pb.ServerUpdate{}

	// Ensure no other character updates can apply while we're ticking
	c.updateLock.Lock()
	defer c.updateLock.Unlock()

	updates = append(updates, c.updatePhysics(level, tickDuration)...)

	// TODO Apply other command updates

	return updates
}

// Perform Physics calculation alongside clients and validate their positions
func (c *Character) updatePhysics(level *Level, tickDuration time.Duration) []*pb.ServerUpdate {
	var previousPosition Position
	previousPosition.X = c.position.X
	previousPosition.Y = c.position.Y

	// If we're moving/falling, Test if we've landed on any platforms
	if !c.physics.grounded {
		for _, platform := range level.platforms {
			if (c.position.Y+c.physics.height >= platform.Y && c.position.Y <= platform.Y+10) && // Character bottom lines up around the platform
				(c.position.X >= platform.X && c.position.X <= platform.X+100) { // Character is over top of the platform
				c.physics.grounded = true
				c.doublejumping = false // Reset any jumps
				c.physics.forces[1] = 0
				break
			}
		}
	}

	// If we didn't find ourselves on a platform, we're moving
	if !c.physics.grounded {
		// Apply Gravity to the Y forces
		c.physics.forces[1] = math.Min(c.physics.forces[1]+GRAVITY_ACCELLERATION, GRAVITY_MAX_FORCE)

		// Apply wind resistance to the X axis forces
		if c.physics.forces[0] > 0 {
			c.physics.forces[0] -= HORIZONTAL_ACCELLERATION / 6
		} else if c.physics.forces[0] < 0 {
			c.physics.forces[0] += HORIZONTAL_ACCELLERATION / 6
		}
	} else {
		// We're standing on a platform, so apply friction to X axis forces
		if c.physics.forces[0] > 0 {
			c.physics.forces[0] -= HORIZONTAL_ACCELLERATION / 2
		} else if c.physics.forces[0] < 0 {
			c.physics.forces[0] += HORIZONTAL_ACCELLERATION / 2
		}
	}

	// Clamp small values to zero so floating point errors don't make you slide around forever
	c.physics.forces[0] = clamp(c.physics.forces[0], 0.2)
	c.physics.forces[1] = clamp(c.physics.forces[1], 0.02)

	// Apply the accumulated horizontal forces to the character, if necessary
	if c.physics.forces[0] != 0 {
		c.physics.grounded = false
		c.position.X += c.physics.forces[0]
		// Correct for level boundaries
		c.position.X = math.Max(0, c.position.X)
		c.position.X = math.Min(level.width-c.physics.width, c.position.X)
	}

	// Apply accumulated Vertical force if necessary
	if c.physics.forces[1] != 0 {
		c.physics.grounded = false
		c.position.Y += c.physics.forces[1]
		// Correct for level boundaries
		c.position.Y = math.Max(0, c.position.Y)
		c.position.Y = math.Min(level.height-c.physics.height, c.position.Y)
	}

	// If required, send the position update out over the wire to the client and all other clients
	if previousPosition.X != c.position.X || previousPosition.Y != c.position.Y {
		positionUpdate := &pb.ServerUpdate{
			UpdatePayload: &pb.ServerUpdate_P{
				P: &pb.Position{
					X: float32(c.position.X),
					Y: float32(c.position.Y),
				},
			},
		}

		// TODO Perhaps we might return collisions here too with Command updates?
		return []*pb.ServerUpdate{positionUpdate}
	}

	return nil
}

// Return the input value unless it is below a certain range, in which case return 0
func clamp(value float64, clampRange float64) float64 {
	if math.Abs(value) <= clampRange {
		return 0
	}
	return value
}
