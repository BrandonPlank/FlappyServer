package main

import (
	"brandonplank.org/FlappyServer/database"
	"brandonplank.org/FlappyServer/global"
	"brandonplank.org/FlappyServer/models"
	"brandonplank.org/FlappyServer/routes"
	"github.com/common-nighthawk/go-figure"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"strconv"
)

const port = 8069

func initDatabase() {
	var err error
	database.DatabaseConnection, err = gorm.Open("sqlite3", "flappybird.db")
	routes.HandleError(err)
	log.Println("[DATABASE] Connection Opened to Database")
	database.DatabaseConnection.AutoMigrate(&models.User{})
}

func setupRoutes(app *fiber.App) {
	app.Use(
		logger.New(
			logger.Config{
				Format:     "${time} [${method}]->${status} Latency->${latency} - ${path} | ${error}\n",
				TimeFormat: "2006/01/02 15:04:05",
				Output:     global.Writer,
			},
		),
		cors.New(cors.Config{
			AllowCredentials: true,
		}),
		func(ctx *fiber.Ctx) error {
			ctx.Append("Access-Control-Allow-Origin", "*")
			ctx.Append("Developer", "crypticplank")
			ctx.Append("License", "BSD 3-Clause License")
			ctx.Append("Source-Url", "https://github.com/crypticplank/FlappyServer")
			return ctx.Next()
		},
	)

	app.Static("/", "./Public")

	v1 := app.Group("/v1")

	authv1 := v1.Group("/auth")
	authv1.Use(basicauth.New(basicauth.Config{
		Authorizer:      models.Auth,
		ContextUsername: "name",
	}))

	app.Get("/", routes.Home)
	app.Get("/bans", routes.Bans)
	app.Get("/user/:name", routes.GetUser)

	v1.Get("/leaderboard/:amount", routes.Leaderboard)
	v1.Get("/globalDeaths", routes.GlobalDeaths)
	v1.Get("/userCount", routes.UserCount)
	v1.Get("/users", routes.GetUsers)
	v1.Get("/getID/:name", routes.GetID)
	v1.Post("/registerUser", routes.RegisterUser)

	authv1.Post("/submitScore", routes.SubmitScore)
	authv1.Post("/submitDeath", routes.SubmitDeath)
	authv1.Post("/isJailbroken", routes.IsJailbroken)
	authv1.Post("/emulator", routes.Emulator)
	authv1.Post("/hasHackedTools", routes.HasHackedTools)
	authv1.Post("/login", routes.Login)

	authv1.Get("/internalUsers", routes.InternalUsers)
	authv1.Get("/ban/:id/:reason", routes.Ban)
	authv1.Get("/makeAdmin/:id", routes.MakeAdmin)
	authv1.Get("/unban/:id", routes.UnBan)
	authv1.Get("/delete/:id", routes.DeleteUser)
	authv1.Get("/restoreScore/:id/:score", routes.RestoreScore)
	authv1.Get("/logs", routes.ServerLogFile)
}

func main() {
	myFigure := figure.NewFigure("FlappyBird Server", "", true)
	myFigure.Print()

	log.Println("[START] Starting the FlappyBird REST server")

	// Setup dotenv
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] Error loading .env file")
	}

	global.SECRET_TOKEN = os.Getenv("SECRET_TOKEN")

	global.OWNER_OVERRIDE = os.Getenv("GLOBAL_OWNER_OVERRIDE_KEY")
	log.Println("[START] Got secret token:", global.SECRET_TOKEN)
	log.Println("[START] Got owner override key: " + global.OWNER_OVERRIDE[:4] + "***************************")

	// Setup Logging
	file, err := os.OpenFile("flappyserver.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	routes.HandleError(err)
	defer file.Close()
	global.Writer = io.MultiWriter(os.Stdout, file)
	log.SetOutput(global.Writer)

	//Setup views
	engine := html.New("./Resources/Views", ".html")
	//engine.Reload(true)
	//engine.Debug(true)

	router := fiber.New(fiber.Config{DisableStartupMessage: true, Views: engine})

	initDatabase()

	log.Println("[START] Setting up routes")
	setupRoutes(router)

	log.Println("[START] Starting server on port", strconv.Itoa(port))
	log.Fatalln(router.Listen(":" + strconv.Itoa(port)))

	defer database.DatabaseConnection.Close()
}
