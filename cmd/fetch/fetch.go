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
	"flag"
	"fmt"

	"github.com/rizhomator/rizhomator/fetcher"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var crawlName string
var crawlSerial string

func main() {
	if flag.NArg() < 1 {
		panic("should give url")
	}

	cmd := fetcher.FetcherCommand{
		CrawlName:       crawlName,
		CrawlSerial:     crawlSerial,
		CrawlGeneration: -1,
		URL:             flag.Arg(0),
	}

	amqpConn, err := amqp.Dial(viper.GetString("amqp_uri"))
	if err != nil {
		panic(err)
	}

	writer := &fetcher.AMQPCommandWriter{Conn: amqpConn}

	writer.Write(cmd)

	fmt.Println("Command sent to queue")
}

func init() {
	flag.StringVar(&crawlName, "crawl", "test", "Crawl name (test)")
	flag.StringVar(&crawlSerial, "serial", "test", "Crawl serial (test)")
	flag.Parse()
	viper.SetDefault("amqp_uri", "CONFIGURE")
	viper.SetDefault("mongodb_uri", "CONFIGURE")
	viper.SetEnvPrefix("REI")
	viper.AutomaticEnv()
}
