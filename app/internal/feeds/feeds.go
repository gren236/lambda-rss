package feeds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Feeds struct {
	Table string
	db    *dynamodb.Client
}

type Feed struct {
	Url           string     `dynamodbav:"url"`
	LastPublished *time.Time `dynamodbav:"last_published,unixtime"`
}

func New(ctx context.Context, table string) *Feeds {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return &Feeds{
		Table: table,
		db:    dynamodb.NewFromConfig(cfg),
	}
}

func (f *Feeds) Retrieve(ctx context.Context) []Feed {
	res, err := f.db.Scan(ctx, &dynamodb.ScanInput{TableName: &f.Table})
	if err != nil {
		log.Fatalf("failed to scan table, %v", err)
	}

	var feeds []Feed
	if err := attributevalue.UnmarshalListOfMaps(res.Items, &feeds); err != nil {
		log.Fatalf("failed to unmarshal QueryItems, %v", err)
	}

	return feeds
}

func (f *Feeds) Update(ctx context.Context, feed Feed) error {
	av, err := attributevalue.MarshalMap(feed)
	if err != nil {
		log.Fatalf("failed to marshal feed, %v", err)
	}

	_, err = f.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &f.Table,
		Item:      av,
	})
	if err != nil {
		log.Fatalf("failed to put item, %v", err)
	}

	return nil
}
