```bash
#!/bin/bash

BASE_URL="https://localhost:8443"

echo "======================================"
echo " Employee API - Day 2 Test Collection"
echo "======================================"

echo -e "\n=== 1. Health Check ==="

curl -k "$BASE_URL/health" \
  -H "Accept: application/json"

echo -e "\n\n=== 2. Create Employee ==="

CREATE_RESPONSE=$(curl -ks -X POST "$BASE_URL/api/v1/employees" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ali Ahmed",
    "email": "ali.ahmed@example.com",
    "department": "IT"
  }')

echo "$CREATE_RESPONSE"

EMPLOYEE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')

echo "Created Employee ID: $EMPLOYEE_ID"

echo -e "\n\n=== 3. Get Employees ==="

curl -k "$BASE_URL/api/v1/employees?page=1&page_size=10" \
  -H "Accept: application/json"

echo -e "\n\n=== 4. Get Employee By ID ==="

curl -k "$BASE_URL/api/v1/employees/$EMPLOYEE_ID" \
  -H "Accept: application/json"

echo -e "\n\n=== 5. Update Employee ==="

curl -k -X PUT "$BASE_URL/api/v1/employees/$EMPLOYEE_ID" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ali Ahmed Updated",
    "email": "ali.updated@example.com",
    "department": "HR"
  }'

echo -e "\n\n=== 6. Get Updated Employee ==="

curl -k "$BASE_URL/api/v1/employees/$EMPLOYEE_ID" \
  -H "Accept: application/json"

echo -e "\n\n=== 7. Delete Employee ==="

curl -k -X DELETE "$BASE_URL/api/v1/employees/$EMPLOYEE_ID" \
  -H "Accept: application/json"

echo -e "\n\n=== 8. OpenAPI Specification ==="

curl -k "$BASE_URL/openapi.json" \
  -H "Accept: application/json"

echo -e "\n\n======================================"
echo " Day 2 Test Collection Completed"
echo "======================================"
```

