[DBus (name = "com.andrewvos.Vbar")]
class MessageServer : Object {
  public MessageServer() {
    Bus.own_name (BusType.SESSION, "com.andrewvos.Vbar", BusNameOwnerFlags.NONE,
                  this.on_bus_aquired,
                  () => {},
                  () => stderr.printf ("Could not aquire name\n"));
  }

  void on_bus_aquired (DBusConnection conn) {
    try {
      conn.register_object ("/com/andrewvos/vbar", this);
    } catch (IOError e) {
      stderr.printf ("Could not register service\n");
    }
  }

  public signal bool update(string block_name);
  public bool trigger_update(string block_name) {
    return update(block_name);
  }

  public static bool send_update(string block_name) {
    Message message;

    try {
      message = Bus.get_proxy_sync(
        BusType.SESSION,
        "com.andrewvos.Vbar",
        "/com/andrewvos/vbar"
      );
    } catch {
      Logger.error("Couldn't connect to vbar. Is it running?");
      return false;
    }

    try {
      return message.trigger_update(block_name);
    } catch {
      Logger.error("Couldn't send message to vbar. Is it running?");
      return false;
    }
  }
}

[DBus (name = "com.andrewvos.Vbar")]
interface Message : Object {
  public abstract bool trigger_update (string block_name) throws IOError;
}

