package elasticsearch

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evoila/scrape-elasticsearch/config"
	elastic "github.com/olivere/elastic"
)

var elasticClient *elastic.Client

// Tweet is sample Object
type Tweet struct {
	User     string    `json:"user"`
	Message  string    `json:"message"`
	Retweets int       `json:"retweets"`
	Image    string    `json:"image,omitempty"`
	Created  time.Time `json:"created,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Location string    `json:"location,omitempty"`
}

// NewElasticsearchClient creates a client instance of the elastic remote
// client
func NewElasticsearchClient(config *config.Config) *elastic.Client {
	var connectionURL = buildConnectionURL(config.Elasticsearch.Hosts, config.Elasticsearch.Port, config.Elasticsearch.HTTPS)
	elasticClient, err := elastic.NewClient(elastic.SetURL(connectionURL))
	if err != nil {
		// Handle error
		panic(err)
	}

	return elasticClient
}

func buildConnectionURL(ips []string, port int, https bool) string {
	var connectionURL []string
	for _, ip := range ips {
		if https {
			connectionURL = append(connectionURL, "https://"+ip+":"+strconv.Itoa(port))
		} else {
			connectionURL = append(connectionURL, "http://"+ip+":"+strconv.Itoa(port))
		}
	}
	return strings.Join(connectionURL, ",")
}

// WriteMetric writes the value of a Metric to Elasticsearch
func WriteMetric(ctx context.Context, elasticClient *elastic.Client) {
	tweet1 := Tweet{User: "olivere", Message: "Take Five", Retweets: 0}
	put1, err := elasticClient.Index().
		Index("test").
		Type("test").
		BodyJson(tweet1).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
}
