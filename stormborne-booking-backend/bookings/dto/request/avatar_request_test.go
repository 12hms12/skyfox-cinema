package request

import (
	"encoding/json"
	"testing"
)

func exactlyOneProvided(r UpdateAvatarRequest) bool {
	hasID := r.PredefinedAvatarID > 0
	hasURL := len(r.AvatarURL) > 0
	return hasID != hasURL // XOR
}

func TestUpdateAvatarRequest_JSON_Unmarshal_ByID(t *testing.T) {
	body := []byte(`{"predefined_avatar_id": 7}`)
	var req UpdateAvatarRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.PredefinedAvatarID != 7 {
		t.Fatalf("want id=7, got %d", req.PredefinedAvatarID)
	}
	if !exactlyOneProvided(req) {
		t.Fatalf("expected exactly one of ID or URL")
	}
}

func TestUpdateAvatarRequest_JSON_Unmarshal_ByURL(t *testing.T) {
	body := []byte(`{"avatar_url":"https://cdn.example.com/u.png","avatar_type":"uploaded"}`)
	var req UpdateAvatarRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.AvatarURL != "https://cdn.example.com/u.png" {
		t.Fatalf("want url=https://cdn.example.com/u.png, got %s", req.AvatarURL)
	}
	if req.AvatarType != "uploaded" {
		t.Fatalf("want type=uploaded, got %s", req.AvatarType)
	}
	if !exactlyOneProvided(req) {
		t.Fatalf("expected exactly one of ID or URL")
	}
}

func TestUpdateAvatarRequest_BothProvided_ShouldBeInvalid(t *testing.T) {
	body := []byte(`{"predefined_avatar_id":1,"avatar_url":"https://x","avatar_type":"uploaded"}`)
	var req UpdateAvatarRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if exactlyOneProvided(req) {
		t.Fatalf("expected BOTH provided to be invalid (not exactly one)")
	}
}

func TestUpdateAvatarRequest_NoneProvided_ShouldBeInvalid(t *testing.T) {
	var req UpdateAvatarRequest
	if exactlyOneProvided(req) {
		t.Fatalf("expected NONE provided to be invalid (not exactly one)")
	}
}
