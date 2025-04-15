package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
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

	// Docker client setup
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Get all nodes
	r.GET("/api/nodes", func(c *gin.Context) {
		c.JSON(http.StatusOK, nodes)
	})

	// Register a new node and launch a container
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

		// Pull the Docker image (Alpine in this case)
		reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", image.PullOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pull image"})
			return
		}
		defer reader.Close()
		io.Copy(os.Stdout, reader) // Log output of the image pull

		// Create the container
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "alpine",
			Cmd:   []string{"tail", "-f", "/dev/null"},
		}, nil, nil, nil, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create container"})
			return
		}

		// Start the container
		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start container"})
			return
		}

		// Store node data
		node := &Node{
			ID:             id,
			Status:         "Running",
			CPU:            body.CPUCores,
			AvailableCPU:   body.CPUCores,
			Pods:           []string{resp.ID}, // Track container ID as a "pod"
			LastUpdateTime: now,
		}
		nodes[id] = node

		c.JSON(http.StatusOK, node)
	})

	// Delete a node and stop its container
	r.DELETE("/api/nodes/:id", func(c *gin.Context) {
		id := c.Param("id")
		if node, exists := nodes[id]; exists {
			// Stop and remove the container (simulate "node" removal)
			noWaitTimeout := 0
			if err := cli.ContainerStop(ctx, node.Pods[0], containertypes.StopOptions{Timeout: &noWaitTimeout}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop container"})
				return
			}

			if err := cli.ContainerRemove(ctx, node.Pods[0], container.RemoveOptions{}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove container"})
				return
			}

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

	// Start the Gin server
	fmt.Println("Server running at http://localhost:5000")
	r.Run(":5000")
}
