class Panel : Gtk.Box {
  private Gtk.Box containerLeft;
  private Gtk.Box containerCenter;
  private Gtk.Box containerRight;
  private Gdk.Screen screen;

  public Panel(Gdk.Screen screen) {
    Object(orientation: Gtk.Orientation.HORIZONTAL, spacing: 0);
    this.screen = screen;
  }

  private void add_blocks_to_container(Gtk.Box container, Json.Array array) {
    for (var i = 0; i < array.get_length(); i++) {
      var element = array.get_object_element(i);
      container.add(new Block(element));
    }
  }

  public void load_from_path(string path) throws GLib.Error {
    this.get_style_context().add_class("panel");

    this.containerLeft = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_start(this.containerLeft);

    this.containerCenter = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_start(this.containerCenter, true, true);

    this.containerRight = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_end(this.containerRight, false, true);

    var parser = new Json.Parser();
    parser.load_from_file(path);

    string css = "";
    var all_styles = parser.get_root().get_object().get_object_member("styles");
    foreach(var css_class in all_styles.get_members()) {
      var styles = all_styles.get_array_member(css_class);
      for (var i = 0; i < styles.get_length(); i++) {
        var style = "." + css_class + " { " + styles.get_string_element(i) + "}";
        css += style + "\n";
      }
    }
    var css_provider = new Gtk.CssProvider();
    Gtk.StyleContext.add_provider_for_screen(screen, css_provider, Gtk.STYLE_PROVIDER_PRIORITY_USER);
    css_provider.load_from_data(css);

    var blocks = parser.get_root().get_object().get_object_member("blocks");

    this.add_blocks_to_container(this.containerLeft, blocks.get_array_member("left"));
    this.add_blocks_to_container(this.containerCenter, blocks.get_array_member("center"));
    this.add_blocks_to_container(this.containerRight, blocks.get_array_member("right"));
  }
}
