package rethinkgo

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type rssEntry struct {
	XMLName   xml.Name `xml:"entry"`
	Summary   string   `xml:"summary"`
	Published string   `xml:"published"`
	Id        int      `xml:"id"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"feed"`
	Entries []rssEntry `xml:"entry"`
}

type Thought struct {
	Content string
	Date    time.Time
}

var baseUrl = "https://rethink.uwu.network"
var client = http.Client{}

func init() {
	if url, ok := os.LookupEnv("RETHINK_API"); ok {
		baseUrl = url
	}
}

func getRssFeed(user string) (*rssFeed, error) {
	url := fmt.Sprintf("%s/~%s/feed.xml", baseUrl, user)
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	return &feed, nil
}

func GetThoughts(user string) ([]Thought, error) {
	feed, err := getRssFeed(user)
	if err != nil {
		return nil, err
	}

	thoughts := make([]Thought, len(feed.Entries))
	for i := range thoughts {
		entry := feed.Entries[i]

		date, err := time.Parse(time.RFC3339, entry.Published)
		if err != nil {
			return nil, err
		}

		thoughts[i] = Thought{
			// Id:      entry.Id,
			Content: entry.Summary,
			Date:    date,
		}
	}

	return thoughts, nil
}

func PutThought(content string, name string, key string) error {
	url := fmt.Sprintf("%s/api/think", baseUrl)
	body := bytes.NewReader([]byte(content))

	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("name", name)
	req.Header.Add("authorization", key)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return errors.New(res.Status)
	}

	return nil
}
