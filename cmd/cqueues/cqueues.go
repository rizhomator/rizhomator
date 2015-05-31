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

package main

import (
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func main() {
	amqpConn, err := amqp.Dial(viper.GetString("amqp_uri"))
	if err != nil {
		panic(err)
	}

	channel, err := amqpConn.Channel()

	_, err = channel.QueueDeclare("crawl.commands.fetch", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	_, err = channel.QueueDeclare("crawl.listen.fetchlog.genlinks", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	err = channel.ExchangeDeclare("crawl.notifications.fetchlog", "fanout", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	err = channel.QueueBind("crawl.listen.fetchlog.genlinks", "", "crawl.notifications.fetchlog", false, nil)
	if err != nil {
		panic(err)
	}
}

func init() {
	viper.SetDefault("amqp_uri", "CONFIGURE")
	viper.SetEnvPrefix("REI")
	viper.AutomaticEnv()
}
