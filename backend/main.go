package main

import (
	"github.com/docker/docker/api/types/container"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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

type Pod struct {
	ID     string `json:"id"`
	CPU    int    `json:"cpu"`
	NodeID string `json:"node_id"`
	Status string `json:"status"`
}

var (
	nodes = make(map[string]*Node)
	pods  = make(map[string]*Pod)
)

func main() {
	r := gin.Default()

	// Enable CORS for frontend interaction
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // or "*" if you're testing
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// Docker client setupadd it to my previous golang code
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

	// Get all pods
	r.GET("/api/pods", func(c *gin.Context) {
		c.JSON(http.StatusOK, pods)
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

		// creating unique container name
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)
		num := r.Intn(900) + 100
		numStr := strconv.Itoa(num)
		container_name := "cont" + numStr

		// Create the container
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "alpine",
			Cmd:   []string{"tail", "-f", "/dev/null"},
		}, nil, nil, nil, container_name)
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

	// Create pods
	r.POST("/api/pods", func(c *gin.Context) {

		var req struct {
			CPUCores int `json:"cpu_cores"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.CPUCores <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CPU request"})
			return
		}

		// Created pod scheduling with First-fit

		var selectedNode *Node
		for _, node := range nodes {
			if node.Status == "Running" && node.AvailableCPU >= req.CPUCores {
				selectedNode = node
				break
			}
		}

		if selectedNode == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available node"})
			return
		}

		// Allocate resources to pod
		podID := uuid.New().String()
		pod := &Pod{
			ID:     podID,
			CPU:    req.CPUCores,
			NodeID: selectedNode.ID,
			Status: "Running",
		}

		// Add pod to selected node
		selectedNode.Pods = append(selectedNode.Pods, podID)
		selectedNode.AvailableCPU -= req.CPUCores

		// Store pod in global map
		pods[podID] = pod

		c.JSON(http.StatusOK, pod)
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

	// Health monitor code
	go func() {
		for {
			time.Sleep(1 * time.Second)
			now := time.Now()

			// Log current status of nodes and pods
			fmt.Println("----- Nodes -----")
			for id, node := range nodes {
				fmt.Printf("Node ID: %s\n", id)
				fmt.Printf("Status: %s\n", node.Status)
				fmt.Printf("CPU: %d\n", node.CPU)
				fmt.Printf("Available CPU: %d\n", node.AvailableCPU)
				fmt.Printf("Last Heartbeat: %.2f seconds ago\n", node.LastHeartbeat)
				fmt.Println("Pods:")
				for _, podID := range node.Pods {
					pod, exists := pods[podID]
					if exists {
						fmt.Printf("  Pod ID: %s, Status: %s, CPU: %d\n", pod.ID, pod.Status, pod.CPU)
					}
				}
				fmt.Println("---------------")
			}

			fmt.Println("----- Pods -----")
			for _, pod := range pods {
				fmt.Printf("Pod ID: %s\n", pod.ID)
				fmt.Printf("Status: %s\n", pod.Status)
				fmt.Printf("CPU: %d\n", pod.CPU)
				fmt.Printf("Node ID: %s\n", pod.NodeID)
				fmt.Println("---------------")
			}

			for id, node := range nodes {

				// if len(node.Pods) == 0 {
				// 	fmt.Println(node.Pods)
				// 	continue
				//  }

				containerID := node.Pods[0]

				// Inspect the container to get its current state
				inspect, err := cli.ContainerInspect(ctx, containerID)
				if err != nil || !inspect.State.Running {

					node.Status = "Unhealthy"

					// Mark all pods as offline
					for _, podID := range node.Pods {
						if pod, exists := pods[podID]; exists {
							pod.Status = "Offline"
							// Try to reschedule
							pod.NodeID = ""
							for _, altNode := range nodes {
								if altNode.ID != node.ID && altNode.Status == "Running" && altNode.AvailableCPU >= pod.CPU {
									pod.NodeID = altNode.ID
									pod.Status = "Running"
									altNode.Pods = append(altNode.Pods, podID)
									altNode.AvailableCPU -= pod.CPU
									nodes[altNode.ID] = altNode
									break
								}
							}
							pods[podID] = pod
						}
					}

					// Clear pods from the unhealthy node

					// node.Pods = []string{}
					node.AvailableCPU = node.CPU
				} else {
					node.Status = "Running"
				}

				node.LastHeartbeat = now.Sub(node.LastUpdateTime).Seconds()
				node.LastUpdateTime = now
				nodes[id] = node
			}
		}
	}()

	// Start the Gin server
	fmt.Println("Server running at http://localhost:5000")
	r.Run(":5000")
}
