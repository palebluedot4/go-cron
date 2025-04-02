package config

import (
	"sync"
	"time"

	"go-cron/internal/environment"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
}

type ServerConfig struct {
	Port               int                     `mapstructure:"port"`
	Env                environment.Environment `mapstructure:"env"`
	ShutdownTimeout    time.Duration           `mapstructure:"shutdown_timeout"`
	AppShutdownTimeout time.Duration           `mapstructure:"app_shutdown_timeout"`
	ReadTimeout        time.Duration           `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration           `mapstructure:"write_timeout"`
	IdleTimeout        time.Duration           `mapstructure:"idle_timeout"`
	LogLevel           string                  `mapstructure:"log_level"`
	LogOutput          LogOutputConfig         `mapstructure:"log_output"`
}

type LogOutputConfig struct {
	Console bool `mapstructure:"console"`
	File    bool `mapstructure:"file"`
}

type StorageConfig struct {
	ConnectTimeout  time.Duration  `mapstructure:"connect_timeout"`
	ShutdownTimeout time.Duration  `mapstructure:"shutdown_timeout"`
	Postgres        PostgresConfig `mapstructure:"postgres"`
	MongoDB         MongoDBConfig  `mapstructure:"mongodb"`
	Redis           RedisConfig    `mapstructure:"redis"`
}

type PostgresConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type MongoDBConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	URI         string `mapstructure:"uri"`
	Database    string `mapstructure:"database"`
	MaxPoolSize uint64 `mapstructure:"max_pool_size"`
}

type RedisConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	MaxRetries   int           `mapstructure:"max_retries"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
}

var (
	instance   *Config
	once       sync.Once
	cfgError   error
	instanceMu sync.RWMutex
)

func Init() error {
	once.Do(func() {
		instance, cfgError = Load()
	})
	return cfgError
}

func Instance() (*Config, error) {
	instanceMu.RLock()
	if instance != nil {
		defer instanceMu.RUnlock()
		return instance, nil
	}
	instanceMu.RUnlock()

	if err := Init(); err != nil {
		return nil, err
	}

	instanceMu.RLock()
	defer instanceMu.RUnlock()
	return instance, nil
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("internal/config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		newConfig := &Config{}
		if err := viper.Unmarshal(newConfig); err == nil {
			instanceMu.Lock()
			instance = newConfig
			instanceMu.Unlock()

			callbackMu.RLock()
			for _, callback := range callbacks {
				callback(instance, e.Name)
			}
			callbackMu.RUnlock()
		}
	})

	config := &Config{}
	err := viper.Unmarshal(config)
	return config, err
}

type ChangeCallback func(*Config, string)

var (
	callbacks  []ChangeCallback
	callbackMu sync.RWMutex
)

func RegisterChangeCallback(callback ChangeCallback) {
	callbackMu.Lock()
	defer callbackMu.Unlock()
	callbacks = append(callbacks, callback)
}

func IsValid(cfg *Config, err error) bool {
	return err == nil && cfg != nil
}
