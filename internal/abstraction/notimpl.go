package abstraction

import "errors"

func newFileDiscovery(cfg ServiceDiscoveryConfig) (ServiceDiscovery, error) {
	return nil, errors.New("file discovery not implemented")
}

func newEtcdDiscovery(cfg ServiceDiscoveryConfig) (ServiceDiscovery, error) {
	return nil, errors.New("etcd discovery not implemented")
}

func newRedisStreamQueue(cfg MessageQueueConfig) (MessageQueue, error) {
	return nil, errors.New("redis stream queue not implemented")
}

func newNATSQueue(cfg MessageQueueConfig) (MessageQueue, error) {
	return nil, errors.New("nats queue not implemented")
}

func newSQLiteCache(cfg CacheStoreConfig) (CacheStore, error) {
	return nil, errors.New("sqlite cache not implemented")
}

func newRedisCache(cfg CacheStoreConfig) (CacheStore, error) {
	return nil, errors.New("redis cache not implemented")
}

func newGRPCCaller(cfg ServiceCallerConfig) (ServiceCaller, error) {
	return nil, errors.New("grpc caller not implemented")
}

func newHTTPCaller(cfg ServiceCallerConfig) (ServiceCaller, error) {
	return nil, errors.New("http caller not implemented")
}

func newLocalStorage(cfg StorageServiceConfig) (StorageService, error) {
	return nil, errors.New("local storage not implemented")
}

func newMinioStorage(cfg StorageServiceConfig) (StorageService, error) {
	return nil, errors.New("minio storage not implemented")
}

func newS3Storage(cfg StorageServiceConfig) (StorageService, error) {
	return nil, errors.New("s3 storage not implemented")
}

func newPGFTS(cfg SearchEngineConfig) (SearchEngine, error) {
	return nil, errors.New("pg fts not implemented")
}

func newBleveSearch(cfg SearchEngineConfig) (SearchEngine, error) {
	return nil, errors.New("bleve search not implemented")
}

func newElasticSearch(cfg SearchEngineConfig) (SearchEngine, error) {
	return nil, errors.New("elastic search not implemented")
}

func newPGJSONB(cfg DocumentStoreConfig) (DocumentStore, error) {
	return nil, errors.New("pg jsonb not implemented")
}

func newSQLiteDocument(cfg DocumentStoreConfig) (DocumentStore, error) {
	return nil, errors.New("sqlite document not implemented")
}

func newMongoDocument(cfg DocumentStoreConfig) (DocumentStore, error) {
	return nil, errors.New("mongo document not implemented")
}