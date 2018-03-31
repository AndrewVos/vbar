class Block {
  private string name;
  private string text;
  private string click_command;
  private string command;
  private double interval;
  private Gtk.Label label;
  private Gtk.EventBox eventBox;

  public Block(Json.Object element) {
    if (element.has_member("name")) {
      this.name = element.get_string_member("name");
    }
    if (element.has_member("text")) {
      this.text = element.get_string_member("text");
    }
    if (element.has_member("click_command")) {
      this.click_command = element.get_string_member("click_command");
    }
    if (element.has_member("command")) {
      this.command = element.get_string_member("command");
    }
    if (element.has_member("interval")) {
      this.interval = element.get_double_member("interval");
    }
  }

  public Gtk.Widget widget() {
    if (this.eventBox != null) {
      return this.eventBox;
    }

    this.label = new Gtk.Label(this.text);
    this.label.get_style_context().add_class("block");
    this.label.get_style_context().add_class(name);

    this.eventBox = new Gtk.EventBox();
    this.eventBox.add(this.label);

    if (this.click_command != null) {
      string command = this.click_command;
      this.eventBox.button_release_event.connect(() => {
        Executor.execute(command);
        return true;
      });
    }

    if (this.interval > 0 && this.command != null) {
      this.start_updating();
    }

    return this.eventBox;
  }

  private void start_updating() {
    this.update_label();

    uint intervalMilliseconds = (uint)(this.interval * 1000);
    GLib.Timeout.add(intervalMilliseconds, () => { this.update_label(); return true; }, GLib.Priority.DEFAULT);
  }

  private void update_label() {
    string text = Executor.execute(this.command);
    this.label.set_text(text);
  }
}
