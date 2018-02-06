package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Set random seed to pick a page and sgf file.
	rand.Seed(time.Now().UTC().UnixNano())

	var goKifuURL = "http://gokifu.com/?p="

	// Intn argument mutating goKifuURL into a specific page with game
	// records is a amount of gokifu pages as of 16/01/2018
	goKifuURL += string(rand.Intn(1992))

	linksToSGF, err := fetchAndParsePage(goKifuURL)
	if err != nil {
		log.Fatalf("An error occured in crawlForSGF func:\n%s\n", err)
	}

	// Pick and fetch an sgf file.
	selector := rand.Intn(len(linksToSGF))
	file, err := http.Get(linksToSGF[selector])
	if err != nil {
		log.Fatal(nil)
	}

	// Read from http client.
	defer file.Body.Close()
	kifu, err := ioutil.ReadAll(file.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare filename.
	var filename string

	// A local scope is created here, because there is no need to
	// pollute global namespace with one unnecessary variable. And i
	// like that initializing a local scope is this easy.
	{
		playerInfo := extractPlayerData(kifu)
		playerInfo = cleanUp(playerInfo)
		filename = playerInfo + ".sgf"
	}

	// Write a file
	ioutil.WriteFile(filename, kifu, 0644)

}

// fetchAndParsePage performs an HTTP GET request for url, parses the
// response as HTML and passes parsed HTML tree to helper function traverse,
// which goes through parse tree returned by html.Parse and feeds a slice
// with links to sgf files.
func fetchAndParsePage(url string) ([]string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	}

	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}

	return traverse(nil, doc), nil
}

// Traverse appends to links each SGF file link found in a node, and returns
// a slice with links to games to choose from.
func traverse(links []string, node *html.Node) []string {
	if node.Type == html.ElementNode && node.Data == "a" {
		// node.Attr is a slice, hence discard index, a first return
		// value.
		for _, a := range node.Attr {
			// The "if" below weeds out links that don't lead to
			// game records. Declaring isSGF condition inline
			// like that is may be bad style. I don't know, it
			// seems a little dense to me. But i've just read
			// about declaration inside forms and since this is
			// all to learn go anyway, i wanted to use this
			// feature.
			if isSGF := strings.HasSuffix(a.Val, ".sgf"); a.Key == "href" && isSGF {
				links = append(links, a.Val)
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		links = traverse(links, c)
	}

	return links
}

// Extracts a "header" with player information from an sgf.
func extractPlayerData(b []byte) (matched string) {

	// Matches a byte slice containing metadata about a downloaded game
	// record. Have look at sgf spec or any sgf file to understand it
	// better why this regexp looks this way and not the other.
	// NOTE: There probably is a better way. I suck at regexp.
	re := regexp.MustCompile(`PB\[.*\]KM`)
	b = re.Find(b)

	matched = string(b)
	matched = matched[:len(matched)-2]

	return
}

// Cleans up a string that is supposed to be spat from a function above -
// feels hacky, but hey, it works!
func cleanUp(s string) (clean string) {
	for _, rune := range s {
		if rune == ']' {
			clean += "-"
		} else {
			clean += string(rune)
		}
	}
	re := regexp.MustCompile(`([A-Z]{2})\[`)
	clean = re.ReplaceAllString(clean, "")
	clean = clean[:len(clean)-1]

	return
}
