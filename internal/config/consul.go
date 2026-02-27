package config

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

func createConsulClient(consulCfg *ConsulConfig) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = consulCfg.Address

	if consulCfg.Token != "" {
		config.Token = consulCfg.Token
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return client, nil
}

func LoadFromConsul(consulCfg *ConsulConfig) (*Config, error) {
	client, err := createConsulClient(consulCfg)
	if err != nil {
		return nil, err
	}

	kv := client.KV()
	pair, _, err := kv.Get(consulCfg.Key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from consul: %w", err)
	}

	if pair == nil {
		return nil, fmt.Errorf("config not found in consul (key: %s)", consulCfg.Key)
	}

	cfg := &Config{
		Providers: make([]*ProviderConfig, 0),
	}

	if err := yaml.Unmarshal(pair.Value, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from consul: %w", err)
	}

	return cfg, nil
}

func SaveToConsul(consulCfg *ConsulConfig, cfg *Config) error {
	client, err := createConsulClient(consulCfg)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	kv := client.KV()
	pair := &api.KVPair{
		Key:   consulCfg.Key,
		Value: data,
	}

	_, err = kv.Put(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to save config to consul: %w", err)
	}

	return nil
}
