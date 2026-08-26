// pprof 性能排查演示
package main

import (
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
)

// main 启动一个带 pprof 端点的 HTTP 服务
// 访问 http://localhost:6060/debug/pprof/ 查看可用的 profile
//
// 注意：pprof 端点只暴露在内部端口（6060），不要在生产对外暴露
func main() {
	if len(os.Args) > 1 && os.Args[1] == "leak" {
		// 演示模式：制造 goroutine 泄漏
		DemoLeak()
		return
	}

	// 注册 pprof 端点
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// 业务端点（演示用）
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, pprof demo"))
	})

	// 模拟业务：制造一些负载（让 pprof 能看到东西）
	go fakeLoad()

	addr := ":6060"
	log.Printf("pprof server listening on %s", addr)
	log.Printf("try: curl http://localhost%s/debug/pprof/", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func fakeLoad() {
	// 持续分配内存（让 heap profile 有数据）
	for {
		s := make([]byte, 1024*1024) // 1MB
		_ = s
		runtime.GC() // 主动 GC，让数据更稳定
	}
}
