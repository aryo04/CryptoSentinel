#!/bin/bash

# Teneo Agent SDK - Test Runner
# This script runs all tests in the organized test structure

echo "🧪 Teneo Agent SDK Test Suite"
echo "============================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests in a directory
run_test_suite() {
    local test_dir=$1
    local test_name=$2
    
    echo -e "\n${BLUE}📋 Running $test_name Tests${NC}"
    echo "----------------------------------------"
    
    if [ -d "$test_dir" ]; then
        cd "$test_dir"
        
        # Check if go.mod exists
        if [ -f "go.mod" ]; then
            # Run tests with verbose output
            if go test -v ./...; then
                echo -e "${GREEN}✅ $test_name tests passed${NC}"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo -e "${RED}❌ $test_name tests failed${NC}"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            echo -e "${YELLOW}⚠️ No go.mod found in $test_dir${NC}"
        fi
        
        cd - > /dev/null
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    else
        echo -e "${YELLOW}⚠️ Test directory $test_dir not found${NC}"
    fi
}

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Run SDK core tests
echo -e "\n${BLUE}🔧 Running SDK Core Tests${NC}"
echo "----------------------------------------"
cd ..
if go test -v ./pkg/...; then
    echo -e "${GREEN}✅ SDK core tests passed${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ SDK core tests failed${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))
cd "$SCRIPT_DIR"

# Run organized test suites
run_test_suite "unit" "Unit"
run_test_suite "integration" "Integration"

# Check if e2e tests exist
if [ -d "e2e" ] && [ "$(ls -A e2e)" ]; then
    run_test_suite "e2e" "End-to-End"
else
    echo -e "\n${YELLOW}⚠️ No E2E tests found (this is normal for now)${NC}"
fi

# Run example tests
echo -e "\n${BLUE}📋 Running Example Tests${NC}"
echo "----------------------------------------"
cd ../examples

# Test standardized messaging example
if [ -d "standardized-messaging" ]; then
    echo "Testing standardized-messaging example..."
    cd standardized-messaging
    if go run main.go > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Standardized messaging example works${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ Standardized messaging example failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    cd ..
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# Test agent naming example
if [ -d "agent-naming" ]; then
    echo "Testing agent-naming example..."
    cd agent-naming
    if timeout 10s go run main.go > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Agent naming example works${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ Agent naming example failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    cd ..
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

cd "$SCRIPT_DIR"

# Display final results
echo -e "\n${BLUE}📊 Test Results Summary${NC}"
echo "========================"
echo -e "Total Test Suites: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}💥 Some tests failed!${NC}"
    exit 1
fi
