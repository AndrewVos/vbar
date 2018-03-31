public class VbarWindow : Gtk.Window {
  private Gdk.Monitor? monitor;
  private int monitor_width;
  private int monitor_height;
  private int monitor_x;
  private int monitor_y;

  private Gtk.Box box;
  private Gtk.Box boxLeft;
  private Gtk.Box boxCenter;
  private Gtk.Box boxRight;

  public VbarWindow() {
    Object(
      app_paintable: true,
      decorated: false,
      resizable: false,
      skip_pager_hint: true,
      skip_taskbar_hint: true,
      type_hint: Gdk.WindowTypeHint.DOCK,
      vexpand: false
    );

    monitor = this.get_display().get_primary_monitor();

    this.screen.size_changed.connect(update_panel_dimensions);
    this.screen.monitors_changed.connect(update_panel_dimensions);
    this.realize.connect(on_realize);

    var css_provider = new Gtk.CssProvider();
    try {
      css_provider.load_from_path("styles.css");
    } catch (GLib.Error error) {
      Logger.fatal(error.message);
    }
    Gtk.StyleContext.add_provider_for_screen(screen, css_provider, Gtk.STYLE_PROVIDER_PRIORITY_USER);

    var block_configuration = BlockConfiguration.from_file("config.json");

    this.box = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.add(this.box);

    this.boxLeft = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.boxLeft.get_style_context().add_class("box");
    this.box.pack_start(this.boxLeft);

    this.boxCenter = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.boxCenter.get_style_context().add_class("box");
    this.box.pack_start(this.boxCenter, true, true);

    this.boxRight = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.boxRight.get_style_context().add_class("box");
    this.box.pack_end(this.boxRight, false, true);

    for (var i = 0; i < block_configuration.left.length; i++) {
      this.boxLeft.add(block_configuration.left[i].widget());
    }
    for (var i = 0; i < block_configuration.center.length; i++) {
      this.boxCenter.add(block_configuration.center[i].widget());
    }
    for (var i = 0; i < block_configuration.right.length; i++) {
      this.boxRight.add(block_configuration.right[i].widget());
    }
  }

  private void on_realize() {
    update_panel_dimensions();
  }

  private void update_panel_dimensions() {
    var monitor_dimensions = this.monitor.get_geometry();

    monitor_width = monitor_dimensions.width;
    monitor_height = monitor_dimensions.height;

    this.set_size_request(monitor_width, -1);

    monitor_x = monitor_dimensions.x;
    monitor_y = monitor_dimensions.y;

    this.move(monitor_x, monitor_y);

    update_struts();
  }

  private void update_struts() {
    if (!this.get_realized()) {
      return;
    }

    long struts[12];

    var box_height = this.boxLeft.get_allocated_height();

    struts = { 0, 0, box_height , 0, /* strut-left, strut-right, strut-top, strut-bottom */
      0, 0, /* strut-left-start-y, strut-left-end-y */
      0, 0, /* strut-right-start-y, strut-right-end-y */
      monitor_x, monitor_x + monitor_width - 1, /* strut-top-start-x, strut-top-end-x */
      0, 0 }; /* strut-bottom-start-x, strut-bottom-end-x */

    var atom = Gdk.Atom.intern("_NET_WM_STRUT_PARTIAL", false);

    Gdk.property_change(this.get_window(), atom, Gdk.Atom.intern("CARDINAL", false),
    32, Gdk.PropMode.REPLACE, (uint8[])struts, 12);
  }
}
