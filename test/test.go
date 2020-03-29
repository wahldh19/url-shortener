package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dwrz/url-shortener/internal/config"
	"github.com/dwrz/url-shortener/internal/db"
	"github.com/dwrz/url-shortener/internal/shorturl"
	"github.com/dwrz/url-shortener/internal/visit"
)

var (
	cfg        = config.New()
	serviceURL = fmt.Sprintf("http://localhost:%s", cfg.Port)
)

type shortURL struct {
	Short string
	Long  string
}

func main() {
	log.Println("starting test")

	// Check status endpoint.
	if err := checkStatus(); err != nil {
		log.Printf("ERROR: failed to check service status: %v", err)
		log.Fatal("failed status check")
	}
	log.Println("checked status endpoint")

	// Test short URL creation.
	var url = "http://dwrz.net"

	short, err := createShortURL(url)
	if err != nil {
		log.Printf("ERROR: failed to create short URL: %v", err)
		log.Fatal("failed to create short URL")
	}
	log.Println("OK: created single short URL")

	// Test duplicate has unique short URL.
	shortDuplicate, err := createShortURL(url)
	if err != nil {
		log.Printf("ERROR: failed to create short URL: %v", err)
		log.Fatal("failed to create short URL")
	}
	if short.Short == shortDuplicate.Short {
		log.Printf("ERROR: duplicate has identical short URL")
		log.Fatal("failed to create unique short URL")
	}
	log.Println("OK: created unique short URL for duplicate long URL")

	// Test a redirect.
	if err := checkRedirect(short); err != nil {
		log.Printf("ERROR: failed redirect: %v", err)
		log.Fatal("failed redirect")
	}
	log.Println("OK: redirect for single short URL")

	// Test stats.
	if err := checkStats(short); err != nil {
		log.Printf("ERROR: failed get stats: %v", err)
		log.Fatal("failed to get stats")
	}
	log.Println("OK: stats for single short URL")

	// Test invalid URLs.
	if _, err := createShortURL("this isn't a URL"); err == nil {
		log.Printf("ERROR: created invalid URL without error")
		log.Fatal("failed URL validation")
	}
	if _, err := createShortURL("ftp://example.com"); err == nil {
		log.Printf("ERROR: created invalid URL without error")
		log.Fatal("failed URL validation")
	}
	if _, err := createShortURL("http://"); err == nil {
		log.Printf("ERROR: created invalid URL without error")
		log.Fatal("failed URL validation")
	}
	log.Println("OK: url validation")

	// Test invalid redirect.
	invalidShort := shortURL{Short: "test", Long: "test"}
	if err := checkRedirect(invalidShort); err == nil {
		log.Printf("ERROR: redirected invalid short URL without error")
		log.Fatal("failed invalid redirection")
	}
	log.Println("OK: invalid redirect")

	// Test concurrent short URL creation.
	shortURLs, err := createShortURLs()
	if err != nil {
		log.Printf(
			"ERROR: failed to concurrently create short URLs: %v",
			err,
		)
		log.Fatal("failed to concurrently create short URLs")
	}
	if len(shortURLs) == 0 {
		log.Fatal("failed to concurrently create short URLs")
	}
	log.Printf("OK: concurrently created %d short URLs", len(shortURLs))

	// Test concurrent redirects.
	if err := checkRedirects(shortURLs); err != nil {
		log.Printf("ERROR: failed concurrent redirects: %v", err)
		log.Fatal("failed concurrent redirects")
	}
	log.Printf("OK: concurrently redirected %d short URLs", len(shortURLs))

	// Test concurrent stats.
	if err := checkAllStats(shortURLs); err != nil {
		log.Printf("ERROR: failed concurrent get stats: %v", err)
		log.Fatal("failed concurrent get stats")
	}
	log.Printf(
		"OK: concurrently got stats for %d short URLs", len(shortURLs),
	)

	if err := testVisitStats(); err != nil {
		log.Printf("ERROR: failed to get correct stats: %v", err)
		log.Fatal("failed to get correct stats")
	}
	log.Println("OK: got correct visit stats across timespans")

	log.Println("successfully completed tests")
}

func checkStatus() error {
	res, err := http.Get(serviceURL)
	if err != nil {
		return fmt.Errorf("failed to get status: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"expected status code of %d, but got %d",
			http.StatusOK, res.StatusCode,
		)
	}

	return nil
}

