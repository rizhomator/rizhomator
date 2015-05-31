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

package main

import (
	"os"

	"github.com/rizhomator/rizhomator/fetcher"

	"github.com/spf13/viper"
)

func main() {
	if len(os.Args) != 2 {
		panic("Should pass the crawl name")
	}

	crawlName := os.Args[1]

	fetcher.StartCrawl("crawlers.yml", crawlName)
}

func init() {
	viper.SetEnvPrefix("REI")
	viper.AutomaticEnv()
}
