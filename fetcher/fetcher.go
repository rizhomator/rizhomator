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
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/rizhomator/rizhomator/crawl"
	"github.com/rizhomator/rizhomator/useragent"

	"github.com/op/go-logging"
	"golang.org/x/net/proxy"
)

var logger = logging.MustGetLogger("fetcher")

type FetcherCommand struct {
	URL             string
	CrawlName       string
	CrawlSerial     string
	CrawlGeneration int32
	Done            chan error `json:"-"`
}

type Fetcher interface {
	Fetch(FetcherCommand) error
}

type StdFetcher struct {
	SOCKS     string
	logs      FetchLogsRepository
	storage   ConstantObjectStorage
	notifier  Notifier
	uas       useragent.Ring
	client    *http.Client
	transport *http.Transport
	config    crawl.Config
}

func NewStdFetcher(logs FetchLogsRepository, storage ConstantObjectStorage, notifier Notifier) (f *StdFetcher) {
	f = new(StdFetcher)
	f.logs = logs
	f.storage = storage
	f.notifier = notifier
	f.uas.LoadFromFile("browseruseragents.json")

	var err error
	f.config, err = crawl.LoadConfig("crawlers.yml")
	if err != nil {
		panic(err)
	}

	return
}

func (f *StdFetcher) Fetch(cmd FetcherCommand) (err error) {
	if len(cmd.URL) == 0 {
		logger.Debug("Fethcer: skiping command with empty URL")
		cmd.Done <- nil
		return nil
	}

	if cmd.CrawlGeneration > f.config[cmd.CrawlName].MaxGeneration {
		logger.Debug("Fetcher discarding fetch command as over MaxGeneration")
		cmd.Done <- nil
		return nil
	}

	if !f.config.ShouldFetch(cmd.CrawlName, cmd.URL) {
		logger.Notice("Fetcher: skiping command due to crawler configuration, %s", cmd.URL)
		cmd.Done <- nil
		return nil
	}

	url, err := url.Parse(cmd.URL)
	if err != nil {
		logger.Error("Fetcher: Error parsing url (%s), %v", cmd.URL, err)
		cmd.Done <- err
		return err
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		logger.Error("Fetcher: Error creating request (%s), %v", cmd.URL, err)
		cmd.Done <- err
		return err
	}
	ua := f.getUserAgent()
	logger.Debug("Using UA: %s", ua)
	req.Header.Set("User-Agent", ua)

	var res *http.Response

	res, err = f.getClient().Do(req)
	if err != nil {
		logger.Error("Fetcher: Error performing request (%s), %v", cmd.URL, err)
		logger.Warning("Fetcher: retrying")
		res, err = f.retryOnce(req)
		if err != nil {
			logger.Error("Fetcher: Error retying to perform request (%s), %v", cmd.URL, err)
			cmd.Done <- err
			return err
		}
	}

	var hash [sha256.Size]byte
	var content []byte

	log := NewFetchLog(cmd, req, res, "")

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		content, _ = ioutil.ReadAll(res.Body)
		res.Body.Close()
		hash = sha256.Sum256(content)
		key := hex.EncodeToString(hash[:])
		log.ContentSHA256 = key
		err = f.storage.Store(string(key[:]), content)
		if err != nil {
			logger.Error("Fetcher: Error storing response body (%s), %v", cmd.URL, err)
			cmd.Done <- err
			return err
		}
	}

	err = f.logs.Log(&log)
	if err != nil {
		logger.Error("Fetcher: Error logging fetch (%s), %v", cmd.URL, err)
		cmd.Done <- err
		return err
	}

	err = f.notifier.Notify(log)
	cmd.Done <- nil

	return err
}

func (f *StdFetcher) getUserAgent() string {
	return f.uas.Get().Identification
}

func (f *StdFetcher) getClient() (client *http.Client) {
	if f.client != nil {
		return f.client
	}
	if f.SOCKS == "" {
		client = &http.Client{}
		return
	}

	logger.Debug("Fetcher: User socks server: %s", f.SOCKS)

	tbProxyURL, err := url.Parse("socks5://" + f.SOCKS)
	if err != nil {
		panic(err)
	}
	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		panic(err)
	}
	f.transport = &http.Transport{Dial: tbDialer.Dial}
	client = &http.Client{Transport: f.transport}

	f.client = client
	return
}

func (f *StdFetcher) Reset() {
	if f.transport != nil {
		f.transport.CloseIdleConnections()
	}
	f.client = nil
	f.transport = nil
}

func (f *StdFetcher) Close() {
	f.Reset()
}

func (f *StdFetcher) retryOnce(req *http.Request) (res *http.Response, err error) {
	f.Reset()
	time.Sleep(time.Second)
	res, err = f.getClient().Do(req)
	return
}
