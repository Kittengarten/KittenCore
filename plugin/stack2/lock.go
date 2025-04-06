package stack2

import "sync"

// 读写锁，用于叠猫猫文件
var mu sync.RWMutex

// Lock 上锁
func Lock() {
	mu.Lock()
}

// Unlock 解锁
func Unlock() {
	mu.Unlock()
}

// RLock 上读锁
func RLock() {
	mu.RLock()
}

// RUnlock 解读锁
func RUnlock() {
	mu.RUnlock()
}
