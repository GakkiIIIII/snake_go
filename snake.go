package main

import (
	"fmt"

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
