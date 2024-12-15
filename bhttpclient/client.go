package bhttpclient

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type HttpClient struct {
	services map[string]*Service
	mu       sync.RWMutex
}

type Service struct {
	name string
	urls map[string]*ServiceURL
	mu   sync.Mutex
}

type ServiceURL struct {
	url          string
	successCount int
	failCount    int
	status       bool // true: healthy, false: unhealthy
	lastFailTime time.Time
	mu           sync.Mutex
}

type Response struct {
	Success bool
	Err     error
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		services: make(map[string]*Service),
	}
}

func (c *HttpClient) AddService(serviceName string, urls []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	service := &Service{
		name: serviceName,
		urls: make(map[string]*ServiceURL),
	}

	for _, url := range urls {
		service.urls[url] = &ServiceURL{
			url:    url,
			status: true,
		}
	}
	c.services[serviceName] = service
}

// getHealthyService 获取健康的服务地址
func (c *HttpClient) getHealthyService(serviceName string) (*ServiceURL, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[serviceName]
	if !exists {
		return nil, errors.New("service not found")
	}

	// 查找健康的 URL
	var healthyURLs []*ServiceURL
	for _, url := range service.urls {
		url.mu.Lock()
		if url.status {
			healthyURLs = append(healthyURLs, url)
		}
		url.mu.Unlock()
	}

	if len(healthyURLs) == 0 {
		return nil, errors.New("no healthy services available")
	}

	// 随机选择一个健康的 URL
	return healthyURLs[rand.Intn(len(healthyURLs))], nil
}

// monitorHealth 健康检查逻辑
func (url *ServiceURL) monitorHealth() {
	url.mu.Lock()
	defer url.mu.Unlock()

	total := url.successCount + url.failCount
	if total > 0 {
		errorRate := float64(url.failCount) / float64(total)
		if errorRate > 0.5 {
			url.status = false
			url.lastFailTime = time.Now()
		}
	} else {
		url.status = true
	}
}

// retryHealthCheck 自动重试健康服务
func (url *ServiceURL) retryHealthCheck() {
	url.mu.Lock()
	defer url.mu.Unlock()

	// 如果失败时间距离当前不足30秒，暂不恢复
	if time.Since(url.lastFailTime) < 30*time.Second {
		return
	}

	// 测试服务是否恢复
	// [TO-DO] 此处 url 应该指向某个固定可访问的 HealthCheck 接口，需要下游接口同步支持
	resp, err := http.Get(url.url)
	if err == nil && resp.StatusCode == http.StatusOK {
		url.status = true
		url.failCount = 0
		url.successCount = 0
	}
}

// sendRequestWithRetry 带重试的请求逻辑
func (url *ServiceURL) sendRequestWithRetry(retries int) bool {
	for i := 0; i < retries; i++ {
		// [TO-DO] 此处应支持多种请求方式，不仅是 GET，后续增加 POST 的请求支持
		resp, err := http.Get(url.url)
		if err == nil && resp.StatusCode == http.StatusOK {
			url.mu.Lock()
			url.successCount++
			if i != 0 {
				url.failCount += i
			}
			url.mu.Unlock()
			return true
		}
		time.Sleep(100 * time.Millisecond) // 间隔后重试
	}
	url.mu.Lock()
	url.failCount += retries
	url.mu.Unlock()
	return false
}

// StartHealthCheck 启动后台健康检查
func (c *HttpClient) StartHealthCheck(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			c.mu.RLock()
			for _, service := range c.services {
				for _, url := range service.urls {
					url.retryHealthCheck()
				}
			}
			c.mu.RUnlock()
		}
	}()
}

// CallSync 同步调用
func (c *HttpClient) CallSync(serviceName string, timeout time.Duration, retries int) (*http.Response, error) {
	serviceURL, err := c.getHealthyService(serviceName)
	if err != nil {
		return nil, err
	}

	// 带超时的请求上下文
	client := &http.Client{Timeout: timeout}
	// [TO-DO] 此处应支持多种请求方式，不仅是 GET，后续增加 POST 的请求支持
	resp, err := client.Get(serviceURL.url)
	if err != nil {
		success := serviceURL.sendRequestWithRetry(retries)
		serviceURL.monitorHealth()
		if !success {
			return nil, fmt.Errorf("request failed after retries to service: %s", serviceName)
		}
	}

	serviceURL.monitorHealth()
	return resp, nil
}
