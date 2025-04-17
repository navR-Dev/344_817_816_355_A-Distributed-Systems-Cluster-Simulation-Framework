import React, { useEffect, useState } from "react";
import axios from "axios";
import "./App.css"; // Import the CSS file

function App() {
  const [nodes, setNodes] = useState({});
  const [pods, setPods] = useState({});

  const [cpuCores, setCpuCores] = useState("");
  const [resourceType, setResourceType] = useState("node");

  const API_URL = "http://localhost:5000/api/nodes";
  const PODS_API_URL = "http://localhost:5000/api/pods";

  const fetchNodes = async () => {
    try {
      const response = await axios.get(API_URL);
      setNodes(response.data);
    } catch (error) {
      console.error("Error fetching nodes:", error);
    }
  };

  const fetchPods = async () => {
    try {
      const response = await axios.get(PODS_API_URL);
      setPods(response.data);
      console.log(response.data);
    } catch (error) {
      console.error("Error fetching pods:", error);
    }
  };

  const addnodepod = async () => {
    const endpoint = resourceType === "node" ? API_URL : PODS_API_URL;

    try {
      // Parse the CPU cores input and validate it
      const cores = parseInt(cpuCores);
      if (isNaN(cores) || cores <= 0) {
        return alert("Invalid CPU core count");
      }

      // Send the request to the appropriate endpoint
      await axios.post(endpoint, {
        cpu_cores: cores, // Send 'cpu_cores' to the server
      });

      setCpuCores(""); // Clear the input field after successful submission
      fetchNodes(); // Refresh the nodes or pods after adding
    } catch (error) {
      console.error("Error adding resource:", error);
    }
  };

  const deleteNode = async (id) => {
    try {
      await axios.delete(`${API_URL}/${id}`);
      fetchNodes();
    } catch (error) {
      console.error("Error deleting node:", error);
    }
  };

  useEffect(() => {
    fetchNodes();
    fetchPods();

    const interval = setInterval(() => {
      fetchNodes();
      fetchPods();
    }, 5000); // every 5 seconds

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="container">
      <h1 className="title">🚀 Node Dashboard</h1>
      <div className="form">
        <select
          value={resourceType}
          onChange={(e) => setResourceType(e.target.value)}
          className="dropdown"
        >
          <option value="node">Node</option>
          <option value="pod">Pod</option>
        </select>

        <input
          type="number"
          placeholder="Enter CPU cores"
          value={cpuCores}
          onChange={(e) => setCpuCores(e.target.value)}
          className="input"
        />

        <button onClick={addnodepod} className="button">
          Add {resourceType.charAt(0).toUpperCase() + resourceType.slice(1)}
        </button>
      </div>

      <h2 className="subtitle">🧠 All Nodes</h2>
      {Object.keys(nodes).length === 0 ? (
        <p className="noNodes">No nodes available.</p>
      ) : (
        <div className="nodeList">
          {Object.entries(nodes).map(([id, node]) => (
            <div key={id} className="nodeCard">
              <p>
                <strong>ID:</strong> {node.id.slice(0, 8) + "..."}
              </p>
              <p>
                <strong>Status:</strong>{" "}
                <span
                  style={{ color: node.status === "Running" ? "green" : "red" }}
                >
                  {node.status}
                </span>
              </p>
              <p>
                <strong>CPU:</strong> {node.cpu}
              </p>
              <p>
                <strong>Available CPU:</strong> {node.available_cpu}
              </p>
              <p>
                <strong>Last Heartbeat:</strong>{" "}
                {node.last_heartbeat.toFixed(2)}s ago
              </p>
              <button onClick={() => deleteNode(id)} className="deleteButton">
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
      <h2 className="subtitle">📦 All Pods</h2>
      {Object.keys(pods).length === 0 ? (
        <p className="noNodes">No pods available.</p>
      ) : (
        <div className="nodeList">
          {Object.entries(pods).map(([id, pod]) => (
            <div key={id} className="nodeCard">
              <p>
                <strong>Pod ID:</strong> {pod.id}
              </p>
              <p>
                <strong>Status:</strong>{" "}
                <span
                  style={{
                    color: pod.status === "Running" ? "green" : "orange",
                  }}
                >
                  {pod.status}
                </span>
              </p>
              <p>
                <strong>Node ID:</strong> {pod.node_id.slice(0, 8) + "..."}
              </p>
              <p>
                <strong>CPU Requested:</strong> {pod.cpu}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default App;
