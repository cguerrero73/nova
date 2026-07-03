#!/usr/bin/env bash
# Integration test: design → publish → assign → resolve → render round-trip
# Requires: running backend with seeded DB, NOVA_TEST_DB_DSN set
#
# This script exercises the full formbuilder lifecycle:
# 1. Create a form (auto-creates default layout)
# 2. Save a draft with sections and fields
# 3. Publish the draft
# 4. Assign a layout to a role
# 5. Resolve the layout for that role
# 6. Verify the resolved JSON matches the published definition
#
# Usage: ./scripts/test-formbuilder-e2e.sh [BASE_URL]
# Default BASE_URL: http://localhost:3000/api/formbuilder

set -euo pipefail

BASE_URL="${1:-http://localhost:3000/api/formbuilder}"
FORM_KEY="e2e_test_$(date +%s)"
LAYOUT_NAME="default"
ROLE_NAME="admin"
COOKIE_JAR=$(mktemp)

cleanup() { rm -f "$COOKIE_JAR"; }
trap cleanup EXIT

echo "=== Form Builder E2E Test ==="
echo "Base URL: $BASE_URL"
echo "Form key: $FORM_KEY"
echo ""

# Helper: authenticated request (uses session cookie)
api() {
  local method="$1" path="$2" data="${3:-}"
  local args=(-s -w '\n%{http_code}' -b "$COOKIE_JAR" -c "$COOKIE_JAR"
              -H 'Content-Type: application/json')
  if [[ -n "$data" ]]; then
    args+=(-d "$data")
  fi
  local response
  response=$(curl "${args[@]}" -X "$method" "${BASE_URL}${path}")
  local http_code
  http_code=$(echo "$response" | tail -1)
  local body
  body=$(echo "$response" | sed '$d')
  if [[ "$http_code" -ge 400 ]]; then
    echo "FAIL: $method $path → $http_code"
    echo "$body" | jq . 2>/dev/null || echo "$body"
    return 1
  fi
  echo "$body"
}

