package webauthn

import "testing"

func TestRegistrationUserMethods(t *testing.T) {
	user := registrationUser{id: []byte("handle"), name: "first@example.org"}
	if string(user.WebAuthnID()) != "handle" {
		t.Fatalf("WebAuthnID = %q", user.WebAuthnID())
	}
	if user.WebAuthnName() != "first@example.org" {
		t.Fatalf("WebAuthnName = %q", user.WebAuthnName())
	}
	if user.WebAuthnDisplayName() != "first@example.org" {
		t.Fatalf("fallback display name = %q", user.WebAuthnDisplayName())
	}
	if credentials := user.WebAuthnCredentials(); credentials != nil {
		t.Fatalf("credentials = %#v, want nil", credentials)
	}
}
