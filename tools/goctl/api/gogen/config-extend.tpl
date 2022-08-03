package config

import (
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/timest/env"
	"gopkg.in/yaml.v3"
	"strings"
)

type config struct {
	Nacos struct {
		Host     string
		UserName string
		PassWd   string
		LogDir   string
		CacheDir string
		Port     uint64 `default:"8848"`
		Group    string `default:"DEFAULT_GROUP"`
	}
}

func (m *Config) loadConfig() (string, error) {
	// 获取nacos配置
	cfg := new(config)
	err := env.Fill(cfg)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Loading config err: %v", err))
	}

	dataId := "{{.serviceKey}}" + "-" + "rpc"

	sc := []constant.ServerConfig{{.serverConfig}}


	cc := constant.ClientConfig{
		NamespaceId:         "", // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              cfg.Nacos.LogDir,
		CacheDir:            cfg.Nacos.CacheDir,
		LogLevel:            "info",
		Username:            cfg.Nacos.UserName,
		Password:            cfg.Nacos.PassWd,
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  &cc,
	})
	if err != nil {
		return "", err
	}

	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  "middleware",
	})
	if err != nil {
		return "", err
	}

	return content, err
}

func (m *Config) readMapFromString(content string, key string) (interface{}, error) {
	resultMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(content), &resultMap); err != nil {
		return nil, err
	}

	var info interface{}
	keyArr := strings.Split(key, ".")
	keyLength := len(keyArr)
	for i := 0; i < keyLength; i++ {
		if val, ok := resultMap[keyArr[i]]; ok {
			info = val
			switch val.(type) {
			case string:
				break
			case []interface{}:
				break
			case interface{}:
				resultMap = val.(map[string]interface{})
			}
		} else {
			return nil, errors.New("查找不到对应配置")
		}
	}

	return info, nil
}

func (m *Config) ConfigInterface(key string) (interface{}, error) {
	content, err := m.loadConfig()
	if err != nil {
		return nil, err
	}

	info, err := m.readMapFromString(content, key)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (m *Config) ConfigString(key string) (string, error) {
	var result string
	content, err := m.loadConfig()
	if err != nil {
		return result, err
	}

	info, err := m.readMapFromString(content, key)
	if err != nil {
		return result, err
	}

	if result, ok := info.(string); !ok {
		return result, errors.New("查找不到对应配置")
	}

	return result, nil
}
