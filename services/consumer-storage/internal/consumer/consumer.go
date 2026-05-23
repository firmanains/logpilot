package consumer

import (
	"fmt"

	"github.com/IBM/sarama"
)

type Consumer struct {
}

func NewConsumer() *Consumer {
	return &Consumer{}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {

	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {

	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			fmt.Println(string(msg.Value))
		case <-session.Context().Done():
			return nil
		}
	}
}
