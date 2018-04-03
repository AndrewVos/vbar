class Block : Gtk.EventBox {
  public string block_name;
  private AddBlockOptions options;

  private Gtk.Label label;
  public Gtk.Menu menu;

  public Block(AddBlockOptions options) {
    Object();

    this.options = options;
    this.block_name = options.name;

    this.label = new Gtk.Label(this.options.text);
    this.label.ellipsize = Pango.EllipsizeMode.END;
    this.label.get_style_context().add_class("block");
    this.label.get_style_context().add_class(this.block_name);
    this.add(this.label);

    if (not_null_or_empty(options.click_command)) {
      this.button_release_event.connect(() => {
        Executor.execute(options.click_command);
        return true;
      });
    }

    if (not_null_or_empty(this.options.command)) {
      this.update_label();

      if (this.options.interval > 0) {
        this.start_updating();
      }
    } else if (not_null_or_empty(this.options.tail_command)) {
      var executor = new Executor();
      executor.line.connect((line) => {
        this.label.set_text(line);
      });
      executor.execute_async_tail(this.options.tail_command);
    }
  }

  private bool not_null_or_empty(string value) {
    return !(value == null || value == "");
  }

  private void start_updating() {
    uint intervalMilliseconds = (uint)(this.options.interval * 1000);
    GLib.Timeout.add(intervalMilliseconds, () => { this.update_label(); return true; }, GLib.Priority.DEFAULT);
  }

  public void update_label() {
    string text = Executor.execute(this.options.command);
    this.label.set_text(text);
  }
}
