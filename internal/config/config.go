/*
Package config: пакет содержит объявления всех структур, содержащих конфигурацию компонентов приложения и метод их
загрузки из файла. Заполнение данными этих структур производится вызовом метода MustLoad, который производит чтение
файла конфигурации. Путь к файлу указывается в командной строке по ключу 'config' или считывается из переменной
окружения 'CONFIG_PATH'

# Структуры конфигурации

1. Config - структура, содержащая все остальные конфигурации

2. Kafka - структура, содержащая названия топиков и брокеры Apache Kafka

3. PersistentStorage - настройки реляционной СУБД, используемой в качестве постоянного хранилища

*/

package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

const (
	EnvironmentLocal      = "local"
	EnvironmentDebug      = "debug"
	EnvironmentProduction = "production"
)

type Config struct {
	Kafka             `yaml:"kafka"`
	PersistentStorage `yaml:"persistent_storage"`
}

type Kafka struct {
	Brokers      []string `yaml:"kafka_brokers" env:"KAFKA_BROKERS"`
	MessageTopic string   `yaml:"kafka_message_topic" env:"KAFKA_MESSAGE_TOPIC"`
	ConfirmTopic string   `yaml:"kafka_confirm_topic" env:"KAFKA_CONFIRM_TOPIC"`
}

type PersistentStorage struct {
	DatabaseLogin              string `yaml:"database_login" env:"DATABASE_LOGIN" env-required:"true"`
	DatabasePassword           string `yaml:"database_password" env:"DATABASE_PASSWORD" env-required:"true"`
	DatabaseAddress            string `yaml:"database_address" env:"DATABASE_ADDRESS" env-required:"true"`
	DatabasePort               int    `yaml:"database_port" env:"DATABASE_PORT" env-required:"true"`
	DatabaseName               string `yaml:"database_name" env:"DATABASE_NAME" env-required:"true"`
	DatabaseSchema             string `yaml:"database_schema" env:"DATABASE_SCHEMA"`
	DatabaseMaxOpenConnections int    `yaml:"database_max_open_connections" env:"DATABASE_MAX_OPEN_CONNECTIONS" env-required:"true"`

	QueryTimeout time.Duration `yaml:"query_timeout" env:"QUERY_TIMEOUT" env-required:"true"`
}

// MustLoad возвращает конфигурацию, считанную из файла, путь к которому передан из командной строки по флагу config или
// содержится в переменной окружения CONFIG_PATH. Для переопределения конфигурационных значений можно использовать
// переменные окружения (описанные в структурах данных в этом файле).
func MustLoad() *Config {
	var configPath = flag.String("config", "", "путь к файлу конфигурации")
	var cfg Config

	flag.Parse()

	if *configPath == "" {
		*configPath = os.Getenv("CONFIG_PATH")
	}

	if *configPath == "" {
		log.Fatal("config path is not set")
	}

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", *configPath)
	}

	if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
