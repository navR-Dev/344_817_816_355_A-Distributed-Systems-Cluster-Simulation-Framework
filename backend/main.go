package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Node struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	CPU            int       `json:"cpu"`
	AvailableCPU   int       `json:"available_cpu"`
	Pods           []string  `json:"pods"`
	LastHeartbeat  float64   `json:"last_heartbeat"`
	LastUpdateTime time.Time `json:"-"`
}

var nodes = make(map[string]*Node)

func main() {

	r := gin.Default()
	// Enable CORS for frontend interaction
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // or "*" if you're testing
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// Get all nodes
	r.GET("/api/nodes", func(c *gin.Context) {
		c.JSON(http.StatusOK, nodes)
	})

	// Register a new node
	r.POST("/api/nodes", func(c *gin.Context) {
		var body struct {
			CPUCores int `json:"cpu_cores"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.CPUCores <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		id := uuid.New().String()
		now := time.Now()
		node := &Node{
			ID:             id,
			Status:         "Running",
			CPU:            body.CPUCores,
			AvailableCPU:   body.CPUCores,
			Pods:           []string{},
			LastUpdateTime: now,
		}
		nodes[id] = node
		c.JSON(http.StatusOK, node)
	})

	// Delete a node
	r.DELETE("/api/nodes/:id", func(c *gin.Context) {
		id := c.Param("id")
		if node, exists := nodes[id]; exists {
			delete(nodes, id)
			c.JSON(http.StatusOK, gin.H{"message": "Node deleted", "node": node})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		}
	})

	// Optional: Heartbeat update
	go func() {
		for {
			time.Sleep(3 * time.Second)
			now := time.Now()
			for id, node := range nodes {
				node.LastHeartbeat = now.Sub(node.LastUpdateTime).Seconds()
				node.LastUpdateTime = now
				nodes[id] = node // Update the node in the map
			}
		}
	}()

	fmt.Println("Server running at http://localhost:5000")
	r.Run(":5000")
}
