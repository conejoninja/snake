package main

import (
	"image/color"
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/tinyfont"
	"github.com/conejoninja/snake/fonts"

	"tinygo.org/x/drivers/shifter"
	"tinygo.org/x/drivers/st7735"
)

const (
	BLACK = iota
	SNAKE
	APPLE
	TEXT
)

const (
	START = iota
	PLAY
	GAMEOVER
)

const (
	WIDTH  = 16
	HEIGHT = 13
)

type Snake struct {
	body      [208][2]int16
	length    int16
	direction int16
}

type Game struct {
	display        st7735.Device
	colors         []color.RGBA
	snake          Snake
	appleX, appleY int16
	status         uint8
}

func main() {
	game := Game{
		colors: []color.RGBA{
			color.RGBA{0, 0, 0, 255},
			color.RGBA{0, 200, 0, 255},
			color.RGBA{250, 0, 0, 255},
			color.RGBA{160, 160, 160, 255},
		},
		snake: Snake{
			body: [208][2]int16{
				{0, 3},
				{0, 2},
				{0, 1},
			},
			length:    3,
			direction: 3,
		},
		appleX: -1,
		appleY: -1,
	}
	scoreStr := []byte("SCORE: 123")

	machine.SPI1.Configure(machine.SPIConfig{
		SCK:       machine.SPI1_SCK_PIN,
		MOSI:      machine.SPI1_MOSI_PIN,
		MISO:      machine.SPI1_MISO_PIN,
		Frequency: 8000000,
	})

	game.display = st7735.New(machine.SPI1, machine.TFT_RST, machine.TFT_DC, machine.TFT_CS, machine.TFT_LITE)
	game.display.Configure(st7735.Config{
		Rotation: st7735.ROTATION_90,
	})

	buttons := shifter.New(shifter.EIGHT_BITS, machine.BUTTON_LATCH, machine.BUTTON_CLK, machine.BUTTON_OUT)
	buttons.Configure()

	game.display.FillScreen(game.colors[BLACK])
	for {
		switch game.status {
		case START:
			game.display.FillScreen(game.colors[BLACK])

			tinyfont.WriteLine(&game.display, &fonts.Bold24pt7b, 0, 50, []byte("SNAKE"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 8, 100, []byte("Press START"), game.colors[TEXT])

			time.Sleep(2 * time.Second)
			for game.status == START {
				pressed, _ := buttons.Read8Input()
				if pressed&machine.BUTTON_START_MASK > 0 {
					game.status = PLAY
				}
			}
			break
		case GAMEOVER:
			game.display.FillScreen(game.colors[BLACK])

			scoreStr[7] = 48 + uint8((game.snake.length-3)/100)
			scoreStr[8] = 48 + uint8(((game.snake.length-3)/10)%10)
			scoreStr[9] = 48 + uint8((game.snake.length-3)%10)

			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 8, 50, []byte("GAME OVER"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 8, 100, []byte("Press START"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &tinyfont.TomThumb, 50, 120, scoreStr, game.colors[TEXT])

			time.Sleep(2 * time.Second)
			for game.status == GAMEOVER {
				pressed, _ := buttons.Read8Input()
				if pressed&machine.BUTTON_START_MASK > 0 {
					game.status = START
				}
			}
			break
		case PLAY:
			game.display.FillScreen(game.colors[BLACK])
			game.snake.body = [208][2]int16{
				{0, 3},
				{0, 2},
				{0, 1},
			}
			game.snake.length = 3
			game.snake.direction = 3
			game.drawSnake()
			game.createApple()
			time.Sleep(2000 * time.Millisecond)
			for game.status == PLAY {

				// Faster
				pressed, _ := buttons.Read8Input()
				if pressed&machine.BUTTON_LEFT_MASK > 0 {
					if game.snake.direction != 3 {
						game.snake.direction = 0
					}
				}
				if pressed&machine.BUTTON_UP_MASK > 0 {
					if game.snake.direction != 2 {
						game.snake.direction = 1
					}
				}
				if pressed&machine.BUTTON_DOWN_MASK > 0 {
					if game.snake.direction != 1 {
						game.snake.direction = 2
					}
				}
				if pressed&machine.BUTTON_RIGHT_MASK > 0 {
					if game.snake.direction != 0 {
						game.snake.direction = 3
					}
				}
				game.moveSnake()
				time.Sleep(100 * time.Millisecond)
			}

			break
		}
	}
}

func (g *Game) collisionWithSnake(x, y int16) bool {
	for i := int16(0); i < g.snake.length; i++ {
		if x == g.snake.body[i][0] && y == g.snake.body[i][1] {
			return true
		}
	}
	return false
}

func (g *Game) createApple() {
	g.appleX = int16(rand.Int31n(16))
	g.appleY = int16(rand.Int31n(13))
	for g.collisionWithSnake(g.appleX, g.appleY) {
		g.appleX = int16(rand.Int31n(16))
		g.appleY = int16(rand.Int31n(13))
	}
	g.drawSnakePartial(g.appleX, g.appleY, g.colors[APPLE])
}

func (g *Game) moveSnake() {
	x := g.snake.body[0][0]
	y := g.snake.body[0][1]

	switch g.snake.direction {
	case 0:
		x--
		break
	case 1:
		y--
		break
	case 2:
		y++
		break
	case 3:
		x++
		break
	}
	if x >= WIDTH {
		x = 0
	}
	if x < 0 {
		x = WIDTH - 1
	}
	if y >= HEIGHT {
		y = 0
	}
	if y < 0 {
		y = HEIGHT - 1
	}

	if g.collisionWithSnake(x, y) {
		g.status = GAMEOVER
	}

	// draw head
	g.drawSnakePartial(x, y, g.colors[SNAKE])
	if x == g.appleX && y == g.appleY {
		g.snake.length++
		g.createApple()
	} else {
		// remove tail
		g.drawSnakePartial(g.snake.body[g.snake.length-1][0], g.snake.body[g.snake.length-1][1], g.colors[BLACK])
	}
	for i := g.snake.length - 1; i > 0; i-- {
		g.snake.body[i][0] = g.snake.body[i-1][0]
		g.snake.body[i][1] = g.snake.body[i-1][1]
	}
	g.snake.body[0][0] = x
	g.snake.body[0][1] = y
}

func (g *Game) drawSnake() {
	for i := int16(0); i < g.snake.length; i++ {
		g.drawSnakePartial(g.snake.body[i][0], g.snake.body[i][1], g.colors[SNAKE])
	}
}

func (g *Game) drawSnakePartial(x, y int16, c color.RGBA) {
	modY := int16(9)
	if y == 12 {
		modY = 8
	}
	g.display.FillRectangle(10*x, 10*y, 9, modY, c)
}
