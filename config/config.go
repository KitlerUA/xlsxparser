package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//Config - contains some configuration data
type Config struct {
	Page          string            `json:"page"`
	PagesNames    map[string]string `json:"pages_names"`
	Name          string            `json:"policy_name"`
	RolesBegin    string            `json:"roles_begin"`
	RolesEnd      string            `json:"roles_end"`
	Type          string            `json:"type"`
	TechGroupName string            `json:"technical_group_name"`
	DisplayName   string            `json:"display_name"`
}

var (
	config *Config
	once   sync.Once
	e      error
)

//Init - read config from file and send error in given channel
func Init() error {
	once.Do(func() {
		config = &Config{}
		e = loadConfig()
	})
	return e
}

//Get - return copy of config
func Get() Config {
	return *config
}

func loadConfig() error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("cannot find location of executable: %s", err)
	}
	data, err := ioutil.ReadFile(dir + "/config.json")
	if err != nil {
		return fmt.Errorf("cannot find config file: %s<br>Please, add config file and restart program", err)
	}
	config = &Config{}
	if err = json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("corrupted data in config file: %s<br>Please, fix config and restart program", err)
	}
	for i := range config.PagesNames {
		if _, ok := config.PagesNames[strings.ToLower(i)]; !ok {
			config.PagesNames[strings.ToLower(i)] = config.PagesNames[i]
			delete(config.PagesNames, i)
		}
	}
	return nil
}