func createShortURLs() (urls []shortURL, err error) {
	var longURLs = []string{
		"https://blog.golang.org/",
		"https://derrickjensen.org/",
		"https://e360.yale.edu/",
		"https://ifconfig.io/",
		"https://www.bloomberg.com/green",
		"https://www.gnu.org/",
		"https://www.marines.mil/",
		"https://www.monbiot.com/",
		"https://www.openstreetmap.org/",
		"https://www.wikipedia.org/",
	}

	var (
		errs      = make(chan error, len(longURLs))
		shortURLs = make(chan shortURL, len(longURLs))
		wg        sync.WaitGroup
	)
	wg.Add(len(longURLs))

	for _, url := range longURLs {
		go func(url string) {
			defer wg.Done()

			shortURL, err := createShortURL(url)
			if err != nil {
				errs <- err
				return
			}
			shortURLs <- shortURL
		}(url)
	}

	wg.Wait()
	close(errs)
	close(shortURLs)

	if len(errs) > 0 {
		for err := range errs {
			log.Printf("ERROR: failed to create short url: %v", err)
		}
		return nil, fmt.Errorf("failed to create all URLs successfully")
	}

	for s := range shortURLs {
		urls = append(urls, s)
	}

	return urls, nil
}

func createShortURL(longURL string) (shortURL, error) {
	res, err := http.PostForm(
		serviceURL, url.Values{"url": {longURL}},
	)
	if err != nil {
		return shortURL{}, fmt.Errorf(
			"failed to post long url: %v", err,
		)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return shortURL{}, fmt.Errorf(
			"expected status code of %d, but got %d",
			http.StatusOK, res.StatusCode,
		)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return shortURL{}, fmt.Errorf(
			"failed to read response body: %v", err,
		)
	}

	return shortURL{Short: string(body), Long: longURL}, nil
}

func checkRedirects(shortURLs []shortURL) error {
	var (
		errs = make(chan error, len(shortURLs))
		wg   sync.WaitGroup
	)
	wg.Add(len(shortURLs))

	for _, short := range shortURLs {
		go func(short shortURL) {
			defer wg.Done()

			if err := checkRedirect(short); err != nil {
				errs <- err
			}
		}(short)
	}

	wg.Wait()
	close(errs)

	if len(errs) > 0 {
		for err := range errs {
			log.Printf(
				"ERROR: failed to redirect short URL: %v", err,
			)
		}
		return fmt.Errorf("failed to redirect all URLs successfully")
	}

	return nil
}

func ignoreRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func checkRedirect(shortURL shortURL) error {
	client := &http.Client{CheckRedirect: ignoreRedirect}

	url := fmt.Sprintf("%s/%s", serviceURL, shortURL.Short)

	log.Printf("requesting %v for %v", shortURL.Short, shortURL.Long)

	now := time.Now()
	res, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get with short url: %v", err)
	}
	defer res.Body.Close()

	elapsed := time.Since(now)

	if res.StatusCode != http.StatusMovedPermanently {
		return fmt.Errorf(
			"expected status code of %v, but got %v",
			http.StatusMovedPermanently, res.StatusCode,
		)
	}

	log.Printf(
		"%s: took ~%v to receive redirect response",
		shortURL.Short, elapsed,
	)
	if elapsed > 10*time.Millisecond {
		return fmt.Errorf("failed to redirect within 10ms")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), shortURL.Long) {
		return fmt.Errorf("redirect does not match expected long url")
	}

	return nil
}

func checkAllStats(shortURLs []shortURL) error {
	var (
		errs = make(chan error, len(shortURLs))
		wg   sync.WaitGroup
	)
	wg.Add(len(shortURLs))

	for _, short := range shortURLs {
		go func(short shortURL) {
			defer wg.Done()

			if err := checkStats(short); err != nil {
				errs <- err
			}
		}(short)
	}

	wg.Wait()
	close(errs)

	if len(errs) > 0 {
		for err := range errs {
			log.Printf(
				"ERROR: failed to check stats: %v", err,
			)
		}
		return fmt.Errorf("failed to check all stats successfully")
	}

	return nil
}

