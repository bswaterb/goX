package stack

import (
	"testing"
)

func TestPushAndPop(t *testing.T) {
	stack := NewBLFStack()

	// 测试入栈
	valuesToPush := []any{1, 2, 3, "four", 5.0}

	for _, v := range valuesToPush {
		stack.Push(v)
	}

	// 测试出栈
	for i := len(valuesToPush) - 1; i >= 0; i-- {
		val, ok := stack.Pop()
		if !ok {
			t.Fatalf("Expected to pop value, but got nothing")
		}
		if val != valuesToPush[i] {
			t.Errorf("Expected %v, but got %v", valuesToPush[i], val)
		}
	}

	// 测试栈空
	_, ok := stack.Pop()
	if ok {
		t.Fatal("Expected to pop nothing from empty stack, but got a value")
	}
}

func TestPopEmptyStack(t *testing.T) {
	stack := NewBLFStack()

	// 尝试从空栈出栈
	val, ok := stack.Pop()
	if val != nil || ok {
		t.Fatalf("Expected nil and false, but got val: %v and ok: %v", val, ok)
	}
}

func TestInterleavedPushPop(t *testing.T) {
	stack := NewBLFStack()

	valuesToPush := []any{1, 2, 3}

	// 第一轮入栈
	stack.Push(valuesToPush[0])
	stack.Push(valuesToPush[1])

	// 第一轮出栈
	val1, ok1 := stack.Pop()
	if !ok1 || val1 != valuesToPush[1] {
		t.Errorf("Expected %v, but got %v", valuesToPush[1], val1)
	}

	// 第二轮入栈
	stack.Push(valuesToPush[2])

	// 第二轮出栈
	val2, ok2 := stack.Pop()
	if !ok2 || val2 != valuesToPush[2] {
		t.Errorf("Expected %v, but got %v", valuesToPush[2], val2)
	}

	// 最后出栈
	val3, ok3 := stack.Pop()
	if !ok3 || val3 != valuesToPush[0] {
		t.Errorf("Expected %v, but got %v", valuesToPush[0], val3)
	}
}
