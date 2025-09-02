#!/bin/bash

# Race Condition Test Script for UKM Slot Reservation
# This script tests the slot reservation system using real authenticated sessions

echo "🚀 UKM Slot Reservation Race Condition Test"
echo "============================================="
echo ""

# Configuration
API_BASE="http://localhost:8080"
UKM_ID="your-ukm-id-here"  # Replace with actual UKM ID
NUM_CONCURRENT=5

echo "📋 Configuration:"
echo "  API Base: $API_BASE"
echo "  UKM ID: $UKM_ID"
echo "  Concurrent Requests: $NUM_CONCURRENT"
echo ""

echo "⚠️  IMPORTANT: Before running this test:"
echo "1. Login to your app in a browser to get a valid session"
echo "2. Open browser developer tools → Application → Cookies"
echo "3. Copy the session cookie value"
echo "4. Replace SESSION_COOKIE below with your actual session cookie"
echo ""

# You need to replace this with an actual session cookie from a logged-in user
SESSION_COOKIE="session=your-actual-session-cookie-here"

if [ "$SESSION_COOKIE" = "session=your-actual-session-cookie-here" ]; then
    echo "❌ ERROR: Please update SESSION_COOKIE with a real session value"
    echo "   1. Login to http://localhost:5173 in your browser"
    echo "   2. Open DevTools → Application → Cookies"
    echo "   3. Copy the session cookie value"
    echo "   4. Update this script"
    exit 1
fi

echo "🧪 Testing with session: ${SESSION_COOKIE:0:30}..."
echo ""

# Function to make reservation request
make_reservation_request() {
    local user_id=$1
    local output_file="/tmp/race_test_user_${user_id}.txt"
    
    curl -s -w "User $user_id: %{http_code} %{time_total}s\n" \
         -H "Content-Type: application/json" \
         -H "Cookie: $SESSION_COOKIE" \
         -d "{\"ukm_id\": \"$UKM_ID\"}" \
         -X POST \
         "$API_BASE/api/registrations/reserve" \
         -o "$output_file" &
    
    echo $! # Return process ID
}

echo "🏃‍♂️ Starting $NUM_CONCURRENT concurrent reservation requests..."
echo ""

# Array to store process IDs
pids=()

# Start all requests simultaneously
for i in $(seq 1 $NUM_CONCURRENT); do
    pid=$(make_reservation_request $i)
    pids+=($pid)
done

# Wait for all requests to complete
for pid in "${pids[@]}"; do
    wait $pid
done

echo ""
echo "📊 Analyzing results..."
echo ""

# Analyze response files
success_count=0
failure_count=0

for i in $(seq 1 $NUM_CONCURRENT); do
    output_file="/tmp/race_test_user_${i}.txt"
    if [ -f "$output_file" ]; then
        response=$(cat "$output_file")
        if echo "$response" | grep -q "reservation_id"; then
            echo "✅ User $i: SUCCESS - Got reservation"
            success_count=$((success_count + 1))
        else
            echo "❌ User $i: FAILED - $response"
            failure_count=$((failure_count + 1))
        fi
        rm "$output_file"
    fi
done

echo ""
echo "📈 Summary:"
echo "  Successful reservations: $success_count"
echo "  Failed reservations: $failure_count"
echo ""

# Race condition analysis
if [ $success_count -gt 2 ]; then
    echo "⚠️  WARNING: More than expected successes ($success_count > 2)"
    echo "   This might indicate a race condition issue!"
elif [ $success_count -eq 0 ]; then
    echo "ℹ️  INFO: No successful reservations"
    echo "   This could mean UKM is full or authentication failed"
else
    echo "✅ GOOD: Expected number of successes ($success_count)"
    echo "   Race condition prevention appears to be working!"
fi

echo ""
echo "🔍 Next steps:"
echo "1. Check database: SELECT * FROM slot_reservations WHERE ukm_id = '$UKM_ID';"
echo "2. Verify UKM slots: SELECT current_slot, max_slot FROM ukms WHERE id = '$UKM_ID';"
echo "3. Test reservation completion with one of the successful reservation IDs"
