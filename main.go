package main

import (
	"context"
	"defense-dashboard/model"
	"defense-dashboard/pkg/cli"
	"defense-dashboard/pkg/helper"
	"defense-dashboard/pkg/route"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Set log to stderr
	log.SetOutput(os.Stderr)

	dsn := "host=127.0.0.1 user=postgres password=hitcon-defense-2021 dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Taipei"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("-", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("-", err)
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Minute)

	if err != nil {
		log.Fatal("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(
		model.Team{},
		model.Quest{},
		model.Event{},
		model.Status{},
	)

	// Seed team data from file
	helper.SeedTeam(db, "data/teams.txt")
	// Seed quest data from file
	helper.SeedQuest(db, "data/quests.csv")

	// Start the menu
	quit := make(chan bool)
	go cli.Menu(db, quit)

	// Start the server
	r := gin.New()
	// r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.GET("/ping", route.PingHandler)
	r.GET("/service/:status", route.ServiceHandler)
	t := r.Group("/team")
	t.GET("/", route.Controller{DB: db}.TeamsHandler)
	t.GET("/:id", route.Controller{DB: db}.TeamHandler)
	t.GET("/view", route.TeamViewHandler)
	t.GET("/events", route.Controller{DB: db}.TeamViewLogsHandler)
	t.GET("/status", route.Controller{DB: db}.TeamViewStatusHandler)
	b := r.Group("/board")
	b.GET("/", route.TeamBoardHandler)
	b.GET("/status", route.Controller{DB: db}.TeamStatusHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("[-] listen: %s\n", err)
		}
	}()
	<-quit // exit
	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("[-] Server forced to shutdown:", err)
	}
	log.Println("[+] Server exiting")
	fmt.Fprint(os.Stdout, "\x1b[?25h")
}
