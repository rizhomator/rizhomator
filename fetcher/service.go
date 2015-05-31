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
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"
)

type Service struct {
	Parallelism int

	started     bool
	initialised bool
	mgoSess     *mgo.Session
	amqpConn    *amqp.Connection
	wait        sync.WaitGroup
}

func (s *Service) init() (err error) {
	s.mgoSess, err = mgo.Dial(viper.GetString("mongodb_uri"))
	if err != nil {
		return
	}

	s.mgoSess.SetSafe(&mgo.Safe{})

	s.amqpConn, err = amqp.Dial(viper.GetString("amqp_uri"))
	if err != nil {
		return
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

	for _, proxy := range s.getProxies() {
		for i := 0; i < s.Parallelism; i++ {
			err = s.thread(proxy)
			if err != nil {
				return
			}
		}
	}

	return
}

func (s *Service) thread(proxy string) (err error) {
	mgoSess := s.mgoSess.Copy()
	logs := &MgoLogs{Collection: mgoSess.DB("crawltest").C("fetchlogs")}
	storage := LocalFS{Root: viper.GetString("store_root")}
	notifier := &AMQPNotifier{Conn: s.amqpConn}

	f := NewStdFetcher(logs, storage, notifier)
	f.SOCKS = proxy

	reader := &AMQPCommandReader{Conn: s.amqpConn, QueueName: "crawl.commands.fetch"}
	cmds, err := reader.Consume()
	if err != nil {
		return
	}

	s.wait.Add(1)

	go func() {
		defer reader.Close()
		defer notifier.Close()
		defer mgoSess.Close()
		defer f.Close()
		defer s.wait.Done()
		defer logger.Info("Fetcher thread stopped")

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		for {
			select {
			case cmd := <-cmds:
				fmt.Printf("Recieved new command for %s\n", cmd.URL)
				err = f.Fetch(cmd)
				if err != nil {
					logger.Error("Fetcher service thread: error from fetching %v, %v", cmd, err)
					panic(err)
				}
				fmt.Println("Command processed")
			case <-interrupt:
				return
			}
		}

	}()

	return
}

func (s *Service) getProxies() (proxies []string) {
	data, err := ioutil.ReadFile("proxies.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &proxies)
	if err != nil {
		panic(err)
	}

	return
}

func (s *Service) Wait() { s.wait.Wait() }
