class Logger {
  public static void error(string message) {
    GLib.print("error: " + message + "\n");
  }

  public static void put(string message) {
    GLib.print(message + "\n");
  }
}
