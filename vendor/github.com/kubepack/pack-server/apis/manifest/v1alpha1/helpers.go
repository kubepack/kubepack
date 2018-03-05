package v1alpha1

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	yc "github.com/appscode/go/encoding/yaml"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

func LoadConfig(configPath string) (*ComponentConfig, error) {
	if _, err := os.Stat(configPath); err != nil {
		return nil, errors.Errorf("failed to find file %s. Reason: %s", configPath, err)
	}
	os.Chmod(configPath, 0600)

	cfg := &ComponentConfig{}
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	jsonData, err := yc.ToJSON(bytes)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonData, cfg)
	return cfg, err
}

func (cfg ComponentConfig) Save(configPath string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(configPath), 0755)
	if err := ioutil.WriteFile(configPath, data, 0600); err != nil {
		return err
	}
	return nil
}

func (cfg ComponentConfig) Validate() error {
	return nil
}
