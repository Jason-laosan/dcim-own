package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	// 创建一个可取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用 WaitGroup 等待协程完成
	var wg sync.WaitGroup

	// 启动协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("协程被取消")
				return
			default:
				time.Sleep(1 * time.Second)
				fmt.Println("Hello World", i)
			}
		}
	}()

	fmt.Println("主程序开始")

	// 等待所有协程执行完毕
	wg.Wait()

	defer func() {
		fmt.Println("程序结束")
	}()
}
