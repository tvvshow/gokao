// +build cgo

package cppbridge

import (
	"unsafe"

	"github.com/redis/go-redis/v9"
)

// createCStringArray 创建安全的C字符串数组
func (b *CppHybridRecommendationBridge) createCStringArray(goStrings []string) (**C.char, C.int) {
	if len(goStrings) == 0 {
		return nil, 0
	}
	
	// 分配C字符串数组内存
	cArray := (**C.char)(C.malloc(C.size_t(len(goStrings)) * C.size_t(unsafe.Sizeof((*C.char)(nil)))))
	
	// 创建安全的slice来访问C数组
	slice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cArray))[:len(goStrings):len(goStrings)]
	
	// 填充C字符串
	for i, s := range goStrings {
		slice[i] = C.CString(s)
	}
	
	return cArray, C.int(len(goStrings))
}

// freeCStringArray 安全释放C字符串数组
func (b *CppHybridRecommendationBridge) freeCStringArray(cArray **C.char, count C.int) {
	if cArray == nil {
		return
	}
	
	// 创建安全的slice来访问C数组
	slice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cArray))[:count:count]
	
	// 释放每个C字符串
	for i := 0; i < int(count); i++ {
		C.free(unsafe.Pointer(slice[i]))
	}
	
	// 释放数组本身
	C.free(unsafe.Pointer(cArray))
}

// createSafeCSlice 创建安全的C数组slice访问器
func createSafeCSlice[T any](ptr unsafe.Pointer, length int) []T {
	if ptr == nil || length == 0 {
		return nil
	}
	
	// 使用最大可能的数组大小创建slice
	return (*[1<<30 - 1]T)(ptr)[:length:length]
}

// copyToCStringArray 安全复制Go字符串切片到C字符串数组
func copyToCStringArray(dest **C.char, src []string) C.int {
	if len(src) == 0 {
		return 0
	}
	
	// 创建安全的slice访问器
	destSlice := createSafeCSlice[*C.char](unsafe.Pointer(dest), len(src))
	
	// 复制字符串
	for i, s := range src {
		destSlice[i] = C.CString(s)
	}
	
	return C.int(len(src))
}

// freeCStringArraySlice 安全释放C字符串数组slice
func freeCStringArraySlice(slice []*C.char) {
	for _, ptr := range slice {
		if ptr != nil {
			C.free(unsafe.Pointer(ptr))
		}
	}
	
	if len(slice) > 0 {
		C.free(unsafe.Pointer(&slice[0]))
	}
}

// SafeMemoryManager 安全内存管理器
type SafeMemoryManager struct {
	allocations map[uintptr]bool
}

// NewSafeMemoryManager 创建新的安全内存管理器
func NewSafeMemoryManager() *SafeMemoryManager {
	return &SafeMemoryManager{
		allocations: make(map[uintptr]bool),
	}
}

// Malloc 安全的内存分配
func (m *SafeMemoryManager) Malloc(size uintptr) unsafe.Pointer {
	ptr := C.malloc(C.size_t(size))
	if ptr != nil {
		m.allocations[uintptr(ptr)] = true
	}
	return ptr
}

// Free 安全的内存释放
func (m *SafeMemoryManager) Free(ptr unsafe.Pointer) {
	if ptr != nil {
		addr := uintptr(ptr)
		if m.allocations[addr] {
			C.free(ptr)
			delete(m.allocations, addr)
		}
	}
}

// FreeAll 释放所有分配的内存
func (m *SafeMemoryManager) FreeAll() {
	for addr := range m.allocations {
		C.free(unsafe.Pointer(addr))
	}
	m.allocations = make(map[uintptr]bool)
}

// Cleanup 清理资源
func (m *SafeMemoryManager) Cleanup() {
	m.FreeAll()
}

// MemoryUsage 获取内存使用情况
func (m *SafeMemoryManager) MemoryUsage() int {
	return len(m.allocations)
}

// IsValidPointer 检查指针是否有效
func (m *SafeMemoryManager) IsValidPointer(ptr unsafe.Pointer) bool {
	if ptr == nil {
		return false
	}
	return m.allocations[uintptr(ptr)]
}

// GC 垃圾回收（手动触发）
func (m *SafeMemoryManager) GC() {
	// 这里可以添加更复杂的内存回收逻辑
	m.FreeAll()
}

// MemoryStats 内存统计信息
type MemoryStats struct {
	TotalAllocations int
	CurrentUsage    int
	PeakUsage       int
}

// GetStats 获取内存统计信息
func (m *SafeMemoryManager) GetStats() MemoryStats {
	return MemoryStats{
		TotalAllocations: len(m.allocations),
		CurrentUsage:     len(m.allocations),
		PeakUsage:        len(m.allocations), // 简化实现
	}
}