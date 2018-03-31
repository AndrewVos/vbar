class Panel : Gtk.Box {
  private Gtk.Box containerLeft;
  private Gtk.Box containerCenter;
  private Gtk.Box containerRight;

  public Panel() {
    Object(orientation: Gtk.Orientation.HORIZONTAL, spacing: 0);
  }

  public void load_from_path(string path) throws GLib.Error {
    this.get_style_context().add_class("box");

    this.containerLeft = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_start(this.containerLeft);

    this.containerCenter = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_start(this.containerCenter, true, true);

    this.containerRight = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_end(this.containerRight, false, true);

    var parser = new Json.Parser();
    parser.load_from_file(path);

    var left = parser.get_root().get_object().get_array_member("left");
    var center = parser.get_root().get_object().get_array_member("center");
    var right = parser.get_root().get_object().get_array_member("right");

    for (var i = 0; i < left.get_length(); i++) {
      var element = left.get_object_element(i);
      this.containerLeft.add(new Block(element));
    }

    for (var i = 0; i < center.get_length(); i++) {
      var element = center.get_object_element(i);
      this.containerCenter.add(new Block(element));
    }

    for (var i = 0; i < right.get_length(); i++) {
      var element = right.get_object_element(i);
      this.containerRight.add(new Block(element));
    }
  }

  public int height() {
    return this.containerLeft.get_allocated_height();
  }
}
