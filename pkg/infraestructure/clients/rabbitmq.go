package clients

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func NewRabbitMQ(url string, userS ports.UsersRepository) (*RabbitMQ, error) {
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
		false,  // auto ack
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

			fmt.Printf("message received: %v\nwith routing key: %v\n", user, d.RoutingKey)
			switch d.RoutingKey {
			case "users.event.register":
				err := userS.Save(&user)
				if err == nil {
					d.Ack(false)
				}
			case "users.event.update":
				fmt.Printf("%v to be implemented", d.RoutingKey)
			case "users.event.delete":
				fmt.Printf("%v to be implemented", d.RoutingKey)
			}
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
