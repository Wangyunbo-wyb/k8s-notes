package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 创建一个接收信号的通道
	sigs := make(chan os.Signal, 1)

	// 使用 signal.Notify 注册要接收的信号
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		// 在 goroutine 中等待信号
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		os.Exit(0)
	}()

	fmt.Println("Awaiting signal")
	// 使主 goroutine 无限等待，直到接收到信号
	select {}
}
