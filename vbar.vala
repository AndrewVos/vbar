public class VbarApplication : Gtk.Application {
  public VbarApplication() {
    Object(application_id: "com.andrewvos.vbar", flags: ApplicationFlags.FLAGS_NONE);
  }

  private VbarWindow window;

  protected override void activate() {
    this.window = new VbarWindow(this);
  }

  private static string update_block_name = null;

  private static string class_name = null;
  [CCode (array_length = false, array_null_terminated = true)]
  private static string[]? css = null;

  private static string name = null;
  private static bool left = false;
  private static bool right = false;
  private static bool center = false;
  private static string text = null;
  private static string command = null;
  private static string tail_command = null;
  private static string click_command = null;
  private static double interval = 0;

  private static string menu_block_name = null;
  private static string menu_text = null;
  private static string menu_command = null;

  [CCode (array_length = false, array_null_terminated = true)]
  private static string[]? hide_block_names = null;

  [CCode (array_length = false, array_null_terminated = true)]
  private static string[]? show_block_names = null;

  public static int main(string[] args) {
    GLib.Intl.setlocale();

    var option_context = new OptionContext("- a bar");
    option_context.set_help_enabled(true);

    var vbarCommand = args[1];
    string[] new_args = new string[]{};
    for (var i = 0; i < args.length; i++) {
      if (i != 1) {
        new_args += args[i];
      }
    }
    args = new_args;

    if (vbarCommand == "add-css") {
      setup_add_css_options(option_context);
    } else if (vbarCommand == "add-block") {
      setup_add_block_options(option_context);
    } else if (vbarCommand == "add-menu") {
      setup_add_menu_options(option_context);
    } else if (vbarCommand == "update") {
      setup_update_options(option_context);
    } else if (vbarCommand == "hide") {
      setup_hide_options(option_context);
    } else if (vbarCommand == "show") {
      setup_show_options(option_context);
    } else if (vbarCommand == null) {
      var application = new VbarApplication();
      return application.run(args);
    } else {
      Logger.put("Run 'vbar --help' to see a full list of available command line options.");
      return 1;
    }

    try {
      option_context.parse(ref args);
    } catch (OptionError e) {
      Logger.put("Run 'vbar --help' to see a full list of available command line options.");
      return 1;
    }

    if (vbarCommand == "add-css") {
      return add_css(option_context);
    } else if (vbarCommand == "add-block") {
      return add_block(option_context);
    } else if (vbarCommand == "add-menu") {
      return add_menu(option_context);
    } else if (vbarCommand == "update") {
      return update(option_context);
    } else if (vbarCommand == "hide") {
      return hide(option_context);
    } else if (vbarCommand == "show") {
      return show(option_context);
    }

    return 1;
  }

  private static void setup_add_css_options(OptionContext context) {
    const GLib.OptionEntry[] options = {
      { "class", 0, 0, OptionArg.STRING, ref class_name, "Class name", "CLASS" },
      { "css", 0, 0, OptionArg.STRING_ARRAY, ref css, "CSS", "CSS..." },
      { null }
    };

    context.add_main_entries(options, "Add css");
  }

  private static void setup_add_block_options(OptionContext context) {
    const GLib.OptionEntry[] options = {
      { "left", 0, 0, OptionArg.NONE, ref left, "Add the block to the left", null },
      { "right", 0, 0, OptionArg.NONE, ref right, "Add the block to the right", null },
      { "center", 0, 0, OptionArg.NONE, ref center, "Add the block to the center", null },
      { "name", 0, 0, OptionArg.STRING, ref name, "Block name", "NAME" },
      { "text", 0, 0, OptionArg.STRING, ref text, "Block text", "TEXT" },
      { "command", 0, 0, OptionArg.STRING, ref command, "Command to execute", "COMMAND" },
      { "tail-command", 0, 0, OptionArg.STRING, ref tail_command, "Command to tail", "COMMAND" },
      { "click-command", 0, 0, OptionArg.STRING, ref click_command, "Command to execute when clicking on the block", "COMMAND" },
      { "interval", 0, 0, OptionArg.DOUBLE, ref interval, "Interval between command executions", "INTERVAL" },
      { null }
    };

    context.add_main_entries(options, "Add a new block");
  }

  private static void setup_add_menu_options(OptionContext context) {
    const GLib.OptionEntry[] options = {
      { "block", 0, 0, OptionArg.STRING, ref menu_block_name, "Block name to add the menu to", "BLOCK_NAME" },
      { "text", 0, 0, OptionArg.STRING, ref menu_text, "The menu text", "TEXT" },
      { "command", 0, 0, OptionArg.STRING, ref menu_command, "The command to execute when clicking the menu item", "COMMAND" },
      { null }
    };

    context.add_main_entries(options, "Add a new block");
  }

  private static void setup_update_options(OptionContext context) {
    const GLib.OptionEntry[] options = {
      { "block", 0, 0, OptionArg.STRING, ref update_block_name, "Block to update", "BLOCK_NAME" },
      { null }
    };

    context.add_main_entries(options, "Update a block");
  }

  private static void setup_hide_options(OptionContext context) {
    const GLib.OptionEntry[] options = {
      { "block", 0, 0, OptionArg.STRING_ARRAY, ref hide_block_names, "Blocks to hide", "BLOCK_NAME..." },
      { null }
    };

    context.add_main_entries(options, "Hide blocks");
  }

  private static void setup_show_options(OptionContext context) {
    const GLib.OptionEntry[] options = {
      { "block", 0, 0, OptionArg.STRING_ARRAY, ref show_block_names, "Blocks to show", "BLOCK_NAMES" },
      { null }
    };

    context.add_main_entries(options, "Show blocks");
  }

  private static int add_css(OptionContext context) {
    if (class_name == null || css == null) {
      Logger.put(context.get_help(true, null));
      return 1;
    }

    bool result = MessageServer.send_add_css(class_name, css);
    if (result) {
      return 0;
    }
    return 1;
  }

  private static int add_block(OptionContext context) {
    if (name == null) {
      Logger.put(context.get_help(true, null));
      return 1;
    }

    var location = BlockLocation.LEFT;
    if (right) {
      location = BlockLocation.RIGHT;
    } else if (center) {
      location = BlockLocation.CENTER;
    }

    var options = AddBlockOptions() {
      location = location,
      name = name,
      text = text,
      command = command,
      tail_command = tail_command,
      click_command = click_command,
      interval = interval
    };

    bool result = MessageServer.send_add_block(options);
    if (result) {
      return 0;
    }
    return 1;
  }

  private static int add_menu(OptionContext context) {
    if (menu_block_name == null) {
      Logger.put(context.get_help(true, null));
      return 1;
    }

    var options = AddMenuOptions() {
      block_name = menu_block_name,
      text = menu_text,
      command = menu_command
    };

    bool result = MessageServer.send_add_menu(options);
    if (result) {
      return 0;
    }
    return 1;
  }

  private static int update(OptionContext context) {
    if (update_block_name == null) {
      Logger.put(context.get_help(true, null));
      return 1;
    }

    bool result = MessageServer.send_update(update_block_name);
    if (result) {
      return 0;
    }
    return 1;
  }

  private static int hide(OptionContext context) {
    if (hide_block_names == null) {
      Logger.put(context.get_help(true, null));
      return 1;
    }

    bool result = MessageServer.send_hide(hide_block_names);
    if (result) {
      return 0;
    }
    return 1;
  }

  private static int show(OptionContext context) {
    if (show_block_names == null) {
      Logger.put(context.get_help(true, null));
      return 1;
    }

    bool result = MessageServer.send_show(show_block_names);
    if (result) {
      return 0;
    }
    return 1;
  }
}
