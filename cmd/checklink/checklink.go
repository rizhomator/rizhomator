package main

import (
	"flag"
	"fmt"
	"github.com/rizhomator/rizhomator/crawl"

	"github.com/spf13/viper"
)

var crawlName string

func main() {
	config, err := crawl.LoadConfig("crawlers.yml")
	if err != nil {
		panic(err)
	}

	if flag.NArg() < 1 {
		panic("should give url")
	}

	rawurl := flag.Arg(0)

	fmt.Printf("Rawurl: %s\n", rawurl)

	if config.ShouldFetch(crawlName, rawurl) {
		fmt.Println("URL will be fetched")
	} else {
		fmt.Println("URL will not be fetched")
	}

}

func init() {
	flag.StringVar(&crawlName, "crawl", "idealista", "Crawl name (idealista)")
	flag.Parse()
	viper.AutomaticEnv()
}
