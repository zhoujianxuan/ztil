package main

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"github.com/urfave/cli/v2"
)

func NewKafkaCommand() *cli.Command {
	return &cli.Command{
		Name:      "kafka_send",
		Aliases:   []string{"ks"},
		Usage:     "Send kafka message",
		UsageText: "ks <topic> <value> [url]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				fmt.Println("Periodic incomplete")
				return nil
			}
			hosts := []string{"192.168.3.54:9092", "192.168.3.53:9092", "192.168.3.37:9092"}
			if c.NArg() == 3 {
				hosts = []string{c.Args().Get(2)}
			}

			client, err := NewClient(hosts)
			if err != nil {
				return err
			}

			topic := c.Args().First()
			value := c.Args().Get(1)

			client.Send(topic, value)
			return nil
		},
	}
}

type Client struct {
	ProducerClient sarama.AsyncProducer
}

func NewClient(server []string) (*Client, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true // 成功交付的消息将在success channel返回
	// 构造一个消息
	// 连接kafka
	client, err := sarama.NewAsyncProducer(server, config)
	if err != nil {
		return nil, err
	}
	return &Client{ProducerClient: client}, nil
}

func (c *Client) Send(topic string, v string) {
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.StringEncoder(v)
	c.ProducerClient.Input() <- msg
	select {
	case m := <-c.ProducerClient.Successes():
		log.Println(m)
	case m := <-c.ProducerClient.Errors():
		log.Println(m)
	}
}

func (c *Client) Close() {
	err := c.ProducerClient.Close()
	if err != nil {
		return
	}
}
