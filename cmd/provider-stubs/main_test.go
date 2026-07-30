package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGroupsIOStubAcceptsAuthorizedMessage(t *testing.T) {
	t.Parallel()

	stub := newGroupsIOStub()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/groups/{group}/messages", stub.postMessage)
	mux.HandleFunc("GET /_stub/groupsio/messages", stub.listMessages)

	unauthorized := httptest.NewRequest(http.MethodPost, "/api/v1/groups/troop900/messages", strings.NewReader(`{"subject":"hi"}`))
	unauthorizedResponse := httptest.NewRecorder()
	mux.ServeHTTP(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorizedResponse.Code)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/troop900/messages", strings.NewReader(`{"subject":"Coverage alert"}`))
	request.Header.Set("Authorization", "Bearer stub-token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d", response.Code)
	}

	list := httptest.NewRecorder()
	mux.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/_stub/groupsio/messages", nil))
	var messages []storedMessage
	if err := json.Unmarshal(list.Body.Bytes(), &messages); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(messages) != 1 || messages[0].Group != "troop900" {
		t.Fatalf("messages = %+v", messages)
	}
}
