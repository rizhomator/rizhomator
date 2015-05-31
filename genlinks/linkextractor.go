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
	"bytes"
	"net/url"
	"strings"

	"github.com/rizhomator/rizhomator/fetcher"

	"github.com/PuerkitoBio/goquery"
	"github.com/op/go-logging"
	"golang.org/x/net/html"
)

var logger = logging.MustGetLogger("genlinks")

type LinkExtractor struct {
	Storage fetcher.ConstantObjectStorage
}

func (g *LinkExtractor) Extract(log fetcher.FetchLog) (links []LinkLog, err error) {

	links = []LinkLog{}

	// Consider content-type

	if log.Res.StatusCode != 200 {
		logger.Debug("Exiting link generation (%x) because status code != 200", log.ID)
		return
	}

	if log.ContentSHA256 == "" {
		logger.Debug("Exiting link generation (%x) because content is empty", log.ID)
		return
	}

	content, err := g.Storage.Retrieve(log.ContentSHA256)
	if err != nil {
		logger.Error("Error retrieving content from storage (fetch id: %x)", log.ID)
		return
	}

	rootNode, err := html.Parse(bytes.NewBuffer(content))
	if err != nil {
		logger.Error("Error parsing HTML (fetch id: %x)", log.ID)
		return
	}

	doc := goquery.NewDocumentFromNode(rootNode)
	doc.Url, err = url.Parse(log.URL)
	if err != nil {
		logger.Error("Error parsing URL (fetch id: %x)", log.ID)
		return
	}

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		link := NewLinkLog(log)
		link.Type = Anchor
		link.Rel, _ = s.Attr("rel")
		link.Dest.nparURL, _ = s.Attr("href")

		if len(link.Dest.nparURL) > 0 && !strings.HasPrefix(link.Dest.nparURL, "#") {
			parsed, urlerr := url.Parse(link.Dest.nparURL)
			if urlerr != nil {
				logger.Error("Error parsing link URL (fetch id: %x)", log.ID)
				return
			}
			link.Dest.URL = doc.Url.ResolveReference(parsed).String()
		}

		links = append(links, link)
	})

	doc.Find("script[src]").Each(func(_ int, s *goquery.Selection) {
		link := NewLinkLog(log)
		link.Type = Script
		link.Rel, _ = s.Attr("rel")
		link.Dest.nparURL, _ = s.Attr("src")

		if len(link.Dest.nparURL) > 0 && !strings.HasPrefix(link.Dest.nparURL, "#") {
			parsed, urlerr := url.Parse(link.Dest.nparURL)
			if urlerr != nil {
				logger.Error("Error parsing link URL (fetch id: %x)", log.ID)
				return
			}
			link.Dest.URL = doc.Url.ResolveReference(parsed).String()
		}

		links = append(links, link)
	})

	doc.Find("link[href]").Each(func(_ int, s *goquery.Selection) {
		link := NewLinkLog(log)
		link.Type = Link
		link.Rel, _ = s.Attr("rel")
		link.Dest.nparURL, _ = s.Attr("href")

		if len(link.Dest.nparURL) > 0 && !strings.HasPrefix(link.Dest.nparURL, "#") {
			parsed, urlerr := url.Parse(link.Dest.nparURL)
			if urlerr != nil {
				logger.Error("Error parsing link URL (fetch id: %x)", log.ID)
				return
			}
			link.Dest.URL = doc.Url.ResolveReference(parsed).String()
		}

		links = append(links, link)
	})

	return
}
