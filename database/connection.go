package database

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
    client   *mongo.Client
    database *mongo.Database
    once     sync.Once
)

// Connect initializes the MongoDB connection
func Connect() {
    once.Do(func() {
        uri := os.Getenv("MONGO_URI")
        if uri == "" {
            log.Fatal("MONGO_URI environment variable is not set")
        }

        clientOptions := options.Client().ApplyURI(uri)
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        var err error
        client, err = mongo.Connect(ctx, clientOptions)
        if err != nil {
            log.Fatal(err)
        }

        // Ping the database
        err = client.Ping(ctx, nil)
        if err != nil {
            log.Fatal(err)
        }

        database = client.Database("nutri_meal")
        
        // Create unique index on email
        _, err = database.Collection("users").Indexes().CreateOne(
            context.Background(),
            mongo.IndexModel{
                Keys:    bson.D{{Key: "email", Value: 1}},
                Options: options.Index().SetUnique(true),
            },
        )
        if err != nil {
            log.Fatal("Error creating unique index on email:", err)
        }

        log.Println("Connected to MongoDB!")
    })
}

// GetCollection returns a collection from the database
func GetCollection(name string) *mongo.Collection {
    if database == nil {
        Connect()
    }
    return database.Collection(name)
}

// GetDatabase returns the database instance
func GetDatabase() *mongo.Database {
    if database == nil {
        Connect()
    }
    return database
}

// Close closes the database connection
func Close() {
    if client != nil {
        if err := client.Disconnect(context.Background()); err != nil {
            log.Printf("Error disconnecting from database: %v", err)
        }
    }
}
