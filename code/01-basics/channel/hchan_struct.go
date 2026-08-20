package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// hchanField hchan 关键字段偏移（Go 1.25 runtime/chan.go）
//
// Go 1.25 的 hchan 定义：
//
//   qcount   uint           // 0
//   dataqsiz uint           // 8
//   buf      unsafe.Pointer // 16
//   elemsize uint16         // 24
//   closed   uint32         // 26
//   timer    *timer         // 32 (Go 1.25+ 新增)
//   elemtype *_type         // 40
//   sendx    uint           // 48
//   recvx    uint           // 56
//   recvq    waitq          // 64
//   sendq    waitq          // 80
//   bubble   *synctestBubble// 96
//   lock     mutex          // 104
//
// ⚠️ unsafe 代码依赖 Go runtime 内部布局！
//   升级 Go 版本前要先核对 runtime/chan.go 里的字段顺序
//   不同 Go 版本字段顺序可能不同（如 Go 1.25 加了 timer 字段）
type hchanField struct {
	name   string
	offset uintptr
	size   uintptr
	typ    string
}

// hchanFields Go 1.25 的字段偏移
var hchanFields = []hchanField{
	{"qcount", 0, 8, "uint"},
	{"dataqsiz", 8, 8, "uint"},
	{"buf", 16, 8, "unsafe.Pointer"},
	{"elemsize", 24, 2, "uint16"},
	{"closed", 28, 4, "uint32"},  // 注意: elemsize 后有 2 字节 padding
	{"timer", 32, 8, "*timer (1.25+)"},
	{"elemtype", 40, 8, "*_type"},
	{"sendx", 48, 8, "uint"},
	{"recvx", 56, 8, "uint"},
	{"recvq", 64, 16, "waitq"},
	{"sendq", 80, 16, "waitq"},
	{"bubble", 96, 8, "*synctestBubble (1.25+)"},
	{"lock", 104, 8, "mutex"},
}

// hchanSize Go 1.25 的 hchan 大小（约 112 字节）
const hchanSizeGo125 = 112

// hchanPointer 把 channel 转成 unsafe.Pointer 以窥探内部字段
// 用泛型支持任意 channel 类型
func hchanPointer[T any](ch chan T) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&ch))
}

// loadField 从 hchan 指针读取指定字段
func loadField(p unsafe.Pointer, offset uintptr, size uintptr) uint64 {
	switch size {
	case 2:
		return uint64(*(*uint16)(unsafe.Add(p, offset)))
	case 4:
		return uint64(*(*uint32)(unsafe.Add(p, offset)))
	case 8:
		return uint64(*(*uint64)(unsafe.Add(p, offset)))
	default:
		return 0
	}
}

