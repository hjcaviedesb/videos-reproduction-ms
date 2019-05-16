package models

import "gopkg.in/mgo.v2/bson"

// Represents a movie, we uses bson keyword to tell the mgo driver how to name
// the properties in mongodb document
type Video struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Thumbnail   string        `bson:"cover_image" json:"cover_image"`
	Description string        `bson:"description" json:"description"`
	URL         string        `bson:"url" json:"url"`
}
