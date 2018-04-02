public class VbarApplication : Gtk.Application {
  public VbarApplication() {
    Object(application_id: "com.andrewvos.vbar", flags: ApplicationFlags.FLAGS_NONE);
  }

  protected override void activate() {
    var window = new VbarWindow(this);
    window.show_all();
  }

  private static string? update_block_name = null;

  private const GLib.OptionEntry[] options = {
    { "update", 'u', 0, OptionArg.STRING, ref update_block_name, "Trigger a block update", "BLOCK_NAME" },
    { null }
  };

  public static int main(string[] args) {
    try {
      var option_context = new OptionContext("- a bar");
      option_context.set_help_enabled(true);
      option_context.add_main_entries(options, null);
      option_context.parse(ref args);
    } catch (OptionError e) {
      Logger.put("Run 'vbar --help' to see a full list of available command line options.");
      return 0;
    }

    if (update_block_name != null) {
      bool result = MessageServer.send_update(update_block_name);
      if (result) {
        return 0;
      }
      return 1;
    }

    var application = new VbarApplication();
    return application.run(args);
  }
}
