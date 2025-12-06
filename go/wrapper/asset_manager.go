package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AssetType 资产类型
type AssetType string

const (
	AssetTypeGeoIP   AssetType = "geoip"
	AssetTypeGeoSite AssetType = "geosite"
)

// AssetManager 资产管理器
type AssetManager struct {
	baseDir    string
	mu         sync.RWMutex
	downloading map[AssetType]bool
}

// AssetInfo 资产信息
type AssetInfo struct {
	Type         AssetType `json:"type"`
	Version      string    `json:"version"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	Path         string    `json:"path"`
	Exists       bool      `json:"exists"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	Type       AssetType `json:"type"`
	Total      int64     `json:"total"`
	Downloaded int64     `json:"downloaded"`
	Percentage float64   `json:"percentage"`
	Status     string    `json:"status"` // downloading, completed, failed
	Error      string    `json:"error,omitempty"`
}

// 默认资产下载地址
const (
	DefaultGeoIPURL   = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat"
	DefaultGeoSiteURL = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat"
)

// NewAssetManager 创建新的资产管理器
func NewAssetManager(baseDir string) *AssetManager {
	return &AssetManager{
		baseDir:     baseDir,
		downloading: make(map[AssetType]bool),
	}
}

// GetAssetPath 获取资产文件路径
func (am *AssetManager) GetAssetPath(assetType AssetType) string {
	filename := string(assetType) + ".dat"
	return filepath.Join(am.baseDir, filename)
}

// GetAssetInfo 获取资产信息
func (am *AssetManager) GetAssetInfo(assetType AssetType) (*AssetInfo, error) {
	path := am.GetAssetPath(assetType)

	info := &AssetInfo{
		Type: assetType,
		Path: path,
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			info.Exists = false
			return info, nil
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	info.Exists = true
	info.Size = fileInfo.Size()
	info.LastModified = fileInfo.ModTime()

	return info, nil
}

// DownloadAsset 下载资产文件
func (am *AssetManager) DownloadAsset(assetType AssetType, url string, progressCallback func(*DownloadProgress)) error {
	am.mu.Lock()
	if am.downloading[assetType] {
		am.mu.Unlock()
		return fmt.Errorf("asset %s is already being downloaded", assetType)
	}
	am.downloading[assetType] = true
	am.mu.Unlock()

	defer func() {
		am.mu.Lock()
		am.downloading[assetType] = false
		am.mu.Unlock()
	}()

	// 使用默认 URL (如果未提供)
	if url == "" {
		switch assetType {
		case AssetTypeGeoIP:
			url = DefaultGeoIPURL
		case AssetTypeGeoSite:
			url = DefaultGeoSiteURL
		default:
			return fmt.Errorf("no default URL for asset type: %s", assetType)
		}
	}

	// 创建目录
	if err := os.MkdirAll(am.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 下载文件
	progress := &DownloadProgress{
		Type:   assetType,
		Status: "downloading",
	}

	if progressCallback != nil {
		progressCallback(progress)
	}

	resp, err := http.Get(url)
	if err != nil {
		progress.Status = "failed"
		progress.Error = err.Error()
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		progress.Status = "failed"
		progress.Error = err.Error()
		if progressCallback != nil {
			progressCallback(progress)
		}
		return err
	}

	progress.Total = resp.ContentLength

	// 创建临时文件
	tmpPath := am.GetAssetPath(assetType) + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		progress.Status = "failed"
		progress.Error = err.Error()
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpPath)

	// 复制数据并报告进度
	buf := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				tmpFile.Close()
				progress.Status = "failed"
				progress.Error = writeErr.Error()
				if progressCallback != nil {
					progressCallback(progress)
				}
				return fmt.Errorf("failed to write: %w", writeErr)
			}

			downloaded += int64(n)
			progress.Downloaded = downloaded
			if progress.Total > 0 {
				progress.Percentage = float64(downloaded) / float64(progress.Total) * 100
			}

			if progressCallback != nil {
				progressCallback(progress)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			progress.Status = "failed"
			progress.Error = err.Error()
			if progressCallback != nil {
				progressCallback(progress)
			}
			return fmt.Errorf("failed to read: %w", err)
		}
	}

	tmpFile.Close()

	// 移动临时文件到最终位置
	finalPath := am.GetAssetPath(assetType)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		progress.Status = "failed"
		progress.Error = err.Error()
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("failed to rename file: %w", err)
	}

	progress.Status = "completed"
	progress.Percentage = 100
	if progressCallback != nil {
		progressCallback(progress)
	}

	return nil
}

// CheckAssetUpdate 检查资产是否需要更新
func (am *AssetManager) CheckAssetUpdate(assetType AssetType, url string) (bool, error) {
	// 获取本地文件信息
	localInfo, err := am.GetAssetInfo(assetType)
	if err != nil {
		return false, err
	}

	if !localInfo.Exists {
		return true, nil // 文件不存在，需要下载
	}

	// 使用默认 URL (如果未提供)
	if url == "" {
		switch assetType {
		case AssetTypeGeoIP:
			url = DefaultGeoIPURL
		case AssetTypeGeoSite:
			url = DefaultGeoSiteURL
		default:
			return false, fmt.Errorf("no default URL for asset type: %s", assetType)
		}
	}

	// 发送 HEAD 请求获取远程文件信息
	resp, err := http.Head(url)
	if err != nil {
		return false, fmt.Errorf("failed to check remote file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// 比较文件大小
	remoteSize := resp.ContentLength
	if remoteSize != localInfo.Size {
		return true, nil
	}

	// 比较最后修改时间
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		remoteTime, err := http.ParseTime(lastModified)
		if err == nil && remoteTime.After(localInfo.LastModified) {
			return true, nil
		}
	}

	return false, nil
}

// DeleteAsset 删除资产文件
func (am *AssetManager) DeleteAsset(assetType AssetType) error {
	path := am.GetAssetPath(assetType)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为成功
		}
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}

// IsDownloading 检查是否正在下载
func (am *AssetManager) IsDownloading(assetType AssetType) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.downloading[assetType]
}

// VerifyAsset 验证资产文件是否有效
func (am *AssetManager) VerifyAsset(assetType AssetType) (bool, error) {
	info, err := am.GetAssetInfo(assetType)
	if err != nil {
		return false, err
	}

	if !info.Exists {
		return false, nil
	}

	// 检查文件大小 (至少应该有几KB)
	if info.Size < 1024 {
		return false, nil
	}

	// 可以添加更多验证逻辑
	// 例如: 检查文件头、校验和等

	return true, nil
}

// GetAllAssets 获取所有资产信息
func (am *AssetManager) GetAllAssets() ([]*AssetInfo, error) {
	assets := make([]*AssetInfo, 0)

	for _, assetType := range []AssetType{AssetTypeGeoIP, AssetTypeGeoSite} {
		info, err := am.GetAssetInfo(assetType)
		if err != nil {
			return nil, err
		}
		assets = append(assets, info)
	}

	return assets, nil
}
