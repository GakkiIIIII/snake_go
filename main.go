package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/eiannone/keyboard"
)

const (
	// 设置方向键
	up    = 'w'
	down  = 's'
	left  = 'a'
	right = 'd'

	keyFormat = "%d,%d"
)

var score int32

type Snake struct {
	Head      *Node
	Tail      *Node
	Direction rune // 运动方向
}

type Node struct {
	Pre  *Node
	Next *Node
	Loc  []int // 坐标
}

type Scene struct {
	Width    int
	Height   int
	Border   map[string]struct{}      // 边界
	Obstacle map[string]struct{}      // 障碍
	Food     chan map[string]struct{} // 食物
	Snake    *Snake
	Speed    int // 越小越快
}

func (s *Snake) Move(border, obstacle map[string]struct{}) bool {
	oldHead := s.Head
	newHead := &Node{}
	switch s.Direction {
	case up:
		newHead.Loc = []int{oldHead.Loc[0] - 1, oldHead.Loc[1]}
	case down:
		newHead.Loc = []int{oldHead.Loc[0] + 1, oldHead.Loc[1]}
	case left:
		newHead.Loc = []int{oldHead.Loc[0], oldHead.Loc[1] - 1}
	case right:
		newHead.Loc = []int{oldHead.Loc[0], oldHead.Loc[1] + 1}
	}

	node := s.Tail
	for node.Pre != nil {
		node.Loc = node.Pre.Loc
		node = node.Pre
	}

	oldHead.Next.Pre = newHead
	newHead.Next = oldHead.Next
	oldHead.Next = nil
	s.Head = newHead

	// 检测碰撞到边界
	if _, ok := border[fmt.Sprintf(keyFormat, newHead.Loc[0], newHead.Loc[1])]; ok {
		return true
	}

	// 检测碰撞到障碍物
	if _, ok := obstacle[fmt.Sprintf(keyFormat, newHead.Loc[0], newHead.Loc[1])]; ok {
		return true
	}

	// 检测咬到自己
	node = newHead.Next
	for node != nil {
		if node.Loc[0] == newHead.Loc[0] && node.Loc[1] == newHead.Loc[1] {
			return true
		}
		node = node.Next
	}

	return false
}

func (s *Snake) Eat(before map[string]struct{}) map[string]struct{} {
	after := make(map[string]struct{})
	for food := range before {
		if food == fmt.Sprintf(keyFormat, s.Head.Loc[0], s.Head.Loc[1]) {
			continue
		}

		after[food] = struct{}{}
	}

	tail := s.Tail
	tailPre := tail.Pre

	newTail := &Node{}
	newTail.Pre = tail
	tail.Next = newTail
	s.Tail = newTail

	if tail.Loc[0] == tailPre.Loc[0] {
		if tail.Loc[1] < tailPre.Loc[1] {
			newTail.Loc = []int{tail.Loc[0], tail.Loc[1] - 1}
		} else {
			newTail.Loc = []int{tail.Loc[0], tail.Loc[1] + 1}
		}
	}

	if tail.Loc[1] == tailPre.Loc[1] {
		if tail.Loc[0] < tailPre.Loc[0] {
			newTail.Loc = []int{tail.Loc[0] - 1, tail.Loc[1]}
		} else {
			newTail.Loc = []int{tail.Loc[0] + 1, tail.Loc[1]}
		}
	}

	score++
	return after
}

func (s *Snake) Control() {
	dirNow := s.Direction
	input, _, _ := keyboard.GetSingleKey()
	// 方向冲突
	if (dirNow == left && input == right) || (dirNow == up && input == down) ||
		(dirNow == right && input == left) || (dirNow == down && input == up) {
		return
	}

	if input != up && input != left && input != down && input != right {
		return
	}

	s.Direction = input
}

func (s *Snake) GetBodySet() map[string]struct{} {
	bodySet := make(map[string]struct{})
	node := s.Head
	for node != nil {
		bodySet[fmt.Sprintf(keyFormat, node.Loc[0], node.Loc[1])] = struct{}{}
		node = node.Next
	}

	return bodySet
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

// GenFood 刷新食物
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
	if len(foodSet) < 5 {
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
	snakeBodySet := s.Snake.GetBodySet()

	foodSet := <-s.Food
	fmt.Printf("score:%d\n", score)
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
					fmt.Print("o")
				} else if isSnake {
					// 打印蛇
					fmt.Print("*")
				} else if isObstacle {
					// 打印障碍物
					fmt.Print("x")
				} else {
					fmt.Print(" ")
				}
			}
		}
	}

	s.Food <- foodSet
}

func main() {
	fmt.Println("请选择难度：1.简单; 2.中等; 3.困难")
	var difficult int
	fmt.Scanln(&difficult)
	var (
		speed       int
		genObstacle bool
	)

	switch difficult {
	case 1:
		speed = 180
	case 2:
		speed = 160
	case 3:
		speed = 140
		genObstacle = true
	default:
		fmt.Println("无效输入，拜拜...")
		return
	}

	scene := InitScene(23, 40, 3, speed)
	snake := scene.Snake

	// 控制移动
	go func() {
		for {
			snake.Control()
			scene.Render()
		}
	}()

	// 生成障碍物
	go func(genObstacle bool) {
		if !genObstacle {
			return
		}

		v := 20000 // 最初20s生成一个障碍物
		vMin := 3000
		for {
			time.Sleep(time.Duration(v) * time.Millisecond)
			scene.GenObstacle()
			v -= 1500
			if v < vMin {
				v = vMin
			}
		}
	}(genObstacle)

	go func() {
		v := 2000 // 最初2s生成一个食物
		vMax := 8000
		for {
			time.Sleep(time.Duration(v) * time.Millisecond)
			scene.GenFood()
			v += 500
			if v > vMax {
				v = vMax
			}
		}
	}()

	for {
		// 让蛇一直动起来
		isOver := snake.Move(scene.Border, scene.Obstacle)
		scene.Render()
		if isOver {
			fmt.Println("************** Game Over **************")
			return
		}

		head := snake.Head.Loc

		before := <-scene.Food
		if _, ok := before[fmt.Sprintf(keyFormat, head[0], head[1])]; ok {
			scene.Food <- snake.Eat(before)
			scene.Render()
		} else {
			scene.Food <- before
		}

		time.Sleep(time.Millisecond * time.Duration(scene.Speed))
	}
}
