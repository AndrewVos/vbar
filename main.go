package main

/*
#cgo pkg-config: gdk-3.0
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include <gdk/gdk.h>

static GtkWindow * toGtkWindow(void *p)
{
  return (GTK_WINDOW(p));
}

#include <gtk/gtk.h>
#include <gdk/gdk.h>

void set_strut_properties(GtkWindow *window,
				long left, long right, long top, long bottom,
 				long left_start_y, long left_end_y,
 				long right_start_y, long right_end_y,
 				long top_start_x, long top_end_x,
 				long bottom_start_x, long bottom_end_x) {
	gulong data[12] = {0};
	data[0] = left; data[1] = right; data[2] = top; data[3] = bottom;
	data[4] = left_start_y; data[5] = left_end_y;
	data[6] = right_start_y; data[7] = right_end_y;
	data[8] = top_start_x; data[9] = top_end_x;
	data[10] = bottom_start_x; data[11] = bottom_end_x;

	gdk_property_change(gtk_widget_get_window(GTK_WIDGET(window)),
				gdk_atom_intern("_NET_WM_STRUT_PARTIAL", FALSE),
				gdk_atom_intern ("CARDINAL", FALSE),
				32, GDK_PROP_MODE_REPLACE, (unsigned char *)data, 12);

	gdk_property_change(gtk_widget_get_window(GTK_WIDGET(window)),
				gdk_atom_intern("_NET_WM_STRUT", FALSE),
				gdk_atom_intern("CARDINAL", FALSE), 32, GDK_PROP_MODE_REPLACE,
				(unsigned char *) &data, 4);

}
*/
import "C"

import (
	"log"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	height := 50

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetAppPaintable(true)
	win.SetDecorated(false)
	win.SetResizable(false)
	win.SetSkipPagerHint(true)
	win.SetSkipTaskbarHint(true)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
	win.SetVExpand(false)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	l, err := gtk.LabelNew("Hello, gotk3!")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	win.Add(l)

	win.Move(0, 0)
	win.SetDefaultSize(400, height)
	win.SetPosition(gtk.WIN_POS_NONE)

	win.ShowAll()

	monitorX := 0
	monitorWidth := 1920

	p := unsafe.Pointer(win.GObject)
	w := C.toGtkWindow(p)

	C.set_strut_properties(
		w,
		0, 0, C.long(height), 0, /* strut-left, strut-right, strut-top, strut-bottom */
		0, 0, /* strut-left-start-y, strut-left-end-y */
		0, 0, /* strut-right-start-y, strut-right-end-y */
		C.long(monitorX), C.long(monitorX+monitorWidth-1), /* strut-top-start-x, strut-top-end-x */
		0, 0, /* strut-bottom-start-x, strut-bottom-end-x */
	)

	gtk.Main()
}
