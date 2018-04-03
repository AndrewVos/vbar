class ScreenThief {
  public static void steal(Gdk.Window window, Gdk.Rectangle monitor_dimensions, int height) {
    long struts[12];

    struts = { 0, 0, height , 0, /* strut-left, strut-right, strut-top, strut-bottom */
      0, 0, /* strut-left-start-y, strut-left-end-y */
      0, 0, /* strut-right-start-y, strut-right-end-y */
      monitor_dimensions.x, monitor_dimensions.x + monitor_dimensions.width - 1, /* strut-top-start-x, strut-top-end-x */
      0, 0 }; /* strut-bottom-start-x, strut-bottom-end-x */

    var atom = Gdk.Atom.intern("_NET_WM_STRUT_PARTIAL", false);

    Gdk.property_change(window, atom, Gdk.Atom.intern("CARDINAL", false),
    32, Gdk.PropMode.REPLACE, (uint8[])struts, 12);
  }
}
