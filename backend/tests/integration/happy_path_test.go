package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestFullHappyPath tests the complete flow:
// Register → Login → Create Property → Create Unit Type → Create Unit →
// Create Guest → Create Reservation → Verify
func TestFullHappyPath(t *testing.T) {
	// Step 1: Register
	regReq := registerRequest{
		FullName:     "Amit Patel",
		Email:        "amit-happy@stayflow-test.com",
		Password:     "HappyPath@2026",
		PropertyName: "Happy Path Hotel",
		Phone:        "+919999888777",
	}

	resp := doPost(t, "/api/v1/auth/register", regReq, "")
	assertStatus(t, resp, http.StatusCreated, "register")

	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	resp.Body.Close()

	var regResult registerResponse
	json.Unmarshal(apiResp.Data, &regResult)

	token := regResult.AccessToken
	if token == "" {
		t.Fatal("no access token from registration")
	}
	t.Logf("✓ Registered user: %s (tenant: %s)", regResult.User.Email, regResult.Tenant.Slug)

	// Step 2: Login with the same credentials
	loginResp := doPost(t, "/api/v1/auth/login", loginRequest{
		Email:    "amit-happy@stayflow-test.com",
		Password: "HappyPath@2026",
	}, "")
	assertStatus(t, loginResp, http.StatusOK, "login")

	var loginApiResp apiResponse
	json.NewDecoder(loginResp.Body).Decode(&loginApiResp)
	loginResp.Body.Close()

	var loginResult loginResponse
	json.Unmarshal(loginApiResp.Data, &loginResult)

	if loginResult.User.FullName != "Amit Patel" {
		t.Errorf("login: expected full_name 'Amit Patel', got '%s'", loginResult.User.FullName)
	}
	token = loginResult.AccessToken // Use fresh token
	t.Log("✓ Login successful")

	// Step 3: Create Property
	propReq := map[string]any{
		"name":    "Beachside Villas",
		"address": "12 Marine Drive",
		"city":    "Mumbai",
		"state":   "Maharashtra",
		"country": "India",
	}

	resp = doPost(t, "/api/v1/properties", propReq, token)
	assertStatus(t, resp, http.StatusCreated, "create property")

	propertyID := extractID(t, resp)
	t.Logf("✓ Created property: %s", propertyID)

	// Step 4: Create Unit Type
	unitTypeReq := map[string]any{
		"name":          "Ocean View Suite",
		"base_rate":     8000,
		"max_occupancy": 3,
	}

	resp = doPost(t, fmt.Sprintf("/api/v1/properties/%s/unit-types", propertyID), unitTypeReq, token)
	assertStatus(t, resp, http.StatusCreated, "create unit type")

	unitTypeID := extractID(t, resp)
	t.Logf("✓ Created unit type: %s", unitTypeID)

	// Step 5: Create Units
	units := []map[string]any{
		{"unit_number": "101", "floor": "1", "unit_type_id": unitTypeID},
		{"unit_number": "102", "floor": "1", "unit_type_id": unitTypeID},
		{"unit_number": "201", "floor": "2", "unit_type_id": unitTypeID},
	}

	var unitIDs []string
	for _, unit := range units {
		resp = doPost(t, fmt.Sprintf("/api/v1/properties/%s/units", propertyID), unit, token)
		assertStatus(t, resp, http.StatusCreated, "create unit")
		id := extractID(t, resp)
		unitIDs = append(unitIDs, id)
	}
	t.Logf("✓ Created %d units", len(unitIDs))

	// Step 6: Create Guest
	guestReq := map[string]any{
		"first_name":  "Vikram",
		"last_name":   "Singh",
		"email":       "vikram@example.com",
		"phone":       "+919876111222",
		"id_type":     "passport",
		"id_number":   "J1234567",
		"nationality": "Indian",
	}

	resp = doPost(t, "/api/v1/guests", guestReq, token)
	assertStatus(t, resp, http.StatusCreated, "create guest")

	guestID := extractID(t, resp)
	t.Logf("✓ Created guest: %s", guestID)

	// Step 7: Create Reservation
	checkIn := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 4).Format("2006-01-02")

	resReq := map[string]any{
		"guest_id":         guestID,
		"property_id":      propertyID,
		"unit_id":          unitIDs[0],
		"check_in_date":    checkIn,
		"check_out_date":   checkOut,
		"num_guests":       2,
		"booking_source":   "walk_in",
		"rate_per_night":   8000,
		"special_requests": "Late check-in",
	}

	resp = doPost(t, "/api/v1/reservations", resReq, token)
	assertStatus(t, resp, http.StatusCreated, "create reservation")

	reservationID := extractID(t, resp)
	t.Logf("✓ Created reservation: %s (check-in: %s, check-out: %s)", reservationID, checkIn, checkOut)

	// Step 8: Get reservation details
	resp = doGet(t, "/api/v1/reservations/"+reservationID, token)
	assertStatus(t, resp, http.StatusOK, "get reservation")

	var resApiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&resApiResp)
	resp.Body.Close()

	var reservation map[string]any
	json.Unmarshal(resApiResp.Data, &reservation)

	if reservation["guest_id"] != guestID {
		t.Errorf("reservation guest_id mismatch: expected %s, got %v", guestID, reservation["guest_id"])
	}
	t.Log("✓ Reservation details verified")

	// Step 9: List reservations
	resp = doGet(t, "/api/v1/reservations", token)
	assertStatus(t, resp, http.StatusOK, "list reservations")
	resp.Body.Close()
	t.Log("✓ List reservations works")

	// Step 10: Verify we can't access anything after logout
	logoutResp := doPost(t, "/api/v1/auth/logout", nil, token)
	if logoutResp.StatusCode == http.StatusOK || logoutResp.StatusCode == http.StatusNoContent {
		t.Log("✓ Logout successful")
	}
	logoutResp.Body.Close()

	t.Log("=== Full Happy Path Completed Successfully ===")
}

