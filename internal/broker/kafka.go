package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	Broker     string
	Topic      string
	Connection *kafka.Conn
	Key        string
	Payload    []byte
}

// ConsumeSalesOrders
func (b *Kafka) ConsumeSalesOrders(ctx context.Context) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{b.Broker},
		Topic:    b.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		log.Printf("read error: %v", err)
	}

	fmt.Printf("consuming key = %s  value = %s", msg.Key, msg.Value)

	return nil

	/*
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		log.Printf("consuming from topic=%q on broker=%q", b.Topic, b.Broker)
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return err // clean shutdown
				}

				log.Printf("read error: %v", err)
				continue
			}

			bal := balances.TxnLog{}
			if err := json.Unmarshal(msg.Value, &bal); err != nil {
				log.Printf("unmarshal error: %v", err)
				continue
			}

			// fmt.Printf("received  partition=%d  offset=%d  product=%-22s  qty=%d  total=$%.2f  region=%s\n",
			// 	msg.Partition, msg.Offset, sale.Product, sale.Quantity, sale.Total, sale.Region)
			fmt.Printf("consuming key = %s \n value = %s \n\n", msg.Key, msg.Value)
		}
	*/
}

func (b *Kafka) StartSalesConsumer(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{b.Broker},
		Topic:    b.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Printf("consumer started  broker=%s  topic=%s", b.Broker, b.Topic)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("consumer stopped")
				return
			}
			log.Printf("read error: %v", err)
			continue
		}

		fmt.Printf("val = %s\n", msg.Value)
		order := SalesOrder{}

		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}

		err = order.ProcessOrder(ctx)
		if err != nil {
			log.Println("salesorder error    failed to process order    err =")
		}

	}
}
