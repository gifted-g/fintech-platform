package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writers map[string]*kafka.Writer
}

func NewProducer(brokers []string) (*Producer, error) {
	return &Producer{
		writers: make(map[string]*kafka.Writer),
	}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, message []byte) error {
	writer, exists := p.writers[topic]
	if !exists {
		writer = &kafka.Writer{
			Addr:     kafka.TCP("localhost:9092"),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
		p.writers[topic] = writer
	}

	err := writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	for _, writer := range p.writers {
		if err := writer.Close(); err != nil {
			return err
		}
	}
	return nil
}
