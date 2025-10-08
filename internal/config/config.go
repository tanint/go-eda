package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Logger LoggerConfig `mapstructure:"logger"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type KafkaConfig struct {
	Brokers          []string          `mapstructure:"brokers"`
	SecurityProtocol string            `mapstructure:"security_protocol"`
	SASLMechanism    string            `mapstructure:"sasl_mechanism"`
	SASLUsername     string            `mapstructure:"sasl_username"`
	SASLPassword     string            `mapstructure:"sasl_password"`
	GroupID          string            `mapstructure:"group_id"`
	Topics           map[string]string `mapstructure:"topics"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"` // json or console
	OutputPath string `mapstructure:"output_path"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Environment variables
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")

	// Kafka defaults for local development
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.security_protocol", "PLAINTEXT")
	v.SetDefault("kafka.group_id", "default-group")
	v.SetDefault("kafka.topics.order_created", "order.created")
	v.SetDefault("kafka.topics.order_confirmed", "order.confirmed")
	v.SetDefault("kafka.topics.inventory_reserved", "inventory.reserved")

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.encoding", "json")
	v.SetDefault("logger.output_path", "stdout")
}
