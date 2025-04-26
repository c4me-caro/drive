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
	client *mongo.Client
	db     string
}

func NewDriveWorker(c *mongo.Client, db string) *DriveWorker {
	return &DriveWorker{
		client: c,
		db:     db,
	}
}

func (cfw *DriveWorker) AddResourceChildren(resource drive.Resource, children string) error {
	if resource.Id == "0" {
		return fmt.Errorf("system update forbiden")
	}

	coll := cfw.client.Database(cfw.db).Collection("resources")
	filter := bson.M{"id": resource.Id, "name": resource.Name}
	update := bson.M{
		"$push": bson.M{"resource.$.content": children},
	}

	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (cfw *DriveWorker) CheckResource(resource drive.Resource) error {
	if resource.Id == "0" {
		return fmt.Errorf("system check forbiden")
	}

	coll := cfw.client.Database(cfw.db).Collection("resources")
	filter := bson.M{"id": resource.Id, "name": resource.Name}

	_, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

func (cfw *DriveWorker) GetResource(search string) (drive.Resource, error) {
	coll := cfw.client.Database(cfw.db).Collection("resources")
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return drive.Resource{}, err
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var resource drive.Resource
		if err := cursor.Decode(&resource); err != nil {
			return drive.Resource{}, err
		}

		if resource.Name == search {
			return resource, nil
		}
	}

	return drive.Resource{}, fmt.Errorf("resource not found: %s", search)
}

func (cfw *DriveWorker) CreateResource(resource drive.Resource) error {
	coll := cfw.client.Database(cfw.db).Collection("resources")
	_, err := coll.InsertOne(context.TODO(), resource)
	if err != nil {
		return err
	}

	return nil
}

func (cfw *DriveWorker) DeleteResource(resource drive.Resource) error {
	if resource.Id == "0" {
		return fmt.Errorf("system deletion forbiden")
	}

	coll := cfw.client.Database(cfw.db).Collection("resources")
	filter := bson.M{"id": resource.Id, "name": resource.Name}

	_, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

func (cfw *DriveWorker) GetUserById(userid string) (drive.User, error) {
	coll := cfw.client.Database(cfw.db).Collection("users")
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return drive.User{}, err
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var user drive.User
		if err := cursor.Decode(&user); err != nil {
			return drive.User{}, err
		}

		if user.Id == userid {
			return user, nil
		}
	}

	return drive.User{}, fmt.Errorf("userid not found: %s", userid)
}

func (cfw *DriveWorker) GetUser(username string, password string) (drive.User, error) {
	coll := cfw.client.Database(cfw.db).Collection("users")
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return drive.User{}, err
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var user drive.User
		if err := cursor.Decode(&user); err != nil {
			return drive.User{}, err
		}

		if user.Name == username && user.Password == password {
			return user, nil
		}
	}

	return drive.User{}, fmt.Errorf("authuser not found: %s", username)
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
