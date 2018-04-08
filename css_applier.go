package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// CSSApplier applies CSS to a gtk.Window.
type CSSApplier struct {
	cssOptions []cssOptions
	provider   *gtk.CssProvider
}

// Apply CSS to a gtk.Window.
func (ca *CSSApplier) Apply(screen *gdk.Screen, options cssOptions) error {
	ca.cssOptions = append(ca.cssOptions, options)

	if ca.provider == nil {
		provider, err := gtk.CssProviderNew()
		if err != nil {
			return err
		}
		ca.provider = provider
		gtk.AddProviderForScreen(screen, provider, 0)
	}

	css := ""
	for _, options := range ca.cssOptions {
		css += fmt.Sprintf(".%s { %s }\n", options.Class, options.Value)
	}
	err := ca.provider.LoadFromData(css)
	if err != nil {
		return err
	}

	return nil
}
