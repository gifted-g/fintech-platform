.PHONY: help run-all test docker-build k8s-deploy

# Variables
SERVICES := credit-scoring risk-engine user-verification fraud-detection notification
DOCKER_REGISTRY := your-registry
VERSION := latest

help:
	@echo "Available commands:"
	@echo "  make run-all          - Run all services locally"
	@echo "  make test             - Run all tests"
	@echo "  make docker-build     - Build all Docker images"
	@echo "  make docker-push      - Push Docker images to registry"
	@echo "  make k8s-deploy       - Deploy to Kubernetes"
	@echo "  make migrate-up       - Run database migrations"
	@echo "  make lint             - Run linters"

# Run all services
run-all:
	@echo "Starting infrastructure dependencies..."
	docker-compose up -d postgres redis kafka
	@echo "Waiting for services to be ready..."
	sleep 10
	@echo "Starting microservices..."
	@$(foreach service,$(SERVICES), \
		cd services/$(service) && go run main.go & \
	)

# Run tests
test:
	@$(foreach service,$(SERVICES), \
		echo "Testing $(service)..." && \
		cd services/$(service) && go test -v -race ./... && cd ../.. || exit 1; \
	)

# Run tests with coverage
test-coverage:
	@$(foreach service,$(SERVICES), \
		echo "Testing $(service) with coverage..." && \
		cd services/$(service) && \
		go test -v -race -coverprofile=coverage.out -covermode=atomic ./... && \
		go tool cover -html=coverage.out -o coverage.html && \
		cd ../.. || exit 1; \
	)

# Build Docker images
docker-build:
	@$(foreach service,$(SERVICES), \
		echo "Building $(service)..." && \
		docker build -t $(DOCKER_REGISTRY)/$(service):$(VERSION) services/$(service) || exit 1; \
	)

# Push Docker images
docker-push:
	@$(foreach service,$(SERVICES), \
		echo "Pushing $(service)..." && \
		docker push $(DOCKER_REGISTRY)/$(service):$(VERSION) || exit 1; \
	)

# Deploy to Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f infrastructure/kubernetes/namespace.yaml
	kubectl apply -f infrastructure/kubernetes/configmap.yaml
	kubectl apply -f infrastructure/kubernetes/secrets.yaml
	kubectl apply -f infrastructure/kubernetes/
	@echo "Waiting for deployments to be ready..."
	kubectl wait --for=condition=available --timeout=300s deployment --all -n fintech-platform

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	@for file in database/migrations/*.sql; do \
		echo "Applying $$file..."; \
		PGPASSWORD=fintech123 psql -h localhost -U fintech -d fintech -f $$file; \
	done

# Lint code
lint:
	@$(foreach service,$(SERVICES), \
		echo "Linting $(service)..." && \
		cd services/$(service) && golangci-lint run ./... && cd ../.. || exit 1; \
	)

# Clean
clean:
	@echo "Cleaning up..."
	docker-compose down -v
	@$(foreach service,$(SERVICES), \
		rm -f services/$(service)/coverage.* ; \
	)
