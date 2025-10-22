package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Client *mongo.Client

func Connect() *mongo.Client {
	// Load mongodb url from config
	cfg := config.LoadConfig()
	MongoDb := cfg.MongoURI

	if MongoDb == "" {
		log.Fatal("MONGODB_URI not set!")
	}

	fmt.Println("MongoDB URI: ",MongoDb)

	// Create Context with Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(MongoDb)
	client, err := mongo.Connect(clientOptions); if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping to verify connection
	err = client.Ping(ctx, nil); if err != nil {
		log.Fatal("Failed to Ping MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB!")

	Client = client
	return client
}

func Disconnect() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := Client.Disconnect(ctx); if err != nil {
		return err
	}

	fmt.Println("Disconnected from MongoDB!")
	return nil
}

func OpenCollection(collectionName string) *mongo.Collection {
	cfg := config.LoadConfig()
	dbName := cfg.DatabaseName

	if Client == nil {
		log.Fatal("MongoDB client is not initialized. Call Connect() first!")
	}

	collection := Client.Database(dbName).Collection(collectionName)
	return collection
}