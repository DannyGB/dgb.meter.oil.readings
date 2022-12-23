package database

import (
	"context"
	"dgb/meter.oil.readings/internal/configuration"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Repository struct {
	config configuration.Configuration
}

func (repository *Repository) GetTotalForYear(year int) []primitive.M {
	connect(repository.config)
	coll := repository.getCollection()
	filter := bson.M{
		"$and": []bson.M{
			{"date": bson.D{{
				Key: "$gte", Value: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			}}},
			{"date": bson.D{{
				Key: "$lte", Value: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
			}}},
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: 1}, {Key: "volume", Value: 1}}).
		SetProjection(bson.D{{Key: "volume", Value: 1}, {Key: "_id", Value: 0}})

	cursor, err := coll.Find(context.TODO(), filter, opts)

	if err == mongo.ErrNoDocuments {
		return nil
	}

	if err != nil {
		panic(err)
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func (repository *Repository) GetAll(pageParams PageParams) []primitive.M {

	connect(repository.config)
	coll := repository.getCollection()
	sortDir := pageParams.GetSortDirection()
	filter := pageParams.GetFilters()

	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: sortDir}, {Key: "volume", Value: sortDir}}).
		SetLimit(int64(pageParams.Take)).SetSkip(int64(pageParams.Skip))

	cursor, err := coll.Find(context.TODO(), filter, opts)

	if err == mongo.ErrNoDocuments {
		return nil
	}

	if err != nil {
		panic(err)
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func (repository *Repository) Count(pageParams PageParams) int64 {
	connect(repository.config)
	coll := repository.getCollection()
	filter := pageParams.GetFilters()

	count, err := coll.CountDocuments(context.TODO(), filter)

	if err != nil {
		panic(err)
	}

	return count
}

func (repository *Repository) GetSingle(id string) bson.M {

	connect(repository.config)
	coll := repository.getCollection()

	filter := bson.M{"_id": id}

	var result bson.M
	res := coll.FindOne(context.TODO(), filter)
	err := res.Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return result
}

func (repository *Repository) Insert(data bson.M) (id interface{}, err error) {

	connect(repository.config)
	coll := repository.getCollection()

	date, _ := time.Parse(time.RFC3339Nano, data["date"].(string))

	data["date"] = date

	result, err := coll.InsertOne(context.TODO(), data)

	if err != nil {
		return nil, errors.New("Could not insert document")
	}

	return result.InsertedID, nil
}

func (repository *Repository) Update(id interface{}, data bson.M) error {

	connect(repository.config)
	coll := repository.getCollection()

	date, _ := time.Parse(time.RFC3339Nano, data["date"].(string))
	data["date"] = date

	filter := bson.D{{Key: "_id", Value: id}}
	_, err := coll.ReplaceOne(context.TODO(), filter, data)

	if err != nil {
		return errors.New("Could not insert document")
	}

	return nil
}

func (repository *Repository) Delete(id interface{}) (deletedCount int, err error) {

	connect(repository.config)
	coll := repository.getCollection()

	filter := bson.D{{"_id", id}}
	result, err := coll.DeleteOne(context.TODO(), filter)

	if result.DeletedCount <= 0 || err != nil {
		return int(result.DeletedCount), errors.New("Could not delete")
	}

	return int(result.DeletedCount), nil
}

func connect(config configuration.Configuration) {

	if client != nil {
		return
	}

	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(config.MONGO_CONNECTION))

	if err != nil {
		panic(err)
	}
}

func (repository *Repository) getCollection() *mongo.Collection {
	return client.Database(repository.config.MONGO_DB).Collection(repository.config.MONGO_COLLECTION)
}

func NewRepository(cfg configuration.Configuration) *Repository {
	return &Repository{
		cfg,
	}
}
