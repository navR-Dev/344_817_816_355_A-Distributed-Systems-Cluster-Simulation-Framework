import docker
import time

docker_client = docker.from_env()


def launch_node_container(cpu_cores):
    try:
        container = docker_client.containers.run(
            "alpine",
            command="tail -f /dev/null",
            detach=True,
            name=f"node_{int(time.time())}",
            cpu_period=100000,
            cpu_quota=cpu_cores * 100000,
            mem_limit="512m",  # Limit memory for simulation
        )
        return container.id
    except Exception as e:
        print(f"Error launching container: {e}")
        return None
