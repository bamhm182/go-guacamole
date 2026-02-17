package guacamole

import (
	"context"
	"net/http"
	"testing"
)

func TestListSharingProfiles(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/sharingProfiles")
		writeJSON(t, w, map[string]SharingProfile{
			"1": {Identifier: "1", Name: "Read-only Share", PrimaryConnectionIdentifier: "5"},
		})
	})
	got, err := c.ListSharingProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListSharingProfiles: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len: got %d, want 1", len(got))
	}
}

func TestCreateSharingProfile(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPost)
		assertPath(t, r, "/api/session/data/postgresql/sharingProfiles")
		var body SharingProfile
		mustReadJSON(t, r, &body)
		if body.PrimaryConnectionIdentifier != "5" {
			t.Errorf("PrimaryConnectionIdentifier: got %q, want %q", body.PrimaryConnectionIdentifier, "5")
		}
		writeJSON(t, w, SharingProfile{
			Identifier:                  "1",
			Name:                        body.Name,
			PrimaryConnectionIdentifier: body.PrimaryConnectionIdentifier,
		})
	})
	sp, err := c.CreateSharingProfile(context.Background(), SharingProfile{
		Name:                        "Read-only Share",
		PrimaryConnectionIdentifier: "5",
		Parameters:                  map[string]string{"read-only": "true"},
	})
	if err != nil {
		t.Fatalf("CreateSharingProfile: %v", err)
	}
	if sp.Identifier != "1" {
		t.Errorf("Identifier: got %q, want %q", sp.Identifier, "1")
	}
}

func TestGetSharingProfile(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/sharingProfiles/1")
		writeJSON(t, w, SharingProfile{
			Identifier:                  "1",
			Name:                        "Read-only Share",
			PrimaryConnectionIdentifier: "5",
		})
	})
	sp, err := c.GetSharingProfile(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetSharingProfile: %v", err)
	}
	if sp.PrimaryConnectionIdentifier != "5" {
		t.Errorf("PrimaryConnectionIdentifier: got %q, want %q", sp.PrimaryConnectionIdentifier, "5")
	}
}

func TestGetSharingProfileParameters(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodGet)
		assertPath(t, r, "/api/session/data/postgresql/sharingProfiles/1/parameters")
		writeJSON(t, w, map[string]string{"read-only": "true"})
	})
	params, err := c.GetSharingProfileParameters(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetSharingProfileParameters: %v", err)
	}
	if params["read-only"] != "true" {
		t.Errorf(`params["read-only"]: got %q, want "true"`, params["read-only"])
	}
}

func TestUpdateSharingProfile(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodPut)
		assertPath(t, r, "/api/session/data/postgresql/sharingProfiles/1")
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.UpdateSharingProfile(context.Background(), "1", SharingProfile{
		Name:                        "Updated Share",
		PrimaryConnectionIdentifier: "5",
	})
	if err != nil {
		t.Fatalf("UpdateSharingProfile: %v", err)
	}
}

func TestDeleteSharingProfile(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, http.MethodDelete)
		assertPath(t, r, "/api/session/data/postgresql/sharingProfiles/1")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteSharingProfile(context.Background(), "1"); err != nil {
		t.Fatalf("DeleteSharingProfile: %v", err)
	}
}
