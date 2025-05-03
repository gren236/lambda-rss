package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	"github.com/sethvargo/go-envconfig"
	"lambda-rss/internal/feeds"
	"lambda-rss/internal/tg"
	"log"
	"slices"
	"sync"
)

type Config struct {
	Table          string `env:"DYNAMODB_TABLE"`
	TelegramToken  string `env:"TELEGRAM_TOKEN"`
	TelegramChatID int64  `env:"TELEGRAM_CHAT_ID"`
}

var (
	conf    Config
	dbFeeds *feeds.Feeds
)

func init() {
	ctx := context.Background()

	if err := envconfig.Process(ctx, &conf); err != nil {
		log.Fatal(err)
	}

	if conf.TelegramToken == "" || conf.TelegramChatID == 0 || conf.Table == "" {
		log.Fatal("missing required environment variables")
	}

	dbFeeds = feeds.New(ctx, conf.Table)
}

func handler(ctx context.Context) error {
	fds := dbFeeds.Retrieve(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(len(fds))

	bot, err := tgbotapi.NewBotAPI(conf.TelegramToken)
	if err != nil {
		log.Panic(err)
	}

	for _, feed := range fds {
		go func() {
			defer wg.Done()
			fp := gofeed.NewParser()
			feedParsed, err := fp.ParseURLWithContext(feed.Url, ctx)
			if err != nil {
				log.Printf("failed to parse feed %s, %v", feed.Url, err)
				return
			}

			if len(feedParsed.Items) == 0 {
				log.Printf("feed %s has no items", feedParsed.Title)
				return
			}

			if feed.LastPublished == nil || feed.LastPublished.IsZero() {
				feed.LastPublished = feedParsed.Items[0].PublishedParsed

				err := dbFeeds.Update(ctx, feed)
				if err != nil {
					log.Printf("failed to update feed %s, %v", feed.Url, err)
				}

				return
			}

			// Collect new posts
			var newPosts []gofeed.Item
			for _, item := range feedParsed.Items {
				if item.PublishedParsed == nil ||
					item.PublishedParsed.Equal(*feed.LastPublished) ||
					item.PublishedParsed.Before(*feed.LastPublished) {
					break
				}

				newPosts = append(newPosts, *item)
			}

			// Iterate over new posts in reverse because we want to send the latest posts last
			for _, item := range slices.Backward(newPosts) {
				post := tg.Post{
					Title:       item.Title,
					Author:      feedParsed.Title,
					Description: item.Description,
					Link:        item.Link,
				}

				htmlPost, err := post.ToHTMLPost()
				if err != nil {
					log.Printf("failed to parse post %s, %v", post.Title, err)
					continue
				}

				msg := tgbotapi.NewMessage(conf.TelegramChatID, htmlPost)
				msg.ParseMode = tgbotapi.ModeHTML
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("failed to send message %s, %v", msg.Text, err)
				}

				log.Printf("sent message with feed by %s", feedParsed.Title)

				feed.LastPublished = item.PublishedParsed

				err = dbFeeds.Update(ctx, feed)
				if err != nil {
					log.Printf("failed to update feed %s, %v", feed.Url, err)
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

func main() {
	lambda.Start(handler)
}
