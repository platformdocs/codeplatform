package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codeplatform/mongodb/config"
	"codeplatform/mongodb/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB *Database
)

type Database struct {
	Mongo *mongo.Client
}

// 初始化
func Init() {
	DB = &Database{
		Mongo: SetConnect(),
	}
}

// 连接设置
func SetConnect() *mongo.Client {

	var retryWrites bool = false

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().SetHosts(config.GetConf().MongoConf.Hosts).
		SetMaxPoolSize(config.GetConf().MongoConf.MaxPoolSize).
		SetHeartbeatInterval(constants.HEART_BEAT_INTERVAL).
		SetConnectTimeout(constants.CONNECT_TIMEOUT).
		SetMaxConnIdleTime(constants.MAX_CONNIDLE_TIME).
		SetRetryWrites(retryWrites)

	username := config.GetConf().MongoConf.Username
	password := config.GetConf().MongoConf.Password

	if len(username) > 0 && len(password) > 0 {
		clientOptions.SetAuth(options.Credential{Username: username, Password: password})
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client
}

func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	DB.Mongo.Disconnect(ctx)
}

func connectDB() {
	opts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	coll := client.Database("codeplatform").Collection("codeList")
	content := "a"
	var result bson.M
	err = coll.FindOne(context.TODO(), bson.D{{"content", content}}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the content %s\n", content)
		return
	}
	if err != nil {
		panic(err)
	}
	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", jsonData)
}
