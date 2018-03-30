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

  private Gtk.Box box;

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

    this.screen.size_changed.connect(update_panel_dimensions);
    this.screen.monitors_changed.connect(update_panel_dimensions);
    this.realize.connect(on_realize);

    var css_provider = new Gtk.CssProvider();
    css_provider.load_from_path("styles.css");
    Gtk.StyleContext.add_provider_for_screen(screen, css_provider, Gtk.STYLE_PROVIDER_PRIORITY_USER);

    var bar_configuration = load_bar_configuration();

    box = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 10);
    box.get_style_context().add_class("box");
    this.add(box);

    for (var i = 0; i < bar_configuration.bars.length; i++) {
      var bar = bar_configuration.bars[i];
      var label = new Gtk.Label(bar.name);
      label.get_style_context().add_class(bar.name);
      box.add(label);
    }
  }

  private string execute_command_sync_get_output (string cmd)
  {
    try {
      int exitCode;
      string std_out;
      Process.spawn_command_line_sync(cmd, out std_out, null, out exitCode);
      return std_out;
    }
    catch (Error e){
      log_error (e.message);
      return "";
    }
  }

  private void log_error(string message) {
    GLib.print("error: " + message + "\n");
  }

  private BarConfiguration load_bar_configuration() {
    var parser = new Json.Parser();
    parser.load_from_file("config.json");
    var bars = parser.get_root().get_array();

    var bar_configuration = new BarConfiguration();
    bar_configuration.bars = new Bar[bars.get_length()];

    for (var i = 0; i < bars.get_length(); i++) {
      var element = bars.get_object_element(i);
      bar_configuration.bars[i] = new Bar();
      bar_configuration.bars[i].name = element.get_string_member("name");
    }

    return bar_configuration;
  }

  private void on_realize() {
    update_panel_dimensions();
  }

  private void update_panel_dimensions() {
    Gdk.Rectangle monitor_dimensions;
    this.screen.get_monitor_geometry(monitor_number, out monitor_dimensions);

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

    var monitor = monitor_number == -1 ? this.screen.get_primary_monitor() : monitor_number;

    var position_top = monitor_y;

    Gdk.Atom atom;
    Gdk.Rectangle primary_monitor_rect;

    long struts[12];

    this.screen.get_monitor_geometry(monitor, out primary_monitor_rect);

    var box_height = this.box.get_allocated_height();

    struts = { 0, 0, box_height , 0, /* strut-left, strut-right, strut-top, strut-bottom */
      0, 0, /* strut-left-start-y, strut-left-end-y */
      0, 0, /* strut-right-start-y, strut-right-end-y */
      monitor_x, monitor_x + monitor_width - 1, /* strut-top-start-x, strut-top-end-x */
      0, 0 }; /* strut-bottom-start-x, strut-bottom-end-x */

    atom = Gdk.Atom.intern("_NET_WM_STRUT_PARTIAL", false);

    Gdk.property_change(this.get_window(), atom, Gdk.Atom.intern("CARDINAL", false),
    32, Gdk.PropMode.REPLACE, (uint8[])struts, 12);
  }
}

class BarConfiguration {
  public Bar[] bars;

  public BarConfiguration() {}
}

class Bar {
  public string name;
  public string command;
  public Bar() {}
}
