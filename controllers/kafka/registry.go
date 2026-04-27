package kafka

import "sync"

/*
Registry bertugas sebagai dependency container
untuk Kafka producer.
*/
type Registry struct {
	brokers  []string
	producer IKafka
	once     sync.Once
	err      error
}

/*
IKafkaRegistry mendefinisikan kontrak registry Kafka.
*/
type IKafkaRegistry interface {
	GetKafkaProducer() (IKafka, error)
}

/*
NewKafkaRegistry membuat registry Kafka.
Producer belum dibuat sampai benar-benar dibutuhkan (lazy init).
*/
func NewKafkaRegistry(brokers []string) IKafkaRegistry {
	return &Registry{
		brokers: brokers,
	}
}

/*
GetKafkaProducer mengembalikan Kafka producer singleton.
Producer hanya dibuat sekali walaupun dipanggil berkali-kali.
*/
func (r *Registry) GetKafkaProducer() (IKafka, error) {
	r.once.Do(func() {
		r.producer, r.err = NewKafkaProducer(r.brokers)
	})
	return r.producer, r.err
}
