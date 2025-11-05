#!/bin/bash

# Test script for project endpoints
# Make sure the mock server is running on port 8080

BASE_URL="http://localhost:8080/v2"

echo "========================================="
echo "Testing Project Endpoints"
echo "========================================="

# Test 1: List all projects (should have default project)
echo ""
echo "1. List all projects:"
curl -s -X GET "$BASE_URL/projects" | jq .

# Test 2: Create a new project
echo ""
echo "2. Create a new project:"
PROJECT_RESPONSE=$(curl -s -X POST "$BASE_URL/projects" \
  -H "Content-Type: application/json" \
  -d '{
    "projectName": "test-project-terraform",
    "plan": "Standard"
  }')
echo $PROJECT_RESPONSE | jq .

PROJECT_ID=$(echo $PROJECT_RESPONSE | jq -r '.data')
echo "Created project ID: $PROJECT_ID"

# Test 3: Get project by ID
echo ""
echo "3. Get project by ID:"
curl -s -X GET "$BASE_URL/projects/$PROJECT_ID" | jq .

# Test 4: List all projects again (should now have 2)
echo ""
echo "4. List all projects (should have 2):"
curl -s -X GET "$BASE_URL/projects" | jq .

# Test 5: Upgrade project plan
echo ""
echo "5. Upgrade project plan to Enterprise:"
curl -s -X PATCH "$BASE_URL/projects/$PROJECT_ID/plan" \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "Enterprise"
  }' | jq .

# Test 6: Get project by ID to verify upgrade
echo ""
echo "6. Get project by ID (verify plan upgraded):"
curl -s -X GET "$BASE_URL/projects/$PROJECT_ID" | jq .

# Test 7: Create project with default plan
echo ""
echo "7. Create project with default plan:"
DEFAULT_PROJECT_RESPONSE=$(curl -s -X POST "$BASE_URL/projects" \
  -H "Content-Type: application/json" \
  -d '{
    "projectName": "test-project-default"
  }')
echo $DEFAULT_PROJECT_RESPONSE | jq .

DEFAULT_PROJECT_ID=$(echo $DEFAULT_PROJECT_RESPONSE | jq -r '.data')
echo ""
echo "8. Get default project (should have Enterprise plan):"
curl -s -X GET "$BASE_URL/projects/$DEFAULT_PROJECT_ID" | jq .

# Test 9: Delete project
echo ""
echo "9. Delete project:"
curl -s -X DELETE "$BASE_URL/projects/$PROJECT_ID" | jq .

# Test 10: Try to get deleted project (should fail)
echo ""
echo "10. Try to get deleted project (should return 404):"
curl -s -X GET "$BASE_URL/projects/$PROJECT_ID" | jq .

# Test 11: List all projects (should not include deleted project)
echo ""
echo "11. List all projects (should not include deleted project):"
curl -s -X GET "$BASE_URL/projects" | jq .

echo ""
echo "========================================="
echo "Project endpoint tests completed!"
echo "========================================="
