-- Race Condition Test Verification Queries
-- Run these queries before and after the race condition test

-- 1. Check UKM status before test
SELECT 
    id,
    name,
    current_slot,
    max_slot,
    (max_slot - COALESCE(current_slot, 0)) as available_slots
FROM ukms 
WHERE id = 'your-ukm-id-here';

-- 2. Check active reservations
SELECT 
    reservation_id,
    nrp,
    ukm_id,
    expires_at,
    created_at,
    CASE 
        WHEN expires_at > NOW() THEN 'ACTIVE'
        ELSE 'EXPIRED'
    END as status
FROM slot_reservations 
WHERE ukm_id = 'your-ukm-id-here'
ORDER BY created_at;

-- 3. Count reservations by status
SELECT 
    CASE 
        WHEN expires_at > NOW() THEN 'ACTIVE'
        ELSE 'EXPIRED'
    END as reservation_status,
    COUNT(*) as count
FROM slot_reservations 
WHERE ukm_id = 'your-ukm-id-here'
GROUP BY reservation_status;

-- 4. Check completed registrations
SELECT 
    id,
    nrp,
    ukm_id,
    created_at
FROM detail_registrations 
WHERE ukm_id = 'your-ukm-id-here'
ORDER BY created_at;

-- 5. Verify no duplicate reservations per user
SELECT 
    nrp,
    COUNT(*) as reservation_count
FROM slot_reservations 
WHERE ukm_id = 'your-ukm-id-here' 
    AND expires_at > NOW()
GROUP BY nrp
HAVING COUNT(*) > 1;

-- 6. Clean up test data (run after test)
-- DELETE FROM slot_reservations WHERE ukm_id = 'your-ukm-id-here';
-- DELETE FROM detail_registrations WHERE ukm_id = 'your-ukm-id-here';

-- 7. Reset UKM slot count for testing (if needed)
-- UPDATE ukms SET current_slot = 98 WHERE id = 'your-ukm-id-here';

-- 8. Create test scenario: UKM with almost full capacity
-- UPDATE ukms SET max_slot = 100, current_slot = 98 WHERE id = 'your-ukm-id-here';
-- This leaves only 2 slots available for race condition testing
