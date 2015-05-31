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

package amqp

import sw "github.com/streadway/amqp"

type Queue struct {
	Conn       *sw.Connection
	Name       string
	FetchCount int
	channel    *sw.Channel
	deliveries <-chan sw.Delivery
}

func (q *Queue) Channel() (channel *sw.Channel, err error) {
	if q.channel != nil {
		return q.channel, nil
	}

	channel, err = q.Conn.Channel()
	if err != nil {
		return nil, err
	}
	q.channel = channel

	return
}

func (q *Queue) Consume() (deliveries <-chan sw.Delivery, err error) {
	if q.deliveries != nil {
		return q.deliveries, nil
	}
	channel, err := q.Channel()
	if err != nil {
		return nil, err
	}

	if q.FetchCount > 0 {
		channel.Qos(q.FetchCount, 0, false)
	}

	deliveries, err = channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q.deliveries = deliveries

	return
}

func (q *Queue) Get() (delivery sw.Delivery, ok bool, err error) {
	channel, err := q.Channel()
	if err != nil {
		return sw.Delivery{}, false, err
	}

	return channel.Get(q.Name, false)
}

func (q *Queue) Publish(msg sw.Publishing) (err error) {
	channel, err := q.Channel()
	if err != nil {
		return
	}

	err = channel.Publish("", q.Name, false, false, msg)
	// TODO confirm it was recieved by the server
	return
}

func (q *Queue) Check() (err error) {
	channel, err := q.Channel()
	if err != nil {
		return
	}
	_, err = channel.QueueDeclarePassive(q.Name, true, false, false, false, nil)
	return
}

func (q *Queue) Reset() (err error) {
	err = q.channel.Close()
	q.deliveries = nil

	return
}
