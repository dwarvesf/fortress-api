#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-"http://localhost:8080"}
ENV=${ENV:-"local"}
API_TOKEN=${API_TOKEN:-"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3OTQzMDMyMTMsImlkIjoiOWQ3MGE1NTMtMmQ1ZC00YTAxLWE3ZWQtZmY4OThhZjgyZjFhIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81NTQ4MzU5MzY2Njc4Nzc2MDgucG5nIiwiZW1haWwiOiJxdWFuZ0BkLmZvdW5kYXRpb24ifQ.U9VIIt8l1durlc3nuavEyvqeFkxeik6n8NR5O7FZ25M"}

use_auth=true
if [[ -z "$API_TOKEN" && "$ENV" != "prod" ]]; then
  use_auth=false
fi

DESC="Consulting services"
TOTAL=1500
INVOICE_DATE=$(date +%Y-%m-%d)
DUE_DATE=$(date -v+30d +%Y-%m-%d 2>/dev/null || date -d "+30 days" +%Y-%m-%d)
MONTH=$(date +%-m)
YEAR=$(date +%Y)

payload=$(cat <<JSON
{
  "projectID": "482435ca-afc7-41c8-8082-7f26793ca2b3",
  "bankID": "e7c056c6-cced-4065-836b-bd4f3052fc3b",
  "sentBy": "9d70a553-2d5d-4a01-a7ed-ff898af82f1a",
  "invoiceMonth": $MONTH,
  "invoiceYear": $YEAR,
  "invoiceDate": "$INVOICE_DATE",
  "dueDate": "$DUE_DATE",
  "email": "quang@d.foundation",
  "description": "$DESC",
  "total": $TOTAL,
  "lineItems": [
    {"description": "$DESC", "quantity": 1, "unitCost": $TOTAL, "discount": 0, "cost": $TOTAL, "isExternal": false}
  ]
}
JSON
)

args=("-sS" "-X" "POST" "$BASE_URL/api/v1/invoices/send" "-H" "Content-Type: application/json" "-d" "$payload")
if $use_auth; then
  args+=("-H" "Authorization: ApiKey $API_TOKEN")
fi

curl "${args[@]}" | jq
