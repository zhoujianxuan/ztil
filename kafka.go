package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/Shopify/sarama"
	"github.com/urfave/cli/v2"
)

func NewKafkaCommand() *cli.Command {
	return &cli.Command{
		Name:      "kafka_send",
		Aliases:   []string{"ks"},
		Usage:     "Send kafka message",
		UsageText: "ks <topic> <value> <times> [url]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 3 {
				fmt.Println("Periodic incomplete")
				return nil
			}
			hosts := []string{"192.168.3.54:9092", "192.168.3.53:9092", "192.168.3.37:9092"}
			if c.NArg() == 4 {
				hosts = []string{c.Args().Get(3)}
			}

			client, err := NewClient(hosts)
			if err != nil {
				return err
			}

			topic := c.Args().First()
			value := c.Args().Get(1)
			times, err := strconv.Atoi(c.Args().Get(2))
			if err != nil {
				return err
			}

			size := 100
			if times > size {
				size = 1000
			}

			p, _ := ants.NewPool(size)
			defer p.Release()
			wg := sync.WaitGroup{}
			for i := 0; i < times; i++ {
				wg.Add(1)
				_ = p.Submit(func() {
					client.Send(topic, value)
					wg.Done()
				})
			}
			wg.Wait()

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

func NewKafkaConsumerCommand() *cli.Command {
	return &cli.Command{
		Name:      "kafka_receive",
		Aliases:   []string{"kr"},
		Usage:     "receive kafka message",
		UsageText: "kr <topic> [url]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				fmt.Println("Periodic incomplete")
				return nil
			}

			hosts := []string{"192.168.3.54:9092", "192.168.3.53:9092", "192.168.3.37:9092"}
			if c.NArg() >= 2 {
				hosts = []string{c.Args().Get(1)}
			}

			config := sarama.NewConfig()
			config.Consumer.Return.Errors = false
			config.Version = sarama.V2_4_0_0
			config.Consumer.Offsets.Initial = sarama.OffsetNewest
			consumer, err := sarama.NewConsumerGroup(hosts, "ztil", config)
			if err != nil {
				panic(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			go consume(ctx, &consumer, []string{c.Args().First()})

			ch := make(chan os.Signal)
			signal.Notify(ch, os.Interrupt, os.Kill)
			<-ch

			cancel()
			err = consumer.Close()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func consume(ctx context.Context, group *sarama.ConsumerGroup, topics []string) {
	handler := consumerHandler{}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := (*group).Consume(ctx, topics, handler)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

type consumerHandler struct{}

func (consumerHandler) Setup(s sarama.ConsumerGroupSession) error {
	fmt.Println("setup")
	return nil
}
func (consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (h consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d  value:%s\n",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		// 手动确认消息
		sess.MarkMessage(msg, "")
	}
	return nil
}
