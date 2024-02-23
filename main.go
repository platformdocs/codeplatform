package main

import (
	"context"
	"fmt"
	"log"
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
	Content string `json:"content"`
	Id      string `json:"id"`
}

var (
	DB *Database
)

type Database struct {
	Mongo *mongo.Client
}

type CodeData struct {
	Id      string `bson:"_id,omitempty" json:"id"`
	Name    string `bson:"name" json:"name"`
	Content string `bson:"content" json:"content"`
	Time    string `bson:"time" json:"time"`
	Type    string `bson:"type" json:"type"`
}

func Init() {
}

func connectDB() {
	opts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	DB = &Database{
		Mongo: client,
	}
}

func insertCode(ctx *gin.Context, codeType string, content string) (*mongo.InsertOneResult, error) {

	client := DB.Mongo
	// 获取数据库和集合
	collection := client.Database("codeplatform").Collection("codeList")

	results, _ := searchCode(ctx, codeType)

	var index int = len(results) + 1

	codedata := CodeData{}
	codedata.Id = uuid.New().String()
	codedata.Time = time.Now().Format("2006-01-02 15:04:05")
	codedata.Content = content
	if codeType == "save" {
		codedata.Name = "main" + strconv.Itoa(index) + ".go"
	} else {
		codedata.Name = "exec" + strconv.Itoa(index) + ".go"
	}

	codedata.Type = codeType
	insertOneResult, err := collection.InsertOne(ctx, &codedata)

	return insertOneResult, err
}

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
	if err != nil {
		log.Println("saved code List error is ", err)
	}

	var results []CodeData
	if err = cursor.All(ctx, &results); err != nil {
		log.Println("saved code List error is ", err)
	}
	return results, err
}

func updateCode(ctx *gin.Context, id string, content string) (*mongo.UpdateResult, error) {
	client := DB.Mongo
	collection := client.Database("codeplatform").Collection("codeList")
	filter := bson.M{"_id": id}
	value := bson.M{"$set": bson.M{
		"Content": content,
		"Time":    time.Now().Format("2006-01-02 15:04:05"),
	}}

	updateOneResult, err := collection.UpdateOne(ctx, filter, value)
	if err != nil {
		log.Println("update user data failed, err is ", err)
	}
	log.Println("update success !")

	return updateOneResult, err
}

/*
* write code into a file
 */
func writeFile(filePath string, content string) bool {

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return false
	}

	fmt.Println("File written successfully.")
	return true
}

/*
* execute file and return the result
 */
func execFile(filePath string) (string, string) {
	cmd := exec.Command("go", "run", filePath)
	// var stdout, stderr bytes.Buffer
	// cmd.Stdout = &stdout // 标准输出
	// cmd.Stderr = &stderr // 标准错误
	// err := cmd.Run()
	// outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	// fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	// if err != nil {
	// 	log.Fatalf("cmd.Run() failed with %s\n", err)
	// }
	doneChan := make(chan []byte, 1)
	errorChan := make(chan string, 1)
	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("execute command failed, output: %s, error: %v\n", string(output), err)
			var msg string = string(output) + err.Error()
			errorChan <- msg
			return
		}
		doneChan <- output
	}()

	select {
	case <-time.After(10 * time.Second):
		// log.Printf("execute command 10s timeout\n")
		cmd.Process.Kill()
		return "timeout", "execute code over 10s timeout"
	case output := <-doneChan:
		// fmt.Printf("out:\n%s", output)
		return "success", string(output)
	case err := <-errorChan:
		// log.Printf("execute command failure, error: %v\n", err)
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
		// test
		api.GET("/hello", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"msg": "world"})
		})

		// run code
		api.POST("/run", func(ctx *gin.Context) {
			var requestData RequestData
			if err := ctx.ShouldBindJSON(&requestData); err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			filePath := "execCode.go"
			// write code to the file
			writeFile(filePath, requestData.Content)

			// run the code and get the result message
			var resultType, message = execFile(filePath)

			_, err := insertCode(ctx, "exec", requestData.Content)

			if err != nil {
				log.Println("insert one error is ", err)
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}

			ctx.JSON(200, gin.H{
				"type":    resultType,
				"message": message,
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

			_, err := insertCode(ctx, "save", requestData.Content)

			if err != nil {
				log.Println("insert one error is ", err)
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(200, gin.H{
				"message": "保存成功",
			})
		})

		// update code
		api.POST("/update", func(ctx *gin.Context) {
			var requestData RequestData
			if err := ctx.ShouldBindJSON(&requestData); err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}

			_, err := updateCode(ctx, requestData.Id, requestData.Content)

			if err != nil {
				log.Println("insert one error is ", err)
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(200, gin.H{
				"message": "更新成功",
			})
		})
	}

	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(404, gin.H{"msg": "not found"}) })

	// 开始监听服务请求
	router.Run(":8080")
}
