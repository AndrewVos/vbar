class Block : Gtk.EventBox {
  public string block_name;
  private string block_text;
  private string block_command;
  private string block_tail_command;
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
    if (element.has_member("command")) {
      this.block_command = element.get_string_member("command");
    }
    if (element.has_member("tail_command")) {
      this.block_tail_command = element.get_string_member("tail_command");
    }
    if (element.has_member("interval")) {
      this.block_interval = element.get_double_member("interval");
    }

    this.label = new Gtk.Label(this.block_text);
    this.label.get_style_context().add_class("block");
    this.label.get_style_context().add_class(this.block_name);
    this.add(this.label);

    if (element.has_member("menu-items")) {
      var menu_items = element.get_array_member("menu-items");
      var menu = new Gtk.Menu();
      menu.get_style_context().add_class("menu");
      for (var i = 0; i < menu_items.get_length(); i++) {
        var menu_item = menu_items.get_object_element(i);

        var text = menu_item.get_string_member("text");
        var command = menu_item.get_string_member("command");

        Gtk.MenuItem item = new Gtk.MenuItem.with_label(text);
        item.activate.connect(() => {
          Executor.execute(command);
        });
        menu.add(item);
      }
      menu.show_all();

      this.button_release_event.connect(() => {
        menu.popup_at_widget(this.label, Gdk.Gravity.SOUTH_WEST, Gdk.Gravity.NORTH_WEST, null);
        return true;
      });
    } else if (element.has_member("click-command")) {
      string command = element.get_string_member("click-command");
      this.button_release_event.connect(() => {
        Executor.execute(command);
        return true;
      });
    }

    if (this.block_command != null) {
      this.update_label();

      if (this.block_interval > 0) {
        this.start_updating();
      }
    } else if (this.block_tail_command != null) {
      var executor = new Executor();
      executor.line.connect((line) => {
        this.label.set_text(line);
      });
      executor.execute_async(this.block_tail_command);
    }
  }

  private void start_updating() {
    uint intervalMilliseconds = (uint)(this.block_interval * 1000);
    GLib.Timeout.add(intervalMilliseconds, () => { this.update_label(); return true; }, GLib.Priority.DEFAULT);
  }

  public void update_label() {
    string text = Executor.execute(this.block_command);
    this.label.set_text(text);
  }
}
