package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

// App config struct
type Config struct {
	Server ServerConfig
	Logger Logger
	Xray   Xray
}

// Server config struct
type ServerConfig struct {
	AppVersion      string
	Mode            string
	DefaultTestPlan string
}

// Logger config
type Logger struct {
	ReportCaller bool
	Encoding     string
	Level        string
	Output       string
}

type Xray struct {
	ClientId     string
	ClientSecret string
}

// Redis config
//type RedisConfig struct {
//	RedisAddr      string
//	RedisPassword  string
//	RedisDB        string
//	RedisDefaultdb string
//	MinIdleConns   int
//	PoolSize       int
//	PoolTimeout    int
//	Password       string
//	DB             int
//}

// MongoDB config
//type MongoDB struct {
//	MongoURI string
//}

// Metrics config
//type Metrics struct {
//	URL         string
//	ServiceName string
//}

//type Jaeger struct {
//	Host        string
//	ServiceName string
//	LogSpans    bool
//}

// Load config file from given path
func LoadConfig(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	return v, nil
}

// Parse config file
func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}
