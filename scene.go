package main

import (
	"fmt"
	"math/rand"
	"snake/utils"
	"time"

	"github.com/fatih/color"
)

var score int32

type Scene struct {
	Width    int
	Height   int
	Border   map[string]struct{}      // 边界
	Obstacle map[string]struct{}      // 障碍
	Food     chan map[string]struct{} // 食物
	Snake    *Snake
	Speed    int // 越小越快
}

func InitScene(bw, bh, snakeLen, speed int) *Scene {
	seed := time.Now().Unix()
	border := make([][]int, 0)
	for i := 0; i < bw; i++ {
		for j := 0; j < bh; j++ {
			if i == 0 || i == bw-1 || j == 0 || j == bh-1 {
				border = append(border, []int{i, j})
			}
		}
	}

	snake := &Snake{
		Direction: right,
	}
	rand.Seed(seed)
	snake.Head = &Node{
		Loc: []int{rand.Intn(bw-2*snakeLen) + snakeLen, rand.Intn(bh / 3)},
	}

	pre := snake.Head
	for i := 1; i < snakeLen; i++ {
		node := &Node{}
		node.Loc = []int{pre.Loc[0], pre.Loc[1] - 1}
		pre.Next = node
		node.Pre = pre

		pre = node
	}
	snake.Tail = pre

	borderSet := make(map[string]struct{})
	for _, b := range border {
		borderSet[fmt.Sprintf(keyFormat, b[0], b[1])] = struct{}{}
	}

	foodChan := make(chan map[string]struct{}, 1)
	foodChan <- map[string]struct{}{"0,0": {}}

	return &Scene{
		Width:    bw,
		Height:   bh,
		Speed:    speed,
		Border:   borderSet,
		Snake:    snake,
		Food:     foodChan,
		Obstacle: make(map[string]struct{}),
	}
}

func (s *Scene) GenFood() {
	rand.Seed(time.Now().Unix())
	bodySet := s.Snake.GetBodySet()

	var key string
	for {
		// 1. 新果实不能出现在边上
		newFood := []int{rand.Intn(s.Width), rand.Intn(s.Height)}
		key = fmt.Sprintf(keyFormat, newFood[0], newFood[1])
		if _, ok := s.Border[key]; ok {
			continue
		}
		// 2. 新果实不能出现在蛇身体上
		if _, ok := bodySet[key]; ok {
			continue
		}
		// 3. 新果实不能出现在障碍物上
		if _, ok := s.Obstacle[key]; !ok {
			break
		}
	}
	foodSet := <-s.Food
	if len(foodSet) < 8 {
		foodSet[key] = struct{}{}
	}
	s.Food <- foodSet
}

func (s *Scene) GenObstacle() {
	rand.Seed(time.Now().Unix())
	bodySet := s.Snake.GetBodySet()

	foodSet := <-s.Food
	var key string
	for {
		obstacle := []int{rand.Intn(s.Width), rand.Intn(s.Height)}
		key = fmt.Sprintf(keyFormat, obstacle[0], obstacle[1])
		// 1. 不能在蛇的身上
		if _, ok := bodySet[key]; ok {
			continue
		}
		// 2. 不能在边界上
		if _, ok := s.Border[key]; ok {
			continue
		}
		// 3. 不允许一个坐标点重复生成障碍
		if _, ok := s.Obstacle[key]; ok {
			continue
		}
		// 4. 不能落在食物上
		if _, ok := foodSet[key]; !ok {
			break
		}

	}
	s.Food <- foodSet
	s.Obstacle[key] = struct{}{}
}

func (s *Scene) Render() {
	utils.ClearTerminal()

	snakeBodySet := s.Snake.GetBodySet()
	foodSet := <-s.Food
	fmt.Printf("-- score:%d\n", score)
	for i := 0; i < s.Width; i++ {
		for j := 0; j < s.Height; j++ {
			// 打印边界
			if i == 0 || j == 0 || i == s.Width-1 || j == s.Height-1 {
				if j == s.Height-1 {
					fmt.Println("#")
				} else {
					fmt.Print("#")
				}
			} else {
				key := fmt.Sprintf(keyFormat, i, j)
				_, isFood := foodSet[key]
				_, isSnake := snakeBodySet[key]
				_, isObstacle := s.Obstacle[key]

				if isFood {
					// 打印食物
					color.New(color.FgGreen).Print("o")
				} else if isSnake {
					// 打印蛇
					color.New(color.FgBlue).Print("*")
				} else if isObstacle {
					// 打印障碍物
					color.New(color.FgRed).Print("x")
				} else {
					fmt.Print(" ")
				}
			}
		}
	}

	s.Food <- foodSet
}
