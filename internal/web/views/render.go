package views

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
)

//go:embed templates/*.gohtml
var templateFiles embed.FS

type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	functions := template.FuncMap{
		"printfIcon": func(name string) Icon {
			return Icon{Name: name}
		},
		"dictButton": func(label, variant, buttonType string) Button {
			return Button{Label: label, Variant: variant, Type: buttonType}
		},
		"buttonVariant": func(variant string) string {
			switch variant {
			case "primary", "secondary", "quiet", "destructive":
				return variant
			default:
				return "secondary"
			}
		},
		"linkVariant": func(variant string) string {
			switch variant {
			case "standalone", "destructive", "external", "default":
				return variant
			default:
				return "default"
			}
		},
		"badgeVariant": func(variant Variant) Variant {
			switch variant {
			case VariantNeutral, VariantInfo, VariantComplete, VariantWarning, VariantCritical, VariantSpecial, VariantProvisional:
				return variant
			default:
				return VariantNeutral
			}
		},
		"alertVariant": func(variant Variant) Variant {
			switch variant {
			case VariantInfo, VariantComplete, VariantWarning, VariantCritical:
				return variant
			default:
				return VariantInfo
			}
		},
	}
	templates, err := template.New("root").Funcs(functions).ParseFS(templateFiles, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("parse web templates: %w", err)
	}
	return &Renderer{templates: templates}, nil
}

func (r *Renderer) Home(ctx context.Context, output io.Writer, data Home) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "home", data); err != nil {
		return fmt.Errorf("render home: %w", err)
	}
	return nil
}

func (r *Renderer) Bootstrap(ctx context.Context, output io.Writer, data BootstrapPage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "bootstrap", data); err != nil {
		return fmt.Errorf("render bootstrap: %w", err)
	}
	return nil
}

func (r *Renderer) SignIn(ctx context.Context, output io.Writer, data SignInPage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "sign-in", data); err != nil {
		return fmt.Errorf("render sign in: %w", err)
	}
	return nil
}

func (r *Renderer) Landing(ctx context.Context, output io.Writer, data LandingPage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "landing", data); err != nil {
		return fmt.Errorf("render landing: %w", err)
	}
	return nil
}

func (r *Renderer) Account(ctx context.Context, output io.Writer, data AccountPage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "account", data); err != nil {
		return fmt.Errorf("render account: %w", err)
	}
	return nil
}

func (r *Renderer) AccountSecurity(ctx context.Context, output io.Writer, data AccountSecurityPage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "account-security", data); err != nil {
		return fmt.Errorf("render account security: %w", err)
	}
	return nil
}

func (r *Renderer) ComponentGallery(ctx context.Context, output io.Writer, data Gallery) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.templates.ExecuteTemplate(output, "gallery", data); err != nil {
		return fmt.Errorf("render component gallery: %w", err)
	}
	return nil
}

func (r *Renderer) ParityResult(ctx context.Context, output io.Writer, data Gallery) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	templateName := "parity-page"
	if data.ParityFragment {
		templateName = "parity-result"
	}
	if err := r.templates.ExecuteTemplate(output, templateName, data); err != nil {
		return fmt.Errorf("render parity result: %w", err)
	}
	return nil
}
