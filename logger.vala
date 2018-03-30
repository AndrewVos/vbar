class Logger {
  public static void error(string message) {
    GLib.print("error: " + message + "\n");
  }

  public static void fatal(string message) {
    GLib.print("fatal: " + message + "\n");
    Gtk.main_quit();
  }
}