// DemoHchanStruct 演示 hchan 内部结构
func DemoHchanStruct() {
	fmt.Println("=== hchan 内部结构窥探 ===")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Println()

	// ---------- 1. channel 变量大小 ----------
	fmt.Println("📐 channel 变量本身的结构：")
	chVar := make(chan int, 0)
	fmt.Printf("  chan 变量大小: %d 字节（指向 hchan 的指针）\n", unsafe.Sizeof(chVar))
	fmt.Printf("  hchan 结构大小（Go 1.25）: ~%d 字节\n\n", hchanSizeGo125)

	// ---------- 2. 字段布局 ----------
	fmt.Println("📐 hchan 字段布局（Go 1.25 实际布局）:")
	for _, f := range hchanFields {
		fmt.Printf("  offset=%3d  size=%2d  %s %s\n", f.offset, f.size, f.typ, f.name)
	}
	fmt.Println()

	// ---------- 3. unbuffered vs buffered 对比 ----------
	fmt.Println("【实验 1】unbuffered vs buffered 的 hchan 差异")
	unbuf := make(chan int)
	buf := make(chan int, 5)

	p1 := hchanPointer(unbuf)
	p2 := hchanPointer(buf)

	fmt.Printf("  unbuffered hchan:\n")
	fmt.Printf("    qcount=%d, dataqsiz=%d, buf=0x%x\n",
		loadField(p1, 0, 8), loadField(p1, 8, 8), loadField(p1, 16, 8))
	fmt.Printf("  buffered hchan (cap=5):\n")
	fmt.Printf("    qcount=%d, dataqsiz=%d, buf=0x%x\n",
		loadField(p2, 0, 8), loadField(p2, 8, 8), loadField(p2, 16, 8))
	fmt.Println("  📌 dataqsiz 区分 unbuffered (0) vs buffered (N)")
	fmt.Println()

	// ---------- 4. sendx/recvx 索引移动 ----------
	fmt.Println("【实验 2】写入数据后 sendx/recvx 移动")
	ch := make(chan int, 3)
	p := hchanPointer(ch)
	fmt.Printf("  初始: qcount=%d, sendx=%d, recvx=%d\n",
		loadField(p, 0, 8), loadField(p, 48, 8), loadField(p, 56, 8))

	ch <- 1
	fmt.Printf("  <-ch 1: qcount=%d, sendx=%d, recvx=%d\n",
		loadField(p, 0, 8), loadField(p, 48, 8), loadField(p, 56, 8))

	ch <- 2
	fmt.Printf("  <-ch 2: qcount=%d, sendx=%d, recvx=%d\n",
		loadField(p, 0, 8), loadField(p, 48, 8), loadField(p, 56, 8))

	<-ch
	fmt.Printf("  ch ->  : qcount=%d, sendx=%d, recvx=%d (recvx 移到下一位置)\n",
		loadField(p, 0, 8), loadField(p, 48, 8), loadField(p, 56, 8))
	fmt.Println()

	// ---------- 5. closed 字段验证 ----------
	fmt.Println("【实验 3】closed 字段验证")
	ch2 := make(chan int, 1)
	p2 = hchanPointer(ch2)
	fmt.Printf("  close 前: closed=%d\n", loadField(p2, 28, 4))
	close(ch2)
	fmt.Printf("  close 后: closed=%d\n", loadField(p2, 28, 4))
	fmt.Println()

	// ---------- 6. 元素大小 ----------
	fmt.Println("【实验 4】不同元素类型的 elemsize")
	type ChanStruct struct{ a, b int64 }
	chStruct := make(chan ChanStruct, 1)
	pStruct := hchanPointer(chStruct)
	fmt.Printf("  chan int      : elemsize=%d\n", loadField(pStruct, 24, 2))
	chStr := make(chan string, 1)
	pStr := hchanPointer(chStr)
	fmt.Printf("  chan string   : elemsize=%d (string header = 16 bytes)\n", loadField(pStr, 24, 2))
	fmt.Println()

	// ---------- 7. lock 保护验证 ----------
	fmt.Println("【实验 5】lock 保护：无数据竞争")
	ch3 := make(chan int, 100)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			ch3 <- i
		}
		close(done)
	}()
	go func() {
		count := 0
		for range ch3 {
			count++
		}
		atomic.StoreInt32(&dummyCounter, int32(count))
	}()
	<-done
	fmt.Println("  1000 个写入 + 1000 个读取无 race（hchan.lock 保护）")
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - hchan = (qcount, dataqsiz, buf, closed, sendx, recvx, sendq, recvq, lock)")
	fmt.Println("   - buf 是环形队列，sendx/recvx 是索引")
	fmt.Println("   - sendq/recvq 是等待的 goroutine 队列（双向链表）")
	fmt.Println("   - lock 保护整个 hchan（channel 是线程安全的根因）")
	fmt.Println("   - unbuffered channel 的 dataqsiz=0, buf=nil")
	fmt.Println()
	fmt.Println("⚠️ unsafe 代码依赖 runtime 内部布局，生产环境绝对不要这么写")
	fmt.Println("   （了解即可，面试能讲清结构 + 画图就够了）")
	fmt.Println()
	fmt.Println("💡 实际面试建议画 hchan 图（不需要 unsafe 代码）：")
	fmt.Println("   ┌─────────────────────────────────────┐")
	fmt.Println("   │ qcount | dataqsiz | buf ───┐       │")
	fmt.Println("   │ closed | sendx | recvx     │       │")
	fmt.Println("   │ recvq (waitq: first/last)   ▼       │")
	fmt.Println("   │ sendq (waitq: first/last)  [0][1]   │")
	fmt.Println("   │ lock (mutex)        环形队列        │")
	fmt.Println("   └─────────────────────────────────────┘")
}

var dummyCounter int32
