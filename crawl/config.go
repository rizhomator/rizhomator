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

package crawl

import (
	"io/ioutil"
	"net/url"
	"regexp"

	"github.com/rizhomator/rizhomator/robotstxt"

	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

var logger = logging.MustGetLogger("genlinks")

type Config map[string]CrawlConfig

var regexpcache = make(map[string]*regexp.Regexp)

type CrawlConfig struct {
	Seeds         []string
	Domains       []string
	PreURLExcl    []string
	URLExcl       []string
	URLIncl       []string
	MaxGeneration int32
}

func LoadConfig(filename string) (config Config, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &config)

	return
}

func (c Config) ShouldFetch(crawlName, rawurl string) bool {
	url, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		logger.Debug("should not fetch because incorrect schema")
		return false
	}

	// TODO Check Domains in Crawl config
	if !c.MatchDomain(crawlName, url.Host) {
		logger.Debug("should not fetch because host does not match")
		return false
	}

	if c.MatchPreURLExcl(crawlName, rawurl) {
		logger.Debug("should not fetch because pre excluded url matched")
		return false
	}

	if c.MatchURLIncl(crawlName, rawurl) {
		logger.Debug("matched url included pattern")
	} else {
		if c.MatchURLExcl(crawlName, rawurl) {
			logger.Debug("should not fetch because excluded url matched")
			return false
		}
	}

	robotsallowed := <-robotstxt.Lookup("myuseragent", url.Host, url.Path)
	if !robotsallowed {
		logger.Debug("should not fetch because ROBOTSTXT")
		return false
	}

	return true
}

func (c Config) MatchDomain(crawlName, domain string) bool {
	config, ok := c[crawlName]
	if !ok {
		panic("crawlName " + crawlName + "not found")
		return false
	}

	for _, pattern := range config.Domains {
		if regexpMatchString(pattern, domain) {
			return true
		}
	}

	return false
}

func (c Config) MatchPreURLExcl(crawlName, url string) bool {
	config, ok := c[crawlName]
	if !ok {
		return false
	}

	for _, pattern := range config.PreURLExcl {
		if regexpMatchString(pattern, url) {
			return true
		}
	}

	return false
}

func (c Config) MatchURLIncl(crawlName, url string) bool {
	config, ok := c[crawlName]
	if !ok {
		return false
	}

	for _, pattern := range config.URLIncl {
		if regexpMatchString(pattern, url) {
			return true
		}
	}

	return false
}

func (c Config) MatchURLExcl(crawlName, url string) bool {
	config, ok := c[crawlName]
	if !ok {
		return false
	}

	for _, pattern := range config.URLExcl {
		if regexpMatchString(pattern, url) {
			return true
		}
	}

	return false
}

func regexpMatchString(pattern, value string) bool {
	var reg *regexp.Regexp
	var err error
	var ok bool

	reg, ok = regexpcache[pattern]
	if !ok {
		reg, err = regexp.Compile(pattern)
		if err != nil {
			panic(err)
		}
		regexpcache[pattern] = reg
	}

	return reg.MatchString(value)
}
