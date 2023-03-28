package service

import (
	"context"
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
	baseRegistry
	client        *mongo.Client
	database      *mongo.Database
	collection    *mongo.Collection
	ctx           context.Context
	updateTimeout time.Duration
}

func NewMongoServiceRegistry(ctx context.Context, client *mongo.Client, database string, collection string) (ServiceRegistry, error) {

	r := &mongoServiceRegistry{
		client:   client,
		database: client.Database(database),
		ctx:      ctx,
	}

	r.collection = r.database.Collection(collection)

	err := r.loadServices()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *mongoServiceRegistry) AddService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.addService(service, func() error {
		_, err := r.collection.InsertOne(r.ctx, service)
		return err
	})
}

func (r *mongoServiceRegistry) UpdateService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.updateService(service, func() error {
		_, err := r.collection.UpdateOne(r.ctx, bson.M{"name": service.Name}, bson.M{"$set": service})
		return err

	})
}

func (r *mongoServiceRegistry) RemoveService(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.removeService(name, func() error {
		_, err := r.collection.DeleteOne(r.ctx, bson.M{"name": name})
		return err
	})
}

func (r *mongoServiceRegistry) GetServices() []*BackendService {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.getServices()
}

func (r *mongoServiceRegistry) loadServices() error {
	var services []*BackendService

	cursor, err := r.collection.Find(r.ctx, bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(r.ctx)

	for cursor.Next(r.ctx) {
		var service BackendService
		err = cursor.Decode(&service)
		if err != nil {
			return err
		}

		service.Init()
		r.services = append(services, &service)
	}

	if err = cursor.Err(); err != nil {
		return err
	}

	return nil
}
