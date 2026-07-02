package kafka

import (
	configApp "payment-service/config"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

/*
Kafka producer bertugas mengirim message ke Kafka cluster.
Producer:
- mengirim data ke topic
- tidak peduli siapa consumer-nya
- fokus pada durability & ordering
*/
type Kafka struct {
	brokers  []string
	producer sarama.SyncProducer
}

/*
IKafka adalah kontrak producer Kafka.
- ProduceMessage: kirim message ke topic tertentu
- Close: menutup koneksi producer (dipanggil saat shutdown)
*/
type IKafka interface {
	ProduceMessage(string, uuid.UUID, []byte) error
	Close() error
}

/*
NewKafkaProducer membuat Kafka producer sekali di awal aplikasi.
Producer bersifat heavy object → WAJIB di-reuse.
*/
func NewKafkaProducer(brokers []string) (IKafka, error) {
	config := sarama.NewConfig()

	// wajib untuk SyncProducer
	config.Producer.Return.Successes = true

	// durability tinggi (cocok untuk payment / financial event)
	config.Producer.RequiredAcks = sarama.WaitForAll

	// retry saat timeout / broker down
	config.Producer.Retry.Max = configApp.Config.Kafka.MaxRetry
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	// ===== idempotent producer =====
	// mencegah duplicate message saat retry
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1 // wajib untuk idempotent

	// versi minimal Kafka untuk idempotent
	config.Version = sarama.V2_5_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logrus.Errorf("Failed to create kafka producer: %v", err)
		return nil, err
	}

	return &Kafka{
		brokers:  brokers,
		producer: producer,
	}, nil
}

/*
ProduceMessage mengirim message ke Kafka topic.

Konsep penting:
- topic     : channel / kategori data (misal: payment.settlement)
- key       : menentukan partition → ordering DIJAMIN untuk key yang sama (pakai payment_id)
- partition : urutan hanya dijamin di dalam satu partition
- offset    : nomor urut message dalam partition
*/
func (k *Kafka) ProduceMessage(topic string, key uuid.UUID, data []byte) error {
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key.String()), // key sama → partition sama → no reorder
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := k.producer.SendMessage(message)
	if err != nil {
		logrus.Errorf("Failed to produce message to kafka: %v", err)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"topic":     topic,
		"partition": partition,
		"offset":    offset,
	}).Debug("kafka message produced")

	return nil
}

/*
Close menutup koneksi Kafka producer.
Dipanggil saat aplikasi shutdown (SIGTERM / SIGINT).
*/
func (k *Kafka) Close() error {
	if k.producer != nil {
		return k.producer.Close()
	}
	return nil
}
