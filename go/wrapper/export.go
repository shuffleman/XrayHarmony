package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
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

// VPN 管理器映射
var (
	vpnManagers   = make(map[int64]*VPNManager)
	vpnManagersMu sync.RWMutex
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
	version := "XrayHarmony v1.1.0 (with VPN support)"
	return C.CString(version)
}

// ================ VPN 相关导出函数 ================

//export VPNNewManager
func VPNNewManager(xrayInstanceID C.longlong) C.longlong {
	instancesMu.RLock()
	xrayInstance, exists := instances[int64(xrayInstanceID)]
	instancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("xray instance not found"))
		return -1
	}

	vpnManagersMu.Lock()
	defer vpnManagersMu.Unlock()

	id := nextID
	nextID++

	manager := NewVPNManager(xrayInstance)
	vpnManagers[id] = manager

	setLastError(nil)
	return C.longlong(id)
}

//export VPNDeleteManager
func VPNDeleteManager(id C.longlong) C.int {
	vpnManagersMu.Lock()
	defer vpnManagersMu.Unlock()

	managerID := int64(id)
	manager, exists := vpnManagers[managerID]
	if !exists {
		setLastError(fmt.Errorf("VPN manager not found"))
		return -1
	}

	// 如果 VPN 正在运行，先停止
	if manager.IsRunning() {
		if err := manager.Stop(); err != nil {
			setLastError(err)
			return -1
		}
	}

	delete(vpnManagers, managerID)
	setLastError(nil)
	return 0
}

//export VPNStart
func VPNStart(id C.longlong, configJSON *C.char) C.int {
	vpnManagersMu.RLock()
	manager, exists := vpnManagers[int64(id)]
	vpnManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("VPN manager not found"))
		return -1
	}

	// 解析配置
	configStr := C.GoString(configJSON)
	var config VPNConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		setLastError(fmt.Errorf("failed to parse VPN config: %w", err))
		return -1
	}

	// 启动 VPN
	if err := manager.Start(&config); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export VPNStop
func VPNStop(id C.longlong) C.int {
	vpnManagersMu.RLock()
	manager, exists := vpnManagers[int64(id)]
	vpnManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("VPN manager not found"))
		return -1
	}

	if err := manager.Stop(); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export VPNIsRunning
func VPNIsRunning(id C.longlong) C.int {
	vpnManagersMu.RLock()
	manager, exists := vpnManagers[int64(id)]
	vpnManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("VPN manager not found"))
		return -1
	}

	if manager.IsRunning() {
		setLastError(nil)
		return 1
	}

	setLastError(nil)
	return 0
}

//export VPNGetStats
func VPNGetStats(id C.longlong) *C.char {
	vpnManagersMu.RLock()
	manager, exists := vpnManagers[int64(id)]
	vpnManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("VPN manager not found"))
		return nil
	}

	stats, err := manager.GetStats()
	if err != nil {
		setLastError(err)
		return nil
	}

	// 将统计信息转换为 JSON
	data, err := json.Marshal(stats)
	if err != nil {
		setLastError(fmt.Errorf("failed to marshal stats: %w", err))
		return nil
	}

	setLastError(nil)
	return C.CString(string(data))
}

func main() {
	// Required for buildmode=c-shared
}
