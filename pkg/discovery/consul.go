package discovery

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"

	"github.com/tvvshow/gokao/pkg/response"
)

// ServiceDiscovery 服务发现接口
type ServiceDiscovery interface {
	// RegisterService 注册服务到服务发现
	RegisterService(serviceName, serviceID, address string, port int, tags []string) error

	// DeregisterService 从服务发现注销服务
	DeregisterService(serviceID string) error

	// DiscoverService 发现服务实例
	DiscoverService(serviceName string) ([]*ServiceInstance, error)

	// WatchService 监听服务变化
	WatchService(serviceName string, callback func(instances []*ServiceInstance))

	// HealthCheck 健康检查
	HealthCheck(serviceID, checkID, notes string) error
}

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	ID        string
	Name      string
	Address   string
	Port      int
	Tags      []string
	Meta      map[string]string
	Healthy   bool
	ServiceID string
}

// ConsulDiscovery Consul服务发现实现
type ConsulDiscovery struct {
	client   *api.Client
	config   *api.Config
	services map[string][]*ServiceInstance
	mutex    sync.RWMutex
	watchers map[string][]func([]*ServiceInstance)
}

// NewConsulDiscovery 创建Consul服务发现实例
func NewConsulDiscovery(address string) (*ConsulDiscovery, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	return &ConsulDiscovery{
		client:   client,
		config:   config,
		services: make(map[string][]*ServiceInstance),
		watchers: make(map[string][]func([]*ServiceInstance)),
	}, nil
}

// RegisterService 注册服务到Consul
func (cd *ConsulDiscovery) RegisterService(serviceName, serviceID, address string, port int, tags []string) error {
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Tags:    tags,
	}

	// 添加健康检查
	registration.Check = &api.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
		Timeout:                        "5s",
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "30s",
	}

	return cd.client.Agent().ServiceRegister(registration)
}

// DeregisterService 从Consul注销服务
func (cd *ConsulDiscovery) DeregisterService(serviceID string) error {
	return cd.client.Agent().ServiceDeregister(serviceID)
}

// DiscoverService 从Consul发现服务实例
func (cd *ConsulDiscovery) DiscoverService(serviceName string) ([]*ServiceInstance, error) {
	services, _, err := cd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	var instances []*ServiceInstance
	for _, service := range services {
		instance := &ServiceInstance{
			ID:        service.Service.ID,
			Name:      service.Service.Service,
			Address:   service.Service.Address,
			Port:      service.Service.Port,
			Tags:      service.Service.Tags,
			Meta:      service.Service.Meta,
			Healthy:   true, // Consul已经过滤了健康实例
			ServiceID: service.Service.ID,
		}
		instances = append(instances, instance)
	}

	// 更新本地缓存
	cd.mutex.Lock()
	cd.services[serviceName] = instances
	cd.mutex.Unlock()

	// 通知观察者
	cd.notifyWatchers(serviceName, instances)

	return instances, nil
}

// WatchService 监听服务变化
func (cd *ConsulDiscovery) WatchService(serviceName string, callback func(instances []*ServiceInstance)) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.watchers[serviceName] = append(cd.watchers[serviceName], callback)

	// 立即获取当前服务实例并回调
	if instances, ok := cd.services[serviceName]; ok {
		go callback(instances)
	}
}

// HealthCheck 执行健康检查
func (cd *ConsulDiscovery) HealthCheck(serviceID, checkID, notes string) error {
	check := &api.AgentCheckRegistration{
		ID:        checkID,
		Name:      "Service Health Check",
		ServiceID: serviceID,
		AgentServiceCheck: api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://localhost:8500/v1/agent/check/%s", checkID),
			Interval: "10s",
			Timeout:  "5s",
			Notes:    notes,
		},
	}

	return cd.client.Agent().CheckRegister(check)
}

// StartBackgroundRefresh 启动后台服务刷新
func (cd *ConsulDiscovery) StartBackgroundRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cd.refreshAllServices()
	}
}

// refreshAllServices 刷新所有服务
func (cd *ConsulDiscovery) refreshAllServices() {
	cd.mutex.RLock()
	serviceNames := make([]string, 0, len(cd.services))
	for name := range cd.services {
		serviceNames = append(serviceNames, name)
	}
	cd.mutex.RUnlock()

	for _, serviceName := range serviceNames {
		instances, err := cd.DiscoverService(serviceName)
		if err != nil {
			log.Printf("Failed to refresh service %s: %v", serviceName, err)
			continue
		}
		log.Printf("Refreshed service %s: %d instances", serviceName, len(instances))
	}
}

// notifyWatchers 通知观察者服务变化
func (cd *ConsulDiscovery) notifyWatchers(serviceName string, instances []*ServiceInstance) {
	cd.mutex.RLock()
	watchers := cd.watchers[serviceName]
	cd.mutex.RUnlock()

	for _, callback := range watchers {
		go callback(instances)
	}
}

// GetServiceAddress 获取服务地址（负载均衡）
func (cd *ConsulDiscovery) GetServiceAddress(serviceName string) (string, error) {
	instances, err := cd.DiscoverService(serviceName)
	if err != nil {
		return "", err
	}

	if len(instances) == 0 {
		return "", fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	// 简单的轮询负载均衡
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	// 使用简单的轮询算法
	currentIndex := 0
	if len(instances) > 1 {
		currentIndex = (currentIndex + 1) % len(instances)
	}

	instance := instances[currentIndex]
	return fmt.Sprintf("http://%s:%d", instance.Address, instance.Port), nil
}

// GetServiceInstances 获取所有服务实例
func (cd *ConsulDiscovery) GetServiceInstances(serviceName string) ([]*ServiceInstance, error) {
	cd.mutex.RLock()
	instances, exists := cd.services[serviceName]
	cd.mutex.RUnlock()

	if !exists {
		return cd.DiscoverService(serviceName)
	}

	return instances, nil
}

// IsServiceHealthy 检查服务是否健康
func (cd *ConsulDiscovery) IsServiceHealthy(serviceName string) (bool, error) {
	instances, err := cd.GetServiceInstances(serviceName)
	if err != nil {
		return false, err
	}

	return len(instances) > 0, nil
}

// ServiceDiscoveryMiddleware Gin中间件，用于服务发现
func (cd *ConsulDiscovery) ServiceDiscoveryMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		address, err := cd.GetServiceAddress(serviceName)
		if err != nil {
			response.AbortWithError(c, http.StatusServiceUnavailable, "service_unavailable",
				fmt.Sprintf("Service %s is unavailable", serviceName), nil)
			return
		}

		// 将服务地址设置到上下文中
		c.Set("service_address", address)
		c.Set("service_name", serviceName)
		c.Next()
	}
}
