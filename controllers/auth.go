package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/helpers"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *fiber.Ctx) error{
	var data map[string]string
	if err:= c.BodyParser(&data); err!=nil {
		return c.Status(400).JSON(fiber.Map{"error":"Invalid input"})
	}
	hashedPassword,_:= bcrypt.GenerateFromPassword([]byte(data["password"]), 14)
	user := models.User{
		ID: uuid.New(),
		Name: data["name"],
		Email: data["email"],
		Password: string(hashedPassword),
		Role: "student",
	}
	config.DB.Create(&user)
	return c.JSON(fiber.Map{"message":"User created successfully"})
}

func Login(c *fiber.Ctx) error{
	var data map[string]string
	if err:= c.BodyParser(&data); err!= nil {
		return c.Status(400).JSON(fiber.Map{"error":"Inavlid input"})
	}
	var user models.User
	config.DB.Where("email=?",data["email"]).First(&user)

	if user.ID==uuid.Nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid email/password"})
	}
	err:= bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(data["password"]))
	if err!= nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid email/password"})
	}
	token, err := helpers.GenerateJWT(user.ID.String(), string(user.Role))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
    }

    return c.JSON(fiber.Map{"token": token})
}