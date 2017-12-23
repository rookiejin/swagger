package model

import "gopkg.in/mgo.v2/bson"

// @def Pets
type Pets struct {
	Id bson.ObjectId `json:"id"`
	Tag []Tag `json:"tag" swag:"Tag"`
}

// @def Tag
type Tag struct {
	Id bson.ObjectId `json:"id"`
	Name string `json:"name"`
}

// @def Error
type Error struct {
	Code int `json:"code"`
	Message string `json:"message"`
}
