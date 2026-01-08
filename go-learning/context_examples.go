package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// 示例1: 使用 WithTimeout 设置超时
func example1_WithTimeout() {
	fmt.Println("\n=== 示例1: WithTimeout - 超时控制 ===")

	// 创建一个3秒超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("协程超时退出:", ctx.Err())
				return
			default:
				time.Sleep(1 * time.Second)
				fmt.Printf("任务执行中... %d\n", i)
			}
		}
	}()

	wg.Wait()
	fmt.Println("示例1完成")
}

// 示例2: 使用 WithDeadline 设置截止时间
func example2_WithDeadline() {
	fmt.Println("\n=== 示例2: WithDeadline - 截止时间控制 ===")

	// 设置5秒后的截止时间
	deadline := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("协程到达截止时间:", ctx.Err())
				return
			default:
				time.Sleep(1 * time.Second)
				fmt.Printf("处理数据 %d\n", i)
			}
		}
	}()

	wg.Wait()
	fmt.Println("示例2完成")
}

// 示例3: 使用 WithValue 传递请求范围的值
func example3_WithValue() {
	fmt.Println("\n=== 示例3: WithValue - 传递请求数据 ===")

	// 创建带有用户信息的 context
	ctx := context.WithValue(context.Background(), "userID", "12345")
	ctx = context.WithValue(ctx, "requestID", "req-abc-123")

	processRequest(ctx)
	fmt.Println("示例3完成")
}

func processRequest(ctx context.Context) {
	userID := ctx.Value("userID")
	requestID := ctx.Value("requestID")

	fmt.Printf("处理请求 - UserID: %v, RequestID: %v\n", userID, requestID)
}

// 示例4: 手动取消协程
func example4_ManualCancel() {
	fmt.Println("\n=== 示例4: 手动取消协程 ===")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("收到取消信号，协程退出")
				return
			default:
				time.Sleep(500 * time.Millisecond)
				fmt.Printf("工作中... %d\n", i)
			}
		}
	}()

	// 2秒后手动取消
	time.Sleep(2 * time.Second)
	fmt.Println("发送取消信号")
	cancel()

	wg.Wait()
	fmt.Println("示例4完成")
}

// 示例5: 多个协程共享 context
func example5_MultipleGoroutines() {
	fmt.Println("\n=== 示例5: 多个协程共享 context ===")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// 启动3个协程
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(ctx, &wg, i)
	}

	wg.Wait()
	fmt.Println("示例5完成")
}

func worker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	for j := 0; j < 10; j++ {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d 退出: %v\n", id, ctx.Err())
			return
		default:
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("Worker %d 执行任务 %d\n", id, j)
		}
	}
}

// 示例6: Context 链式传递
func example6_ContextChain() {
	fmt.Println("\n=== 示例6: Context 链式传递 ===")

	// 根 context
	rootCtx := context.Background()

	// 添加值
	ctx1 := context.WithValue(rootCtx, "layer", "1")

	// 添加超时
	ctx2, cancel := context.WithTimeout(ctx1, 2*time.Second)
	defer cancel()

	// 再添加值
	ctx3 := context.WithValue(ctx2, "requestID", "xyz-789")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		layer := ctx3.Value("layer")
		reqID := ctx3.Value("requestID")

		fmt.Printf("Context 值 - Layer: %v, RequestID: %v\n", layer, reqID)

		for i := 0; i < 5; i++ {
			select {
			case <-ctx3.Done():
				fmt.Println("Context 链超时")
				return
			default:
				time.Sleep(500 * time.Millisecond)
				fmt.Printf("处理中... %d\n", i)
			}
		}
	}()

	wg.Wait()
	fmt.Println("示例6完成")
}

func main() {
	fmt.Println("========== Go Context 使用示例 ==========")

	// 运行所有示例
	example1_WithTimeout()
	example2_WithDeadline()
	example3_WithValue()
	example4_ManualCancel()
	example5_MultipleGoroutines()
	example6_ContextChain()

	fmt.Println("\n========== 所有示例完成 ==========")
}
