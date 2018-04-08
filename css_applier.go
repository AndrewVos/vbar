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
}

// Apply CSS to a gtk.Window.
func (ca *CSSApplier) Apply(screen *gdk.Screen, addCSS AddCSS) error {
	ca.addCSS = append(ca.addCSS, addCSS)

	if ca.provider == nil {
		provider, err := gtk.CssProviderNew()
		if err != nil {
			return err
		}
		ca.provider = provider
		gtk.AddProviderForScreen(screen, provider, 0)
	}

	css := ""
	for _, addCSS := range ca.addCSS {
		css += fmt.Sprintf(".%s { %s }\n", addCSS.Class, addCSS.Value)
	}
	err := ca.provider.LoadFromData(css)
	if err != nil {
		return err
	}

	return nil
}
