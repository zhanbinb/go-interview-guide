package main

import "fmt"

// DemoVsJava 演示 Go vs Java 的核心区别
//
// 核心对比：
//   1. 部署：Go 单二进制 vs Java 需 JVM
//   2. 并发：goroutine vs thread
//   3. 类型：接口隐式 vs implements
//   4. 生态：Java 成熟 vs Go 云原生
func DemoVsJava() {
	fmt.Println("=== Go vs Java ===")
	fmt.Println()

	fmt.Println("【1】部署")
	fmt.Println("  Go:")
	fmt.Println("    - go build 输出单一二进制")
	fmt.Println("    - 几 MB～几十 MB，零依赖")
	fmt.Println("    - 直接 ./binary 运行（无运行时）")
	fmt.Println("  Java:")
	fmt.Println("    - jar/war + JVM 运行时")
	fmt.Println("    - JVM 本身几百 MB")
	fmt.Println("    - 需要预装 JRE/JDK")
	fmt.Println()

	fmt.Println("【2】并发模型")
	fmt.Println("  Go:")
	fmt.Println("    - goroutine: 2KB 栈，按需增长")
	fmt.Println("    - 轻松创建百万级")
	fmt.Println("    - 通过 channel 通信")
	fmt.Println("  Java:")
	fmt.Println("    - thread: 1MB 栈（默认）")
	fmt.Println("    - 创建数千个就吃力")
	fmt.Println("    - 通过共享内存 + synchronized/Lock 通信")
	fmt.Println()

	fmt.Println("【3】类型系统")
	fmt.Println("  Go:")
	fmt.Println("    - 接口隐式实现（鸭子类型）")
	fmt.Println("    - 没有 implements 关键字")
	fmt.Println("    - 没有 class，只有 struct")
	fmt.Println("  Java:")
	fmt.Println("    - 必须显式 implements")
	fmt.Println("    - 完整的 class 体系")
	fmt.Println("    - 抽象类/接口严格区分")
	fmt.Println()

	fmt.Println("【4】性能（典型 benchmark）")
	fmt.Println("  Go:")
	fmt.Println("    - 启动快（毫秒级）")
	fmt.Println("    - 内存占用低（goroutine 2KB）")
	fmt.Println("    - 适合云原生/微服务")
	fmt.Println("  Java:")
	fmt.Println("    - 启动慢（秒级）")
	fmt.Println("    - JVM 内存大（百 MB 起）")
	fmt.Println("    - 长时间跑热后性能高")
	fmt.Println()

	fmt.Println("【5】生态")
	fmt.Println("  Go:")
	fmt.Println("    - 云原生（Docker/K8s 大量 Go）")
	fmt.Println("    - DevOps 工具（Terraform, K8s）")
	fmt.Println("    - 区块链（Ethereum, Cosmos）")
	fmt.Println("  Java:")
	fmt.Println("    - 企业应用（Spring 全家桶）")
	fmt.Println("    - Android（虽然现在 Kotlin 主导）")
	fmt.Println("    - 大数据（Hadoop, Kafka）")
	fmt.Println()

	fmt.Println("📌 选择建议:")
	fmt.Println("   - 新服务/微服务/云原生：Go")
	fmt.Println("   - 大型企业应用/需要丰富库：Java")
	fmt.Println("   - 性能敏感：Go")
	fmt.Println("   - 已有 Java 生态：Java")
}
