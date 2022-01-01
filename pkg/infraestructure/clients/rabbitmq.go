package clients

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err = ch.ExchangeDeclare(
		"users",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	events := []string{"users.event.register", "users.event.update", "users.event.delete"}

	for _, event := range events {
		if err = ch.QueueBind(
			q.Name,  // queue name
			event,   // routing key
			"users", // exchange
			false,
			nil,
		); err != nil {
			return nil, err
		}
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range msgs {
			body := d.Body
			reader := bytes.NewReader(body)
			var user domain.User

			gob.NewDecoder(reader).Decode(&user)
			// TODO: implement all the user login related to user and wire it up with
			// 			the received messages
			fmt.Printf("message received: %v\nwith routing key: %v\n", user, d.RoutingKey)
		}
	}()

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	r.Connection.Close()
}
