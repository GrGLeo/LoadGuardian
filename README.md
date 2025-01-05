# LoadGuardian Service
![Logo](doc/logo.webp)

## Overview
This project is a Work-In-Progress (WIP) replacement for Docker Compose, designed to include advanced features like autoscaling across multi-cluster environments. It offers a distinctive command execution method through a socket interface, enabling streamlined management of services and containers. The project aims to provide a more robust and scalable solution for container orchestration, addressing the complexities of multi-cluster deployments while maintaining simplicity and efficiency.

## Features
- **Command Execution via Socket:** All commands are transmitted through a socket interface, enabling seamless integration and remote control of services.
- **Command `up`:**
  - `-f`: Specify a YAML file containing the configuration and service descriptions.
  - `-s`: Schedule a delay in hours before the command execution.
  - **Description:** This command initiates the services as outlined in the specified YAML file, ensuring they are up and running according to the defined parameters.
- **Command `update`:**
  - `-f`: Provide a YAML file with the new configuration, allowing for the removal, addition, or updating of services.
  - `-s`: Set a delay in hours before executing the update.
  - **Description:** The update command modifies the currently running services based on the new configuration provided, with flexibility in service management.
  - **Rollback Mechanism:** In the event of an unhealthy container, the system automatically reverts to the previous configuration, replacing containers sequentially to minimize downtime and maintain service continuity.
- **Command `info`:**
  - **Output:** This command delivers a detailed table displaying information about services and their container replicas, including:
    - Service Name
    - Container Name
    - Health Status
    - CPU Usage
    - Memory Usage
- **Command `down`:**
  - `-s`: Schedule a delay in hours prior to executing the command.
  - **Description:** The down command halts and cleans up all active containers, thereby shutting down all services in an orderly manner.

## Future Features
- **Autoscaling Mechanism:** A robust autoscaling feature will be integrated, enhancing the system's capability to adjust resource allocation dynamically, following the implementation of a comprehensive health monitoring routine.
- **Container Restart:** An automatic restart function for containers identified as dead or unhealthy will be incorporated to ensure ongoing service availability and reliability.

## Usage
### Up Command
```bash
up -f <path_to_yaml_file> -s <schedule_delay_in_hours>
```

### Update Command
```bash
update -f <path_to_yaml_file> -s <schedule_delay_in_hours>
```

### Info Command
```bash
info
```

### Down Command
```bash
down -s <schedule_delay_in_hours>
```

## Contribution
Contributions are encouraged and welcomed! Whether through opening issues or submitting pull requests, your input will help enhance the project's functionality and stability.

## License
This project is licensed under the MIT License. For more details, please refer to the [LICENSE](LICENSE) file.
