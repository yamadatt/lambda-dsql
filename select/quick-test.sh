#!/bin/bash

echo "Aurora DSQL Lambda Quick Test - Button Clicks"
echo "============================================="
echo ""

# ローカルでLambda関数を直接実行
echo "1. Testing Lambda function locally..."
echo "-------------------------------------"
sam local invoke DSQLVersionFunction \
  --event events/test-event.json \
  --region ap-northeast-1 \
  2>/dev/null | jq '.body | fromjson' | jq '.'

echo ""
echo "2. Testing with simple curl command..."
echo "--------------------------------------"
echo "Run the following command after starting local API:"
echo ""
echo "# Start local API (in another terminal):"
echo "sam local start-api --region ap-northeast-1 --port 3000"
echo ""
echo "# Test with curl:"
echo "curl -X GET http://localhost:3000/version | jq '.'"
echo ""
echo "Expected response:"
echo '{'
echo '  "button_clicks": ['
echo '    {'
echo '      "id": 1,'
echo '      "action": "record",'
echo '      "timestamp": "2025-09-18T22:56:03Z",'
echo '      "ip_address": "192.168.1.1",'
echo '      "user_agent": "Mozilla/5.0 (Test Browser)"'
echo '    },'
echo '    ...'
echo '  ],'
echo '  "count": 3,'
echo '  "message": "Successfully retrieved button clicks",'
echo '  "timestamp": "2025-01-19T..."'
echo '}'