package database

import (
	"context"
	"encoding/json"
	"fmt"
	"golaco/model"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const uri = "mongodb://localhost:27017"
const databaseName = "api"

func collection(name string, jsonSchema func() bson.M) *mongo.Collection {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("\033[31m", fmt.Sprintf("%s %d\t%s", time.Now().Format("2006-01-02 15:04:05"), 500, err))
		return nil
	}
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		fmt.Println("\033[31m", fmt.Sprintf("%s %d\t%s", time.Now().Format("2006-01-02 15:04:05"), 500, err))
		return nil
	}
	db := client.Database(databaseName)
	collections, err := db.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println("\033[31m", fmt.Sprintf("%s %d\t%s", time.Now().Format("2006-01-02 15:04:05"), 500, err))
		return nil
	}
	for _, v := range collections {
		if v == name {
			return db.Collection(name)
		}
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema(),
	}
	opts := options.CreateCollection().SetValidator(validator)

	err = db.CreateCollection(context.TODO(), name, opts)
	if err != nil {
		fmt.Println("\033[31m", fmt.Sprintf("%s %d\t%s", time.Now().Format("2006-01-02 15:04:05"), 500, err))
		return nil
	}
	return db.Collection(name)
}

func Create(inter interface{}, finder string) model.Callback {
	mod := reflect.ValueOf(inter)
	db := collection(strings.ToLower(mod.Type().Elem().Name()), mod.MethodByName("Schema").Interface().(func() bson.M))
	err := db.FindOne(context.TODO(), bson.D{{Key: strings.ToLower(finder), Value: mod.Elem().FieldByName(finder).Interface()}}).Decode(mod.Interface())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			mod.Elem().FieldByName("ID").Set(reflect.Zero(mod.Elem().FieldByName("ID").Type()))
			post, err := db.InsertOne(context.TODO(), mod.Interface())
			if err != nil {
				return model.Callback{
					Code:   400,
					Result: nil,
					Err:    err,
				}
			}
			json, err := json.Marshal(post.InsertedID)
			if err != nil {
				return model.Callback{
					Code:   500,
					Result: nil,
					Err:    err,
				}
			}
			return model.Callback{
				Code:   201,
				Result: json,
				Err:    nil,
			}
		}
		return model.Callback{
			Code:   400,
			Result: nil,
			Err:    err,
		}
	}
	return model.Callback{
		Code:   409,
		Result: nil,
		Err:    nil,
	}
}

func Delete(inter interface{}) model.Callback {
	mod := reflect.ValueOf(inter)
	db := collection(strings.ToLower(mod.Type().Elem().Name()), mod.MethodByName("Schema").Interface().(func() bson.M))
	err := db.FindOneAndDelete(context.TODO(), bson.D{{Key: "_id", Value: mod.Elem().FieldByName("ID").Interface()}}).Decode(mod.Interface())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Callback{
				Code:   404,
				Result: nil,
				Err:    nil,
			}
		}
		return model.Callback{
			Code:   400,
			Result: nil,
			Err:    err,
		}
	}
	return model.Callback{
		Code:   204,
		Result: nil,
		Err:    nil,
	}
}

func Read(inter interface{}, finder string) model.Callback {
	mod := reflect.ValueOf(inter)
	db := collection(strings.ToLower(mod.Type().Elem().Name()), mod.MethodByName("Schema").Interface().(func() bson.M))
	if mod.Elem().FieldByName(finder).IsZero() {
		cursor, err := db.Find(context.TODO(), bson.D{})
		if err != nil {
			return model.Callback{
				Code:   400,
				Result: nil,
				Err:    err,
			}
		}
		var mods []interface{}
		for cursor.Next(context.TODO()) {
			err = cursor.Decode(mod.Interface())
			if err != nil {
				return model.Callback{
					Code:   400,
					Result: nil,
					Err:    err,
				}
			}
			mods = append(mods, mod.Elem().Interface())
		}
		json, err := json.Marshal(mods)
		if err != nil {
			return model.Callback{
				Code:   500,
				Result: nil,
				Err:    err,
			}
		}
		return model.Callback{
			Code:   200,
			Result: json,
			Err:    nil,
		}
	}
	err := db.FindOne(context.TODO(), bson.D{{Key: strings.ToLower(finder), Value: mod.Elem().FieldByName(finder).Interface()}}).Decode(mod.Interface())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Callback{
				Code:   404,
				Result: nil,
				Err:    nil,
			}
		}
		return model.Callback{
			Code:   400,
			Result: nil,
			Err:    err,
		}
	}
	json, err := json.Marshal(mod.Elem().Interface())
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	return model.Callback{
		Code:   200,
		Result: json,
		Err:    nil,
	}
}

func Update(inter interface{}) model.Callback {
	mod := reflect.ValueOf(inter)
	db := collection(strings.ToLower(mod.Type().Elem().Name()), mod.MethodByName("Schema").Interface().(func() bson.M))
	err := db.FindOneAndUpdate(context.TODO(), bson.D{{Key: "_id", Value: mod.Elem().FieldByName("ID").Interface()}}, bson.D{{Key: "$set", Value: mod.Elem().Interface()}}).Decode(mod.Interface())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Callback{
				Code:   404,
				Result: nil,
				Err:    nil,
			}
		}
		return model.Callback{
			Code:   400,
			Result: nil,
			Err:    err,
		}
	}
	return model.Callback{
		Code:   204,
		Result: nil,
		Err:    nil,
	}
}
