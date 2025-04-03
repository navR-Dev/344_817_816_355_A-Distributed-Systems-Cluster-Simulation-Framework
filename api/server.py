from flask import Flask, jsonify, request
import time
import threading
import logging
from node_manager import NodeManager, HEARTBEAT_INTERVAL, HEARTBEAT_TIMEOUT
from docker_utils import launch_node_container

# Initialize Flask app
app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize components
node_manager = NodeManager()

@app.route('/')
def health_check():
    """Endpoint to check API server health"""
    return jsonify({
        "status": "API Server running",
        "node_count": len(node_manager.get_nodes()),
        "timestamp": time.time()
    })

@app.route('/nodes', methods=['GET'])
def list_nodes():
    """List all nodes in the cluster"""
    return jsonify(node_manager.get_nodes())

@app.route('/nodes', methods=['POST'])
def add_node():
    """Add a new node to the cluster"""
    data = request.json
    cpu_cores = data.get('cpu_cores', 1)
    
    if cpu_cores <= 0:
        return jsonify({"error": "CPU cores must be positive"}), 400
    
    try:
        # Launch container (or simulate if Docker not available)
        container_id = launch_node_container(cpu_cores)
        if not container_id:
            logger.warning("Docker not available - using simulated node")
            container_id = f"simulated_node_{int(time.time())}"
        
        # Register node
        node_manager.register_node(container_id, cpu_cores)
        
        # Start heartbeat simulation
        threading.Thread(
            target=simulate_node_heartbeat,
            args=(container_id,),
            name=f"Heartbeat-{container_id[:8]}",
            daemon=True
        ).start()
        
        logger.info(f"Added new node: {container_id} with {cpu_cores} CPU cores")
        return jsonify({
            "message": "Node added successfully",
            "node_id": container_id,
            "cpu_cores": cpu_cores,
            "is_simulated": container_id.startswith("simulated_node_")
        }), 201
    
    except Exception as e:
        logger.error(f"Error adding node: {str(e)}")
        return jsonify({"error": str(e)}), 500

@app.route('/pods', methods=['POST'])
def launch_pod():
    """Launch a new pod in the cluster"""
    data = request.json
    cpu_required = data.get('cpu_required')
    
    if not cpu_required or cpu_required <= 0:
        return jsonify({"error": "CPU requirement must be positive"}), 400
    
    # Week 2 will implement actual pod scheduling
    return jsonify({
        "message": "Pod scheduling will be implemented in Week 2",
        "warning": "This is a placeholder implementation"
    }), 501  # 501 Not Implemented

def simulate_node_heartbeat(node_id):
    """Simulate periodic heartbeats from a node"""
    while node_manager.node_exists(node_id):
        try:
            time.sleep(HEARTBEAT_INTERVAL)
            node_manager.update_heartbeat(node_id)
            logger.debug(f"Heartbeat from node {node_id[:8]}...")
        except Exception as e:
            logger.error(f"Heartbeat error for node {node_id[:8]}: {str(e)}")
            break

def start_health_monitor():
    """Background thread to monitor node health"""
    def monitor():
        while True:
            try:
                time.sleep(HEARTBEAT_INTERVAL)
                node_manager.check_node_health()
                logger.debug("Health monitor check completed")
            except Exception as e:
                logger.error(f"Health monitor error: {str(e)}")
    
    threading.Thread(
        target=monitor,
        name="HealthMonitor",
        daemon=True
    ).start()
    logger.info("Started health monitor thread")

if __name__ == '__main__':
    # Start background services
    start_health_monitor()
    
    # Start Flask server
    logger.info("Starting API server on port 5000")
    app.run(host='0.0.0.0', port=5000, debug=False)