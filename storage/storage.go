package storage

import (
	"encoding/json"
	"fmt"
)

// NewStorage 根据配置创建存储实例
func NewStorage(config *Config) (Storage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	switch config.Type {
	case StorageTypeLocal:
		localConfig, err := parseLocalConfig(config.Settings)
		if err != nil {
			return nil, fmt.Errorf("invalid local storage config: %w", err)
		}
		return NewLocalStorage(localConfig)

	case StorageTypeDB:
		dbConfig, err := parseDBConfig(config.Settings)
		if err != nil {
			return nil, fmt.Errorf("invalid database storage config: %w", err)
		}
		return NewDatabaseStorage(dbConfig)

	case StorageTypeCOS:
		cosConfig, err := parseCOSConfig(config.Settings)
		if err != nil {
			return nil, fmt.Errorf("invalid COS storage config: %w", err)
		}
		return NewCOSStorage(cosConfig)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// parseLocalConfig 解析本地存储配置
func parseLocalConfig(settings map[string]interface{}) (*LocalConfig, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	var config LocalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// parseDBConfig 解析数据库存储配置
func parseDBConfig(settings map[string]interface{}) (*DBConfig, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	var config DBConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// parseCOSConfig 解析COS存储配置
func parseCOSConfig(settings map[string]interface{}) (*COSConfig, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	var config COSConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// StorageManager 存储管理器
type StorageManager struct {
	storages map[string]Storage
}

// NewStorageManager 创建存储管理器
func NewStorageManager() *StorageManager {
	return &StorageManager{
		storages: make(map[string]Storage),
	}
}

// Register 注册存储实例
func (sm *StorageManager) Register(name string, storage Storage) {
	sm.storages[name] = storage
}

// Get 获取存储实例
func (sm *StorageManager) Get(name string) (Storage, error) {
	storage, exists := sm.storages[name]
	if !exists {
		return nil, fmt.Errorf("storage '%s' not found", name)
	}
	return storage, nil
}

// Close 关闭所有存储连接
func (sm *StorageManager) Close() error {
	var lastErr error
	for _, storage := range sm.storages {
		if err := storage.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// List 列出所有注册的存储
func (sm *StorageManager) List() []string {
	names := make([]string, 0, len(sm.storages))
	for name := range sm.storages {
		names = append(names, name)
	}
	return names
}
