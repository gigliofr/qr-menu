package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureMongoIndices crea gli indici necessari per performance multi-tenant
func EnsureMongoIndices(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.Database("qr-menu")

	// Indici per la collection menus
	menusCollection := db.Collection("menus")
	menuIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "restaurant_id", Value: 1}},
			Options: options.Index().SetName("idx_restaurant_id"),
		},
		{
			Keys: bson.D{
				{Key: "restaurant_id", Value: 1},
				{Key: "active", Value: 1},
			},
			Options: options.Index().SetName("idx_restaurant_active"),
		},
		{
			Keys: bson.D{
				{Key: "restaurant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_restaurant_created"),
		},
	}

	_, err := menusCollection.Indexes().CreateMany(ctx, menuIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create menu indices: %w", err)
	}

	// Indici per la collection restaurants
	restaurantsCollection := db.Collection("restaurants")
	restaurantIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_username_unique"),
		},
		{
			Keys: bson.D{{Key: "owner_id", Value: 1}},
			Options: options.Index().SetName("idx_owner_id"),
		},
		{
			Keys: bson.D{
				{Key: "owner_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_owner_created"),
		},
	}

	_, err = restaurantsCollection.Indexes().CreateMany(ctx, restaurantIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create restaurant indices: %w", err)
	}

	// Indici per la collection users
	usersCollection := db.Collection("users")
	userIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_email_unique"),
		},
		{
			Keys: bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_username_unique"),
		},
	}

	_, err = usersCollection.Indexes().CreateMany(ctx, userIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create user indices: %w", err)
	}

	fmt.Println("[DB] Mongo indices created successfully")
	return nil
}
