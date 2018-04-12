package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// CSSApplier applies CSS to a gtk.Window.
type CSSApplier struct {
	addCSS   []AddCSS
	provider *gtk.CssProvider
	css      string
}

// Apply CSS to a gtk.Window.
func (ca *CSSApplier) Apply(screen *gdk.Screen, addCSS AddCSS) error {
	if ca.provider == nil {
		provider, err := gtk.CssProviderNew()
		if err != nil {
			return err
		}
		ca.provider = provider
		gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	}

	ca.css += fmt.Sprintf(".%s { %s }\n", addCSS.Class, addCSS.Value)

	err := ca.provider.LoadFromData(ca.css)
	if err != nil {
		return err
	}

	return nil
}
