package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	googleURL     = "https://www.google.fr/search?q="
	subTextTag    = "<span class=\"st\">"
	subTextEndTag = "</span>"
)

func main() {
	flag.Usage = func() {
		// http://patorjk.com/software/taag/#p=display&f=Big
		fmt.Fprintf(os.Stderr, ""+
			`googleit [OPTIONS]

---------------------------------------------
   _____                   _      _____ _   
  / ____|                 | |    |_   _| |  
 | |  __  ___   ___   ____| | ___  | | | |_ 
 | | |_ |/ _ \ / _ \ / _  | |/ _ \ | | | __|
 | |__| | (_) | (_) | (_| | |  __/_| |_| |_ 
  \_____|\___/ \___/ \__, |_|\___|_____|\__|
                      __/ |                 
                     |___/                  
---------------------------------------------

Usage:

  googleit -q "what's google?"

starts "googleit.exe"
 - for the given query and prints the first result description.

Options:
`)
		flag.PrintDefaults()
	}
	verbose := flag.Bool("v", false, "verbose mode (flags value)")
	query := flag.String("q", "", "string query to make in the google search engine")
	timeout := flag.Duration("t", 10*time.Second, "http Get request timeout")
	flag.Parse()
	if len(*query) == 0 {
		log.Fatalf("type something to search")
	}
	if *verbose {
		log.Println("query (-q)", *query)
		log.Println("timeout (-t)", *timeout)
	}

	// create client with timeout
	client := &http.Client{
		Timeout: *timeout,
	}

	// make the query request
	resp, err := client.Get(googleURL + url.QueryEscape(*query))
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer resp.Body.Close()

	// read response
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// get first response
	split := strings.Split(string(bytes), subTextTag)
	if len(split) < 2 {
		log.Fatalf("could not find html token in response")
	}
	split = strings.Split(split[1], subTextEndTag)
	if len(split) < 2 {
		log.Fatalf("could not find html end token in response")
	}

	fmt.Println(regexp.MustCompile("</?[A-Za-z]+>").ReplaceAllString(html.UnescapeString(split[0]), ""))
}
