package scraper

import (
	"context"
	"frag-aggra/internal/database"
	"frag-aggra/internal/models"
	"log"
	"os"
	"regexp"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var wtsRe = regexp.MustCompile(`(?i)\[wts\]`)

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

func (r *RedditScraper) FetchPost(subreddit string, repo database.Repository, limit int) ([]models.Post, error) {

	if limit <= 0 {
		limit = 5
	}

	posts, _, err := r.client.Subreddit.NewPosts(context.Background(), subreddit, &reddit.ListOptions{
		Limit: limit,
	})

	if err != nil {
		return nil, err
	}

	log.Print("Grabbing ", limit, " posts")
	var job_postings []models.Post
	for _, post := range posts {

		// only include posts that contain [WTS] (case-insensitive) in title or body
		if !containsWTS(post.Title) && !containsWTS(post.Body) {
			log.Printf("Skipping post %s without [WTS] in title or body", post.ID)
			continue
		}

		//grab post_id first
		exists, err := repo.PostExists(context.Background(), post.ID)
		if err != nil {
			//log error but continue processing other posts
			log.Printf("Error checking post existence for ID %s: %v", post.ID, err)
			continue
		}
		if exists {
			log.Printf("Skipping already seen post with ID %s", post.ID)
			continue
		}

		job_posting := models.Post{
			PostID:         post.ID,
			URL:            post.URL,
			Title:          post.Title,
			Body:           post.Body,
			SellerUsername: post.Author,
		}

		job_postings = append(job_postings, job_posting)
	}

	return job_postings, nil

}

func containsWTS(s string) bool {
	return wtsRe.MatchString(s)
}
