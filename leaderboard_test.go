package predictclob

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

func TestGetLeaderboardUserStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify request body
		var req types.GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.OperationName != "GetLeaderboardUserStats" {
			t.Errorf("Expected operationName 'GetLeaderboardUserStats', got '%s'", req.OperationName)
		}

		addr, ok := req.Variables["address"]
		if !ok {
			t.Error("Expected 'address' in variables")
		}
		if addr != "0x84D3b0a13c40Eb113dC81dc4ba998F499A0D564c" {
			t.Errorf("Expected address '0x84D3b0a13c40Eb113dC81dc4ba998F499A0D564c', got '%s'", addr)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"account": {
					"name": "DukeGiganticBets",
					"address": "0x84D3b0a13c40Eb113dC81dc4ba998F499A0D564c",
					"imageUrl": null,
					"imageStatus": "PUBLIC",
					"leaderboard": {
						"allocationRoundPoints": 87.35298404113188,
						"totalPoints": 88.0138880590596,
						"rank": 21712
					}
				}
			}
		}`))
	}))
	defer server.Close()

	client := &Client{
		graphqlHost: server.URL,
		reqClient:   CreateReqClientWithProxy(nil, "", 30*time.Second),
	}

	address := common.HexToAddress("0x84D3b0a13c40Eb113dC81dc4ba998F499A0D564c")
	account, err := client.GetLeaderboardUserStats(address)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if account.Name != "DukeGiganticBets" {
		t.Errorf("Expected name 'DukeGiganticBets', got '%s'", account.Name)
	}
	if account.Address != address {
		t.Errorf("Expected address %s, got %s", address.Hex(), account.Address.Hex())
	}
	if account.ImageURL != nil {
		t.Errorf("Expected nil imageUrl, got '%v'", account.ImageURL)
	}
	if account.ImageStatus != "PUBLIC" {
		t.Errorf("Expected imageStatus 'PUBLIC', got '%s'", account.ImageStatus)
	}
	if account.Leaderboard == nil {
		t.Fatal("Expected non-nil leaderboard")
	}
	if account.Leaderboard.AllocationRoundPoints != 87.35298404113188 {
		t.Errorf("Expected allocationRoundPoints 87.35298404113188, got %f", account.Leaderboard.AllocationRoundPoints)
	}
	if account.Leaderboard.TotalPoints != 88.0138880590596 {
		t.Errorf("Expected totalPoints 88.0138880590596, got %f", account.Leaderboard.TotalPoints)
	}
	if account.Leaderboard.Rank != 21712 {
		t.Errorf("Expected rank 21712, got %d", account.Leaderboard.Rank)
	}
}

func TestGetLeaderboardUserStats_AccountNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"account":null}}`))
	}))
	defer server.Close()

	client := &Client{
		graphqlHost: server.URL,
		reqClient:   CreateReqClientWithProxy(nil, "", 30*time.Second),
	}

	address := common.HexToAddress("0x0000000000000000000000000000000000000001")
	_, err := client.GetLeaderboardUserStats(address)
	if err == nil {
		t.Fatal("Expected error for non-existent account")
	}
}

func TestGetLeaderboardUserStats_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":null,"errors":[{"message":"Variable '$address' got invalid value '0xinvalid'"}]}`))
	}))
	defer server.Close()

	client := &Client{
		graphqlHost: server.URL,
		reqClient:   CreateReqClientWithProxy(nil, "", 30*time.Second),
	}

	address := common.HexToAddress("0x0000000000000000000000000000000000000001")
	_, err := client.GetLeaderboardUserStats(address)
	if err == nil {
		t.Fatal("Expected error for GraphQL error response")
	}
}

func TestGetLeaderboardUserStats_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	client := &Client{
		graphqlHost: server.URL,
		reqClient:   CreateReqClientWithProxy(nil, "", 30*time.Second),
	}

	address := common.HexToAddress("0x0000000000000000000000000000000000000001")
	_, err := client.GetLeaderboardUserStats(address)
	if err == nil {
		t.Fatal("Expected error for server error")
	}
}
