package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// 全局实例管理器
var (
	instances   = make(map[int64]*XrayInstance)
	instancesMu sync.RWMutex
	nextID      int64 = 1
)

// 错误信息缓存
var (
	lastError   string
	lastErrorMu sync.RWMutex
)

func setLastError(err error) {
	lastErrorMu.Lock()
	defer lastErrorMu.Unlock()
	if err != nil {
		lastError = err.Error()
	} else {
		lastError = ""
	}
}

//export XrayGetLastError
func XrayGetLastError() *C.char {
	lastErrorMu.RLock()
	defer lastErrorMu.RUnlock()
	if lastError == "" {
		return nil
	}
	return C.CString(lastError)
}

//export XrayFreeString
func XrayFreeString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export XrayNewInstance
func XrayNewInstance() C.longlong {
	instancesMu.Lock()
	defer instancesMu.Unlock()

	id := nextID
	nextID++

	instance := NewXrayInstance()
	instances[id] = instance

	setLastError(nil)
	return C.longlong(id)
}

//export XrayDeleteInstance
func XrayDeleteInstance(id C.longlong) C.int {
	instancesMu.Lock()
	defer instancesMu.Unlock()

	instanceID := int64(id)
	instance, exists := instances[instanceID]
	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	// 如果实例正在运行，先停止
	if instance.IsRunning() {
		if err := instance.Stop(); err != nil {
			setLastError(err)
			return -1
		}
	}

	delete(instances, instanceID)
	setLastError(nil)
	return 0
}

//export XrayLoadConfig
func XrayLoadConfig(id C.longlong, configJSON *C.char) C.int {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	config := C.GoString(configJSON)
	if err := instance.LoadConfig(config); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export XrayLoadConfigFromFile
func XrayLoadConfigFromFile(id C.longlong, filePath *C.char) C.int {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	path := C.GoString(filePath)
	if err := instance.LoadConfigFromFile(path); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export XrayStart
func XrayStart(id C.longlong) C.int {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	if err := instance.Start(); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export XrayStop
func XrayStop(id C.longlong) C.int {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	if err := instance.Stop(); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export XrayIsRunning
func XrayIsRunning(id C.longlong) C.int {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	if instance.IsRunning() {
		setLastError(nil)
		return 1
	}

	setLastError(nil)
	return 0
}

//export XrayGetStats
func XrayGetStats(id C.longlong) *C.char {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return nil
	}

	stats, err := instance.GetStats()
	if err != nil {
		setLastError(err)
		return nil
	}

	setLastError(nil)
	return C.CString(stats)
}

//export XrayTestConfig
func XrayTestConfig(id C.longlong, configJSON *C.char) C.int {
	instancesMu.RLock()
	instance, exists := instances[int64(id)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("instance not found"))
		return -1
	}

	config := C.GoString(configJSON)
	if err := instance.TestConfig(config); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

// 版本信息
//export XrayGetVersion
func XrayGetVersion() *C.char {
	version := "XrayHarmony v1.0.0"
	return C.CString(version)
}

func main() {
	// Required for buildmode=c-shared
}
