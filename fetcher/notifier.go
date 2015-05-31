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

	"github.com/streadway/amqp"
)

type Notifier interface {
	Notify(FetchLog) error
}

type AMQPNotifier struct {
	Conn     *amqp.Connection
	Exchange string
	channel  *amqp.Channel
}

func (n *AMQPNotifier) Notify(log FetchLog) (err error) {
	if n.channel == nil {
		n.channel, err = n.Conn.Channel()
		if err != nil {
			return err
		}
		if n.Exchange == "" {
			n.Exchange = "crawl.notifications.fetchlog"
		}

		err = n.channel.ExchangeDeclarePassive(n.Exchange, "fanout", true, false, false, false, nil)
		if err != nil {
			return err
		}
	}

	encd, err := json.Marshal(log)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         encd,
	}

	err = n.channel.Publish(n.Exchange, "", false, false, msg)
	if err != nil {
		return err
	}

	// TODO NotifyConfirm

	return nil
}

func (n *AMQPNotifier) Close() (err error) {
	if n.channel == nil {
		return
	}

	err = n.channel.Close()
	n.channel = nil
	return
}
