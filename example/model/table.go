package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type File struct {
	Id bson.ObjectId `json:"_id" bson:"_id"`
	UserId int64 `json:"user_id"`
	Filename string `json:"filename"`
	Md5 string `json:"md5"`
	Url string `json:"url"`
}


type APIError struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt time.Time
}

type APISuccess struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt time.Time
}
