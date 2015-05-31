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

package robotstxt

import (
	"net/http"
	"sync"

	"github.com/op/go-logging"
	temoto "github.com/temoto/robotstxt-go"
)

var logger = logging.MustGetLogger("robotstxt")

var RobotsTXTDB DB

type DB struct {
	dbs     map[string]*temoto.RobotsData
	running bool
	once    sync.Once
	in      chan request
}

type request struct {
	useragent string
	domain    string
	path      string
	result    chan bool
}

func Lookup(useragent, domain, path string) (out chan bool) {
	return RobotsTXTDB.Lookup(useragent, domain, path)
}

func (db *DB) Lookup(useragent, domain, path string) (out chan bool) {
	db.Start()
	out = make(chan bool)

	req := request{
		useragent: useragent,
		domain:    domain,
		path:      path,
		result:    out,
	}

	db.in <- req

	return
}

func (db *DB) Start() {
	db.once.Do(db.start)
}

func (db *DB) Close() {
	close(db.in)
}

func (db *DB) start() {
	db.in = make(chan request, 10)

	handler := func() {
		for req := range db.in {
			data, ok := db.dbs[req.domain]
			if !ok {
				data = db.loadDomain(req.domain)
			}

			resp := data.TestAgent(req.path, req.useragent)
			req.result <- resp
		}
	}

	go handler()
}

func (db *DB) loadDomain(domain string) *temoto.RobotsData {
	var data *temoto.RobotsData
	resp, err := http.Get("http://" + domain + "/robots.txt")
	if err != nil {
		data, _ = temoto.FromStatusAndBytes(400, []byte{})
		logger.Error("%s", err)
		panic(err)
		return data
	}
	defer resp.Body.Close()

	data, err = temoto.FromResponse(resp)
	if err != nil {
		panic(err)
	}

	return data
}
