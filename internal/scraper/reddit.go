package scraper

import (
	"context"
	"fmt"
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

func (r *RedditScraper) FetchPost(subreddit string) error {

	posts, _, err := r.client.Subreddit.NewPosts(context.Background(), subreddit, &reddit.ListOptions{
		Limit: 5,
	})

	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: https://www.reddit.com%s\n\n", post.Permalink)
	}

	return nil

}
