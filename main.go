package main

import (
	"database/sql"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tamago0224/rest-app-backend/controllers"
	"github.com/tamago0224/rest-app-backend/repository"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	todoRepository := repository.NewTodoMariaDBRepository(db)
	todoController := controllers.NewTodoController(todoRepository)
	userRepository := repository.NewUserMariaDBRepository(db)
	userController := controllers.NewUserController(userRepository)
	authController := controllers.NewAuthController(userRepository)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.POST("/login", authController.Login)
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte("secret"),
	}))
	apiGroup.GET("/todos", todoController.GetTodoList)
	apiGroup.POST("/todos", todoController.AddTodo)
	apiGroup.GET("/todos/:id", todoController.GetTodo)
	apiGroup.DELETE("/todos/:id", todoController.DeleteTodo)

	apiGroup.GET("/users", userController.SearchUser)
	apiGroup.POST("/users", userController.CreateUser)

	e.Logger.Fatal(e.Start(":8080"))
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "todo:hogehoge@tcp(db:3306)/todo")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}
