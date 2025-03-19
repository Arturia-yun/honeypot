package policy

import (
	"gopkg.in/yaml.v2"
	"honeypot/Agent/pkg/vars"
	"os"
	"sync"
)

var (
	policyMutex sync.RWMutex
)

// LoadPolicy 从YAML文件加载策略
func LoadPolicy(policyPath string) error {
	policyMutex.Lock()
	defer policyMutex.Unlock()

	// 读取策略文件
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return err
	}

	// 解析YAML
	policyData := &vars.PolicyData{}
	if err := yaml.Unmarshal(data, policyData); err != nil {
		return err
	}

	// 更新全局策略
	vars.GlobalPolicyData = policyData
	return nil
}

// GetPolicy 获取当前策略的副本
func GetPolicy() *vars.PolicyData {
	policyMutex.RLock()
	defer policyMutex.RUnlock()

	if vars.GlobalPolicyData == nil {
		return nil
	}

	// 返回策略的副本以避免并发访问问题
	policyCopy := *vars.GlobalPolicyData
	return &policyCopy
}
