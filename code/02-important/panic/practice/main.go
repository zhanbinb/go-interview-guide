package main

import "fmt"

func main() {
	DemoPanic()
}

func DemoPanic() {
	//场景1 数组越界
	tryRun("数组越界", func() {
		arr := []int{1, 2, 3}
		fmt.Println(arr[3])
	})

	//场景2 指针解引用
	tryRun("指针解引用", func() {
		var p *int = nil
		fmt.Println(*p)
	})

	//场景3 类型断言失败
	tryRun("类型断言失败", func() {
		var i interface{} = nil
		fmt.Println(i.(int))
	})
	tryRun("类型断言尝试", func() {
		var i interface{} = "hello"
		s, ok := i.(string)
		if !ok {
			panic("类型断言失败")
		}
		fmt.Println(s)
	})

	//场景4 关闭已经关闭的channel
	tryRun("关闭已经关闭的channel", func() {
		ch := make(chan int)
		close(ch)
		fmt.Println(ch)
		close(ch)
	})

	//场景5 向已经关闭的channel写入数据
	tryRun("向已经关闭的channel写入数据", func() {
		ch := make(chan int)
		close(ch)
		ch <- 1
	})

	//场景6 除数为0
	// tryRun("除数为0", func() {
	// 	fmt.Println(1 / 0)
	// })
}

func tryRun(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  💥 %s: panic 捕获 = %v\n", label, r)
		}
	}()
	fn()
}
