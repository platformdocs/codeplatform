package main

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RequestData struct {
	Content    string `json:"content"`
	Id         string `json:"id"`
	Result     string `json:"result"`
	ResultType string `json:"resultType"`
}

type Database struct {
	Mongo *mongo.Client
}

type CodeData struct {
	Id         string `bson:"_id,omitempty" json:"id"`
	Name       string `bson:"name" json:"name"`
	Content    string `bson:"content" json:"content"`
	Time       string `bson:"time" json:"time"`
	Type       string `bson:"type" json:"type"`
	Result     string `bson:"result" json:"result"`
	ResultType string `bson:"resultType" json:"resultType"`
}

func Init() {
}

var (
	DB *Database
)

func connectDB() {
	opts := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	DB = &Database{
		Mongo: client,
	}
}

// insert the code to the database, two types: save and exec
func insertCode(ctx *gin.Context, codeType string, content string, result string, resultType string) (*mongo.InsertOneResult, error) {

	// connect to the database
	client := DB.Mongo
	collection := client.Database("codeplatform").Collection("codeList")

	results, _ := searchCode(ctx, codeType)

	var index int = len(results) + 1

	codedata := CodeData{}
	codedata.Id = uuid.New().String()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	codedata.Time = time.Now().In(loc).Format("2006-01-02 15:04:05")
	codedata.Content = content
	codedata.Result = result
	codedata.ResultType = resultType
	if codeType == "save" {
		codedata.Name = "main" + strconv.Itoa(index) + ".go"
	} else if codeType == "exec" {
		codedata.Name = "exec" + strconv.Itoa(index) + ".go"
	}

	codedata.Type = codeType
	insertOneResult, err := collection.InsertOne(ctx, &codedata)

	return insertOneResult, err
}

// search the code according to the code type
func searchCode(ctx *gin.Context, codeType string) ([]CodeData, error) {
	client := DB.Mongo
	collection := client.Database("codeplatform").Collection("codeList")
	filter := bson.M{}
	if codeType == "save" {
		filter = bson.M{"type": "save"}
	} else if codeType == "exec" {
		filter = bson.M{"type": "exec"}
	}
	cursor, err := collection.Find(ctx, filter)

	var results []CodeData
	err = cursor.All(ctx, &results)
	return results, err
}

// update the code user saved
func updateCode(ctx *gin.Context, id string, content string, result string, resultType string) (*mongo.UpdateResult, error) {
	client := DB.Mongo
	collection := client.Database("codeplatform").Collection("codeList")
	filter := bson.M{"_id": id}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	value := bson.M{"$set": bson.M{
		"Content":    content,
		"Result":     result,
		"ResultType": resultType,
		"Time":       time.Now().In(loc).Format("2006-01-02 15:04:05"),
	}}

	updateOneResult, err := collection.UpdateOne(ctx, filter, value)

	return updateOneResult, err
}

/*
* write code into a file
 */
func writeFile(filePath string, content string) bool {

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return false
	}

	return true
}

/*
* execute file and return the result
 */
func execFile(filePath string) (string, string) {
	// execute the golang file and output the result and error
	cmd := exec.Command("go", "run", filePath)
	doneChan := make(chan []byte, 1)
	errorChan := make(chan string, 1)
	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			var msg string = string(output) + err.Error()
			errorChan <- msg
			return
		}
		doneChan <- output
	}()

	select {
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		return "timeout", "execute code over 30s timeout"
	case output := <-doneChan:
		return "success", string(output)
	case err := <-errorChan:
		return "error", err
	}
}

func main() {

	router := gin.Default()

	// connect to database
	connectDB()

	// /app/dist is the path of frontend source files
	router.Use(static.Serve("/", static.LocalFile("./app/dist", true)))

	// create api router
	api := router.Group("/api")
	{
		// run code
		api.POST("/run", func(ctx *gin.Context) {
			var requestData RequestData
			if err := ctx.ShouldBindJSON(&requestData); err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			filePath := "execCode.go"
			// write code to the file
			writeResult := writeFile(filePath, requestData.Content)

			if !writeResult {
				ctx.JSON(500, gin.H{"error": "server write file error"})
				return
			}

			// run the code and get the result message
			var resultType, message = execFile(filePath)

			_, err := insertCode(ctx, "exec", requestData.Content, message, resultType)

			if err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}

			ctx.JSON(200, gin.H{
				"resultType": resultType,
				"message":    message,
			})

		})

		// search code List
		api.GET("/search", func(ctx *gin.Context) {

			results, _ := searchCode(ctx, "")

			ctx.JSON(200, results)
		})

		// save code
		api.POST("/save", func(ctx *gin.Context) {
			var requestData RequestData
			if err := ctx.ShouldBindJSON(&requestData); err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}

			_, err := insertCode(ctx, "save", requestData.Content, requestData.Result, requestData.ResultType)

			if err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(200, gin.H{
				"message": "code save successfully",
			})
		})

		// update code
		api.POST("/update", func(ctx *gin.Context) {
			var requestData RequestData
			if err := ctx.ShouldBindJSON(&requestData); err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}

			_, err := updateCode(ctx, requestData.Id, requestData.Content, requestData.Result, requestData.ResultType)

			if err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(200, gin.H{
				"message": "code update successfully",
			})
		})
	}

	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(404, gin.H{"msg": "not found"}) })

	// start server
	router.Run(":8080")
}
