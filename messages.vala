[DBus (name = "com.andrewvos.Vbar")]
class MessageServer : Object {
  public MessageServer() {
    Bus.own_name (BusType.SESSION, "com.andrewvos.Vbar", BusNameOwnerFlags.NONE,
                  this.on_bus_aquired,
                  () => {},
                  () => Logger.error("Could not acquire bus name."));
  }

  private void on_bus_aquired(DBusConnection conn) {
    try {
      conn.register_object ("/com/andrewvos/vbar", this);
    } catch (IOError e) {
      Logger.error("Could not register service.");
    }
  }

  private static Message? get_proxy() {
    try {
      Message? message;
      message = Bus.get_proxy_sync(
        BusType.SESSION,
        "com.andrewvos.Vbar",
        "/com/andrewvos/vbar"
        );
      return message;
    } catch {
      Logger.error("Couldn't connect to vbar. Is it running?");
      return null;
    }
  }

  public signal void add_css(string class_name, string[] css);
  public void trigger_add_css(string class_name, string[] css) throws GLib.Error {
    this.add_css(class_name, css);
  }

  public signal void add_block(AddBlockOptions options);
  public void trigger_add_block(AddBlockOptions options) throws GLib.Error {
    this.add_block(options);
  }

  public signal void add_menu(AddMenuOptions options);
  public void trigger_add_menu(AddMenuOptions options) throws GLib.Error {
    this.add_menu(options);
  }

  public signal void update(string block_name);
  public void trigger_update(string block_name) throws GLib.Error {
    this.update(block_name);
  }

  public static bool send_add_css(string class_name, string[] css) {
    var message = get_proxy();
    if (message == null) {
      return false;
    }

    try {
      message.trigger_add_css(class_name, css);
    } catch {
      Logger.error("Couldn't send message to vbar. Is it running?");
      return false;
    }

    return true;
  }

  public static bool send_add_block(AddBlockOptions options) {
    var message = get_proxy();
    if (message == null) {
      return false;
    }

    options.remove_nulls();

    try {
      message.trigger_add_block(options);
    } catch {
      Logger.error("Couldn't send message to vbar. Is it running?");
      return false;
    }

    return true;
  }

  public static bool send_add_menu(AddMenuOptions options) {
    var message = get_proxy();
    if (message == null) {
      return false;
    }

    options.remove_nulls();

    try {
      message.trigger_add_menu(options);
    } catch {
      Logger.error("Couldn't send message to vbar. Is it running?");
      return false;
    }

    return true;
  }

  public static bool send_update(string block_name) {
    var message = get_proxy();
    if (message == null) {
      return false;
    }

    try {
      message.trigger_update(block_name);
    } catch {
      Logger.error("Couldn't send message to vbar. Is it running?");
      return false;
    }

    return true;
  }
}

[DBus (name = "com.andrewvos.Vbar")]
interface Message : Object {
  public abstract void trigger_add_css(string class_name, string[] css) throws GLib.Error;
  public abstract void trigger_add_block(AddBlockOptions options) throws GLib.Error;
  public abstract void trigger_add_menu(AddMenuOptions options) throws GLib.Error;
  public abstract void trigger_update(string block_name) throws GLib.Error;
}
