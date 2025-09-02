# Race Condition Test Script for UKM Slot Reservation (PowerShell)
# This script tests the slot reservation system using real authenticated sessions

Write-Host "🚀 UKM Slot Reservation Race Condition Test" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green
Write-Host ""

# Configuration
$API_BASE = "http://localhost:8080"
$UKM_ID = "your-ukm-id-here"  # Replace with actual UKM ID
$NUM_CONCURRENT = 5

Write-Host "📋 Configuration:" -ForegroundColor Yellow
Write-Host "  API Base: $API_BASE"
Write-Host "  UKM ID: $UKM_ID"
Write-Host "  Concurrent Requests: $NUM_CONCURRENT"
Write-Host ""

Write-Host "⚠️  IMPORTANT: Before running this test:" -ForegroundColor Red
Write-Host "1. Login to your app in a browser to get a valid session"
Write-Host "2. Open browser developer tools → Application → Cookies"
Write-Host "3. Copy the session cookie value"
Write-Host "4. Update the SESSION_COOKIE variable below"
Write-Host ""

# You need to replace this with an actual session cookie from a logged-in user
$SESSION_COOKIE = "session=your-actual-session-cookie-here"

if ($SESSION_COOKIE -eq "session=your-actual-session-cookie-here") {
    Write-Host "❌ ERROR: Please update SESSION_COOKIE with a real session value" -ForegroundColor Red
    Write-Host "   1. Login to http://localhost:5173 in your browser"
    Write-Host "   2. Open DevTools → Application → Cookies"
    Write-Host "   3. Copy the session cookie value"
    Write-Host "   4. Update this script at line 20"
    exit 1
}

Write-Host "🧪 Testing with session: $($SESSION_COOKIE.Substring(0, [Math]::Min(30, $SESSION_COOKIE.Length)))..." -ForegroundColor Cyan
Write-Host ""

# Function to make reservation request
function Make-ReservationRequest {
    param([int]$UserId)
    
    $headers = @{
        "Content-Type" = "application/json"
        "Cookie" = $SESSION_COOKIE
    }
    
    $body = @{
        ukm_id = $UKM_ID
    } | ConvertTo-Json
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    
    try {
        $response = Invoke-RestMethod -Uri "$API_BASE/api/registrations/reserve" -Method POST -Headers $headers -Body $body
        $stopwatch.Stop()
        
        return @{
            UserId = $UserId
            Success = $true
            Response = $response
            StatusCode = 200
            Duration = $stopwatch.ElapsedMilliseconds
        }
    }
    catch {
        $stopwatch.Stop()
        $statusCode = 0
        $errorMessage = $_.Exception.Message
        
        if ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorMessage = $reader.ReadToEnd()
            $reader.Close()
        }
        
        return @{
            UserId = $UserId
            Success = $false
            Error = $errorMessage
            StatusCode = $statusCode
            Duration = $stopwatch.ElapsedMilliseconds
        }
    }
}

Write-Host "🏃‍♂️ Starting $NUM_CONCURRENT concurrent reservation requests..." -ForegroundColor Cyan
Write-Host ""

# Array to store jobs
$jobs = @()

# Start all requests simultaneously using PowerShell jobs
for ($i = 1; $i -le $NUM_CONCURRENT; $i++) {
    $job = Start-Job -ScriptBlock ${function:Make-ReservationRequest} -ArgumentList $i
    $jobs += $job
}

# Wait for all jobs to complete
$results = $jobs | Wait-Job | Receive-Job

# Clean up jobs
$jobs | Remove-Job

Write-Host "📊 Individual Results:" -ForegroundColor Yellow
Write-Host "--------------------------------------------------"

$successCount = 0
$failureCount = 0

foreach ($result in $results) {
    if ($result.Success) {
        Write-Host "✅ User $($result.UserId): SUCCESS (200) - Got reservation [$($result.Duration)ms]" -ForegroundColor Green
        $successCount++
    }
    else {
        Write-Host "❌ User $($result.UserId): FAILED ($($result.StatusCode)) - $($result.Error) [$($result.Duration)ms]" -ForegroundColor Red
        $failureCount++
    }
}

Write-Host "--------------------------------------------------"
Write-Host "📈 Summary:" -ForegroundColor Yellow
Write-Host "  Successful reservations: $successCount"
Write-Host "  Failed reservations: $failureCount"
Write-Host ""

# Race condition analysis
Write-Host "🔍 Race Condition Analysis:" -ForegroundColor Yellow

if ($successCount -gt 2) {
    Write-Host "⚠️  WARNING: More than expected successes ($successCount > 2)" -ForegroundColor Red
    Write-Host "   This might indicate a race condition issue!"
}
elseif ($successCount -eq 0) {
    Write-Host "ℹ️  INFO: No successful reservations" -ForegroundColor Blue
    Write-Host "   This could mean UKM is full or authentication failed"
    
    # Check if all failures are 401 (auth issues)
    $authFailures = ($results | Where-Object { $_.StatusCode -eq 401 }).Count
    if ($authFailures -eq $NUM_CONCURRENT) {
        Write-Host "   → All requests failed with 401: Authentication issue" -ForegroundColor Red
        Write-Host "   → Please check your session cookie" -ForegroundColor Red
    }
}
else {
    Write-Host "✅ GOOD: Expected number of successes ($successCount)" -ForegroundColor Green
    Write-Host "   Race condition prevention appears to be working!"
}

# Error analysis
if ($failureCount -gt 0) {
    Write-Host ""
    Write-Host "📋 Failure Analysis:" -ForegroundColor Yellow
    $errorCounts = @{}
    
    foreach ($result in $results | Where-Object { -not $_.Success }) {
        $errorKey = "$($result.StatusCode): $($result.Error)"
        if ($errorCounts.ContainsKey($errorKey)) {
            $errorCounts[$errorKey]++
        }
        else {
            $errorCounts[$errorKey] = 1
        }
    }
    
    foreach ($error in $errorCounts.Keys) {
        Write-Host "  - $error ($($errorCounts[$error]) users)"
    }
}

Write-Host ""
Write-Host "🎯 Next steps:" -ForegroundColor Cyan
Write-Host "1. Check database: SELECT * FROM slot_reservations WHERE ukm_id = '$UKM_ID';"
Write-Host "2. Verify UKM slots: SELECT current_slot, max_slot FROM ukms WHERE id = '$UKM_ID';"
Write-Host "3. Test reservation completion with one of the successful reservation IDs"

if ($successCount -gt 0) {
    Write-Host ""
    Write-Host "💡 Successful reservations can be completed by calling:" -ForegroundColor Green
    Write-Host "POST /api/registrations/with-reservation/{reservation_id}"
    Write-Host "This will increment the UKM's current_slot counter"
}
