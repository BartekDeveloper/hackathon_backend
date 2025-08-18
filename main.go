package main

import (
	"database/sql"
	s "hack/backend/server"
	u "hack/backend/utils"
	"net/http"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type Shit struct {
	ID           int64  `json:"id"`
	NAME         string `json:"name"`
	DATE_CREATED string `json:"date_created"`
}

var db s.Database

func Router() *gin.Engine {
	router := gin.Default()

	// Removed for better dev-exp
	// allowedHosts: string[] ={
	// 	"127.0.0.1:8080"
	// 	"0.0.0.0:8080",
	// 	"localhost:8080",
	// }

	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length")
		c.Next()
	})

	indexHandler := func(c *gin.Context) {
		if !db.IsConnected {
			db.Connect()
		}

		if !db.IsConnected {
			u.Error("Database not present")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database not present"})
			return
		}

		shits := []Shit{}

		var rows *sql.Rows = nil
		_, err := db.Query("SELECT * FROM created_tables;", &rows)
		if err != nil {
			u.Error("Query Execution Failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
			return
		}

		for rows.Next() {
			var shit = Shit{}

			err := rows.Scan(&shit.ID, &shit.NAME, &shit.DATE_CREATED)
			if err != nil {
				u.ErrorF("%d", err)
			}

			shits = append(shits, shit)
		}
		defer rows.Close()

		c.JSON(http.StatusOK, gin.H{"data": shits})
	}

	indexRoute := router.Group("/")
	indexRoute.GET("/", indexHandler)

	router.Use(static.Serve("/assets", static.LocalFile("./public/", true)))

	return router
}

func main() {
	db = s.Database{}
	db.Config = s.DatabaseConfig{}
	db.Config.Host = "127.0.0.1"
	db.Config.Port = 5432
	db.Config.User = "postgres"
	db.Config.Password = "postgres"
	db.Config.Database = "postgres"
	db.Config.SSL = "disable"
	db.Config.Driver = s.Driver(s.POSTGRESQL)

	db.Connect()
	defer db.Close()

	router := Router()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       15 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil {
		u.ErrorF("Failed to start server!:\t%s\n", err.Error())
	}
}
