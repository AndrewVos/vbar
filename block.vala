class Block : Gtk.EventBox {
  private string block_name;
  private string block_text;
  private string block_click_command;
  private string block_command;
  private double block_interval;

  private Gtk.Label label;

  public Block(Json.Object element) {
    Object();

    if (element.has_member("name")) {
      this.block_name = element.get_string_member("name");
    }
    if (element.has_member("text")) {
      this.block_text = element.get_string_member("text");
    }
    if (element.has_member("click_command")) {
      this.block_click_command = element.get_string_member("click_command");
    }
    if (element.has_member("command")) {
      this.block_command = element.get_string_member("command");
    }
    if (element.has_member("interval")) {
      this.block_interval = element.get_double_member("interval");
    }

    this.label = new Gtk.Label(this.block_text);
    this.label.get_style_context().add_class("block");
    this.label.get_style_context().add_class(this.block_name);
    this.add(this.label);

    if (this.block_click_command != null) {
      string command = this.block_click_command;
      this.button_release_event.connect(() => {
        Executor.execute(command);
        return true;
      });
    }

    if (this.block_interval > 0 && this.block_command != null) {
      this.start_updating();
    }
  }

  private void start_updating() {
    this.update_label();

    uint intervalMilliseconds = (uint)(this.block_interval * 1000);
    GLib.Timeout.add(intervalMilliseconds, () => { this.update_label(); return true; }, GLib.Priority.DEFAULT);
  }

  private void update_label() {
    string text = Executor.execute(this.block_command);
    this.label.set_text(text);
  }
}
