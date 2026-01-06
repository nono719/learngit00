package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nono-system/oracle/internal/config"
	"nono-system/oracle/internal/models"
)

// APIDataSource API数据源
type APIDataSource struct {
	name   string
	url    string
	apiKey string
	client *http.Client
}

// NewAPIDataSource 创建API数据源
func NewAPIDataSource(cfg config.DataSourceConfig) (*APIDataSource, error) {
	return &APIDataSource{
		name:   cfg.Name,
		url:    cfg.URL,
		apiKey: cfg.APIKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name 返回数据源名称
func (ds *APIDataSource) Name() string {
	return ds.name
}

// FetchDeviceStatuses 获取所有设备状态
func (ds *APIDataSource) FetchDeviceStatuses() ([]models.DeviceStatus, error) {
	req, err := http.NewRequest("GET", ds.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if ds.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+ds.apiKey)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var statuses []models.DeviceStatus
	if err := json.Unmarshal(body, &statuses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 设置数据源名称
	for i := range statuses {
		statuses[i].Source = ds.name
		statuses[i].Timestamp = time.Now()
	}

	return statuses, nil
}

// FetchDeviceStatus 获取指定设备状态
func (ds *APIDataSource) FetchDeviceStatus(did string) (*models.DeviceStatus, error) {
	url := fmt.Sprintf("%s/%s", ds.url, did)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if ds.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+ds.apiKey)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device not found: %s", did)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var status models.DeviceStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	status.Source = ds.name
	status.Timestamp = time.Now()

	return &status, nil
}

// HealthCheck 健康检查
func (ds *APIDataSource) HealthCheck() error {
	req, err := http.NewRequest("GET", ds.url+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

