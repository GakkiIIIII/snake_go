package main

import (
	"fmt"
	"snake/utils"
	"time"
)

func main() {
	// 初始化屏幕刷新器
	utils.Init()

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
