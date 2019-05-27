package mongodb

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/evoila/scrape-elasticsearch/config"
	"github.com/evoila/scrape-elasticsearch/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client
var database string
var collection string

// NewMongoDBClient creates a new Client instance for
// MongoDB to read the Services Instances
func NewMongoDBClient(ctx context.Context, config *config.Config) *mongo.Client {
	var connectionURL = buildConnectionURL(config.MongoDB.Username, config.MongoDB.Password,
		config.MongoDB.Hosts, config.MongoDB.Port)
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(connectionURL))
	database = config.MongoDB.Database
	collection = config.MongoDB.Collection
	if err != nil {
		// Handle error
		panic(err)
	}

	err = mongoClient.Connect(ctx)
	if err != nil {
		panic(err)
	}

	return mongoClient
}

func buildConnectionURL(username string, password string, ips []string, port int) string {
	var connectionURL = "mongodb://" + username + ":" + password + "@"
	for _, ip := range ips {
		connectionURL += ip + ":" + strconv.Itoa(port)
	}
	return connectionURL
}

// LoadServiceInstances retrieves all active Service Instances
// from the database to create the Service Workers
func LoadServiceInstances(ctx context.Context, mongoClient *mongo.Client, scrapeType string) []model.ServiceInstance {
	var serviceInstances []model.ServiceInstance
	cur, err := mongoClient.Database(database).Collection(collection).Find(ctx, bson.D{{"hosts.type", scrapeType}})
	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		serviceInstance := model.ServiceInstance{}
		err := cur.Decode(&serviceInstance)
		if err != nil {
			log.Fatal(err)
		}
		serviceInstance.Hosts = removeAddresses(serviceInstance.Hosts, scrapeType)
		serviceInstances = append(serviceInstances, serviceInstance)
	}
	fmt.Printf("Service Instances %v", serviceInstances)
	return serviceInstances
}

func removeAddresses(hosts []model.ServerAddress, scrapeType string) []model.ServerAddress {
	for i, host := range hosts {
		if host.Type != scrapeType {
			hosts = append(hosts[:i], hosts[i+1:]...)
		}
	}
	return hosts
}
