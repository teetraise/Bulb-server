package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Initializing Gin...")
	r := gin.Default()

	log.Println("Setting up routes...")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	log.Println("Starting Bulb API Server on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
