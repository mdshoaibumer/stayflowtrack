package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestPropertyManagement(t *testing.T) {
	token := registerAndGetToken(t, "prop-admin@stayflow-test.com", "Property Admin", "Prop Test Hotel")

	var propertyID string

	t.Run("create property", func(t *testing.T) {
		req := map[string]any{
			"name":    "Sunrise Apartments",
			"address": "456 Beach Road",
			"city":    "Goa",
			"state":   "Goa",
			"country": "India",
		}

		resp := doPost(t, "/api/v1/properties", req, token)
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		var result map[string]any
		json.Unmarshal(apiResp.Data, &result)
		id, ok := result["id"].(string)
		if !ok || id == "" {
			t.Fatal("expected property id in response")
		}
		propertyID = id
		t.Logf("Created property: %s", propertyID)
	})

	t.Run("list properties", func(t *testing.T) {
		resp := doGet(t, "/api/v1/properties", token)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("get property by id", func(t *testing.T) {
		if propertyID == "" {
			t.Skip("no property created")
		}
		resp := doGet(t, "/api/v1/properties/"+propertyID, token)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	var unitTypeID string

	t.Run("create unit type", func(t *testing.T) {
		if propertyID == "" {
			t.Skip("no property created")
		}
		req := map[string]any{
			"name":          "Deluxe Room",
			"base_rate":     4500,
			"max_occupancy": 2,
		}

		resp := doPost(t, fmt.Sprintf("/api/v1/properties/%s/unit-types", propertyID), req, token)
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		var result map[string]any
		json.Unmarshal(apiResp.Data, &result)
		id, ok := result["id"].(string)
		if !ok || id == "" {
			t.Fatal("expected unit_type id in response")
		}
		unitTypeID = id
		t.Logf("Created unit type: %s", unitTypeID)
	})

	t.Run("create unit", func(t *testing.T) {
		if propertyID == "" || unitTypeID == "" {
			t.Skip("prerequisites not met")
		}
		req := map[string]any{
			"unit_number":  "101",
			"floor":        "1",
			"unit_type_id": unitTypeID,
		}

		resp := doPost(t, fmt.Sprintf("/api/v1/properties/%s/units", propertyID), req, token)
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("list units", func(t *testing.T) {
		if propertyID == "" {
			t.Skip("no property created")
		}
		resp := doGet(t, fmt.Sprintf("/api/v1/properties/%s/units", propertyID), token)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("unauthorized without token", func(t *testing.T) {
		resp := doGet(t, "/api/v1/properties", "")
		if resp.StatusCode != http.StatusUnauthorized {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 401, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})
}

func TestGuestManagement(t *testing.T) {
	token := registerAndGetToken(t, "guest-admin@stayflow-test.com", "Guest Admin", "Guest Test Hotel")

	var guestID string

	t.Run("create guest", func(t *testing.T) {
		req := map[string]any{
			"first_name":  "Priya",
			"last_name":   "Sharma",
			"email":       "priya@example.com",
			"phone":       "+919876000111",
			"id_type":     "aadhaar",
			"id_number":   "1234-5678-9012",
			"nationality": "Indian",
		}

		resp := doPost(t, "/api/v1/guests", req, token)
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		var result map[string]any
		json.Unmarshal(apiResp.Data, &result)
		id, ok := result["id"].(string)
		if !ok || id == "" {
			t.Fatal("expected guest id in response")
		}
		guestID = id
		t.Logf("Created guest: %s", guestID)
	})

	t.Run("list guests", func(t *testing.T) {
		resp := doGet(t, "/api/v1/guests", token)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("get guest by id", func(t *testing.T) {
		if guestID == "" {
			t.Skip("no guest created")
		}
		resp := doGet(t, "/api/v1/guests/"+guestID, token)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})
}
