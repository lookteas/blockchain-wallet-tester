package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task 表示一个工作任务
type Task interface {
	Execute(ctx context.Context) error
	GetID() string
}

// TaskResult 表示任务执行结果
type TaskResult struct {
	TaskID    string
	Success   bool
	Error     error
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// WorkerPool 工作池
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	resultQueue chan TaskResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	rateLimit   *RateLimiter
	retryConfig *RetryConfig
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(workerCount int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
		resultQueue: make(chan TaskResult, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// SetRateLimit 设置速率限制
func (wp *WorkerPool) SetRateLimit(rps int) {
	wp.rateLimit = NewRateLimiter(rps)
}

// SetRetryConfig 设置重试配置
func (wp *WorkerPool) SetRetryConfig(config *RetryConfig) {
	wp.retryConfig = config
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	close(wp.resultQueue)
}

// SubmitTask 提交任务
func (wp *WorkerPool) SubmitTask(task Task) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("工作池已停止")
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// GetResults 获取结果通道
func (wp *WorkerPool) GetResults() <-chan TaskResult {
	return wp.resultQueue
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}
			
			// 速率限制
			if wp.rateLimit != nil {
				wp.rateLimit.Wait(wp.ctx)
			}
			
			result := wp.executeTask(task)
			
			select {
			case wp.resultQueue <- result:
			case <-wp.ctx.Done():
				return
			}
			
		case <-wp.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务（包含重试逻辑）
func (wp *WorkerPool) executeTask(task Task) TaskResult {
	startTime := time.Now()
	
	var lastErr error
	maxRetries := 1
	if wp.retryConfig != nil {
		maxRetries = wp.retryConfig.MaxRetries + 1
	}
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 && wp.retryConfig != nil {
			// 重试延迟
			delay := wp.retryConfig.GetDelay(attempt - 1)
			select {
			case <-time.After(delay):
			case <-wp.ctx.Done():
				break
			}
		}
		
		err := task.Execute(wp.ctx)
		if err == nil {
			return TaskResult{
				TaskID:    task.GetID(),
				Success:   true,
				Error:     nil,
				Duration:  time.Since(startTime),
				StartTime: startTime,
				EndTime:   time.Now(),
			}
		}
		
		lastErr = err
		
		// 检查是否应该重试
		if wp.retryConfig != nil && !wp.retryConfig.ShouldRetry(err) {
			break
		}
	}
	
	return TaskResult{
		TaskID:    task.GetID(),
		Success:   false,
		Error:     lastErr,
		Duration:  time.Since(startTime),
		StartTime: startTime,
		EndTime:   time.Now(),
	}
}

// GetStats 获取工作池统计信息
func (wp *WorkerPool) GetStats() PoolStats {
	return PoolStats{
		WorkerCount:     wp.workerCount,
		QueuedTasks:     len(wp.taskQueue),
		QueueCapacity:   cap(wp.taskQueue),
		ResultsQueued:   len(wp.resultQueue),
		ResultsCapacity: cap(wp.resultQueue),
	}
}

// PoolStats 工作池统计信息
type PoolStats struct {
	WorkerCount     int
	QueuedTasks     int
	QueueCapacity   int
	ResultsQueued   int
	ResultsCapacity int
}