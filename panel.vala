class Panel : Gtk.Box {
  public Gtk.Box container_left;
  public Gtk.Box container_center;
  public Gtk.Box container_right;

  public Panel() {
    Object(orientation: Gtk.Orientation.HORIZONTAL, spacing: 0);
    this.get_style_context().add_class("panel");

    this.container_left = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_start(this.container_left);

    this.container_center = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_start(this.container_center, true, true);

    this.container_right = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 0);
    this.pack_end(this.container_right, false, true);
  }
}
