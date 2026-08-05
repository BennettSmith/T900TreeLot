package views_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/web/views"
)

func TestBootstrapPageRendersEntryEnrollAndPasskeyStages(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	for _, test := range []struct {
		name  string
		page  views.BootstrapPage
		wants []string
	}{
		{
			name: "entry",
			page: views.BootstrapPage{
				PageTitle: "Bootstrap first Admin",
				Brand:     "Troop 900 Tree Lot",
				Stage:     views.BootstrapStageEntry,
				CSRFToken: "csrf-token",
				Fields: []views.Field{{
					ID: "bootstrap-token", Name: "bootstrap_token", Label: "Bootstrap token", Type: "text",
				}},
			},
			wants: []string{`data-bootstrap-entry`, `name="bootstrap_token"`, `action="/bootstrap/start"`, `/static/passkeys.js`},
		},
		{
			name: "enroll",
			page: views.BootstrapPage{
				PageTitle: "Claim first Admin profile",
				Brand:     "Troop 900 Tree Lot",
				Stage:     views.BootstrapStageEnroll,
				CSRFToken: "csrf-token",
				Token:     "bootstrap-token",
				Fields: []views.Field{
					{ID: "bootstrap-email", Name: "email", Label: "Email address", Type: "email", Autocomplete: "email"},
					{ID: "bootstrap-first-name", Name: "first_name", Label: "First name", Type: "text", Autocomplete: "given-name"},
				},
			},
			wants: []string{`action="/bootstrap/claim"`, `name="email"`, `autocomplete="given-name"`, `value="bootstrap-token"`},
		},
		{
			name: "passkey",
			page: views.BootstrapPage{
				PageTitle:            "Register Admin passkey",
				Brand:                "Troop 900 Tree Lot",
				Stage:                views.BootstrapStagePasskey,
				CSRFToken:            "csrf-token",
				Token:                "bootstrap-token",
				Email:                "first@example.org",
				FirstName:            "First",
				LastName:             "Admin",
				PreferredDisplayName: "First Admin",
			},
			wants: []string{`data-bootstrap-passkey`, `data-begin-url="/bootstrap/passkey/begin"`, `data-finish-url="/bootstrap/passkey/finish"`, `first@example.org`},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := renderer.Bootstrap(context.Background(), &output, test.page); err != nil {
				t.Fatalf("Bootstrap: %v", err)
			}
			html := output.String()
			for _, want := range test.wants {
				if !strings.Contains(html, want) {
					t.Errorf("%s missing %q in %s", test.name, want, html)
				}
			}
		})
	}
}

func TestAccountPageRendersAdminWelcome(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	var output bytes.Buffer
	if err := renderer.Account(context.Background(), &output, views.AccountPage{
		PageTitle:   "Admin account",
		Brand:       "Troop 900 Tree Lot",
		DisplayName: "Ada Admin",
		CSRFToken:   "csrf-token",
	}); err != nil {
		t.Fatalf("Account: %v", err)
	}
	html := output.String()
	for _, want := range []string{"Welcome, Ada Admin", `aria-label="Primary"`, "/static/app.css", `action="/sign-out"`, `value="csrf-token"`, "Sign out"} {
		if !strings.Contains(html, want) {
			t.Fatalf("account page missing %q in %s", want, html)
		}
	}
}

func TestAccountSecurityPageRendersStepUp(t *testing.T) {
	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := renderer.AccountSecurity(context.Background(), &output, views.AccountSecurityPage{
		PageTitle: "Account security", Brand: "Troop 900 Tree Lot", DisplayName: "Ada",
		PrimaryEmail: "ada@example.org", CSRFToken: "csrf-token", StepUpRequired: true,
		Navigation: []views.Link{{Label: "Security", Href: "/account/security", Current: true}},
	}); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	if !strings.Contains(body, `data-account-step-up`) || strings.Contains(body, `data-account-passkeys`) {
		t.Fatalf("body=%q", body)
	}
}
