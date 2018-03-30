class Block {
  private string name;
  private string text;
  private string command;
  private double interval;
  public Gtk.Label label;

  public Block(Json.Object element) {
    if (element.has_member("name")) {
      this.name = element.get_string_member("name");
    }
    if (element.has_member("text")) {
      this.text = element.get_string_member("text");
    }
    if (element.has_member("command")) {
      this.command = element.get_string_member("command");
    }
    if (element.has_member("interval")) {
      this.interval = element.get_double_member("interval");
    }

    this.label = new Gtk.Label(this.text);
    this.label.get_style_context().add_class("block");
    this.label.get_style_context().add_class(name);

    if (this.interval > 0 && this.command != null) {
      this.start_updating();
    }
  }

  public void start_updating() {
    this.update_label();

    uint intervalMilliseconds = (uint)(this.interval * 1000);
    GLib.Timeout.add(intervalMilliseconds, () => { this.update_label(); return true; }, GLib.Priority.DEFAULT);
  }

  private void update_label() {
    string text = Executor.execute(this.command);
    this.label.set_text(text);
  }
}
