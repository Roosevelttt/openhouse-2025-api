# Race Condition Testing Guide

This guide helps you test the slot reservation system to ensure it properly prevents race conditions during UKM registration.

## Overview

The slot reservation system prevents race conditions by:
1. Using database transactions with `SELECT FOR UPDATE`
2. Checking available slots atomically (registered + reserved vs quota)
3. Creating time-limited reservations (10 minutes)
4. Converting reservations to registrations atomically

## Test Setup

### 1. Prepare Test Environment

```sql
-- Set up a UKM with limited slots for testing
UPDATE ukms SET max_slot = 100, current_slot = 98 
WHERE id = 'your-test-ukm-id';

-- This leaves only 2 slots available for testing
```

### 2. Ensure Server is Running

```bash
cd openhouse-2025-api
make run-server
```

## Running the Race Condition Test

### Method 1: Automated Test Runner

```bash
# Run the automated race condition test
cd openhouse-2025-api
go run cmd/race_test/main.go
```

This will:
- Send 5 concurrent slot reservation requests
- Measure response times
- Analyze success/failure patterns
- Check for race condition issues

### Method 2: Manual Testing with Multiple Browser Tabs

1. Open 5+ browser tabs to: `http://localhost:5173`
2. Login with different test users in each tab
3. Navigate to the same UKM registration page
4. Fill out the registration forms
5. Click "Reserve Slot & Register" simultaneously in all tabs
6. Observe results

### Method 3: Database Query Verification

Run these queries before and after testing:

```sql
-- Before test: Check initial state
SELECT name, current_slot, max_slot, 
       (max_slot - COALESCE(current_slot, 0)) as available_slots
FROM ukms WHERE id = 'your-test-ukm-id';

-- After test: Check reservations
SELECT COUNT(*) as active_reservations 
FROM slot_reservations 
WHERE ukm_id = 'your-test-ukm-id' AND expires_at > NOW();

-- Check for duplicates (should be 0)
SELECT nrp, COUNT(*) as reservation_count
FROM slot_reservations 
WHERE ukm_id = 'your-test-ukm-id' AND expires_at > NOW()
GROUP BY nrp
HAVING COUNT(*) > 1;
```

## Expected Results

### ✅ Correct Behavior (No Race Condition)

- **Only N users succeed** where N = available slots
- **Remaining users get "no slots available"** error
- **No duplicate reservations** for the same user
- **Database consistency**: reservations count ≤ available slots
- **Fast response times** (< 100ms typically)

### ❌ Race Condition Issues

- **More users succeed than available slots**
- **Database inconsistency**: reservations > available slots
- **Duplicate reservations** for the same user
- **Inconsistent error messages**

## Test Scenarios

### Scenario 1: High Contention (10 users, 1 slot)
```sql
UPDATE ukms SET max_slot = 100, current_slot = 99 WHERE id = 'test-ukm';
```
Expected: 1 success, 9 failures

### Scenario 2: Medium Contention (5 users, 2 slots)
```sql
UPDATE ukms SET max_slot = 100, current_slot = 98 WHERE id = 'test-ukm';
```
Expected: 2 successes, 3 failures

### Scenario 3: No Slots Available (5 users, 0 slots)
```sql
UPDATE ukms SET max_slot = 100, current_slot = 100 WHERE id = 'test-ukm';
```
Expected: 0 successes, 5 failures with "no slots available"

## Troubleshooting

### Issue: Authentication Errors
```
Error: {"error":"User not logged in"}
```
**Solution**: Ensure users are properly authenticated. Check session middleware.

### Issue: UKM Not Found
```
Error: {"error":"UKM not found"}
```
**Solution**: Verify the UKM ID exists in database and the `max_slot` column is not NULL.

### Issue: Database Connection Errors
**Solution**: Ensure MySQL is running and database connections are available.

## Performance Testing

Monitor these metrics during testing:

1. **Response Times**: Should be < 100ms for reservation requests
2. **Database Connections**: Monitor active connections
3. **Error Rates**: Should be predictable based on available slots
4. **Concurrent Transactions**: Check for deadlocks or timeouts

## Cleanup After Testing

```sql
-- Clean up test reservations
DELETE FROM slot_reservations WHERE ukm_id = 'your-test-ukm-id';

-- Reset UKM slots for next test
UPDATE ukms SET current_slot = 98 WHERE id = 'your-test-ukm-id';
```

## Advanced Testing

### Load Testing with Artillery

Create `race-test.yml`:
```yaml
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 10
      arrivalRate: 5
scenarios:
  - name: "Concurrent slot reservation"
    requests:
      - post:
          url: "/api/registrations/reserve"
          headers:
            Content-Type: "application/json"
          json:
            ukm_id: "your-test-ukm-id"
```

Run: `artillery run race-test.yml`

### Stress Testing

Test with varying loads:
- 10 concurrent users
- 50 concurrent users  
- 100 concurrent users

Monitor for:
- Response time degradation
- Error rate increases
- Database performance issues

## Verification Checklist

- [ ] Only correct number of users get reservations
- [ ] All excess users get proper error messages
- [ ] No duplicate reservations in database
- [ ] Reservation expiry times are set correctly (10 minutes)
- [ ] Database transactions complete successfully
- [ ] No deadlocks or race conditions detected
- [ ] Performance remains acceptable under load
