/*
Package config: пакет содержит объявления всех структур, содержащих конфигурацию компонентов приложения и метод их
загрузки из файла. Заполнение данными этих структур производится вызовом метода MustLoad, который производит чтение
файла конфигурации. Путь к файлу указывается в командной строке по ключу 'config' или считывается из переменной
окружения 'CONFIG_PATH'

# Структуры конфигурации

0. Instance - строка, являющаяся уникальным идентификатором экземпляра приложения в системе

1. Env - уровень запуска приложения (EnvironmentLocal, EnvironmentDebug или EnvironmentProduction)

2. Config - структура, содержащая все остальные конфигурации

3. Kafka - структура, содержащая названия топиков и брокеры Apache Kafka

4. PersistentStorage - настройки реляционной СУБД, используемой в качестве постоянного хранилища

5. HttpServer - конфигурация http-сервера

6. Service - конфигурация сервисной логики

7. Redis - конфигурация redis-сервера

8. Prometheus - конфигурация метрик

9. Outbox - используемый для хранения не сохраненных данных метод - Naive (простое сохранение в память) или Redis (в списке Redis)

*/

package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"io/ioutil"
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
	HttpServer        `yaml:"http_server"`
	Service           `yaml:"service"`
	Prometheus        `yaml:"prometheus"`
	Redis             `yaml:"redis"`
	Outbox            string `yaml:"outbox" env-required:"true"`
	Instance          string `yaml:"instance" env-required:"true"`
	Env               string `yaml:"env" env:"ENV" env-required:"true"`
}

type Kafka struct {
	Brokers                  []string      `yaml:"kafka_brokers" env:"KAFKA_BROKERS"`
	MessageTopic             string        `yaml:"kafka_message_topic" env:"KAFKA_MESSAGE_TOPIC"`
	ConfirmTopic             string        `yaml:"kafka_confirm_topic" env:"KAFKA_CONFIRM_TOPIC"`
	KafkaWriteTimeout        time.Duration `yaml:"kafka_write_timeout" env:"KAFKA_WRITE_TIMEOUT" env-required:"true"`
	KafkaTimeBetweenAttempts time.Duration `yaml:"kafka_time_between_attempts" env:"KAFKA_TIME_BETWEEN_ATTEMPTS" env-required:"true"`
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

type HttpServer struct {
	HttpPort        string        `yaml:"http_port" env:"HTTP_PORT" env-required:"true"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-required:"true"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-required:"true"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-required:"true"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-required:"true"`
	RequestTimeout  time.Duration `yaml:"request_timeout" env:"REQUEST_TIMEOUT" env-required:"true"`
	EnableProfiler  bool          `yaml:"enable_profiler" env:"ENABLE_PROFILER"`
	SecureKey       string        `yaml:"secure_key" env:"SECURE_KEY" env-required:"true"`
}

type Prometheus struct {
	PrometheusPort       string `yaml:"prometheus_port" env:"PROMETHEUS_PORT"`
	PrometheusMetricsURL string `yaml:"prometheus_metrics_url" env:"PROMETHEUS_METRICS_URL"`
}

type Redis struct {
	RedisAddress  string `yaml:"redis_address" env:"REDIS_ADDRESS"`
	RedisUser     string `yaml:"redis_user" env:"REDIS_USER"`
	RedisPassword string `yaml:"redis_password" env:"REDIS_PWD"`
	RedisDB       int    `yaml:"redis_db" env:"REDIS_DB"`
}

type Service struct {
	RetryTimeout time.Duration `yaml:"retry_timeout" env:"RETRY_TIMEOUT" env-required:"true"`
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

// ReadSecretsToEnv считывает Docker secrets из папки /run/secrets/ и заносит их в переменные окружения. В качестве
// параметра функция принимает карту, где ключами служат названия переменных окружения, а значениями - названия файлов,
// содержащих секреты, которые необходимо занести в эти переменные окружения.
func ReadSecretsToEnv(secrets map[string]string) {
	for envName, fileName := range secrets {
		secret, err := ioutil.ReadFile(fmt.Sprintf("/run/secrets/%s", fileName))
		if err != nil {
			continue
		}

		_ = os.Setenv(envName, string(secret))
	}
}
