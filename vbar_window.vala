public class VbarWindow : Gtk.ApplicationWindow {
  private Gdk.Monitor? monitor;
  private Panel panel;

  public VbarWindow(Gtk.Application application) {
    Object(
      application: application,
      app_paintable: true,
      decorated: false,
      resizable: false,
      skip_pager_hint: true,
      skip_taskbar_hint: true,
      type_hint: Gdk.WindowTypeHint.DOCK,
      vexpand: false
    );

    var config = GLib.Environment.get_user_config_dir();
    string config_path = config + "/vbar/config.json";
    if (!FileUtils.test(config_path, FileTest.EXISTS)) {
      Logger.error("Couldn't find config at \"" + config_path + "\"");
      this.application.quit();
      return;
    }

    this.panel = new Panel(this.screen);
    try {
      this.panel.load_from_path(config_path);
    } catch (GLib.Error error) {
      Logger.error(error.message);
      this.application.quit();
      return;
    }
    this.add(this.panel);

    this.monitor = this.get_display().get_primary_monitor();

    this.screen.size_changed.connect(update_panel_dimensions);
    this.screen.monitors_changed.connect(update_panel_dimensions);
    this.realize.connect(update_panel_dimensions);
  }

  private void update_panel_dimensions() {
    var monitor_dimensions = this.monitor.get_geometry();

    this.set_size_request(monitor_dimensions.width, -1);
    this.move(monitor_dimensions.x, monitor_dimensions.y);
    this.update_struts(monitor_dimensions);
  }

  private void update_struts(Gdk.Rectangle monitor_dimensions) {
    if (!this.get_realized()) {
      return;
    }

    long struts[12];

    var panel_height = this.panel.get_allocated_height();

    struts = { 0, 0, panel_height , 0, /* strut-left, strut-right, strut-top, strut-bottom */
      0, 0, /* strut-left-start-y, strut-left-end-y */
      0, 0, /* strut-right-start-y, strut-right-end-y */
      monitor_dimensions.x, monitor_dimensions.x + monitor_dimensions.width - 1, /* strut-top-start-x, strut-top-end-x */
      0, 0 }; /* strut-bottom-start-x, strut-bottom-end-x */

    var atom = Gdk.Atom.intern("_NET_WM_STRUT_PARTIAL", false);

    Gdk.property_change(this.get_window(), atom, Gdk.Atom.intern("CARDINAL", false),
    32, Gdk.PropMode.REPLACE, (uint8[])struts, 12);
  }
}
