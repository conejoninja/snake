package main

import (
	"image/color"
	"machine"
	"math/rand"
	"time"

	"github.com/conejoninja/snake/fonts"
	"tinygo.org/x/tinyfont"

	"tinygo.org/x/drivers/st7735"
)

const (
	BLACK = iota
	SNAKE
	APPLE
	TEXT
)

const (
	LEFT = iota
	UP
	RIGHT
	DOWN
)

const (
	START = iota
	PLAY
	GAMEOVER
)

const (
	WIDTH  = 13
	HEIGHT = 16
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

	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		MOSI:      machine.SPI0_MOSI_PIN,
		MISO:      machine.SPI0_MISO_PIN,
		Frequency: 8000000,
	})

	game.display = st7735.New(machine.SPI0, machine.D13, machine.D7, machine.D5, machine.A5)
	game.display.Configure(st7735.Config{
		Rotation:     st7735.NO_ROTATION,
		RowOffset:    1,
		ColumnOffset: 2,
	})

	game.display.FillScreen(game.colors[BLACK])

	buttons := [4]machine.Pin{
		machine.D11,
		machine.D12,
		machine.D9,
		machine.D10,
	}
	b := 0
	for b = 0; b < 4; b++ {
		buttons[b].Configure(machine.PinConfig{Mode: machine.PinInput})
	}

	game.status = START
	for {
		switch game.status {
		case START:
			game.display.FillScreen(game.colors[BLACK])

			tinyfont.WriteLine(&game.display, &fonts.Bold12pt7b, 26, 50, []byte("SNAKE"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 8, 100, []byte("Press any"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 14, 124, []byte("BUTTON"), game.colors[TEXT])

			time.Sleep(2 * time.Second)
			for game.status == START {
				for b = 0; b < 4; b++ {
					if !buttons[b].Get() {
						game.status = PLAY
						break
					}
				}
			}
			break
		case GAMEOVER:
			game.display.FillScreen(game.colors[BLACK])

			scoreStr[7] = 48 + uint8((game.snake.length-3)/100)
			scoreStr[8] = 48 + uint8(((game.snake.length-3)/10)%10)
			scoreStr[9] = 48 + uint8((game.snake.length-3)%10)

			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 8, 50, []byte("GAME OVER"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &fonts.Regular12pt7b, 8, 100, []byte("Press ANY"), game.colors[TEXT])
			tinyfont.WriteLine(&game.display, &tinyfont.TomThumb, 50, 120, scoreStr, game.colors[TEXT])

			time.Sleep(2 * time.Second)
			for game.status == GAMEOVER {
				for b = 0; b < 4; b++ {
					if !buttons[b].Get() {
						game.status = START
						break
					}
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

				if !buttons[LEFT].Get() {
					if game.snake.direction != 3 {
						game.snake.direction = 0
					}
				}
				if !buttons[UP].Get() {
					if game.snake.direction != 2 {
						game.snake.direction = 1
					}
				}
				if !buttons[DOWN].Get() {
					if game.snake.direction != 1 {
						game.snake.direction = 2
					}
				}
				if !buttons[RIGHT].Get() {
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
	g.appleX = int16(rand.Int31n(WIDTH))
	g.appleY = int16(rand.Int31n(HEIGHT))
	for g.collisionWithSnake(g.appleX, g.appleY) {
		g.appleX = int16(rand.Int31n(WIDTH))
		g.appleY = int16(rand.Int31n(HEIGHT))
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
	modX := int16(9)
	if x == 12 {
		modX = 8
	}
	g.display.FillRectangle(10*x, 10*y, modX, 9, c)
}
