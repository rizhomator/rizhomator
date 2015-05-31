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
	"strings"
	"syscall"
	"time"

	"github.com/rizhomator/rizhomator/crawl"
	"github.com/rizhomator/rizhomator/fetcher"

	"github.com/garyburd/redigo/redis"
	"github.com/streadway/amqp"
)

type Follower struct {
	AMQP        *amqp.Connection
	amqpChannel *amqp.Channel
	// TODO use global
	Config    crawl.Config
	Redis     redis.Conn
	interrupt chan os.Signal
}

func (f *Follower) FollowLinks(links []LinkLog) error {
	if f.interrupt == nil {
		f.interrupt = make(chan os.Signal, 1)
		signal.Notify(f.interrupt, os.Interrupt, syscall.SIGTERM)
	}

	count := 0
	count_followed := 0
	f.checkAMQPQueue()
	for _, link := range links {
		select {
		case <-f.interrupt:
			return nil
		default:
			followed, err := f.FollowLink(link)
			if err != nil {
				panic(err)
			}
			count_followed += followed
			count++
		}
	}

	logger.Info("Folower: Procesed %d links, %d followed", count, count_followed)

	return nil
}

func (f *Follower) FollowLink(link LinkLog) (count int, err error) {
	if len(link.Dest.URL) == 0 {
		logger.Debug("Follower ignoring link with empty destination")
		return
	}

	if link.Crawl.Generation >= f.Config[link.Crawl.Name].MaxGeneration {
		logger.Debug("Follower: Not following link becuse reached max generation %s (%s, %s, %d)", link.Dest.URL, link.Crawl.Name, link.Crawl.Serial, link.Crawl.Generation)
		return
	}

	logger.Debug("Follower: Checking %s (%s, %s, %d)", link.Dest.URL, link.Crawl.Name, link.Crawl.Serial, link.Crawl.Generation)

	already, err := f.isAlreadyFollowed(link)
	if err != nil {
		logger.Debug("Follower: Error reading redis %s, %v", link.Dest.URL, err)
		return 0, err
	}
	if already {
		logger.Debug("Follower ignoring link that was already followed %s", link.Dest.URL)
		return
	}

	if !f.shouldFollow(link) {
		logger.Debug("Follower link should not be followed %s by how it is linked", link.Dest.URL)
		return
	}

	if !f.Config.ShouldFetch(link.Crawl.Name, link.Dest.URL) {
		logger.Debug("Follower link should not be followed %s by crawler config", link.Dest.URL)
		_, err = f.setAsFollowed(link)
		if err != nil {
			return 0, err
		}
		return
	}

	already, err = f.setAsFollowed(link)
	if err != nil {
		return 0, err
	}

	// TODO - do we risk to double download or
	// do we risk loosing one link, currently the later

	if !already {
		logger.Info("Follower following link %s", link.Dest.URL)
		count = 1
		err = f.sendFetchCommand(link)
	} else {
		logger.Debug("Follower ignoring link that was already followed %s", link.Dest.URL)
	}

	return
}

func (f Follower) shouldFollow(link LinkLog) bool {
	if strings.Contains(link.Rel, "nofollow") {
		logger.Debug("Follower link should not be followed because contains `rel=nofollow` %s", link.Dest.URL)
		return false
	}
	if link.Type == Link && !strings.Contains(link.Rel, "stylesheet") {
		logger.Debug("Follower link should not be followed because is not a link to stylesheet %s", link.Dest.URL)
		return false
	}
	return true
}

func (f Follower) isAlreadyFollowed(link LinkLog) (already bool, err error) {
	set := "follower::followedset::" + link.Crawl.Name + "::" + link.Crawl.Serial
	count, err := redis.Int(f.Redis.Do("SISMEMBER", set, link.Dest.URL))
	return count == 1, err
}

func (f Follower) setAsFollowed(link LinkLog) (already bool, err error) {
	set := "follower::followedset::" + link.Crawl.Name + "::" + link.Crawl.Serial
	count, err := redis.Int(f.Redis.Do("SADD", set, link.Dest.URL))
	return count == 0, err
}

func (f *Follower) getAMQPChannel() (*amqp.Channel, error) {
	if f.amqpChannel != nil {
		return f.amqpChannel, nil
	}
	c, err := f.AMQP.Channel()
	if err != nil {
		return nil, err
	}
	f.amqpChannel = c

	return f.amqpChannel, nil
}

func (f *Follower) checkAMQPQueue() (err error) {
	channel, err := f.getAMQPChannel()
	if err != nil {
		return
	}
	_, err = channel.QueueDeclarePassive("crawl.commands.fetch", true, false, false, false, nil)
	return
}

func (f *Follower) sendFetchCommand(link LinkLog) (err error) {
	channel, err := f.getAMQPChannel()
	if err != nil {
		return
	}

	var cmd fetcher.FetcherCommand
	cmd.CrawlName = link.Crawl.Name
	cmd.CrawlSerial = link.Crawl.Serial
	cmd.CrawlGeneration = link.Crawl.Generation + 1
	cmd.URL = link.Dest.URL

	encd, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Transient,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         encd,
	}

	err = channel.Publish("", "crawl.commands.fetch", false, false, msg)
	if err != nil {
		panic(err)
	}

	return
}
