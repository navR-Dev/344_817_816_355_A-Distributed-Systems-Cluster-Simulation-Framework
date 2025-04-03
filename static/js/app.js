document.addEventListener("DOMContentLoaded", function () {
  // Initialize Socket.IO connection
  const socket = io();

  // Chart initialization
  const ctx = document.getElementById("resourceChart").getContext("2d");
  const chart = new Chart(ctx, {
    type: "doughnut",
    data: {
      labels: ["Used CPU", "Available CPU"],
      datasets: [
        {
          data: [0, 0],
          backgroundColor: ["#ff6384", "#36a2eb"],
        },
      ],
    },
  });

  // Form submission for adding nodes
  document.getElementById("addNodeForm").addEventListener("submit", function (e) {
    e.preventDefault();
    const cpuCores = parseInt(document.getElementById("cpuCores").value);

    fetch("/api/nodes", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ cpu_cores: cpuCores }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.status === "success") {
          alert(`Node added: ${data.node_id}`);
        }
      });
  });

  // Socket.IO event listeners
  socket.on("node_update", function (nodes) {
    updateNodeTable(nodes);
    updateChart(nodes);
    document.getElementById("node-count").textContent = Object.keys(nodes).length;
  });

  socket.on("heartbeat", function (data) {
    console.log(`Heartbeat from node: ${data.node_id}`);
  });

  // Helper functions
  function updateNodeTable(nodes) {
    const tbody = document.getElementById("nodeTableBody");
    tbody.innerHTML = "";

    for (const [node_id, node] of Object.entries(nodes)) {
      const row = document.createElement("tr");

      row.innerHTML = `
                <td>${node_id.substring(0, 8)}...</td>
                <td><span class="badge bg-${node.status === "healthy" ? "success" : "danger"}">${node.status}</span></td>
                <td>${node.cpu}</td>
                <td>${node.available_cpu}</td>
                <td>${node.pods.length}</td>
                <td>${node.last_heartbeat.toFixed(2)}</td>
            `;

      tbody.appendChild(row);
    }
  }

  function updateChart(nodes) {
    let usedCPU = 0;
    let availableCPU = 0;

    Object.values(nodes).forEach((node) => {
      usedCPU += node.cpu - node.available_cpu;
      availableCPU += node.available_cpu;
    });

    chart.data.datasets[0].data = [usedCPU, availableCPU];
    chart.update();
  }
});
