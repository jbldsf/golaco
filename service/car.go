package service

import (
	"encoding/json"
	"golaco/database"
	"golaco/model"
	"io"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Car struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Model        string             `bson:"model,omitempty" json:"model"`
	Manufacturer string             `bson:"manufacturer,omitempty" json:"manufacturer"`
	Year         int                `bson:"year,omitempty" json:"year"`
	Maintenance  map[int]float64    `bson:"maintenance,omitempty" json:"maintenance"`
	Tank         int                `bson:"tank,omitempty" json:"tank"`
	Consumption  map[string]float64 `bson:"consumption,omitempty" json:"consumption"`
}

func (car Car) Schema() bson.M {
	return bson.M{}
}

func (car Car) delete(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&car)
	if err != nil {
		if err != io.EOF {
			return model.Callback{
				Code:   500,
				Result: nil,
				Err:    err,
			}
		}
	}
	return database.Delete(&car)
}

func (car Car) get(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&car)
	if err != nil {
		if err != io.EOF {
			return model.Callback{
				Code:   500,
				Result: nil,
				Err:    err,
			}
		}
	}
	return database.Read(&car, "Model")
}

func (car Car) post(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&car)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	return database.Create(&car, "Model")
}

func (car Car) put(data model.Data) model.Callback {
	err := json.NewDecoder(data.Body).Decode(&car)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	return database.Update(&car)
}

func Cars(data model.Data) model.Callback {
	var car Car
	function := map[string]func(model.Data) model.Callback{
		"DELETE": car.delete,
		"GET":    car.get,
		"POST":   car.post,
		"PUT":    car.put,
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
