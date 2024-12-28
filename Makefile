testUpdateSuccess:
	@echo "Building the Go application..."
	go build -o ./cmd ./src/cmd
	@echo "Starting services with service.yml and streaming logs..."
	(./cmd up service.yml > up.log 2>&1 & echo $$! > up.pid) & \
	tail -f up.log & \
	PID=$$!; \
	sleep 10; \
	echo "Updating services with service_2.yml..."; \
	./cmd update service_2.yml; \
	sleep 5; \
	echo "Shutting down services..."; \
	./cmd down; \
	echo "Stopping the log streaming..."; \
	kill $$PID; \
	kill `cat up.pid`; \
	rm up.pid; \
	echo "Verifying no containers are running..."; \
	if [ -z "$$(docker ps -aq)" ]; then \
		echo "All containers successfully stopped and removed."; \
	else \
		echo "Some containers are still running:"; \
		docker ps -a; \
		exit 1; \
	fi; \
	echo "Test update sequence completed successfully."

