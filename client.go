package notifier

import (
	"crypto/tls"
	"fmt"

	"github.com/nsqio/go-nsq"
)

type NotifyClient struct {
	hostPort string
	queue    *nsq.Consumer
	stop     chan struct{}
}

func (nc *NotifyClient) Close() {
	close(nc.stop)
}

func (nc *NotifyClient) AddHandler(cb func(message *nsq.Message) error) {
	nc.queue.AddHandler(nsq.HandlerFunc(cb))
}

func (nc *NotifyClient) Connect(hostPort string) error {
	err := nc.queue.ConnectToNSQD(hostPort)
	if err != nil {
		return fmt.Errorf("can not connect to queue server. error: %s", err.Error())
	}
	go func() {
		<-nc.stop
		nc.queue.DisconnectFromNSQD(hostPort)
	}()
	return nil
}

func New(topic, channel string) *NotifyClient {
	config := nsq.NewConfig()
	q, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		panic(err)
	}
	return &NotifyClient{
		queue: q,
		stop:  make(chan struct{}),
	}
}

func NewWithTLS(topic, channel, certPath, keyPath string) *NotifyClient {
	config := nsq.NewConfig()
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			panic(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	config.TlsV1 = true
	config.TlsConfig = tlsConfig

	q, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		panic(err)
	}
	return &NotifyClient{
		queue: q,
		stop:  make(chan struct{}),
	}
}
