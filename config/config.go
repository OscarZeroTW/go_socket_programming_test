package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Proxy  ProxyConfig  `yaml:"proxy"`
	Server ServerConfig `yaml:"server"`
	Client ClientConfig `yaml:"client"`
}


type ProxyConfig struct {
	UDPProxy1IP         string `yaml:"udp_proxy1_ip"`
	UDPProxy2IP         string `yaml:"udp_proxy2_ip"`
	UDPProxy1ListenPort string `yaml:"udp_proxy1_listen_port"`
	UDPProxy2ListenPort string `yaml:"udp_proxy2_listen_port"`
}

type ServerConfig struct {
	ServerIP   string `yaml:"server_ip"`
	ServerListenPort string `yaml:"server_listen_port"`
}

type ClientConfig struct {
	ClientIP   string `yaml:"client_ip"`
	ClientListenPort string `yaml:"client_listen_port"`
	Client1ListenPort string `yaml:"client1_listen_port"`
	Client2ListenPort string `yaml:"client2_listen_port"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = findConfigFile()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	return &config, nil
}

func findConfigFile() string {
	possiblePaths := []string{
		"config.yaml",
		"../config.yaml",
		"../../config.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return "config.yaml"
}

func (c *Config) GetProxyConfig() ProxyConfig {
	return c.Proxy
}

func (c *Config) GetServerConfig() ServerConfig {
	return c.Server
}

func (c *Config) GetClientConfig() ClientConfig {
	return c.Client
}

