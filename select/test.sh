#!/bin/bash

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================================"
echo "Aurora DSQL Lambda Test Script"
echo "================================================"

# Configuration
LOCAL_API_URL="http://localhost:3000/version"
STACK_NAME="dsql-version-stack"
REGION="ap-northeast-1"

# Function to test local API
test_local() {
    echo -e "\n${YELLOW}Testing Local API...${NC}"
    echo "URL: $LOCAL_API_URL"
    echo "----------------------------------------"

    response=$(curl -s -X GET "$LOCAL_API_URL" -H "Accept: application/json")
    status=$?

    if [ $status -eq 0 ]; then
        echo -e "${GREEN}✓ Request successful${NC}"
        echo "Response:"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ Request failed${NC}"
        echo "Error code: $status"
    fi
}

# Function to test deployed API
test_deployed() {
    echo -e "\n${YELLOW}Testing Deployed API...${NC}"
    echo "Stack: $STACK_NAME"
    echo "Region: $REGION"
    echo "----------------------------------------"

    # Get API endpoint from CloudFormation
    API_ENDPOINT=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query "Stacks[0].Outputs[?OutputKey=='ApiEndpoint'].OutputValue" \
        --output text 2>/dev/null)

    if [ -z "$API_ENDPOINT" ]; then
        echo -e "${RED}✗ Could not find API endpoint. Is the stack deployed?${NC}"
        return 1
    fi

    echo "API Endpoint: $API_ENDPOINT"
    echo "----------------------------------------"

    response=$(curl -s -X GET "$API_ENDPOINT" -H "Accept: application/json")
    status=$?

    if [ $status -eq 0 ]; then
        echo -e "${GREEN}✓ Request successful${NC}"
        echo "Response:"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ Request failed${NC}"
        echo "Error code: $status"
    fi
}

# Function to test with timing
test_with_timing() {
    local url=$1
    local name=$2

    echo -e "\n${YELLOW}Performance Test: $name${NC}"
    echo "----------------------------------------"

    for i in {1..3}; do
        echo -n "Request $i: "
        time_output=$(curl -s -w "\nTime: %{time_total}s" -X GET "$url" -H "Accept: application/json" -o /dev/null)
        echo "$time_output"
    done
}

# Function to test error handling
test_error_handling() {
    local url=$1
    local name=$2

    echo -e "\n${YELLOW}Error Handling Test: $name${NC}"
    echo "----------------------------------------"

    # Test with invalid method
    echo "Testing with POST method (should fail):"
    curl -s -X POST "$url" -H "Accept: application/json" | jq '.' 2>/dev/null
}

# Main menu
show_menu() {
    echo -e "\n${YELLOW}Select test option:${NC}"
    echo "1) Test Local API"
    echo "2) Test Deployed API"
    echo "3) Performance Test (Local)"
    echo "4) Performance Test (Deployed)"
    echo "5) Error Handling Test (Local)"
    echo "6) Run All Tests"
    echo "0) Exit"
    echo -n "Enter choice: "
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON formatting will be disabled.${NC}"
fi

# Main loop
while true; do
    show_menu
    read choice

    case $choice in
        1)
            test_local
            ;;
        2)
            test_deployed
            ;;
        3)
            test_with_timing "$LOCAL_API_URL" "Local API"
            ;;
        4)
            API_ENDPOINT=$(aws cloudformation describe-stacks \
                --stack-name "$STACK_NAME" \
                --region "$REGION" \
                --query "Stacks[0].Outputs[?OutputKey=='ApiEndpoint'].OutputValue" \
                --output text 2>/dev/null)
            if [ -n "$API_ENDPOINT" ]; then
                test_with_timing "$API_ENDPOINT" "Deployed API"
            else
                echo -e "${RED}✗ Could not find deployed API endpoint${NC}"
            fi
            ;;
        5)
            test_error_handling "$LOCAL_API_URL" "Local API"
            ;;
        6)
            test_local
            test_deployed
            ;;
        0)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid option${NC}"
            ;;
    esac
done