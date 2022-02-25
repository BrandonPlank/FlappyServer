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
	database.DB, err = gorm.Open("sqlite3", "flappybird.db")
	routes.HandleError(err)
	log.Println("[DATABASE] Connection Opened to Database")
	database.DB.AutoMigrate(&models.User{})
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
			ctx.Append("Developer", "Brandon Plank")
			ctx.Append("License", "BSD 3-Clause License")
			ctx.Append("Source-Url", "https://github.com/brandonplank/FlappyServer")
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

	v1.Get("/leaderboard/:amount", routes.V1Leaderboard)
	v1.Get("/globalDeaths", routes.V1GlobalDeaths)
	v1.Get("/userCount", routes.V1UserCount)
	v1.Get("/users", routes.V1GetUsers)
	v1.Get("/getID/:name", routes.V1GetID)
	v1.Post("/registerUser", routes.V1RegisterUser)

	authv1.Post("/submitScore", routes.V1SubmitScore)
	authv1.Post("/submitDeath", routes.V1SubmitDeath)
	authv1.Post("/isJailbroken", routes.V1IsJailbroken)
	authv1.Post("/emulator", routes.V1Emulator)
	authv1.Post("/hasHackedTools", routes.V1HasHackedTools)
	authv1.Post("/login", routes.V1Login)

	authv1.Get("/internalUsers", routes.V1InternalUsers)
	authv1.Get("/ban/:id/:reason", routes.V1Ban)
	authv1.Get("/makeAdmin/:id", routes.V1MakeAdmin)
	authv1.Get("/unban/:id", routes.V1UnBan)
	authv1.Get("/delete/:id", routes.V1DeleteUser)
	authv1.Get("/restoreScore/:id/:score", routes.V1RestoreScore)
	authv1.Get("/logs", routes.V1ServerLogFile)
}

func main() {

	// Setup Logging
	file, err := os.OpenFile("flappyserver.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	routes.HandleError(err)
	defer file.Close()
	global.Writer = io.MultiWriter(os.Stdout, file)
	log.SetOutput(global.Writer)

	myFigure := figure.NewFigure("FlappyBird Server", "", true)
	myFigure.Print()

	log.Println("[START] Starting the FlappyBird REST server")

	// Setup dotenv
	err = godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] Error loading .env file")
	}

	global.SecretToken = os.Getenv("SECRET_TOKEN")
	global.OwnerOverride = os.Getenv("GLOBAL_OWNER_OVERRIDE_KEY")

	if len(global.SecretToken) < 1 {
		log.Fatal("To start the server, you must have SECRET_TOKEN defined in .env")
	}

	if len(global.OwnerOverride) <= 4 {
		log.Fatal("To start the server, you must have OWNER_OVERRIDE defined in .env, must be more than 4 letters")
	}

	log.Println("[START] Got secret token:", global.SecretToken)
	log.Println("[START] Got owner override key: " + global.OwnerOverride[:4] + "***************************")

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

	defer database.DB.Close()
}
