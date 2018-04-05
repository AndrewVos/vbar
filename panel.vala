class Panel : Gtk.Grid {
  private Block last_left;
  private Block last_center;
  private Block last_right;

  public Panel() {
    Object();
    this.get_style_context().add_class("panel");
  }

  public void add_left(Block block) {
    block.halign = Gtk.Align.START;

    this.attach_next_to(block, this.last_left, Gtk.PositionType.RIGHT);
    this.last_left = block;
  }

  public void add_center(Block block) {
    block.label.ellipsize = Pango.EllipsizeMode.END;
    block.halign = Gtk.Align.CENTER;
    block.hexpand = true;

    if (this.last_center == null) {
      this.attach_next_to(block, this.last_left, Gtk.PositionType.RIGHT);
    } else {
      this.attach_next_to(block, this.last_center, Gtk.PositionType.RIGHT);
    }
    this.last_center = block;
  }

  public void add_right(Block block) {
    block.halign = Gtk.Align.END;

    if (this.last_right == null) {
      if (this.last_center == null) {
        this.attach_next_to(block, this.last_left, Gtk.PositionType.RIGHT);
      } else {
        this.attach_next_to(block, this.last_center, Gtk.PositionType.RIGHT);
      }
    } else {
      this.attach_next_to(block, this.last_right, Gtk.PositionType.RIGHT);
    }
    this.last_right = block;
  }
}
