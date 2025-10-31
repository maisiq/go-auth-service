package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/maisiq/go-auth-service/internal/logger"
	"github.com/spf13/viper"
)

var config *Config

type DatabaseConfig struct {
	DSN     string `mapstructure:"dsn"`
	MaxConn int    `mapstrcture:"maxconn"`
}

type MemoryDBConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type VaultConfig struct {
	BaseURL string `mapstructure:"base_url"`
	TOKEN   string `mapstructure:"token"`
}

type AppConfig struct {
	Debug   bool   `mapstructure:"debug"`
	Addr    string `mapstructure:"addr"`
	Limiter struct {
		Limit int `mapstructure:"limit"`
		Burst int `mapstructure:"burst"`
	} `mapstructure:"limiter"`
}

type Config struct {
	App      *AppConfig            `mapstructure:"app"`
	Database *DatabaseConfig       `mapstructure:"database"`
	MemoryDB *MemoryDBConfig       `mapstructure:"memorydb"`
	Vault    *VaultConfig          `mapstructure:"vault"`
	Yandex   *YandexProviderConfig `mapstructure:"yandex"`
}

func initConfig(path string, onChange chan<- interface{}) (*viper.Viper, error) {

	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
	)

	v.SetConfigFile(path)
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	v.WatchConfig()

	var lastChange time.Time

	// it won't work on Windows/macOS as docker host with volumes
	// you must change config file manually or create your own config watcher
	// or use ConfigMap instead
	v.OnConfigChange(func(e fsnotify.Event) {
		// prevent multi invoke
		if time.Since(lastChange) < 500*time.Millisecond {
			return
		}
		lastChange = time.Now()
		log := logger.GetLogger()
		log.Infof("file config changed")

		tmpCfg := new(Config)

		if err := v.Unmarshal(tmpCfg); err != nil {
			log.Errorf("failed to reload config: %s. Using previous version", err)
			return
		}
		config = tmpCfg
		log.Info("config updated")
		// send signal to rebuild container or similar
		onChange <- 1

		//TODO: maybe it needs refactor due config init bound to container init
		// have to close because it will be recreated in di-container
		close(onChange)
		v.OnConfigChange(nil)
	})
	return v, nil
}

func LoadConfig(path string, onChange chan<- interface{}) (*Config, error) {
	v, err := initConfig(path, onChange)
	if err != nil {
		panic(err)
	}

	cfg := new(Config)
	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	config = cfg
	return config, nil
}

func GetConfig() *Config {
	if nil == config {
		panic("config is not initialized")
	}
	return config
}
