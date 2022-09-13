package service

import (
	"encoding/json"
	"fmt"
	"golaco/database"
	"golaco/model"
	"golaco/security"
	"io"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email,omitempty" json:"email"`
	Password string             `bson:"password,omitempty" json:"password"`
	Access   string             `bson:"access,omitempty" json:"access"`
	Car      Car                `bson:"car,omitempty" json:"car"`
}

func (user User) Schema() bson.M {
	return bson.M{
		"bsonType": "object",
		"properties": bson.M{
			"email": bson.M{
				"bsonType": "string",
				"pattern":  `[\w-\.]+@([\w-]+\.)+[\w-]{2,4}`,
			},
			"password": bson.M{
				"bsonType": "string",
				"pattern":  `(.|\s)*\S(.|\s)*`,
			},
			"access": bson.M{
				"bsonType": "string",
				"enum":     []string{"admnistrator", "user"},
			},
		},
	}
}

func (user User) delete(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&user)
	if err != nil {
		if err != io.EOF {
			return model.Callback{
				Code:   500,
				Result: nil,
				Err:    err,
			}
		}
	}
	return database.Delete(&user)
}

func (user User) get(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&user)
	if err != nil {
		if err != io.EOF {
			return model.Callback{
				Code:   500,
				Result: nil,
				Err:    err,
			}
		}
	}
	return database.Read(&user, "Email")
}

func (user User) post(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&user)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	password, err := regexp.MatchString(`(.|\s)*\S(.|\s)*`, user.Password)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	car, err := regexp.MatchString(`(.|\s)*\S(.|\s)*`, user.Car.Model)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	if car {
		callback := database.Read(&Car{Model: user.Car.Model}, "Model")
		if callback.Result != nil {
			err = json.Unmarshal(callback.Result, &user.Car)
			if err != nil {
				return model.Callback{
					Code:   500,
					Result: nil,
					Err:    err,
				}
			}
		} else {
			user.Car = Car{}
		}
	} else {
		user.Car = Car{}
	}
	if password && len(user.Password) > 5 && len(user.Password) < 13 {
		encrypt := security.Encrypt(user.Password)
		if encrypt.Code == 500 {
			return encrypt
		}
		user.Password = fmt.Sprintf("%x", encrypt.Result)
	}
	return database.Create(&user, "Email")
}

func (user User) put(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&user)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	encrypt := security.Encrypt(user.Password)
	if encrypt.Code == 500 {
		return encrypt
	}
	user.Password = fmt.Sprintf("%x", encrypt.Result)
	return database.Update(&user)
}

func Login(data model.Data) model.Callback {
	var user User
	err := json.NewDecoder(data.Body).Decode(&user)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	password := user.Password
	read := database.Read(&user, "Email")
	if read.Code == 200 {
		decrypt := security.Decrypt(user.Password)
		if decrypt.Code == 500 {
			return decrypt
		}
		if password == string(decrypt.Result) {
			return model.Callback{
				Code:   200,
				Result: nil,
				Err:    nil,
			}
		}
	}
	return model.Callback{
		Code:   401,
		Result: nil,
		Err:    nil,
	}
}

func Users(data model.Data) model.Callback {
	var user User
	function := map[string]func(model.Data) model.Callback{
		"DELETE": user.delete,
		"GET":    user.get,
		"POST":   user.post,
		"PUT":    user.put,
	}
	for k := range function {
		if k == data.Method {
			return function[k](data)
		}
	}
	return model.Callback{
		Code:   405,
		Result: nil,
		Err:    nil,
	}
}
