package users_test

import (
	"context"
	"testing"

	"voco/internal/domain"
	"voco/internal/usecase/users"
)

func TestEnsureAndNicknameUnique(t *testing.T) {
	ctx := context.Background()
	store := users.NewMemoryStore()
	uc := users.New(store, nil)

	a, err := uc.EnsureFromAuth(ctx, "sub-a", "a@x.com", "A", "alice_login")
	if err != nil {
		t.Fatal(err)
	}
	b, err := uc.EnsureFromAuth(ctx, "sub-b", "b@x.com", "B", "bob_login")
	if err != nil {
		t.Fatal(err)
	}
	if a.Nickname != "alice_login" || b.Nickname != "bob_login" {
		t.Fatalf("nickname should equal login: a=%q b=%q", a.Nickname, b.Nickname)
	}

	if _, err := uc.UpdateMe(ctx, a.ID, "alice", "Alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.UpdateMe(ctx, b.ID, "alice", "Bob"); err != domain.ErrNicknameTaken {
		t.Fatalf("want nickname taken, got %v", err)
	}

	found, err := uc.Search(ctx, "ali", 10)
	if err != nil || len(found) != 1 || found[0].Nickname != "alice" {
		t.Fatalf("search: %+v err=%v", found, err)
	}
}

type fakeDir struct{ users []users.DirectoryUser }

func (f fakeDir) ListUsers(context.Context) ([]users.DirectoryUser, error) { return f.users, nil }

func TestSyncUpsertBySub(t *testing.T) {
	ctx := context.Background()
	store := users.NewMemoryStore()
	uc := users.New(store, fakeDir{users: []users.DirectoryUser{
		{Sub: "s1", Email: "1@x.com", DisplayName: "One", Username: "one"},
		{Sub: "s1", Email: "1b@x.com", DisplayName: "OneB", Username: "one"},
	}})
	n, err := uc.SyncFromKeycloak(ctx)
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	u, ok, err := store.GetByKeycloakSub(ctx, "s1")
	if err != nil || !ok || u.Email != "1b@x.com" {
		t.Fatalf("got %+v ok=%v err=%v", u, ok, err)
	}
}
