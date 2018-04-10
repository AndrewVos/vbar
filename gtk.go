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

static GtkWidget * toGtkWidget(void *p)
{
  return (GTK_WIDGET(p));
}

static GtkMenu * toGtkMenu(void *p)
{
  return (GTK_MENU(p));
}

static GdkDisplay * toGdkDisplay(void *p)
{
	return (GDK_DISPLAY(p));
}

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

	"github.com/gotk3/gotk3/gtk"
)

// Rectangle is just a rectangle.
type Rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

func getMonitorDimensions(window *gtk.Window) (Rectangle, error) {
	screen, err := window.GetScreen()
	if err != nil {
		return Rectangle{}, err
	}
	display, err := screen.GetDisplay()
	if err != nil {
		return Rectangle{}, err
	}

	geometry := C.GdkRectangle{}
	displayPointer := unsafe.Pointer(display.GObject)
	gdkDisplay := C.toGdkDisplay(displayPointer)
	monitor := C.gdk_display_get_primary_monitor(gdkDisplay)
	C.gdk_monitor_get_geometry(monitor, &geometry)
	return Rectangle{
		X:      int(geometry.x),
		Y:      int(geometry.y),
		Width:  int(geometry.width),
		Height: int(geometry.height),
	}, nil
}

func updateDimensions(window *gtk.Window, bar *gtk.Widget) error {
	monitorDimensions, err := getMonitorDimensions(window)
	if err != nil {
		return err
	}

	window.SetSizeRequest(monitorDimensions.Width, -1)

	p := unsafe.Pointer(window.GObject)
	w := C.toGtkWindow(p)

	C.set_strut_properties(
		w,
		0, 0, C.long(bar.GetAllocatedHeight()), 0, /* strut-left, strut-right, strut-top, strut-bottom */
		0, 0, /* strut-left-start-y, strut-left-end-y */
		0, 0, /* strut-right-start-y, strut-right-end-y */
		C.long(monitorDimensions.X), C.long(monitorDimensions.X+monitorDimensions.Width-1), /* strut-top-start-x, strut-top-end-x */
		0, 0, /* strut-bottom-start-x, strut-bottom-end-x */
	)
	return nil
}

func popupMenuAt(widget *gtk.Widget, menu *gtk.Menu) {
	menuPointer := unsafe.Pointer(menu.GObject)
	gtkMenu := C.toGtkMenu(menuPointer)

	widgetPointer := unsafe.Pointer(widget.GObject)
	gtkWidget := C.toGtkWidget(widgetPointer)

	C.gtk_menu_popup_at_widget(
		gtkMenu,
		gtkWidget,
		C.GDK_GRAVITY_SOUTH_WEST,
		C.GDK_GRAVITY_NORTH_WEST,
		nil,
	)
}

func enableTransparency(window *gtk.Window) error {
	screen, err := window.GetScreen()
	if err != nil {
		return err
	}

	visual, err := screen.GetRGBAVisual()
	if err != nil {
		return err
	}

	if visual != nil && screen.IsComposited() {
		window.SetVisual(visual)
	}

	return nil
}

func applyClass(widget *gtk.Widget, class string) {
	styleContext, err := widget.GetStyleContext()
	if err != nil {
		log.Fatal(err)
	}
	styleContext.AddClass(class)
}
