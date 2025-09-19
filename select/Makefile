.PHONY: build clean deploy test local-api install-deps

# Variables
STACK_NAME = dsql-version-stack
REGION = ap-northeast-1
S3_BUCKET = your-sam-deployment-bucket
GO_VERSION = 1.23.0

# Build Lambda function
build:
	@echo "Building Lambda function..."
	cd lambda && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	chmod +x bootstrap
	@echo "Build complete"

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	cd lambda && go mod download
	@echo "Dependencies installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f lambda/bootstrap
	rm -rf .aws-sam
	@echo "Clean complete"

# Build using SAM
sam-build: clean install-deps
	@echo "Building with SAM..."
	sam build --use-container
	@echo "SAM build complete"

# Build using SAM (no clean)
sam-build-fast:
	@echo "Building with SAM (using cache)..."
	sam build --use-container --cached
	@echo "SAM build complete"

# Deploy to AWS
deploy: sam-build
	@echo "Deploying to AWS..."
	sam deploy \
		--stack-name $(STACK_NAME) \
		--region $(REGION) \
		--capabilities CAPABILITY_IAM \
		--resolve-s3 \
		--no-confirm-changeset \
		--no-fail-on-empty-changeset
	@echo "Deployment complete"

# Start local API for testing
local-api: build
	@echo "Starting local API..."
	sam local start-api \
		--region $(REGION) \
		--port 3000

# Test the deployed API
test-deployed:
	@echo "Testing deployed API..."
	@API_ENDPOINT=$$(aws cloudformation describe-stacks \
		--stack-name $(STACK_NAME) \
		--region $(REGION) \
		--query "Stacks[0].Outputs[?OutputKey=='ApiEndpoint'].OutputValue" \
		--output text) && \
	echo "API Endpoint: $$API_ENDPOINT" && \
	curl -X GET "$$API_ENDPOINT"

# Test local API
test-local:
	@echo "Testing local API..."
	curl -X GET http://localhost:3000/version

# Create test event
create-test-event:
	@echo "Creating test event..."
	@mkdir -p events
	@echo '{\
		"httpMethod": "GET",\
		"path": "/version",\
		"headers": {\
			"Accept": "application/json"\
		},\
		"body": null\
	}' > events/test-event.json
	@echo "Test event created at events/test-event.json"

# Invoke Lambda function locally
invoke-local: build create-test-event
	@echo "Invoking Lambda function locally..."
	sam local invoke DSQLVersionFunction \
		--event events/test-event.json \
		--region $(REGION)

# Show stack outputs
show-outputs:
	@echo "Stack outputs:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_NAME) \
		--region $(REGION) \
		--query "Stacks[0].Outputs" \
		--output table

# Delete the stack
delete:
	@echo "Deleting stack $(STACK_NAME)..."
	sam delete \
		--stack-name $(STACK_NAME) \
		--region $(REGION) \
		--no-prompts
	@echo "Stack deleted"

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build Lambda function binary"
	@echo "  make install-deps   - Install Go dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make sam-build      - Build using SAM"
	@echo "  make deploy         - Deploy to AWS"
	@echo "  make local-api      - Start local API for testing"
	@echo "  make test-deployed  - Test deployed API"
	@echo "  make test-local     - Test local API"
	@echo "  make invoke-local   - Invoke Lambda locally"
	@echo "  make show-outputs   - Show stack outputs"
	@echo "  make delete         - Delete the stack"
	@echo "  make help           - Show this help message"