func checkStats(shortURL shortURL) error {
	url := fmt.Sprintf("%s/%s/stats", serviceURL, shortURL.Short)

	res, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("failed to get stats: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var stats visit.Stats
	if err := json.Unmarshal(body, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if stats.Day != 1 {
		return fmt.Errorf(
			"expected 1 visit in last day, but got %d: %v",
			stats.Day, err,
		)
	}
	if stats.Week != 1 {
		return fmt.Errorf(
			"expected 1 visit in last week, but got %d: %v",
			stats.Week, err,
		)
	}
	if stats.Year != 1 {
		return fmt.Errorf(
			"expected 1 visit in last year, but got %d: %v",
			stats.Year, err,
		)
	}

	return nil
}

func testVisitStats() error {
	// Create a new short URL with no visits.
	short, err := createShortURL("https://go-proverbs.github.io/")
	if err != nil {
		return fmt.Errorf("failed to create short URL")
	}

	db, err := db.Connect(context.TODO(), cfg.MongoURI)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}

	shortURL, err := shorturl.Get(context.TODO(), shorturl.GetParams{
		DB:          db,
		Environment: cfg.Environment,
		Short:       short.Short,
	})
	if err != nil {
		return fmt.Errorf("failed to get short URL")
	}

	// Check stats.
	stats, err := visit.GetStats(context.TODO(), visit.GetStatsParams{
		DB:          db,
		Environment: cfg.Environment,
		ShortID:     shortURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get stats")
	}
	if stats.Day != 0 || stats.Week != 0 || stats.Year != 0 {
		return fmt.Errorf("non-zero visits on new short url")
	}

	// Insert some visits to test timespans.
	coll := db.Database(cfg.Environment).Collection(visit.Collection)

	// Insert a visit from more than one year ago.
	if _, err := coll.InsertOne(context.TODO(), visit.Visit{
		ShortID: shortURL.ID,
		Time:    time.Now().AddDate(-1, 0, -1),
	}); err != nil {
		return fmt.Errorf("failed to insert access record: %v", err)
	}
	stats, err = visit.GetStats(context.TODO(), visit.GetStatsParams{
		DB:          db,
		Environment: cfg.Environment,
		ShortID:     shortURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get stats")
	}
	if stats.Day != 0 || stats.Week != 0 || stats.Year != 0 {
		return fmt.Errorf(
			"got visits for visit older than 1 year",
		)
	}

	// Insert a visit from six months ago.
	if _, err := coll.InsertOne(context.TODO(), visit.Visit{
		ShortID: shortURL.ID,
		Time:    time.Now().AddDate(0, -6, 0),
	}); err != nil {
		return fmt.Errorf("failed to insert access record: %v", err)
	}
	stats, err = visit.GetStats(context.TODO(), visit.GetStatsParams{
		DB:          db,
		Environment: cfg.Environment,
		ShortID:     shortURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get stats")
	}
	if stats.Day != 0 {
		return fmt.Errorf(
			"expected 0 visits for past day, but got %d", stats.Day,
		)
	}
	if stats.Week != 0 {
		return fmt.Errorf(
			"expected 0 visits for past week, but got %d",
			stats.Week,
		)
	}
	if stats.Year != 1 {
		return fmt.Errorf(
			"expected 1 visit for past year, but got %d",
			stats.Year,
		)
	}

	// Insert a visit from 8 days ago.
	if _, err := coll.InsertOne(context.TODO(), visit.Visit{
		ShortID: shortURL.ID,
		Time:    time.Now().AddDate(0, 0, -8),
	}); err != nil {
		return fmt.Errorf("failed to insert access record: %v", err)
	}
	stats, err = visit.GetStats(context.TODO(), visit.GetStatsParams{
		DB:          db,
		Environment: cfg.Environment,
		ShortID:     shortURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get stats")
	}
	if stats.Day != 0 {
		return fmt.Errorf(
			"expected 0 visits for past day, but got %d", stats.Day,
		)
	}
	if stats.Week != 0 {
		return fmt.Errorf(
			"expected 0 visits for past week, but got %d",
			stats.Week,
		)
	}
	if stats.Year != 2 {
		return fmt.Errorf(
			"expected 2 visits for past year, but got %d",
			stats.Year,
		)
	}

	// Insert a visit from 3 days ago.
	if _, err := coll.InsertOne(context.TODO(), visit.Visit{
		ShortID: shortURL.ID,
		Time:    time.Now().AddDate(0, 0, -3),
	}); err != nil {
		return fmt.Errorf("failed to insert access record: %v", err)
	}
	stats, err = visit.GetStats(context.TODO(), visit.GetStatsParams{
		DB:          db,
		Environment: cfg.Environment,
		ShortID:     shortURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get stats")
	}
	if stats.Day != 0 {
		return fmt.Errorf(
			"expected 0 visits for past day, but got %d", stats.Day,
		)
	}
	if stats.Week != 1 {
		return fmt.Errorf(
			"expected 1 visit for past week, but got %d",
			stats.Week,
		)
	}
	if stats.Year != 3 {
		return fmt.Errorf(
			"expected 3 visits for past year, but got %d",
			stats.Year,
		)
	}

	// Insert a visit from today.
	if _, err := coll.InsertOne(context.TODO(), visit.Visit{
		ShortID: shortURL.ID,
		Time:    time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to insert access record: %v", err)
	}
	stats, err = visit.GetStats(context.TODO(), visit.GetStatsParams{
		DB:          db,
		Environment: cfg.Environment,
		ShortID:     shortURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get stats")
	}
	if stats.Day != 1 {
		return fmt.Errorf(
			"expected 1 visits for past day, but got %d", stats.Day,
		)
	}
	if stats.Week != 2 {
		return fmt.Errorf(
			"expected 2 visits for past week, but got %d",
			stats.Week,
		)
	}
	if stats.Year != 4 {
		return fmt.Errorf(
			"expected 4 visits for past year, but got %d",
			stats.Year,
		)
	}

	// TODO: test time boundaries.

	return nil
}
