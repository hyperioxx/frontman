package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type mongoServiceRegistry struct {
	client        *mongo.Client
	database      *mongo.Database
	collection    *mongo.Collection
	services      map[string]*BackendService
	ctx           context.Context
	updateTimeout time.Duration
}

func NewMongoServiceRegistry(ctx context.Context, client *mongo.Client, database string, collection string) (ServiceRegistry, error) {
	db := client.Database(database)
	col := db.Collection(collection)
	return &mongoServiceRegistry{
		client:        client,
		database:      db,
		collection:    col,
		ctx: ctx,
	}, nil
}

func (r *mongoServiceRegistry) AddService(service *BackendService) error {
	_, err := r.collection.InsertOne(r.ctx, service)
	return err
}

func (r *mongoServiceRegistry) UpdateService(service *BackendService) error {
	_, err := r.collection.UpdateOne(r.ctx, bson.M{"name": service.Name}, bson.M{"$set": service})
	return err
}

func (r *mongoServiceRegistry) RemoveService(name string) error {
	_, err := r.collection.DeleteOne(r.ctx, bson.M{"name": name})
	return err
}

func (r *mongoServiceRegistry) GetServices() []*BackendService {
	var services []*BackendService
	
	cursor, err := r.collection.Find(r.ctx, bson.M{})
	if err != nil {
		fmt.Printf("error finding services: %v\n", err)
		return nil
	}
	defer cursor.Close(r.ctx)
	for cursor.Next(r.ctx) {
		var service BackendService
		err := cursor.Decode(&service)
		if err != nil {
			log.Printf("error decoding service: %v\n", err)
			return nil
		}
		services = append(services, &service)
	}
	if err := cursor.Err(); err != nil {
		log.Printf("error iterating over cursor: %v\n", err)
		return nil
	}
	log.Printf("found %d services\n", len(services))
	return services
}