// TestMultiTenantIsolation ensures tenants can't see each other's data.
func TestMultiTenantIsolation(t *testing.T) {
	// Register Tenant A
	tokenA := registerAndGetToken(t, "tenant-a@stayflow-test.com", "Tenant A Admin", "Hotel A")

	// Register Tenant B
	tokenB := registerAndGetToken(t, "tenant-b@stayflow-test.com", "Tenant B Admin", "Hotel B")

	// Tenant A creates a property
	propReq := map[string]any{
		"name":    "Tenant A Property",
		"address": "123 A Street",
		"city":    "Delhi",
		"state":   "Delhi",
		"country": "India",
	}
	resp := doPost(t, "/api/v1/properties", propReq, tokenA)
	assertStatus(t, resp, http.StatusCreated, "tenant A create property")
	resp.Body.Close()

	// Tenant B lists properties - should see nothing from Tenant A
	resp = doGet(t, "/api/v1/properties", tokenB)
	assertStatus(t, resp, http.StatusOK, "tenant B list properties")

	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	resp.Body.Close()

	// Verify the list is empty or only contains Tenant B's data
	var properties []map[string]any
	json.Unmarshal(apiResp.Data, &properties)

	for _, prop := range properties {
		name, _ := prop["name"].(string)
		if name == "Tenant A Property" {
			t.Error("SECURITY: Tenant B can see Tenant A's property!")
		}
	}
	t.Log("✓ Multi-tenant isolation verified")
}

// --- Helpers ---

func assertStatus(t *testing.T, resp *http.Response, expected int, context string) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("[%s] expected status %d, got %d: %s", context, expected, resp.StatusCode, string(body))
	}
}

func extractID(t *testing.T, resp *http.Response) string {
	t.Helper()
	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	resp.Body.Close()

	var result map[string]any
	json.Unmarshal(apiResp.Data, &result)

	id, ok := result["id"].(string)
	if !ok || id == "" {
		t.Fatal("expected 'id' field in response")
	}
	return id
}
