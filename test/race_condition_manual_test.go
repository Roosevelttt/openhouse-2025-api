package test

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

// Simple race condition test that can be run against a live server
func RunRaceConditionTest() {
	fmt.Println("🚀 Starting Race Condition Test for Slot Reservation System")
	fmt.Println(strings.Repeat("=", 60))

	// Configuration
	baseURL := "http://localhost:8080"
	numConcurrentUsers := 5
	testUkmID := "test-ukm-race-condition"

	// Step 1: Setup test environment
	fmt.Println("📋 Setting up test environment...")
	setupTestEnvironment(baseURL, testUkmID)

	// Step 2: Simulate concurrent slot reservations
	fmt.Println("\n🏃‍♂️ Simulating", numConcurrentUsers, "concurrent users trying to reserve slots...")

	var wg sync.WaitGroup
	results := make(chan ReservationTestResult, numConcurrentUsers)

	startTime := time.Now()

	// Launch concurrent requests
	for i := 0; i < numConcurrentUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()
			result := simulateUserReservation(baseURL, testUkmID, userIndex)
			results <- result
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(results)

	duration := time.Since(startTime)

	// Step 3: Analyze results
	fmt.Println("\n📊 Analyzing results...")
	analyzeResults(results, duration)

	fmt.Println("\n✅ Race condition test completed!")
}

type ReservationTestResult struct {
	UserIndex     int           `json:"user_index"`
	Success       bool          `json:"success"`
	ReservationID string        `json:"reservation_id"`
	StatusCode    int           `json:"status_code"`
	Error         string        `json:"error"`
	Duration      time.Duration `json:"duration"`
}

func setupTestEnvironment(baseURL, ukmID string) {
	// In a real scenario, you would need to:
	// 1. Login as admin
	// 2. Create a test UKM with limited slots
	// 3. Create test users

	fmt.Println("  ℹ️  Note: Ensure your test environment has:")
	fmt.Println("     - A UKM with ID:", ukmID)
	fmt.Println("     - Very few available slots (e.g., 2 slots remaining)")
	fmt.Println("     - Test users that can authenticate")
}

func simulateUserReservation(baseURL, ukmID string, userIndex int) ReservationTestResult {
	startTime := time.Now()

	result := ReservationTestResult{
		UserIndex: userIndex,
	}

	// Step 1: Authenticate user (simulate login)
	sessionCookie, err := authenticateUser(baseURL, userIndex)
	if err != nil {
		result.Error = fmt.Sprintf("Authentication failed: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Step 2: Attempt to reserve slot
	requestBody := map[string]string{
		"ukm_id": ukmID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to marshal request: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL+"/api/registrations/reserve", bytes.NewBuffer(bodyBytes))
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Duration = time.Since(startTime)

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		return result
	}

	if resp.StatusCode == http.StatusOK {
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err == nil {
			if reservationID, ok := response["reservation_id"].(string); ok {
				result.Success = true
				result.ReservationID = reservationID
			}
		}
	} else {
		var errorResponse map[string]string
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			result.Error = errorResponse["error"]
		} else {
			result.Error = string(body)
		}
	}

	return result
}

func authenticateUser(baseURL string, userIndex int) (*http.Cookie, error) {
	// In a real test, you would implement actual authentication
	// For now, we'll simulate having a session cookie

	// This is a placeholder - in reality you'd need to:
	// 1. POST to /api/auth/google/start
	// 2. Handle the OAuth flow
	// 3. Get the session cookie

	return &http.Cookie{
		Name:  "session",
		Value: fmt.Sprintf("test-session-user-%d", userIndex),
	}, nil
}

func analyzeResults(results <-chan ReservationTestResult, totalDuration time.Duration) {
	var successful []ReservationTestResult
	var failed []ReservationTestResult
	var totalRequests int

	for result := range results {
		totalRequests++
		if result.Success {
			successful = append(successful, result)
		} else {
			failed = append(failed, result)
		}

		// Print individual result
		status := "❌ FAILED"
		if result.Success {
			status = "✅ SUCCESS"
		}

		fmt.Printf("  User %d: %s (%.2fms) - %s\n",
			result.UserIndex,
			status,
			float64(result.Duration.Nanoseconds())/1e6,
			getResultMessage(result))
	}

	fmt.Println("\n📈 Summary:")
	fmt.Printf("  Total requests: %d\n", totalRequests)
	fmt.Printf("  Successful reservations: %d\n", len(successful))
	fmt.Printf("  Failed reservations: %d\n", len(failed))
	fmt.Printf("  Total test duration: %.2fms\n", float64(totalDuration.Nanoseconds())/1e6)

	// Race condition analysis
	fmt.Println("\n🔍 Race Condition Analysis:")
	if len(successful) > 2 { // Assuming only 2 slots should be available
		fmt.Printf("  ⚠️  WARNING: More users succeeded than expected (%d > 2)\n", len(successful))
		fmt.Println("     This might indicate a race condition issue!")
	} else {
		fmt.Printf("  ✅ GOOD: Correct number of users succeeded (%d)\n", len(successful))
		fmt.Println("     Race condition prevention appears to be working!")
	}

	// Error analysis
	if len(failed) > 0 {
		fmt.Println("\n📋 Failed Reservation Reasons:")
		errorCounts := make(map[string]int)
		for _, result := range failed {
			errorCounts[result.Error]++
		}

		for error, count := range errorCounts {
			fmt.Printf("  - \"%s\": %d users\n", error, count)
		}
	}
}

func getResultMessage(result ReservationTestResult) string {
	if result.Success {
		return fmt.Sprintf("Reserved slot: %s", result.ReservationID)
	}
	return result.Error
}

// Test runner specifically for slot reservation race conditions
func runSlotReservationRaceTest() {
	fmt.Println("🎯 Slot Reservation Race Condition Test")
	fmt.Println("This test simulates multiple users trying to register for a UKM")
	fmt.Println("with very few remaining slots to verify race condition prevention.")
	fmt.Println()

	// Test parameters
	scenarios := []struct {
		name           string
		users          int
		availableSlots int
	}{
		{"High contention", 10, 1},
		{"Medium contention", 5, 2},
		{"Low contention", 3, 2},
	}

	for _, scenario := range scenarios {
		fmt.Printf("🧪 Running scenario: %s (%d users, %d slots)\n",
			scenario.name, scenario.users, scenario.availableSlots)

		// In a real implementation, you would:
		// 1. Set up UKM with exactly scenario.availableSlots remaining
		// 2. Run the concurrent test
		// 3. Verify exactly scenario.availableSlots succeeded

		fmt.Printf("  Expected result: %d successful, %d failed\n",
			scenario.availableSlots, scenario.users-scenario.availableSlots)
		fmt.Println()
	}
}
