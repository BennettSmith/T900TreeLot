package views_test

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/web/views"
)

func TestComponentGalleryRendersAccessibleSignalFoundation(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	var output bytes.Buffer
	if err := renderer.ComponentGallery(context.Background(), &output, views.GalleryData()); err != nil {
		t.Fatalf("ComponentGallery: %v", err)
	}

	html := output.String()
	for _, want := range []string{
		`<!doctype html>`,
		`<main id="main-content"`,
		`<h1`,
		`href="/static/app.css"`,
		`aria-label="Design system sections"`,
		`Shift cards`,
		`Person selector`,
		`Attendance row`,
		`Delivery status`,
		`Scout Bucks summary`,
		`aria-live="polite"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered gallery does not contain %q", want)
		}
	}
}

func TestGalleryStaticAccessibilityAndResponsiveContracts(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	var output bytes.Buffer
	if err := renderer.ComponentGallery(context.Background(), &output, views.GalleryData()); err != nil {
		t.Fatalf("ComponentGallery: %v", err)
	}
	html := output.String()

	if count := strings.Count(html, "<h1"); count != 1 {
		t.Errorf("page-level heading count = %d, want 1", count)
	}
	for _, want := range []string{
		`<html lang="en">`,
		`<aside class="rail">`,
		`<nav aria-label="Design system sections">`,
		`<label for="display-name">`,
		`aria-invalid="true"`,
		`aria-describedby="reason-signup-disabled"`,
		`id="reason-signup-disabled"`,
		`<dialog class="dialog" id="cancel-dialog" open`,
		`<caption>Recent assignments`,
		`scope="col"`,
		`prefers-reduced-motion`,
		`(opens in a new tab)`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("accessibility contract missing %q", want)
		}
	}
	if strings.Contains(html, `style="`) || strings.Contains(html, "<script") {
		t.Error("gallery contains inline style or script instead of static progressive HTML")
	}
}

func TestGalleryFragmentsResolveAndScoutBucksStatesStayPrecise(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	var output bytes.Buffer
	if err := renderer.ComponentGallery(context.Background(), &output, views.GalleryData()); err != nil {
		t.Fatalf("ComponentGallery: %v", err)
	}
	html := output.String()

	ids := make(map[string]bool)
	for _, match := range regexp.MustCompile(`\sid="([^"]+)"`).FindAllStringSubmatch(html, -1) {
		ids[match[1]] = true
	}
	for _, match := range regexp.MustCompile(`\shref="#([^"]+)"`).FindAllStringSubmatch(html, -1) {
		if !ids[match[1]] {
			t.Errorf("fragment link #%s has no target", match[1])
		}
	}

	provisional := regexp.MustCompile(`(?s)data-scout-bucks-state="provisional".*?</article>`).FindString(html)
	if provisional == "" {
		t.Fatal("gallery does not render a provisional Scout Bucks state")
	}
	if strings.Contains(provisional, "$") || !strings.Contains(provisional, "Provisional credited hours") {
		t.Error("provisional Scout Bucks state exposes dollar data or lacks precise provisional language")
	}

	finalized := regexp.MustCompile(`(?s)data-scout-bucks-state="finalized".*?</article>`).FindString(html)
	if finalized == "" {
		t.Fatal("gallery does not render a finalized Scout Bucks state")
	}
	for _, want := range []string{"Finalized credited hours", "$12,500.00", "Exact-pool reconciliation"} {
		if !strings.Contains(finalized, want) {
			t.Errorf("finalized Scout Bucks state does not contain %q", want)
		}
	}
}

func TestUnknownPresentationVariantsFallBackToSafeDefaults(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	data := views.GalleryData()
	data.Buttons[0].Variant = "invented"
	data.Badges[0].Variant = views.Variant("invented")
	data.Alerts[0].Variant = views.Variant("invented")
	data.Navigation[0].Variant = "invented"

	var output bytes.Buffer
	if err := renderer.ComponentGallery(context.Background(), &output, data); err != nil {
		t.Fatalf("ComponentGallery: %v", err)
	}
	html := output.String()
	if strings.Contains(html, "--invented") {
		t.Error("unknown presentation variant leaked into a generated CSS class")
	}
	for _, want := range []string{"button--secondary", "badge--neutral", "alert--info", "link--default"} {
		if !strings.Contains(html, want) {
			t.Errorf("safe fallback class %q is missing", want)
		}
	}
}

func TestParityRendererUsesSameResultForPageAndFragment(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	data := views.GalleryData()
	data.ParityMessage = `<Signal & safe>`

	var page bytes.Buffer
	if err := renderer.ParityResult(context.Background(), &page, data); err != nil {
		t.Fatalf("full page: %v", err)
	}
	data.ParityFragment = true
	var fragment bytes.Buffer
	if err := renderer.ParityResult(context.Background(), &fragment, data); err != nil {
		t.Fatalf("fragment: %v", err)
	}

	for name, rendered := range map[string]string{"page": page.String(), "fragment": fragment.String()} {
		if !strings.Contains(rendered, `&lt;Signal &amp; safe&gt;`) {
			t.Errorf("%s did not safely render the same message", name)
		}
	}
	if !strings.Contains(page.String(), "<!doctype html>") || strings.Contains(fragment.String(), "<!doctype html>") {
		t.Error("page and fragment framing did not differ as expected")
	}
}

func TestRendererHonorsCancelledContext(t *testing.T) {
	t.Parallel()

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := renderer.ComponentGallery(ctx, &bytes.Buffer{}, views.GalleryData()); !errors.Is(err, context.Canceled) {
		t.Errorf("ComponentGallery error = %v, want context.Canceled", err)
	}
	if err := renderer.ParityResult(ctx, &bytes.Buffer{}, views.GalleryData()); !errors.Is(err, context.Canceled) {
		t.Errorf("ParityResult error = %v, want context.Canceled", err)
	}
}
