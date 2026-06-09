package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// apiResponse matches the backend's APIResponse structure.
type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *apiError       `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type registerRequest struct {
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	PropertyName string `json:"property_name"`
	Phone        string `json:"phone,omitempty"`
}

type registerResponse struct {
	Tenant struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"tenant"`
	User struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
		TenantID string `json:"tenant_id"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
		TenantID string `json:"tenant_id"`
	} `json:"user"`
}

var testEnv *TestEnv

func TestMain(m *testing.M) {
	var err error
	testEnv, err = SetupTestEnv()
	if err != nil {
		fmt.Printf("FATAL: failed to setup test env: %v\n", err)
		fmt.Println("Embedded PostgreSQL will be downloaded on first run (may take a minute).")
		return
	}

	code := m.Run()

	testEnv.Teardown()
	if code != 0 {
		fmt.Printf("Tests failed with exit code %d\n", code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(testEnv.Server.URL + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRegisterTenant(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		req := registerRequest{
			FullName:     "Rajesh Kumar",
			Email:        "rajesh@stayflow-test.com",
			Password:     "SecurePass@2026",
			PropertyName: "Grand Palace Hotel",
			Phone:        "+919876543210",
		}

		resp := doPost(t, "/api/v1/auth/register", req, "")
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		if !apiResp.Success {
			t.Fatalf("expected success=true, got error: %v", apiResp.Error)
		}

		var result registerResponse
		json.Unmarshal(apiResp.Data, &result)

		// Verify all fields are correct
		if result.User.FullName != "Rajesh Kumar" {
			t.Errorf("expected full_name 'Rajesh Kumar', got '%s'", result.User.FullName)
		}
		if result.User.Email != "rajesh@stayflow-test.com" {
			t.Errorf("expected email 'rajesh@stayflow-test.com', got '%s'", result.User.Email)
		}
		if result.User.Role != "super_admin" {
			t.Errorf("expected role 'super_admin', got '%s'", result.User.Role)
		}
		if result.Tenant.Name != "Grand Palace Hotel" {
			t.Errorf("expected tenant name 'Grand Palace Hotel', got '%s'", result.Tenant.Name)
		}
		if result.AccessToken == "" {
			t.Error("expected non-empty access_token")
		}
		if result.RefreshToken == "" {
			t.Error("expected non-empty refresh_token")
		}
		if result.User.TenantID == "" {
			t.Error("expected non-empty tenant_id")
		}
	})

	t.Run("single word name", func(t *testing.T) {
		req := registerRequest{
			FullName:     "Madonna",
			Email:        "madonna@stayflow-test.com",
			Password:     "SecurePass@2026",
			PropertyName: "Star Hotel",
		}

		resp := doPost(t, "/api/v1/auth/register", req, "")
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		var result registerResponse
		json.Unmarshal(apiResp.Data, &result)

		if result.User.FullName != "Madonna" {
			t.Errorf("expected full_name 'Madonna', got '%s'", result.User.FullName)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		req := registerRequest{
			FullName:     "Duplicate User",
			Email:        "rajesh@stayflow-test.com", // already registered
			Password:     "SecurePass@2026",
			PropertyName: "Another Hotel",
		}

		resp := doPost(t, "/api/v1/auth/register", req, "")
		if resp.StatusCode != http.StatusConflict {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 409, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("missing required fields", func(t *testing.T) {
		req := registerRequest{
			Email:    "incomplete@test.com",
			Password: "SecurePass@2026",
			// Missing FullName and PropertyName
		}

		resp := doPost(t, "/api/v1/auth/register", req, "")
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 400 or 422, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("password too short", func(t *testing.T) {
		req := registerRequest{
			FullName:     "Short Pass",
			Email:        "shortpass@test.com",
			Password:     "short",
			PropertyName: "Short Hotel",
		}

		resp := doPost(t, "/api/v1/auth/register", req, "")
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 400 or 422, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})
}

func TestLogin(t *testing.T) {
	// First register a user
	regReq := registerRequest{
		FullName:     "Login Tester",
		Email:        "login-tester@stayflow-test.com",
		Password:     "SecurePass@2026",
		PropertyName: "Login Test Hotel",
	}
	regResp := doPost(t, "/api/v1/auth/register", regReq, "")
	if regResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(regResp.Body)
		t.Fatalf("setup failed: register returned %d: %s", regResp.StatusCode, string(body))
	}
	regResp.Body.Close()

	t.Run("successful login", func(t *testing.T) {
		req := loginRequest{
			Email:    "login-tester@stayflow-test.com",
			Password: "SecurePass@2026",
		}

		resp := doPost(t, "/api/v1/auth/login", req, "")
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		if !apiResp.Success {
			t.Fatalf("expected success, got error: %v", apiResp.Error)
		}

		var result loginResponse
		json.Unmarshal(apiResp.Data, &result)

		if result.AccessToken == "" {
			t.Error("expected non-empty access_token")
		}
		if result.RefreshToken == "" {
			t.Error("expected non-empty refresh_token")
		}
		if result.User.Email != "login-tester@stayflow-test.com" {
			t.Errorf("expected email 'login-tester@stayflow-test.com', got '%s'", result.User.Email)
		}
		if result.User.FullName != "Login Tester" {
			t.Errorf("expected full_name 'Login Tester', got '%s'", result.User.FullName)
		}
		if result.User.Role != "super_admin" {
			t.Errorf("expected role 'super_admin', got '%s'", result.User.Role)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		req := loginRequest{
			Email:    "login-tester@stayflow-test.com",
			Password: "WrongPassword!",
		}

		resp := doPost(t, "/api/v1/auth/login", req, "")
		if resp.StatusCode != http.StatusUnauthorized {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 401, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})

	t.Run("nonexistent email", func(t *testing.T) {
		req := loginRequest{
			Email:    "nonexistent@test.com",
			Password: "AnyPassword123!",
		}

		resp := doPost(t, "/api/v1/auth/login", req, "")
		if resp.StatusCode != http.StatusUnauthorized {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 401, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()
	})
}

func TestRefreshToken(t *testing.T) {
	// Register and get tokens
	regReq := registerRequest{
		FullName:     "Refresh Tester",
		Email:        "refresh-tester@stayflow-test.com",
		Password:     "SecurePass@2026",
		PropertyName: "Refresh Hotel",
	}

	resp := doPost(t, "/api/v1/auth/register", regReq, "")
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("setup failed: %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	resp.Body.Close()

	var regResult registerResponse
	json.Unmarshal(apiResp.Data, &regResult)

	t.Run("valid refresh", func(t *testing.T) {
		body := map[string]string{"refresh_token": regResult.RefreshToken}
		resp := doPost(t, "/api/v1/auth/refresh", body, "")
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
		}

		var apiResp apiResponse
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		if !apiResp.Success {
			t.Fatalf("expected success, got: %v", apiResp.Error)
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		body := map[string]string{"refresh_token": "totally-invalid-token"}
		resp := doPost(t, "/api/v1/auth/refresh", body, "")
		if resp.StatusCode != http.StatusUnauthorized {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 401, got %d: %s", resp.StatusCode, string(b))
		}
		resp.Body.Close()
	})
}

func TestLogout(t *testing.T) {
	token := registerAndGetToken(t, "logout-user@stayflow-test.com", "Logout User", "Logout Hotel")

	resp := doPost(t, "/api/v1/auth/logout", nil, token)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200/204, got %d: %s", resp.StatusCode, string(body))
	}
	resp.Body.Close()
}

// --- Helpers ---

func registerAndGetToken(t *testing.T, email, name, property string) string {
	t.Helper()
	req := registerRequest{
		FullName:     name,
		Email:        email,
		Password:     "SecurePass@2026",
		PropertyName: property,
	}

	resp := doPost(t, "/api/v1/auth/register", req, "")
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("registerAndGetToken failed: %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	resp.Body.Close()

	var result registerResponse
	json.Unmarshal(apiResp.Data, &result)
	return result.AccessToken
}

func doPost(t *testing.T, path string, body any, token string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(http.MethodPost, testEnv.Server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func doGet(t *testing.T, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, testEnv.Server.URL+path, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}
