package middle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type Config struct {
	Name      string `json:"name"`
	UploadKey string `json:"key"`
}

func getConfigPath() string {
	path, err := xdg.ConfigFile("rethink/config.json")
	if err == nil {
		return path
	} else {
		fmt.Printf("Something went wrong while finding the config file: %s", err.Error())
	}
	return path
}

func ReadConfig() Config {
	cfg := &Config{}
	cfgFile, err := os.Open(getConfigPath())
	if err != nil {
		fmt.Printf("Failed to open config: %s\n", err.Error())
	}
	defer cfgFile.Close()
	if json.NewDecoder(cfgFile).Decode(cfg) != nil {
		fmt.Printf("Failed to decode config\n")
		return *cfg
	}
	WriteConfig(*cfg)
	return *cfg
}

func WriteConfig(cfg Config) {
	cfgPath := getConfigPath()
	os.MkdirAll(filepath.Dir(cfgPath), os.ModePerm)
	cfgFile, err := os.OpenFile(cfgPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to write config: %s\n", err.Error())
		return
	}
	defer cfgFile.Close()
	if json.NewEncoder(cfgFile).Encode(cfg) != nil {
		fmt.Printf("Failed to encode config\n")
	}
}
