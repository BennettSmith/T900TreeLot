package views_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/web/views"
)

func TestHomeRespectsCanceledContext(t *testing.T) {
	t.Parallel()
	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := renderer.Home(ctx, new(bytes.Buffer), views.Home{}); err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestHomeRendersBrandNavigationAndCSRFField(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	var output bytes.Buffer
	err = renderer.Home(context.Background(), &output, views.Home{
		PageTitle:  "Troop 900 Tree Lot",
		Brand:      "Troop 900 Tree Lot",
		Headline:   "Volunteer shift scheduling",
		Supporting: "Families coordinate tree-lot shifts.",
		Navigation: []views.Link{{Label: "Home", Href: "/", Current: true}},
		CSRFToken:  "csrf-token",
	})
	if err != nil {
		t.Fatalf("Home: %v", err)
	}
	html := output.String()
	for _, want := range []string{
		"Troop 900 Tree Lot",
		`aria-label="Primary"`,
		`name="csrf_token" value="csrf-token"`,
		"/static/app.css",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("home missing %q", want)
		}
	}
}
