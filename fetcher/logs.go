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
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type FetchLogType int

const (
	SuccessfulNewContent FetchLogType = iota
	SuccessfulRepeatedContent
	SuccessfulNoContent
	Redirect
	Error
)

type FetchLog struct {
	ID      string `bson:"_id,omitempty"`
	Type    FetchLogType
	URL     string
	ncanURL string
	Crawl   struct {
		Name       string
		Serial     string
		Generation int32
	}
	Req struct {
		Header map[string][]string
	}
	Res struct {
		Header     map[string][]string
		StatusCode int
	}
	ContentSHA256  string
	SourceRedirect string
	On             time.Time
}

func NewFetchLog(cmd FetcherCommand, req *http.Request, res *http.Response, sha256 string) (log FetchLog) {
	log.URL = cmd.URL
	log.Crawl.Name = cmd.CrawlName
	log.Crawl.Serial = cmd.CrawlSerial
	log.Crawl.Generation = cmd.CrawlGeneration
	log.Req.Header = req.Header
	log.Res.Header = res.Header
	log.Res.StatusCode = res.StatusCode
	log.ContentSHA256 = sha256

	return
}

type FetchLogsRepository interface {
	Log(*FetchLog) error
}

type DumpLogs struct {
}

func (logs DumpLogs) Log(log *FetchLog) error {
	spew.Dump(log)

	return nil
}

type MgoLogs struct {
	Collection *mgo.Collection
}

func (logs *MgoLogs) Log(log *FetchLog) error {
	log.ID = bson.NewObjectId().Hex()
	log.On = time.Now()
	return logs.Collection.Insert(log)
}
