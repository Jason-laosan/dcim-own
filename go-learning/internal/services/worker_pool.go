package services

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// TaskJob 任务作业结构
type TaskJob struct {
	ID          int
	Title       string
	Description string
}

// WorkerPool 工作池，管理多个并发 worker
type WorkerPool struct {
	jobs       chan TaskJob        // 任务通道
	results    chan TaskJob        // 结果通道
	wg         sync.WaitGroup      // 等待组
	workerCount int                // worker 数量
	quit       chan bool           // 退出信号
	updateFunc func(int, string)  // 任务状态更新函数
}

// NewWorkerPool 创建新的工作池
// workerCount: worker 数量
// updateFunc: 任务状态更新回调函数
func NewWorkerPool(workerCount int, updateFunc func(int, string)) *WorkerPool {
	return &WorkerPool{
		jobs:        make(chan TaskJob, 100),  // 带缓冲的任务通道
		results:     make(chan TaskJob, 100),  // 带缓冲的结果通道
		workerCount: workerCount,
		quit:        make(chan bool),
		updateFunc:  updateFunc,
	}
}

// Start 启动所有 worker
func (wp *WorkerPool) Start() {
	log.Printf("启动 Worker Pool，worker 数量: %d", wp.workerCount)

	// 启动多个 worker goroutine
	for i := 1; i <= wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// 启动结果处理器
	go wp.resultHandler()
}

// worker 单个 worker 的执行逻辑
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	log.Printf("Worker %d 已启动", id)

	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				// 通道已关闭，退出 worker
				log.Printf("Worker %d 退出", id)
				return
			}

			log.Printf("Worker %d 开始处理任务 #%d: %s", id, job.ID, job.Title)

			// 更新任务状态为 processing
			if wp.updateFunc != nil {
				wp.updateFunc(job.ID, "processing")
			}

			// 模拟任务处理（耗时操作）
			// 随机休眠 1-5 秒
			sleepDuration := time.Duration(rand.Intn(5)+1) * time.Second
			time.Sleep(sleepDuration)

			// 模拟任务成功或失败（90% 成功率）
			var status string
			if rand.Float32() < 0.9 {
				status = "completed"
				log.Printf("Worker %d 完成任务 #%d (耗时: %v)", id, job.ID, sleepDuration)
			} else {
				status = "failed"
				log.Printf("Worker %d 任务 #%d 失败", id, job.ID)
			}

			// 更新任务状态
			if wp.updateFunc != nil {
				wp.updateFunc(job.ID, status)
			}

			// 将结果发送到结果通道
			wp.results <- job

		case <-wp.quit:
			log.Printf("Worker %d 收到退出信号", id)
			return
		}
	}
}

// resultHandler 处理任务结果
func (wp *WorkerPool) resultHandler() {
	for result := range wp.results {
		log.Printf("任务结果: #%d - %s", result.ID, result.Title)
	}
}

// Submit 提交任务到工作池
func (wp *WorkerPool) Submit(job TaskJob) error {
	select {
	case wp.jobs <- job:
		log.Printf("任务 #%d 已提交到工作池", job.ID)
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("submit timeout: worker pool is busy")
	}
}

// Shutdown 优雅关闭工作池
func (wp *WorkerPool) Shutdown() {
	log.Println("开始关闭 Worker Pool...")

	// 关闭任务通道，通知 workers 不再接收新任务
	close(wp.jobs)

	// 等待所有 worker 完成当前任务
	wp.wg.Wait()

	// 关闭结果通道
	close(wp.results)

	log.Println("Worker Pool 已关闭")
}

// GetStats 获取工作池统计信息
func (wp *WorkerPool) GetStats() map[string]int {
	return map[string]int{
		"worker_count":   wp.workerCount,
		"pending_jobs":   len(wp.jobs),
		"pending_results": len(wp.results),
	}
}

// 演示：使用 channel 进行协程间通信的示例函数
func DemoChannelCommunication() {
	// 创建无缓冲通道
	messages := make(chan string)
	done := make(chan bool)

	// 启动接收者 goroutine
	go func() {
		for {
			msg, ok := <-messages
			if !ok {
				log.Println("通道已关闭")
				done <- true
				return
			}
			log.Printf("接收到消息: %s", msg)
		}
	}()

	// 发送消息
	for i := 1; i <= 5; i++ {
		messages <- fmt.Sprintf("消息 #%d", i)
		time.Sleep(500 * time.Millisecond)
	}

	// 关闭通道
	close(messages)

	// 等待接收者完成
	<-done
	log.Println("演示完成")
}

// 演示：使用 select 处理多个 channel
func DemoSelectMultiplexing() {
	c1 := make(chan string)
	c2 := make(chan string)

	// goroutine 1
	go func() {
		time.Sleep(1 * time.Second)
		c1 <- "来自 channel 1"
	}()

	// goroutine 2
	go func() {
		time.Sleep(2 * time.Second)
		c2 <- "来自 channel 2"
	}()

	// 使用 select 等待两个 channel
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			log.Println("收到:", msg1)
		case msg2 := <-c2:
			log.Println("收到:", msg2)
		case <-time.After(3 * time.Second):
			log.Println("超时")
		}
	}
}
