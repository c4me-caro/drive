package database

import (
	"context"
	"fmt"

	"github.com/c4me-caro/drive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DriveWorker struct {
	client     *mongo.Client
	db         string
	collection string
}
type Document struct {
	ID        string           `bson:"_id"`
	Users     []drive.User     `bson:"users"`
	Resources []drive.Resource `bson:"resources"`
}

func NewDriveWorker(c *mongo.Client, db string, collection string) *DriveWorker {
	return &DriveWorker{
		client:     c,
		db:         db,
		collection: collection,
	}
}

func (cfw *DriveWorker) RemoveFileResource(parent string) error {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	filter := bson.M{"resources.content": bson.M{"$exists": true}}
	update := bson.M{"$pull": bson.M{"resources.$.content": parent}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("error updating resource: %v", err)
	}

	return nil
}

func (cfw *DriveWorker) UpdateFolderContent(res drive.Resource, newId string) error {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	filter := bson.M{"resources": res}
	update := bson.M{"$push": bson.M{"resources.$.content": newId}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("error updating resource: %v", err)
	}

	return nil
}

func (cfw *DriveWorker) DeleteResource(res drive.Resource) error {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	filter := bson.M{"resources": bson.M{"$exists": true}}
	update := bson.M{"$pull": bson.M{"resources": res}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("error removing resource: %v", err)
	}

	return nil
}

func (cfw *DriveWorker) CreateResource(res drive.Resource) error {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	filter := bson.M{"resources": bson.M{"$exists": true}}
	update := bson.M{"$push": bson.M{"resources": res}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("error adding resource: %v", err)
	}

	return nil
}

func (cfw *DriveWorker) GetUser(username string, password string) (string, error) {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return "", err
	}

	for cursor.Next(context.TODO()) {
		var doc Document
		err := cursor.Decode(&doc)
		if err != nil {
			return "", err
		}

		for _, user := range doc.Users {
			if user.Name == username && user.Password == password {
				return user.Id, nil
			}
		}
	}

	return "", fmt.Errorf("user not found")
}

func (cfw *DriveWorker) FindUser(id string) (drive.User, error) {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return drive.User{}, err
	}

	for cursor.Next(context.TODO()) {
		var doc Document
		err := cursor.Decode(&doc)
		if err != nil {
			return drive.User{}, err
		}

		for _, user := range doc.Users {
			if user.Id == id {
				return user, nil
			}
		}
	}

	return drive.User{}, fmt.Errorf("user not found")
}

func (cfw *DriveWorker) GetResource(resName string) (drive.Resource, error) {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return drive.Resource{}, err
	}

	for cursor.Next(context.TODO()) {
		var doc Document
		err := cursor.Decode(&doc)
		if err != nil {
			return drive.Resource{}, err
		}

		for _, res := range doc.Resources {
			if res.Name == resName {
				return res, nil
			}
		}
	}

	return drive.Resource{}, fmt.Errorf("resource not found")
}

func (cfw *DriveWorker) GetResourceId(id string) (drive.Resource, error) {
	coll := cfw.client.Database(cfw.db).Collection(cfw.collection)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return drive.Resource{}, err
	}

	for cursor.Next(context.TODO()) {
		var doc Document
		err := cursor.Decode(&doc)
		if err != nil {
			return drive.Resource{}, err
		}

		for _, res := range doc.Resources {
			if res.Id == id {
				return res, nil
			}
		}
	}

	return drive.Resource{}, fmt.Errorf("resource not found")
}

func (cfw *DriveWorker) Start() error {
	// Verificar la conexión
	err := cfw.client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	return nil
}

func ConnectDB(url string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))
	if err != nil {
		panic(err)
	}

	return client, nil
}
