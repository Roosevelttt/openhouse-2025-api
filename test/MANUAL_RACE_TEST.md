# Manual Race Condition Testing Guide

Since the automated tests require proper authentication, here's a simple manual testing approach:

## Quick Browser Test (Recommended)

### Setup
1. **Set UKM to near capacity:**
   ```sql
   UPDATE ukms SET max_slot = 100, current_slot = 98 WHERE id = 'your-ukm-id';
   -- This leaves only 2 slots available
   ```

2. **Open multiple browser windows:**
   - Open 5 separate browser windows (not tabs)
   - Login with different test users in each window
   - Navigate to the same UKM registration page

### Test Execution
1. **Fill forms simultaneously** in all 5 windows
2. **Click "Reserve Slot & Register" at the same time**
3. **Observe results**

### Expected Results ✅
- **2 users should succeed** (get reservation confirmations)
- **3 users should fail** with "No slots available" or similar error
- **No more than 2 active reservations** in database

### Check Database
```sql
-- Check reservations created
SELECT COUNT(*) as active_reservations 
FROM slot_reservations 
WHERE ukm_id = 'your-ukm-id' AND expires_at > NOW();

-- Should show exactly 2 reservations

-- Check for duplicates (should be 0)
SELECT nrp, COUNT(*) as count 
FROM slot_reservations 
WHERE ukm_id = 'your-ukm-id' AND expires_at > NOW() 
GROUP BY nrp 
HAVING COUNT(*) > 1;
```

## Automated Test with Real Session

### Step 1: Get Session Cookie
1. Login to your app: http://localhost:5173
2. Open Browser DevTools (F12)
3. Go to Application → Cookies
4. Copy the session cookie value

### Step 2: Run PowerShell Test
```powershell
# Edit the script first to add your session cookie
cd openhouse-2025-api/test
# Update race_test.ps1 line 20 with your session cookie
.\race_test.ps1
```

### Step 3: Run with Real UKM ID
```powershell
# First, find a real UKM ID from your database
mysql -u root openhouse_2025 -e "SELECT id, name, current_slot, max_slot FROM ukms LIMIT 5;"

# Update the UKM_ID in race_test.ps1
# Then run the test
```

## Test Scenarios

### Scenario 1: High Contention
```sql
-- 10 users, 1 slot
UPDATE ukms SET max_slot = 100, current_slot = 99 WHERE id = 'test-ukm';
```
**Expected:** 1 success, 9 failures

### Scenario 2: Medium Contention  
```sql
-- 5 users, 2 slots
UPDATE ukms SET max_slot = 100, current_slot = 98 WHERE id = 'test-ukm';
```
**Expected:** 2 successes, 3 failures

### Scenario 3: Full Capacity
```sql
-- 5 users, 0 slots
UPDATE ukms SET max_slot = 100, current_slot = 100 WHERE id = 'test-ukm';
```
**Expected:** 0 successes, 5 failures

## Debugging Common Issues

### Issue: All requests fail with 401
**Cause:** Authentication problem
**Solution:** 
- Ensure users are logged in
- Check session cookies
- Verify middleware is working

### Issue: More successes than available slots
**Cause:** Race condition not prevented
**Solution:**
- Check database transaction isolation
- Verify `SELECT FOR UPDATE` is working
- Check for deadlocks in MySQL logs

### Issue: UKM not found errors
**Cause:** Wrong UKM ID or missing max_slot
**Solution:**
- Use real UKM ID from database
- Ensure max_slot is not NULL

## Performance Monitoring

Watch for these metrics during testing:

```sql
-- Active database connections
SHOW PROCESSLIST;

-- Check for deadlocks
SHOW ENGINE INNODB STATUS;

-- Monitor reservation table
SELECT COUNT(*) FROM slot_reservations WHERE expires_at > NOW();
```

## Cleanup After Testing

```sql
-- Remove test reservations
DELETE FROM slot_reservations WHERE ukm_id = 'your-test-ukm-id';

-- Reset UKM slots
UPDATE ukms SET current_slot = 0 WHERE id = 'your-test-ukm-id';
```

## Success Indicators

✅ **Race condition prevented if:**
- Exactly N users succeed (where N = available slots)
- Remaining users get proper error messages  
- No duplicate reservations in database
- Response times remain fast (< 100ms)
- Database stays consistent

❌ **Race condition exists if:**
- More users succeed than slots available
- Database shows inconsistent state
- Duplicate reservations for same user
- Sporadic or inconsistent errors
