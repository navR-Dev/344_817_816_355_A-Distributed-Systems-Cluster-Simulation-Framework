
# Project Setup Guide

This project consists of two main components:

- **Backend**: A RESTful API built with Go (Golang)
- **Frontend**: A user interface built with React

## Prerequisites

Before you begin, ensure you have the following installed on your local machine:

- Go (version 1.x)
- Node.js (version 16.x or above)
- npm (or yarn)

## Project Structure


## Backend Setup (Go)

### 1. Navigate to the backend folder

```bash
    cd backend
    go mod tidy
    go run main.go
```


## Frontend Setup (React)

1. Navigate to the frontend folder

2. Install Node.js dependencies

3. Run the frontend server

Start the React development server by running:

```bash
    cd frontend
    npm install
    npm start
```




## TO-DO


# Test 1 { node stops , then comes back online after some time }

* node stops 
* all pods assigned to node go offline/pending 
* clear pods array in stopped NODE 
* clear nodeID in POD

* restart node
* pod scheduler should keep checking if a valid node exists
* if valid node exists assign pods to node

