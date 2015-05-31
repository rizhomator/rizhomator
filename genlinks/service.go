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

package genlinks

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rizhomator/rizhomator/amqp"
	"github.com/rizhomator/rizhomator/crawl"
	"github.com/rizhomator/rizhomator/fetcher"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	sw "github.com/streadway/amqp"
)

type Service struct {
	Parallelism int
	started     bool
	initialised bool
	amqpConn    *sw.Connection
	wait        sync.WaitGroup
}

func (s *Service) init() (err error) {

	s.amqpConn, err = sw.Dial(viper.GetString("amqp_uri"))
	if err != nil {
		panic(err)
	}

	s.initialised = true

	return
}

func (s *Service) Start() (err error) {

	if s.started {
		return
	}

	if !s.initialised {
		s.init()
	}

	if s.Parallelism < 1 {
		s.Parallelism = 1
	}

	for i := 0; i < s.Parallelism; i++ {
		err = s.thread()
		if err != nil {
			return
		}
	}

	return
}

func (s *Service) thread() (err error) {
	config, err := crawl.LoadConfig("crawlers.yml")
	if err != nil {
		panic(err)
	}

	storage := fetcher.LocalFS{Root: viper.GetString("store_root")}

	redisConn, err := redis.Dial("tcp", viper.GetString("redis_uri"))
	if err != nil {
		panic(err)
	}

	extractor := &LinkExtractor{Storage: storage}
	follower := &Follower{AMQP: s.amqpConn, Config: config, Redis: redisConn}

	queue := amqp.Queue{Conn: s.amqpConn, Name: "crawl.listen.fetchlog.genlinks", FetchCount: 3}

	queue.Check()

	deliveries, err := queue.Consume()

	s.wait.Add(1)

	go func() {
		defer s.wait.Done()
		defer queue.Reset()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		for {
			select {
			case delivery := <-deliveries:

				var log fetcher.FetchLog
				json.Unmarshal(delivery.Body, &log)

				links, err := extractor.Extract(log)
				if err != nil {
					panic(err)
				}

				err = follower.FollowLinks(links)
				if err != nil {
					panic(err)
				}

				delivery.Ack(false)
			case <-c:
				return
			}
		}

	}()

	return
}

func (s *Service) RunOnce() {
	if s.started {
		return
	}

	if !s.initialised {
		s.init()
	}

	config, err := crawl.LoadConfig("crawlers.yml")
	if err != nil {
		panic(err)
	}

	storage := fetcher.LocalFS{Root: viper.GetString("store_root")}

	redisConn, err := redis.Dial("tcp", viper.GetString("redis_uri"))
	if err != nil {
		panic(err)
	}

	extractor := &LinkExtractor{Storage: storage}
	follower := &Follower{AMQP: s.amqpConn, Config: config, Redis: redisConn}

	queue := amqp.Queue{Conn: s.amqpConn, Name: "crawl.listen.fetchlog.genlinks", FetchCount: 3}

	queue.Check()

	deliveries, err := queue.Consume()

	delivery := <-deliveries

	var log fetcher.FetchLog
	json.Unmarshal(delivery.Body, &log)

	links, err := extractor.Extract(log)
	if err != nil {
		panic(err)
	}

	err = follower.FollowLinks(links)
	if err != nil {
		panic(err)
	}

	delivery.Ack(false)

	queue.Reset()

	return
}

func (s *Service) Wait() { s.wait.Wait() }
