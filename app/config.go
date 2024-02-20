package app

import (
	"embed"
	"io/fs"
	"math/big"
	"os"
	"path"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	//go:embed embed
	embedFS      embed.FS
	unwrapFSOnce sync.Once
	unwrappedFS  fs.FS
)

func FS() fs.FS {
	unwrapFSOnce.Do(func() {
		fsys, err := fs.Sub(embedFS, "embed")
		if err != nil {
			panic(err)
		}
		unwrappedFS = fsys
	})
	return unwrappedFS
}

type Config struct {
	Service string
	Env     string

	Debug     bool   `yaml:"debug"`
	SecretKey string `yaml:"secret_key"`

	DB struct {
		DSN      string `yaml:"dsn"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"db"`
	ChainID *big.Int `yaml:"chain_id"`
	Client  string   `yaml:"client"`
}

func ReadConfig(fsys fs.FS, service, env string) (*Config, error) {
	configContent, err := fs.ReadFile(fsys, path.Join("config", env+".yaml"))

	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	configContent = []byte(os.Expand(string(configContent), defaultValueMapper))

	if err := yaml.Unmarshal(configContent, cfg); err != nil {
		return nil, err
	}

	cfg.Service = service
	cfg.Env = env

	return cfg, nil
}

func defaultValueMapper(placeholderName string) string {
	split := strings.SplitN(placeholderName, ":", 2)
	defValue := ""
	if len(split) == 2 {
		placeholderName = split[0]
		defValue = split[1]
	}
	val, ok := os.LookupEnv(placeholderName)
	if !ok || val == "" {
		return defValue
	}

	return val
}
