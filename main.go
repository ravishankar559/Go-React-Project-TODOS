package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello world")
	app := fiber.New()

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")

	MONGODB_URI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(MONGODB_URI)

	client, err := mongo.Connect(context.Background(), clientOptions)

	defer client.Disconnect(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("golang_db").Collection("todos")

	app.Get("/api/todos", getTodos)

	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	log.Fatal(app.Listen(":" + PORT))
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}
	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return nil
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "TODO body should not be empty"})
	}
	insertResponse, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err
	}
	todo.ID = insertResponse.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": objectId}
	updateOperation := bson.M{"$set": bson.M{"completed": true}}
	_, err = collection.UpdateOne(context.Background(), filter, updateOperation)

	if err != nil {
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": "true"})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	filter := bson.M{"_id": objectId}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}
