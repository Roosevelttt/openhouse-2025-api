package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("🚀 Race Condition Test for UKM Slot Reservation")
	fmt.Println(strings.Repeat("=", 50))

	// Test configuration
	baseURL := "http://localhost:8080"
	testUkmID := "b8f9e5d0-1234-5678-9abc-def012345678" // Use an actual UKM ID from your database
	numUsers := 5

	fmt.Printf("📋 Test Configuration:\n")
	fmt.Printf("  - Server: %s\n", baseURL)
	fmt.Printf("  - UKM ID: %s\n", testUkmID)
	fmt.Printf("  - Concurrent Users: %d\n", numUsers)
	fmt.Printf("  - Expected: Only users with available slots should succeed\n\n")

	// Run the race condition test
	runConcurrentSlotReservationTest(baseURL, testUkmID, numUsers)
}

type TestResult struct {
	UserID     int
	Success    bool
	StatusCode int
	Response   string
	Duration   time.Duration
	Error      string
}

func runConcurrentSlotReservationTest(baseURL, ukmID string, numUsers int) {
	fmt.Printf("🏃‍♂️ Starting %d concurrent slot reservation attempts...\n\n", numUsers)

	var wg sync.WaitGroup
	results := make(chan TestResult, numUsers)
	startTime := time.Now()

	// Launch concurrent requests
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()
			result := attemptSlotReservation(baseURL, ukmID, userIndex)
			results <- result
		}(i)
	}

	// Wait for all to complete
	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime)

	// Analyze results
	analyzeTestResults(results, totalDuration)
}

func attemptSlotReservation(baseURL, ukmID string, userIndex int) TestResult {
	userStartTime := time.Now()

	result := TestResult{
		UserID: userIndex,
	}

	// Create request payload
	payload := map[string]string{
		"ukm_id": ukmID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		result.Error = fmt.Sprintf("JSON marshal error: %v", err)
		result.Duration = time.Since(userStartTime)
		return result
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL+"/api/registrations/reserve", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Error = fmt.Sprintf("Request creation error: %v", err)
		result.Duration = time.Since(userStartTime)
		return result
	}

	req.Header.Set("Content-Type", "application/json")

	// Add mock session (you'll need to modify this based on your auth system)
	// For testing, you might need to manually set cookies or use a test user session
	req.Header.Set("Authorization", fmt.Sprintf("Bearer test-user-%d", userIndex))

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP request error: %v", err)
		result.Duration = time.Since(userStartTime)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Duration = time.Since(userStartTime)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Response read error: %v", err)
		return result
	}

	result.Response = string(body)

	// Check if successful
	if resp.StatusCode == http.StatusOK {
		result.Success = true
	} else {
		// Parse error message
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errMsg, ok := errorResp["error"].(string); ok {
				result.Error = errMsg
			}
		}
	}

	return result
}

func analyzeTestResults(results <-chan TestResult, totalDuration time.Duration) {
	var allResults []TestResult
	var successCount, failureCount int

	fmt.Println("📊 Individual Results:")
	fmt.Println(strings.Repeat("-", 50))

	for result := range results {
		allResults = append(allResults, result)

		status := "❌ FAILED"
		message := result.Error
		if result.Success {
			status = "✅ SUCCESS"
			message = "Slot reserved successfully"
			successCount++
		} else {
			failureCount++
		}

		fmt.Printf("User %d: %s (%d) - %s [%.2fms]\n",
			result.UserID,
			status,
			result.StatusCode,
			message,
			float64(result.Duration.Nanoseconds())/1e6)
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("  Total requests: %d\n", len(allResults))
	fmt.Printf("  Successful: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failureCount)
	fmt.Printf("  Total duration: %.2fms\n", float64(totalDuration.Nanoseconds())/1e6)

	// Race condition analysis
	fmt.Printf("\n🔍 Race Condition Analysis:\n")

	if failureCount > 0 {
		fmt.Printf("✅ GOOD: %d users were properly rejected\n", failureCount)

		// Analyze failure reasons
		errorCounts := make(map[string]int)
		for _, result := range allResults {
			if !result.Success && result.Error != "" {
				errorCounts[result.Error]++
			}
		}

		fmt.Printf("📋 Rejection reasons:\n")
		for reason, count := range errorCounts {
			fmt.Printf("  - \"%s\": %d users\n", reason, count)
		}
	}

	if successCount > 0 {
		fmt.Printf("✅ SUCCESS: %d users got reservations\n", successCount)
	}

	fmt.Printf("\n💡 Expected behavior:\n")
	fmt.Printf("  - If UKM has available slots: Some users should succeed\n")
	fmt.Printf("  - If UKM is full: All users should fail with 'no slots available'\n")
	fmt.Printf("  - No race condition: Success count should match available slots\n")

	fmt.Printf("\n🎯 Test completed! Check your database to verify:\n")
	fmt.Printf("  1. slot_reservations table has correct number of active reservations\n")
	fmt.Printf("  2. UKM current_slot hasn't been incremented yet (happens on completion)\n")
	fmt.Printf("  3. No duplicate reservations for the same user\n")
}
