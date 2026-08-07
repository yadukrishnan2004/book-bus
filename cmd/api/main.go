package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	g := gin.Default()

	g.GET("/health",func(c *gin.Context) {
		c.JSON(http.StatusOK,gin.H{
			"status" : "ok",
			"message" : "Bus Ticket Booking API is running!",
		})
	})

	g.Run(":8080")
}