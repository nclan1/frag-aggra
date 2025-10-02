package scraper

import (
	"os"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type RedditScraper struct {
	client *reddit.Client
}

func New() (*RedditScraper, error) {

	credentials := reddit.Credentials{
		ID:       os.Getenv("REDDIT_CLIENT_ID"),
		Secret:   os.Getenv("REDDIT_CLIENT_SECRET"),
		Username: os.Getenv("REDDIT_USERNAME"),
		Password: os.Getenv("REDDIT_PASSWORD"),
	}

	client, err := reddit.NewClient(credentials)
	if err != nil {
		return nil, err
	}

	return &RedditScraper{
		client: client,
	}, nil
}

func (r *RedditScraper) FetchPost(subreddit string) {

}
