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

import "github.com/rizhomator/rizhomator/fetcher"

type LinkLogType int32

const (
	Anchor LinkLogType = 1
	Script LinkLogType = 2
	Link   LinkLogType = 3
)

type LinkLog struct {
	Type  LinkLogType
	Rel   string
	Crawl struct {
		Name       string
		Serial     string
		Generation int32
	}
	Src struct {
		ID  string
		URL string
	}
	Dest struct {
		URL     string
		nparURL string
	}
}

func NewLinkLog(log fetcher.FetchLog) (link LinkLog) {
	link.Crawl.Name = log.Crawl.Name
	link.Crawl.Serial = log.Crawl.Serial
	link.Crawl.Generation = log.Crawl.Generation
	link.Src.ID = log.ID
	link.Src.URL = log.URL

	return
}
