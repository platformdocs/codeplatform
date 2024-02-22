package mongodb

type CodeData struct {
	Id      string `bson:"_id,omitempty" json:"id"`
	Name    string `bson:"name" json:"name"`
	Content string `bson:"content" json:"content"`
	Time    string `bson:"time" json:"time"`
}
