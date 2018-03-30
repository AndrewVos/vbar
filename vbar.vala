public class VbarApplication : Gtk.Application {
  public VbarApplication() {
    Object(application_id: "vbar", flags: ApplicationFlags.FLAGS_NONE);
  }

  protected override void activate() {
    this.hold();
    var window = new VbarWindow();
    window.show_all();
  }

  public static int main(string[] args) {
    var application = new VbarApplication();
    return application.run(args);
  }
}

public class VbarWindow : Gtk.Window {
  private int monitor_number;
  private int monitor_width;
  private int monitor_height;
  private int monitor_x;
  private int monitor_y;
  private int panel_height;

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

    monitor_number = screen.get_primary_monitor();

    var style_context = get_style_context();
    style_context.add_class(Gtk.STYLE_CLASS_MENUBAR);

    this.screen.size_changed.connect(update_panel_dimensions);
    this.screen.monitors_changed.connect(update_panel_dimensions);
    this.realize.connect(on_realize);

    panel_height = 20;

    var css_provider = new Gtk.CssProvider();

    string path = "styles.css";
    css_provider.load_from_path(path);
    Gtk.StyleContext.add_provider_for_screen(screen, css_provider, Gtk.STYLE_PROVIDER_PRIORITY_USER);

    var box = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 10);
    this.add(box);

    var label = new Gtk.Label("\uf240");
    label.get_style_context().add_class("my_class");
    box.add(label);

    var label2 = new Gtk.Label("100%");
    label2.get_style_context().add_class("my_class");
    box.add(label2);
  }

  private void on_realize() {
    update_panel_dimensions();
  }

  private void update_panel_dimensions() {
    Gdk.Rectangle monitor_dimensions;
    this.screen.get_monitor_geometry(monitor_number, out monitor_dimensions);

    monitor_width = monitor_dimensions.width;
    monitor_height = monitor_dimensions.height;

    this.set_size_request(monitor_width, panel_height);

    monitor_x = monitor_dimensions.x;
    monitor_y = monitor_dimensions.y;

    this.move(monitor_x, monitor_y);

    update_struts();
  }

  private void update_struts() {
    if (!this.get_realized()) {
      return;
    }

    var monitor = monitor_number == -1 ? this.screen.get_primary_monitor() : monitor_number;

    var position_top = monitor_y;

    Gdk.Atom atom;
    Gdk.Rectangle primary_monitor_rect;

    long struts[12];

    this.screen.get_monitor_geometry(monitor, out primary_monitor_rect);

    // We need to manually include the scale factor here as GTK gives us unscaled sizes for widgets
    struts = { 0, 0, panel_height , 0, /* strut-left, strut-right, strut-top, strut-bottom */
      0, 0, /* strut-left-start-y, strut-left-end-y */
      0, 0, /* strut-right-start-y, strut-right-end-y */
      monitor_x, monitor_x + monitor_width - 1, /* strut-top-start-x, strut-top-end-x */
      0, 0 }; /* strut-bottom-start-x, strut-bottom-end-x */

    atom = Gdk.Atom.intern("_NET_WM_STRUT_PARTIAL", false);

    Gdk.property_change(this.get_window(), atom, Gdk.Atom.intern("CARDINAL", false),
    32, Gdk.PropMode.REPLACE, (uint8[])struts, 12);
  }
}
