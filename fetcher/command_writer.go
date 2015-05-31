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
	"time"

	"github.com/rizhomator/rizhomator/amqp"

	sw "github.com/streadway/amqp"
)

type AMQPCommandWriter struct {
	Conn      *sw.Connection
	QueueName string
	queue     *amqp.Queue
}

func (w *AMQPCommandWriter) Write(cmd FetcherCommand) (err error) {
	if w.queue == nil {
		if w.QueueName == "" {
			w.QueueName = "crawl.commands.fetch"
		}
		w.queue = &amqp.Queue{Conn: w.Conn, Name: w.QueueName}
		w.queue.Check()
	}
	encd, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}

	msg := sw.Publishing{
		DeliveryMode: sw.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         encd,
	}

	err = w.queue.Publish(msg)

	return
}
