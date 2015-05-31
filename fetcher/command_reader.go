/*  Copyright (C) 2015 Leopoldo Lara Vazquez.

    This program is free software: you can redistribute it and/or  modify
    it under the terms of the GNU Affero General Public License, version 3,
	  as published by the Free Software Foundation.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package fetcher

import (
	"encoding/json"

	"github.com/rizhomator/rizhomator/amqp"

	sw "github.com/streadway/amqp"
)

type AMQPCommandReader struct {
	Conn      *sw.Connection
	QueueName string
	//channel   *amqp.Channel
	out   chan FetcherCommand
	queue amqp.Queue
}

func (r *AMQPCommandReader) Consume() (out <-chan FetcherCommand, err error) {
	if r.out != nil {
		return r.out, nil
	}

	r.queue = amqp.Queue{Conn: r.Conn, Name: r.QueueName, FetchCount: 1}
	r.queue.Check()
	deliveries, err := r.queue.Consume()

	r.out = make(chan FetcherCommand)

	logger.Notice("Starting consumer of FetchCommands on queue %s", r.QueueName)
	go func() {
		for delivery := range deliveries {
			var cmd FetcherCommand
			json.Unmarshal(delivery.Body, &cmd)

			logger.Debug("Fetcher command reader recived a command for url %s", cmd.URL)

			cmd.Done = make(chan error)

			r.out <- cmd

			err = <-cmd.Done
			if err == nil {
				delivery.Ack(false)
			} else {
				delivery.Reject(true)
			}
		}
	}()

	return r.out, nil
}

func (r *AMQPCommandReader) Close() error {
	r.out = nil
	return r.queue.Reset()
}
