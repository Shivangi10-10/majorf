package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client                *mongo.Client
	usersCollection       *mongo.Collection
	connectionsCollection *mongo.Collection
)

func init() {
	// This will be executed when the package is imported
	ctx := context.Background()
	var err error

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		// This is just a fallback in case .env isn't loaded yet
		return
	}

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return
	}

	// Initialize collections
	db := client.Database("referral_db")
	usersCollection = db.Collection("users")
	connectionsCollection = db.Collection("connections")
}

// RegisterUser registers a user with role and company
func RegisterUser(username, role, company string) error {
	if client == nil {
		return errors.New("MongoDB client not initialized")
	}

	ctx := context.Background()

	// Format role if it's a comma-separated string
	var formattedRole string
	if strings.Contains(role, ",") {
		formattedRole = role
	} else {
		formattedRole = role
	}

	// Update user or insert if not exists
	filter := bson.M{"name": username}
	update := bson.M{
		"$set": bson.M{
			"name":    username,
			"role":    formattedRole,
			"company": company,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := usersCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

// ConnectUsers creates a connection between two users
func ConnectUsers(user1, user2 string) error {
	if client == nil {
		return errors.New("MongoDB client not initialized")
	}

	ctx := context.Background()

	// Create connection
	filter := bson.M{"user1": user1, "user2": user2}
	update := bson.M{
		"$set": bson.M{
			"user1": user1,
			"user2": user2,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := connectionsCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetAllUsers retrieves all users from the database
func GetAllUsers() ([]map[string]interface{}, error) {
	if client == nil {
		return nil, errors.New("MongoDB client not initialized")
	}

	ctx := context.Background()

	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []map[string]interface{}
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	// Remove _id field from each user
	for i := range users {
		delete(users[i], "_id")
	}

	return users, nil
}

// GetUserDetails retrieves details for a specific user
func GetUserDetails(username string) (map[string]interface{}, error) {
	if client == nil {
		return nil, errors.New("MongoDB client not initialized")
	}

	ctx := context.Background()

	var result map[string]interface{}
	err := usersCollection.FindOne(ctx, bson.M{"name": username}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		return nil, err
	}

	// Remove _id field
	delete(result, "_id")
	return result, nil
}

// GetConnections retrieves all connections from the database
func GetConnections() ([]map[string]interface{}, error) {
	if client == nil {
		return nil, errors.New("MongoDB client not initialized")
	}

	ctx := context.Background()

	cursor, err := connectionsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []map[string]interface{}
	err = cursor.All(ctx, &connections)
	if err != nil {
		return nil, err
	}

	// Remove _id field from each connection
	for i := range connections {
		delete(connections[i], "_id")
	}

	return connections, nil
}
