package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"nono-system/oracle/internal/config"
	"nono-system/oracle/internal/models"
)

// CertificateDataSource 证书数据源
type CertificateDataSource struct {
	name   string
	url    string
	apiKey string
	client *http.Client
}

// NewCertificateDataSource 创建证书数据源
func NewCertificateDataSource(cfg config.DataSourceConfig) (*CertificateDataSource, error) {
	return &CertificateDataSource{
		name:   cfg.Name,
		url:    cfg.URL,
		apiKey: cfg.APIKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name 返回数据源名称
func (ds *CertificateDataSource) Name() string {
	return ds.name
}

// FetchDeviceStatuses 获取所有设备状态
func (ds *CertificateDataSource) FetchDeviceStatuses() ([]models.DeviceStatus, error) {
	req, err := http.NewRequest("GET", ds.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 如果配置了API Key，添加到请求头
	if ds.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+ds.apiKey)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from %s: %w", ds.url, err)
	}
	defer resp.Body.Close()

	// 读取响应体以便获取错误信息
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed (401): API Key may be invalid or user not found. Response: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s: %s", resp.StatusCode, ds.url, string(body))
	}

	var statuses []models.DeviceStatus
	if err := json.Unmarshal(body, &statuses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 设置数据源名称和时间戳
	for i := range statuses {
		statuses[i].Source = ds.name
		statuses[i].Timestamp = time.Now()
	}

	return statuses, nil
}

// FetchDeviceStatus 获取指定设备状态
func (ds *CertificateDataSource) FetchDeviceStatus(did string) (*models.DeviceStatus, error) {
	url := fmt.Sprintf("%s/%s", ds.url, did)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 如果配置了API Key，添加到请求头
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
func (ds *CertificateDataSource) HealthCheck() error {
	// 尝试访问实际的数据源URL（而不是/health端点）
	// 因为后端API可能没有/health端点，或者需要认证
	req, err := http.NewRequest("GET", ds.url, nil)
	if err != nil {
		return err
	}

	// 如果配置了API Key，添加到请求头
	if ds.apiKey != "" {
		authHeader := "Bearer " + ds.apiKey
		req.Header.Set("Authorization", authHeader)
		log.Printf("[HealthCheck] Data source %s: Setting Authorization header: Bearer %s", ds.name, ds.apiKey)
	} else {
		log.Printf("[HealthCheck] Data source %s: WARNING - API Key is empty, request will fail authentication", ds.name)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体以便获取错误信息
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed (401): API Key may be invalid. Response: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

