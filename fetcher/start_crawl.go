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
	"time"

	"github.com/rizhomator/rizhomator/crawl"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func StartCrawl(configFile string, crawlName string) {
	amqpConn, err := amqp.Dial(viper.GetString("amqp_uri"))
	if err != nil {
		panic(err)
	}

	config, err := crawl.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	crawlConfig, ok := config[crawlName]
	if !ok {
		panic("Not found crawl " + crawlName)
	}

	crawlSerial := time.Now().Format(time.RFC3339)

	cmd := FetcherCommand{CrawlName: crawlName, CrawlSerial: crawlSerial}

	cmdWriter := &AMQPCommandWriter{Conn: amqpConn}

	for _, seed := range crawlConfig.Seeds {
		cmd.URL = seed
		err = cmdWriter.Write(cmd)
		if err != nil {
			panic(err)
		}
	}
}
