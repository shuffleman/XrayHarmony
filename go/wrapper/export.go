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
	version := "XrayHarmony v2.0.0"
	return C.CString(version)
}

// ===============================
// 协议工具函数导出
// ===============================

//export XrayParseShareURL
func XrayParseShareURL(shareURL *C.char) *C.char {
	url := C.GoString(shareURL)

	config, err := ParseShareURL(url)
	if err != nil {
		setLastError(err)
		return nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		setLastError(err)
		return nil
	}

	setLastError(nil)
	return C.CString(string(data))
}

//export XrayGenerateShareURL
func XrayGenerateShareURL(configJSON *C.char) *C.char {
	configStr := C.GoString(configJSON)

	var config ServerConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		setLastError(err)
		return nil
	}

	url, err := GenerateShareURL(&config)
	if err != nil {
		setLastError(err)
		return nil
	}

	setLastError(nil)
	return C.CString(url)
}

// ===============================
// Tun2Socks 函数导出
// ===============================

var (
	tun2socksInstances   = make(map[int64]*Tun2SocksInstance)
	tun2socksInstancesMu sync.RWMutex
	nextTun2SocksID      int64 = 1
)

//export Tun2SocksNew
func Tun2SocksNew(configJSON *C.char) C.longlong {
	configStr := C.GoString(configJSON)

	var config Tun2SocksConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		setLastError(err)
		return -1
	}

	instance := NewTun2SocksInstance(&config)

	tun2socksInstancesMu.Lock()
	defer tun2socksInstancesMu.Unlock()

	id := nextTun2SocksID
	nextTun2SocksID++
	tun2socksInstances[id] = instance

	setLastError(nil)
	return C.longlong(id)
}

//export Tun2SocksStart
func Tun2SocksStart(id C.longlong) C.int {
	tun2socksInstancesMu.RLock()
	instance, exists := tun2socksInstances[int64(id)]
	tun2socksInstancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("tun2socks instance not found"))
		return -1
	}

	if err := instance.Start(); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export Tun2SocksStop
func Tun2SocksStop(id C.longlong) C.int {
	tun2socksInstancesMu.RLock()
	instance, exists := tun2socksInstances[int64(id)]
	tun2socksInstancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("tun2socks instance not found"))
		return -1
	}

	if err := instance.Stop(); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export Tun2SocksDelete
func Tun2SocksDelete(id C.longlong) C.int {
	tun2socksInstancesMu.Lock()
	defer tun2socksInstancesMu.Unlock()

	instanceID := int64(id)
	instance, exists := tun2socksInstances[instanceID]
	if !exists {
		setLastError(fmt.Errorf("tun2socks instance not found"))
		return -1
	}

	if instance.IsRunning() {
		if err := instance.Stop(); err != nil {
			setLastError(err)
			return -1
		}
	}

	delete(tun2socksInstances, instanceID)
	setLastError(nil)
	return 0
}

//export Tun2SocksIsRunning
func Tun2SocksIsRunning(id C.longlong) C.int {
	tun2socksInstancesMu.RLock()
	instance, exists := tun2socksInstances[int64(id)]
	tun2socksInstancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("tun2socks instance not found"))
		return -1
	}

	if instance.IsRunning() {
		setLastError(nil)
		return 1
	}

	setLastError(nil)
	return 0
}

//export Tun2SocksGetStats
func Tun2SocksGetStats(id C.longlong) *C.char {
	tun2socksInstancesMu.RLock()
	instance, exists := tun2socksInstances[int64(id)]
	tun2socksInstancesMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("tun2socks instance not found"))
		return nil
	}

	bytesUp, bytesDown := instance.GetStats()

	stats := map[string]interface{}{
		"bytes_up":   bytesUp,
		"bytes_down": bytesDown,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		setLastError(err)
		return nil
	}

	setLastError(nil)
	return C.CString(string(data))
}

// ===============================
// 资产管理函数导出
// ===============================

var (
	assetManagers   = make(map[int64]*AssetManager)
	assetManagersMu sync.RWMutex
	nextAssetMgrID  int64 = 1
)

//export AssetManagerNew
func AssetManagerNew(baseDir *C.char) C.longlong {
	dir := C.GoString(baseDir)

	manager := NewAssetManager(dir)

	assetManagersMu.Lock()
	defer assetManagersMu.Unlock()

	id := nextAssetMgrID
	nextAssetMgrID++
	assetManagers[id] = manager

	setLastError(nil)
	return C.longlong(id)
}

//export AssetManagerDelete
func AssetManagerDelete(id C.longlong) C.int {
	assetManagersMu.Lock()
	defer assetManagersMu.Unlock()

	instanceID := int64(id)
	if _, exists := assetManagers[instanceID]; !exists {
		setLastError(fmt.Errorf("asset manager not found"))
		return -1
	}

	delete(assetManagers, instanceID)
	setLastError(nil)
	return 0
}

//export AssetManagerGetInfo
func AssetManagerGetInfo(id C.longlong, assetType *C.char) *C.char {
	assetManagersMu.RLock()
	manager, exists := assetManagers[int64(id)]
	assetManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("asset manager not found"))
		return nil
	}

	aType := AssetType(C.GoString(assetType))
	info, err := manager.GetAssetInfo(aType)
	if err != nil {
		setLastError(err)
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		setLastError(err)
		return nil
	}

	setLastError(nil)
	return C.CString(string(data))
}

//export AssetManagerDownload
func AssetManagerDownload(id C.longlong, assetType *C.char, url *C.char) C.int {
	assetManagersMu.RLock()
	manager, exists := assetManagers[int64(id)]
	assetManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("asset manager not found"))
		return -1
	}

	aType := AssetType(C.GoString(assetType))
	downloadURL := C.GoString(url)

	if err := manager.DownloadAsset(aType, downloadURL, nil); err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export AssetManagerCheckUpdate
func AssetManagerCheckUpdate(id C.longlong, assetType *C.char, url *C.char) C.int {
	assetManagersMu.RLock()
	manager, exists := assetManagers[int64(id)]
	assetManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("asset manager not found"))
		return -1
	}

	aType := AssetType(C.GoString(assetType))
	checkURL := C.GoString(url)

	needsUpdate, err := manager.CheckAssetUpdate(aType, checkURL)
	if err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	if needsUpdate {
		return 1
	}
	return 0
}

//export AssetManagerVerify
func AssetManagerVerify(id C.longlong, assetType *C.char) C.int {
	assetManagersMu.RLock()
	manager, exists := assetManagers[int64(id)]
	assetManagersMu.RUnlock()

	if !exists {
		setLastError(fmt.Errorf("asset manager not found"))
		return -1
	}

	aType := AssetType(C.GoString(assetType))

	valid, err := manager.VerifyAsset(aType)
	if err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	if valid {
		return 1
	}
	return 0
}

func main() {
	// Required for buildmode=c-shared
}
