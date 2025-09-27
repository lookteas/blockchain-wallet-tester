package concurrency

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int                    // 最大重试次数
	BaseDelay     time.Duration          // 基础延迟时间
	MaxDelay      time.Duration          // 最大延迟时间
	BackoffFactor float64               // 退避因子
	RetryableErrors []string             // 可重试的错误关键词
	ShouldRetryFunc func(error) bool     // 自定义重试判断函数
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		BaseDelay:     time.Second,
		MaxDelay:      time.Minute,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"network unreachable",
			"temporary failure",
			"rate limit",
			"too many requests",
			"service unavailable",
			"internal server error",
		},
	}
}

// GetDelay 计算重试延迟时间（指数退避）
func (rc *RetryConfig) GetDelay(attempt int) time.Duration {
	if attempt < 0 {
		return rc.BaseDelay
	}
	
	delay := float64(rc.BaseDelay) * math.Pow(rc.BackoffFactor, float64(attempt))
	
	if delay > float64(rc.MaxDelay) {
		return rc.MaxDelay
	}
	
	return time.Duration(delay)
}

// ShouldRetry 判断是否应该重试
func (rc *RetryConfig) ShouldRetry(err error) bool {
	if err == nil {
		return false
	}
	
	// 使用自定义判断函数
	if rc.ShouldRetryFunc != nil {
		return rc.ShouldRetryFunc(err)
	}
	
	// 检查错误消息是否包含可重试的关键词
	errMsg := strings.ToLower(err.Error())
	for _, keyword := range rc.RetryableErrors {
		if strings.Contains(errMsg, strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}

// RetryableError 可重试错误类型
type RetryableError struct {
	Err       error
	Retryable bool
}

func (re *RetryableError) Error() string {
	return re.Err.Error()
}

func (re *RetryableError) Unwrap() error {
	return re.Err
}

// NewRetryableError 创建可重试错误
func NewRetryableError(err error) *RetryableError {
	return &RetryableError{
		Err:       err,
		Retryable: true,
	}
}

// NewNonRetryableError 创建不可重试错误
func NewNonRetryableError(err error) *RetryableError {
	return &RetryableError{
		Err:       err,
		Retryable: false,
	}
}

// IsRetryableError 检查是否为可重试错误
func IsRetryableError(err error) bool {
	var retryableErr *RetryableError
	if errors.As(err, &retryableErr) {
		return retryableErr.Retryable
	}
	return false
}

// RetryExecutor 重试执行器
type RetryExecutor struct {
	config *RetryConfig
}

// NewRetryExecutor 创建重试执行器
func NewRetryExecutor(config *RetryConfig) *RetryExecutor {
	if config == nil {
		config = DefaultRetryConfig()
	}
	
	return &RetryExecutor{
		config: config,
	}
}

// Execute 执行函数并重试
func (re *RetryExecutor) Execute(fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= re.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := re.config.GetDelay(attempt - 1)
			time.Sleep(delay)
		}
		
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 检查是否应该重试
		if !re.config.ShouldRetry(err) {
			break
		}
	}
	
	return fmt.Errorf("重试失败，最后错误: %w", lastErr)
}

// ExecuteWithContext 带上下文的重试执行
func (re *RetryExecutor) ExecuteWithContext(ctx context.Context, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= re.config.MaxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if attempt > 0 {
			delay := re.config.GetDelay(attempt - 1)
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 检查是否应该重试
		if !re.config.ShouldRetry(err) {
			break
		}
	}
	
	return fmt.Errorf("重试失败，最后错误: %w", lastErr)
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	maxFailures   int           // 最大失败次数
	resetTimeout  time.Duration // 重置超时时间
	failures      int           // 当前失败次数
	lastFailTime  time.Time     // 上次失败时间
	state         CBState       // 熔断器状态
}

// CBState 熔断器状态
type CBState int

const (
	CBClosed CBState = iota // 关闭状态（正常）
	CBOpen                  // 开启状态（熔断）
	CBHalfOpen             // 半开状态（试探）
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CBClosed,
	}
}

// Execute 执行函数（带熔断保护）
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == CBOpen {
		// 检查是否可以进入半开状态
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CBHalfOpen
		} else {
			return errors.New("熔断器开启，拒绝执行")
		}
	}
	
	err := fn()
	
	if err != nil {
		cb.onFailure()
		return err
	}
	
	cb.onSuccess()
	return nil
}

// onSuccess 成功回调
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	cb.state = CBClosed
}

// onFailure 失败回调
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.failures >= cb.maxFailures {
		cb.state = CBOpen
	}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CBState {
	return cb.state
}