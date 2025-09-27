package concurrency

import (
	"context"
	"sync"
	"time"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	rate     int           // 每秒允许的请求数
	interval time.Duration // 请求间隔
	lastTime time.Time     // 上次请求时间
	mutex    sync.Mutex    // 互斥锁
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(rps int) *RateLimiter {
	if rps <= 0 {
		rps = 1
	}
	
	return &RateLimiter{
		rate:     rps,
		interval: time.Second / time.Duration(rps),
		lastTime: time.Now(),
	}
}

// Wait 等待直到可以执行下一个请求
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastTime)
	
	if elapsed < rl.interval {
		waitTime := rl.interval - elapsed
		
		select {
		case <-time.After(waitTime):
			rl.lastTime = time.Now()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	rl.lastTime = now
	return nil
}

// TryAcquire 尝试获取许可，不阻塞
func (rl *RateLimiter) TryAcquire() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastTime)
	
	if elapsed >= rl.interval {
		rl.lastTime = now
		return true
	}
	
	return false
}

// GetRate 获取当前速率
func (rl *RateLimiter) GetRate() int {
	return rl.rate
}

// SetRate 设置新的速率
func (rl *RateLimiter) SetRate(rps int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	if rps <= 0 {
		rps = 1
	}
	
	rl.rate = rps
	rl.interval = time.Second / time.Duration(rps)
}

// TokenBucket 令牌桶算法实现
type TokenBucket struct {
	capacity    int           // 桶容量
	tokens      int           // 当前令牌数
	refillRate  int           // 每秒补充令牌数
	lastRefill  time.Time     // 上次补充时间
	mutex       sync.Mutex    // 互斥锁
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// TryConsume 尝试消费令牌
func (tb *TokenBucket) TryConsume(tokens int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.refill()
	
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}
	
	return false
}

// Wait 等待直到有足够的令牌
func (tb *TokenBucket) Wait(ctx context.Context, tokens int) error {
	for {
		if tb.TryConsume(tokens) {
			return nil
		}
		
		// 计算需要等待的时间
		tb.mutex.Lock()
		needed := tokens - tb.tokens
		waitTime := time.Duration(needed) * time.Second / time.Duration(tb.refillRate)
		tb.mutex.Unlock()
		
		if waitTime > time.Second {
			waitTime = time.Second
		}
		
		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// refill 补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	if elapsed > 0 {
		tokensToAdd := int(elapsed.Seconds() * float64(tb.refillRate))
		tb.tokens += tokensToAdd
		
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		
		tb.lastRefill = now
	}
}

// GetTokens 获取当前令牌数
func (tb *TokenBucket) GetTokens() int {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.refill()
	return tb.tokens
}