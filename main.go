package main

import (
	"log"

	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tamago0224/rest-app-backend/controller"
	"github.com/tamago0224/rest-app-backend/domain/service"
	"github.com/tamago0224/rest-app-backend/infra"
	"github.com/tamago0224/rest-app-backend/infra/mariadb"
	"github.com/tamago0224/rest-app-backend/usecase"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := infra.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	todoRepository := mariadb.NewTodoMariaDBRepository(db)
	userRepository := mariadb.NewUserMariaDBRepository(db)

	todoService := service.NewTodoService(userRepository, todoRepository)
	userService := service.NewUserService(userRepository)

	todoUsecase := usecase.NewTodoUsecase(todoService)
	userUsecase := usecase.NewUserUsecase(userService)

	// oidc設定
	config, err := controller.NewOAuthConfig()
	if err != nil {
		log.Fatal(err)
	}
	todoController := controller.NewTodoController(todoUsecase)
	authController := controller.NewAuthController(userUsecase, config)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
	}))

	e.GET("/login", authController.Login)
	e.GET("/callback", authController.Callback)
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(controller.JwtCustomClaims)
		},
		SigningKey:  []byte("secret"),
		TokenLookup: "header:Authorization:Bearer ,cookie:auth_token",
	}))
	apiGroup.GET("/todos", todoController.GetTodoList)
	apiGroup.POST("/todos", todoController.AddTodo)
	apiGroup.GET("/todos/:id", todoController.GetTodo)
	apiGroup.DELETE("/todos/:id", todoController.DeleteTodo)

	e.Logger.Fatal(e.Start(":8080"))
}
