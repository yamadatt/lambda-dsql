#!/bin/bash

# Simple local test script for Aurora DSQL Lambda
# Usage: ./test-local.sh

echo "🚀 Aurora DSQL Lambda Local Test"
echo "================================="

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Check if template.yaml exists
if [ ! -f "template.yaml" ]; then
    echo "❌ Error: template.yaml not found. Run this script from the project root."
    exit 1
fi

echo "✅ Prerequisites check passed"
echo ""

# Build the function
echo "🔨 Building Lambda function with SAM..."
if sam build --use-container --cached; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

echo ""

# Test the function
echo "🧪 Testing Lambda function..."
echo "Executing: sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1"
echo ""

RESULT=$(sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ Lambda function executed successfully"
    echo ""
    echo "📊 Response:"
    echo "============"
    echo "$RESULT" | jq '.body | fromjson' 2>/dev/null || echo "$RESULT"
    echo ""

    # Extract count for summary
    COUNT=$(echo "$RESULT" | jq '.body | fromjson | .count' 2>/dev/null)
    if [ "$COUNT" != "null" ] && [ "$COUNT" != "" ]; then
        echo "📈 Summary: Retrieved $COUNT button click records"
    fi
else
    echo "❌ Lambda function execution failed"
    echo ""
    echo "🔍 Debug output:"
    sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1
    exit 1
fi

echo ""
echo "✨ Test completed successfully!"
echo ""
echo "💡 Next steps:"
echo "   • Start local API: sam local start-api --region ap-northeast-1 --port 3000"
echo "   • Test API: curl -X GET http://localhost:3000/version | jq '.'"
echo "   • Deploy to AWS: make deploy"