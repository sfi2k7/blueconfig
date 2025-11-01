#!/bin/bash

# Test script for BlueConfig Admin API
# Start the server first: ./configadmin

BASE_URL="http://localhost:8213"
API_URL="$BASE_URL/api"

echo "=================================="
echo "BlueConfig Admin API Test"
echo "=================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS=0
PASSED=0
FAILED=0

# Helper function to test API
test_api() {
    TESTS=$((TESTS + 1))
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    echo -e "${YELLOW}Test $TESTS: $description${NC}"
    echo "  $method $endpoint"

    if [ -z "$data" ]; then
        response=$(curl -s -X $method "$API_URL$endpoint")
    else
        response=$(curl -s -X $method -H "Content-Type: application/json" -d "$data" "$API_URL$endpoint")
    fi

    echo "  Response: $response"

    # Check if success field is true
    if echo "$response" | grep -q '"success":true'; then
        echo -e "  ${GREEN}✓ PASSED${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "  ${RED}✗ FAILED${NC}"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

# Wait for server to be ready
echo "Checking if server is running..."
if ! curl -s "$BASE_URL" > /dev/null 2>&1; then
    echo -e "${RED}Error: Server is not running on $BASE_URL${NC}"
    echo "Please start the server first: ./configadmin"
    exit 1
fi
echo -e "${GREEN}Server is running!${NC}"
echo ""

# Test 1: Get root node
test_api "GET" "/node/" "" "Get root node children"

# Test 2: Create a test path
test_api "POST" "/node/root/users" "" "Create users node"

# Test 3: Create another node
test_api "POST" "/node/root/users/john" "" "Create john user node"

# Test 4: Set a property
test_api "POST" "/properties/root/users/john" '{"key":"name","value":"John Doe"}' "Set name property"

# Test 5: Set another property
test_api "POST" "/properties/root/users/john" '{"key":"email","value":"john@example.com"}' "Set email property"

# Test 6: Set special __title property
test_api "POST" "/properties/root/users/john" '{"key":"__title","value":"John D."}' "Set __title property"

# Test 7: Set __color property
test_api "POST" "/properties/root/users/john" '{"key":"__color","value":"#3498db"}' "Set __color property"

# Test 8: Set __icon property
test_api "POST" "/properties/root/users/john" '{"key":"__icon","value":"fa-user"}' "Set __icon property"

# Test 9: Get all properties
test_api "GET" "/properties/root/users/john" "" "Get all properties for john"

# Test 10: Get children of users node
test_api "GET" "/node/root/users" "" "Get users node children"

# Test 11: Create another user
test_api "POST" "/node/root/users/jane" "" "Create jane user node"

# Test 12: Set properties for jane
test_api "POST" "/properties/root/users/jane" '{"key":"name","value":"Jane Smith"}' "Set jane name"
test_api "POST" "/properties/root/users/jane" '{"key":"__title","value":"Jane S."}' "Set jane __title"

# Test 13: Get users children again
test_api "GET" "/node/root/users" "" "Get users children (should have john and jane)"

# Test 14: Delete a property
test_api "DELETE" "/properties/root/users/john?key=email" "" "Delete email property"

# Test 15: Get john's properties (email should be gone)
test_api "GET" "/properties/root/users/john" "" "Get john properties after deletion"

# Test 16: Create nested structure
test_api "POST" "/node/root/config" "" "Create config node"
test_api "POST" "/node/root/config/database" "" "Create database config node"
test_api "POST" "/properties/root/config/database" '{"key":"host","value":"localhost"}' "Set database host"
test_api "POST" "/properties/root/config/database" '{"key":"port","value":"5432"}' "Set database port"

# Test 17: Get nested properties
test_api "GET" "/properties/root/config/database" "" "Get database config properties"

# Test 18: Create a node with JSON data
test_api "POST" "/properties/root/users/john" '{"key":"metadata","value":"{\"created\":\"2024-01-01\",\"updated\":\"2024-01-15\"}"}' "Set JSON metadata"

# Test 19: Create a node with markdown content
test_api "POST" "/properties/root/users/john" '{"key":"bio","value":"# John Doe\\n\\nSoftware Engineer at Example Corp.\\n\\n## Skills\\n- Go\\n- React\\n- Docker"}' "Set markdown bio"

# Summary
echo "=================================="
echo "Test Summary"
echo "=================================="
echo -e "Total Tests: $TESTS"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    exit 1
fi
