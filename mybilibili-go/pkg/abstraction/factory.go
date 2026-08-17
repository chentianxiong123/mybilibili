package abstraction

import (
	"errors"
	"time"
)

type Config struct {
	ServiceDiscovery ServiceDiscoveryConfig `yaml:"service_discovery"`
	MessageQueue     MessageQueueConfig     `yaml:"message_queue"`
	CacheStore       CacheStoreConfig       `yaml:"cache_store"`
	ServiceCaller    ServiceCallerConfig    `yaml:"service_caller"`
	StorageService   StorageServiceConfig   `yaml:"storage_service"`
	SearchEngine     SearchEngineConfig     `yaml:"search_engine"`
	DocumentStore    DocumentStoreConfig    `yaml:"document_store"`
}

type ServiceDiscoveryConfig struct {
	Type      string   `yaml:"type"` // memory / file / etcd
	Endpoints []string `yaml:"endpoints"`
	Prefix    string   `yaml:"prefix"`
}

type MessageQueueConfig struct {
	Type         string `yaml:"type"` // memory / file / redis-stream / nats
	Path         string `yaml:"path"` // file 队列目录
	RedisAddr    string `yaml:"redis_addr"`
	StreamPrefix string `yaml:"stream_prefix"`
	NATSURL      string `yaml:"nats_url"`
}

type CacheStoreConfig struct {
	Type       string        `yaml:"type"` // memory / sqlite / redis
	Addr       string        `yaml:"addr"`
	Password   string        `yaml:"password"`
	DB         int           `yaml:"db"`
	Path       string        `yaml:"path"`
	TableName  string        `yaml:"table_name"`
	MaxItems   int           `yaml:"max_items"`
	DefaultTTL time.Duration `yaml:"default_ttl"`
}

type ServiceCallerConfig struct {
	Type    string        `yaml:"type"` // memory / grpc / http
	Timeout time.Duration `yaml:"timeout"`
	Retries int           `yaml:"retries"`
}

type StorageServiceConfig struct {
	Type      string `yaml:"type"` // local / minio / s3
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	Region    string `yaml:"region"`
	RootPath  string `yaml:"root_path"`
}

type SearchEngineConfig struct {
	Type      string   `yaml:"type"` // pg-fts / bleve / elasticsearch
	IndexPath string   `yaml:"index_path"`
	Addresses []string `yaml:"addresses"`
	DSN       string   `yaml:"dsn"`
}

type DocumentStoreConfig struct {
	Type string `yaml:"type"` // pg-jsonb / sqlite / mongodb
	Path string `yaml:"path"`
	DSN  string `yaml:"dsn"`
}

func NewServiceDiscovery(cfg ServiceDiscoveryConfig) (ServiceDiscovery, error) {
	switch cfg.Type {
	case "memory":
		return newMemoryDiscovery(), nil
	case "file":
		return newFileDiscovery(cfg)
	case "etcd":
		return newEtcdDiscovery(cfg)
	default:
		return nil, errors.New("unknown service discovery type: " + cfg.Type)
	}
}

func NewMessageQueue(cfg MessageQueueConfig) (MessageQueue, error) {
	switch cfg.Type {
	case "memory":
		return newMemoryQueue(), nil
	case "file":
		return newFileQueue(cfg)
	case "redis-stream":
		return newRedisStreamQueue(cfg)
	case "nats":
		return newNATSQueue(cfg)
	default:
		return nil, errors.New("unknown message queue type: " + cfg.Type)
	}
}

func NewCacheStore(cfg CacheStoreConfig) (CacheStore, error) {
	switch cfg.Type {
	case "memory":
		return newMemoryCache(cfg)
	case "sqlite":
		return newSQLiteCache(cfg)
	case "redis":
		return newRedisCache(cfg)
	default:
		return nil, errors.New("unknown cache store type: " + cfg.Type)
	}
}

func NewServiceCaller(cfg ServiceCallerConfig) (ServiceCaller, error) {
	switch cfg.Type {
	case "memory":
		return newMemoryCaller(), nil
	case "ollama":
		return newOllamaCaller(), nil
	case "grpc":
		return newGRPCCaller(cfg)
	case "http":
		return newHTTPCaller(cfg)
	default:
		return nil, errors.New("unknown service caller type: " + cfg.Type)
	}
}

func NewStorageService(cfg StorageServiceConfig) (StorageService, error) {
	switch cfg.Type {
	case "memory":
		return newMemoryStorage(), nil
	case "local":
		return newLocalStorage(cfg)
	case "minio":
		return newMinioStorage(cfg)
	case "s3":
		return newS3Storage(cfg)
	default:
		return nil, errors.New("unknown storage service type: " + cfg.Type)
	}
}

func NewSearchEngine(cfg SearchEngineConfig) (SearchEngine, error) {
	switch cfg.Type {
	case "memory":
		return newMemorySearch(), nil
	case "pg-fts":
		return newPGFTS(cfg)
	case "bleve":
		return newBleveSearch(cfg)
	case "elasticsearch":
		return newElasticSearch(cfg)
	default:
		return nil, errors.New("unknown search engine type: " + cfg.Type)
	}
}

func NewDocumentStore(cfg DocumentStoreConfig) (DocumentStore, error) {
	switch cfg.Type {
	case "memory":
		return newMemoryDocStore(), nil
	case "pg-jsonb":
		return newPGJSONB(cfg)
	case "sqlite":
		return newSQLiteDocument(cfg)
	case "mongodb":
		return newMongoDocument(cfg)
	default:
		return nil, errors.New("unknown document store type: " + cfg.Type)
	}
}
