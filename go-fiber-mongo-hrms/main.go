package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:3000/" + dbName

type Employee struct {
	ID     string  `json: "id,omitempty" bson: "_id,omitempty"`
	Name   string  `json: "name"`
	Salary float64 `json: "salary"`
	Age    float64 `json: "age"`
}

func Connect() error {
	//how to create  client using mongo
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))

	//How to create a timeout incase program not responding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	err = client.Connect(ctx)
	db := client.Database(dbName)

	//how to handle err
	if err != nil {
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
	return nil
}

// the control area of the program
func main() {

	//Error function for connect
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	//ROUTES FOR PROJECT
	app := fiber.New()

	// GET FUNCTION:  all the employee from database
	app.Get("/employee", func(c *fiber.Ctx) error {
		query := bson.D{{}}

		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		//create a var for employee using a slice
		var employees []Employee = make([]Employee, 0)

		//receives all the data from function "cursor" converts it into a struct
		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(employees)
	})
	// POST FUCNTION: print out employee details
	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")

		//defining varaiable "employee"
		employee := new(Employee)

		//formats into the format you need which is a struct that golang understands
		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		employee.ID = ""

		insertionResult, err := collection.InsertOne(c.Context(), employee)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)

		createdEmployeee := &Employee{}
		createdRecord.Decode(createdEmployeee)

		return c.Status(201).JSON(createdEmployeee)
	})
	// UPDATE FUNCTION:
	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeID, err := primitive.ObjectIDFromHex(idParam)

		if err != nil {
			return c.SendStatus(400)
		}

		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		query := bson.D{{Key: "_id", Value: employeeID}}
		update := bson.D{{
			//Update query
			Key: "$set",
			Value: bson.D{
				{Key: "name", Value: employee.Name},
				{Key: "age", Value: employee.Age},
				{Key: "salary", Value: employee.Salary},
			},
		},
		}

		err = mg.Db.Collection("employees").FindOneAndUpdate(c.Context(), query, update).Err()

		// if file is not found or exist sent error message
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.SendStatus(400)
			}
			// regular error message
			return c.SendStatus(500)
		}

		employee.ID = idParam

		return c.Status(200).JSON(employee)
	})

	// DELETE FUNCTION
	app.Delete("/emploee/:id", func(c *fiber.Ctx) error {

		employeeID, err := primitive.ObjectIDFromHex(c.Params("id"))

		// error message
		if err != nil {
			return c.SendStatus(400)
		}

		query := bson.D{{Key: "_id", Value: employeeID}}
		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), query)

		// error message
		if err != nil {
			return c.SendStatus(500)
		}

		if result.DeletedCount < 1 {
			return c.SendStatus(404)
		}

		return c.Status(200).JSON("record deleted")
	})

	log.Fatal(app.Listen(":3000"))

}
