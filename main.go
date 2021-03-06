// RSS feed reader that outputs plain text, werc/apps/barf
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SlyMarbo/rss"
)

type Article struct {
	Title	string
	Link	string
	Date	time.Time
	Content	string
	Tags	[]string
}

type renderer func(articles []Article)
type filter func(article *Article)

var (
	debug	= flag.Bool("d", false, "print debug msgs to stderr")
	format	= flag.String("f", "", "output format")
	root    = flag.String("r", "", "output root")
	dest	string
	links 	string
)
var filters = map[string]filter {
	"http://feeds.feedburner.com/Explosm": filterexplosm,
}


func usage() {
	os.Stderr.WriteString("usage: rrss [-d] [-f barf|blagh] [-r root] <feed file>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func fetchfeed(url string) (resp *http.Response, err error) {
	client := http.DefaultClient
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; hjdicks)")
	return client.Do(req)
}

func isold(date time.Time, link string, path string) bool {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0775)
	if err != nil {
		return true
	}
	defer file.Close()
	s := fmt.Sprintf("%d_%s", date.Unix(), link);
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(s, scanner.Text()) {
			return true
		}
	}
	return false
}

func makeold(date time.Time, link string, path string) (int, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	defer f.Close()
	check(err)
	if link == "" {
		link = "empty"
	}
	s := fmt.Sprintf("%d_%s", date.Unix(), link);
	return f.WriteString(s + "\n")
}

func writef(dir, filename, content string) {
	err := ioutil.WriteFile(path.Join(dir, filename), []byte(content+"\n"), 0775)
	check(err)
}

func ensuredir(dir string) {
	err := os.MkdirAll(dir, 0775)
	check(err)
}

// http://code.9front.org/hg/barf
func barf(articles []Article) {
	dest = path.Join(*root, "src")
	ensuredir(dest)
	n := lastarticle(dest)
	for _, a := range articles {
		n = n + 1
		d := fmt.Sprintf("%s/%d", dest, n)
		ensuredir(d)
		writef(d, "title", a.Title)
		writef(d, "link", a.Link)
		writef(d, "date", a.Date.String())
		writef(d, "body", a.Content)
		if a.Tags != nil {
			ensuredir(path.Join(d, "tags"))
			for _, j := range a.Tags {
				f, err := os.Create(d + "/tags/" + j)
				f.Close()
				check(err)
			}
		}
		_, err := makeold(a.Date, a.Link, links)
		check(err)
	}
}

// http://werc.cat-v.org/apps/blagh
func blagh(articles []Article) {
	var err error
	for _, a := range articles  {
		dest = path.Join(*root, fmt.Sprintf("%d/%02d/%02d", a.Date.Year(), a.Date.Month(), a.Date.Day()));
		ensuredir(dest)
		f, _ := os.Open(dest) // directory will usually not exist yet
		defer f.Close()
		n, _ := f.Readdirnames(0)
		d := fmt.Sprintf("%s/%d", dest, len(n))
		ensuredir(d)
		writef(d, "index.md", fmt.Sprintf("%s\n===\n\n%s\n", a.Title, a.Content))
		_, err = makeold(a.Date, a.Link, links)
		check(err)
	}
}

func stdout(articles []Article) {
	for _, a := range articles {
		fmt.Printf("title: %s\nlink: %s\ndate: %s\n%s\n\n",
			a.Title, a.Link, a.Date, a.Content)
	}
}

func lastarticle(dir string) int {
	f, err := os.Open(dir)
	defer f.Close()
	check(err)
	dn, err := f.Readdirnames(0)
	check(err)
	var di []int
	for _, j := range dn {
		k, _ := strconv.Atoi(j)
		di = append(di, k)
	}
	sort.Ints(di)
	n := 0
	if di != nil {
		n = di[len(di)-1]
	}
	return n
}

func filterexplosm(article *Article) {
	r, err := http.Get(article.Link)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer r.Body.Close()
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "main-comic") {
			s := strings.Replace(scanner.Text(), "src=\"", "src=\"http:", 1)
			article.Content = s
			break;
		}
	}
}

func loadfeed(url string, tags []string) []Article {
	var articles []Article
	if *debug {
		log.Printf("Fetching feed '%s'", url)
	}
	feed, err := rss.FetchByFunc(fetchfeed, url)
	if err != nil {
		log.Printf("Cannot load feed '%s': %v", url, err)
		return nil
	}
	for _, i := range feed.Items {
		if isold(i.Date, i.Link, links) {
			continue
		}
		a := Article{i.Title, i.Link, i.Date, conorsum(i), tags}
		if f, ok := filters[url]; ok {
			f(&a)
		}
		articles = append(articles, a)
	}
	if *debug {
		log.Printf("Loaded %d items", len(articles))
	}
	return articles
}

func conorsum(i *rss.Item) string {
	s := ""
	switch{
	case len(i.Content) > 0:
		s = i.Content
	case len(i.Summary) > 0:
		s = i.Summary
	}
	return html.UnescapeString(s)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.Arg(0) == "" {
		usage()
	}
	var render renderer
	switch *format {
	case "barf":
		render = barf
	case "blagh":
		render = blagh
	case "":
		render = stdout
	default:
		usage()
	}		
	links = path.Join(*root, "links")
	file, err := os.Open(flag.Arg(0))
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var articles []Article
	var tags []string
	for scanner.Scan() {
		t := scanner.Text()
		if len(t) <= 0 {
			continue
		}
		l := strings.Split(t, " ")
		if len(l) > 1 {
			tags = l[1:]
		}
		a := loadfeed(l[0], tags)
		if a != nil {
			articles = append(articles, a...)
		}
	}
	sort.Slice(articles, func(i, j int) bool { return articles[i].Date.Before(articles[j].Date) })
	render(articles)
}