# Step 0: Login (get session cookie)
echo "--- Step 0: Authenticate ---"
LOGIN_RESP=$(curl -s -c "$COOKIE_JAR" -X POST \
  "http://localhost:3000/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"user_code":"admin","password":"admin123"}')
echo "Login: $(echo "$LOGIN_RESP" | jq -r '.success' 2>/dev/null || echo 'raw response')"
echo ""

# Step 1: Create form
echo "--- Step 1: Create form ---"
CREATE_RESP=$(api POST "/forms" "{\"key\":\"$FORM_KEY\",\"name\":\"E2E Test Form\",\"description\":\"Integration test\"}")
echo "Created form: $(echo "$CREATE_RESP" | jq -r '.data.frm_key' 2>/dev/null)"
echo ""

# Step 2: Save draft with sections and fields
echo "--- Step 2: Save draft ---"
DRAFT_DEF=$(cat <<EOF
{
  "formKey": "$FORM_KEY",
  "layoutName": "$LAYOUT_NAME",
  "sections": [
    {
      "name": "personal_info",
      "title": "Personal Information",
      "order": 0,
      "fields": [
        {
          "type": "text",
          "name": "full_name",
          "ui": { "label": "Full Name", "placeholder": "Enter your name", "width": "full" },
          "validators": [{ "kind": "required" }]
        },
        {
          "type": "text",
          "name": "email",
          "ui": { "label": "Email", "width": "half" },
          "validators": [{ "kind": "required" }, { "kind": "email" }]
        },
        {
          "type": "number",
          "name": "age",
          "ui": { "label": "Age", "width": "half" },
          "validators": [{ "kind": "min", "value": 18 }, { "kind": "max", "value": 120 }]
        }
      ]
    },
    {
      "name": "preferences",
      "title": "Preferences",
      "order": 1,
      "fields": [
        {
          "type": "select",
          "name": "department",
          "ui": { "label": "Department", "width": "full" },
          "options": [
            { "label": "Engineering", "value": "eng" },
            { "label": "Sales", "value": "sales" },
            { "label": "HR", "value": "hr" }
          ],
          "validators": [{ "kind": "required" }]
        },
        {
          "type": "checkbox",
          "name": "agree_terms",
          "ui": { "label": "I agree to the terms", "width": "full" },
          "validators": [{ "kind": "required" }]
        }
      ]
    }
  ],
  "rules": [
    { "operator": "hiddenIf", "source": "agree_terms", "target": "department" }
  ]
}
EOF
)
SAVE_RESP=$(api PUT "/forms/$FORM_KEY/layouts/$LAYOUT_NAME/draft" "$DRAFT_DEF")
echo "Draft saved: $(echo "$SAVE_RESP" | jq -r '.data.sections | length' 2>/dev/null) sections"
echo ""

# Step 3: Publish
echo "--- Step 3: Publish ---"
PUBLISH_RESP=$(api POST "/forms/$FORM_KEY/layouts/$LAYOUT_NAME/publish" '{"description":"Initial version from E2E test"}')
echo "Published version: $(echo "$PUBLISH_RESP" | jq -r '.data.flv_version_number' 2>/dev/null)"
echo ""

# Step 4: Assign layout to role
echo "--- Step 4: Assign to role ---"
ASSIGN_RESP=$(api PUT "/forms/$FORM_KEY/assignments/$ROLE_NAME" "{\"layoutName\":\"$LAYOUT_NAME\"}")
echo "Assignment: $(echo "$ASSIGN_RESP" | jq -r '.data.fra_role_name' 2>/dev/null) → $(echo "$ASSIGN_RESP" | jq -r '.data.fra_layout_name' 2>/dev/null)"
echo ""

# Step 5: Resolve (GET the runtime form)
echo "--- Step 5: Resolve ---"
RESOLVE_RESP=$(api GET "/forms/$FORM_KEY")
RESOLVED_DEF=$(echo "$RESOLVE_RESP" | jq -c '.data' 2>/dev/null)
echo "Resolved layout: $(echo "$RESOLVED_DEF" | jq -r '.layoutName' 2>/dev/null)"
echo "Sections: $(echo "$RESOLVED_DEF" | jq -r '.sections | length' 2>/dev/null)"
echo "Fields in section 0: $(echo "$RESOLVED_DEF" | jq -r '.sections[0].fields | length' 2>/dev/null)"
echo "Rules: $(echo "$RESOLVED_DEF" | jq -r '.rules | length' 2>/dev/null)"
echo ""

# Step 6: Verify audit log
echo "--- Step 6: Audit log ---"
AUDIT_RESP=$(api GET "/forms/$FORM_KEY/audit")
echo "Audit entries: $(echo "$AUDIT_RESP" | jq -r '.data.total' 2>/dev/null)"
echo "Actions: $(echo "$AUDIT_RESP" | jq -r '[.data.items[].action] | join(", ")' 2>/dev/null)"
echo ""

# Step 7: Verify draft matches published
echo "--- Step 7: Verify round-trip integrity ---"
DRAFT_SECTIONS=$(echo "$DRAFT_DEF" | jq -c '.sections')
RESOLVED_SECTIONS=$(echo "$RESOLVED_DEF" | jq -c '.sections')
if [[ "$DRAFT_SECTIONS" == "$RESOLVED_SECTIONS" ]]; then
  echo "✅ PASS: Published definition matches draft"
else
  echo "❌ FAIL: Published definition does not match draft"
  exit 1
fi

# Verify rules preserved
DRAFT_RULES=$(echo "$DRAFT_DEF" | jq -c '.rules')
RESOLVED_RULES=$(echo "$RESOLVED_DEF" | jq -c '.rules')
if [[ "$DRAFT_RULES" == "$RESOLVED_RULES" ]]; then
  echo "✅ PASS: Cross-field rules preserved"
else
  echo "❌ FAIL: Cross-field rules not preserved"
  exit 1
fi

echo ""
echo "=== All E2E checks passed ==="
