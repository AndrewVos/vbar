public class VbarWindow : Gtk.ApplicationWindow {
  private Panel panel;
  private MessageServer message_server;
  private string all_css = "";
  private Gtk.CssProvider css_provider;
  private Block[] blocks;

  public VbarWindow(Gtk.Application application) {
    Object(
      application: application,
      app_paintable: true,
      decorated: false,
      resizable: false,
      skip_pager_hint: true,
      skip_taskbar_hint: true,
      type_hint: Gdk.WindowTypeHint.DOCK,
      vexpand: false
      );

    this.enable_transparency();
    this.create_panel();

    this.listen_for_messages();
    this.initialize_css();

    this.screen.size_changed.connect(update_dimensions);
    this.screen.monitors_changed.connect(update_dimensions);
    this.realize.connect(update_dimensions);

    this.execute_config();
  }

  private void listen_for_messages() {
    this.message_server = new MessageServer();

    this.message_server.add_css.connect((class_name, css) => {
      for (var i = 0; i < css.length; i++) {
        var style = "." + class_name + " { " + css[i] + "}";
        this.all_css += style + "\n";
      }
      try {
        css_provider.load_from_data(this.all_css);
      } catch (GLib.Error e) {
        Logger.error(e.message);
      }
    });

    this.message_server.add_block.connect((options) => {
      var block = new Block(options);
      this.blocks += block;

      if (options.location == BlockLocation.LEFT) {
        this.panel.container_left.add(block);
      } else if (options.location == BlockLocation.RIGHT) {
        this.panel.container_right.add(block);
      } else if (options.location == BlockLocation.CENTER) {
        this.panel.container_center.add(block);
      }

      this.update_dimensions();
    });

    this.message_server.add_menu.connect((options) => {
      for (var i = 0; i < this.blocks.length; i++) {
        var block = this.blocks[i];
        if (block.block_name == options.block_name) {
          if (block.menu == null) {
            block.menu = new Gtk.Menu();
            block.menu.get_style_context().add_class("menu");

            block.button_release_event.connect(() => {
              block.menu.popup_at_widget(block, Gdk.Gravity.SOUTH_WEST, Gdk.Gravity.NORTH_WEST, null);
              return true;
            });
          }

          Gtk.MenuItem item = new Gtk.MenuItem.with_label(options.text);
          item.activate.connect(() => {
            Executor.execute(options.command);
          });
          block.menu.add(item);
          block.menu.show_all();
        }
      }
    });

    this.message_server.update.connect((block_name) => {
      if (this.blocks != null) {
        for (var i = 0; i < this.blocks.length; i++) {
          var block = this.blocks[i];
          if (block.block_name == block_name) {
            block.update_label();
          }
        }
      }
    });
  }

  private void create_panel() {
    this.panel = new Panel();
    this.add(this.panel);
  }

  private void update_dimensions() {
    this.show_all();

    var monitor = this.get_display().get_primary_monitor();
    var monitor_dimensions = monitor.get_geometry();

    this.set_size_request(monitor_dimensions.width, -1);
    this.move(monitor_dimensions.x, monitor_dimensions.y);

    ScreenThief.steal(
      this.get_window(),
      monitor_dimensions,
      this.panel.get_allocated_height()
      );
  }

  private void initialize_css() {
    this.css_provider = new Gtk.CssProvider();
    Gtk.StyleContext.add_provider_for_screen(this.screen, this.css_provider, Gtk.STYLE_PROVIDER_PRIORITY_USER);
  }

  private void enable_transparency() {
    var screen = this.get_screen();
    var visual = screen.get_rgba_visual();

    if(visual != null && screen.is_composited()) {
      this.set_visual(visual);
    }
  }

  private void execute_config() {
    var config = GLib.Environment.get_user_config_dir();
    string config_path = config + "/vbar/vbarrc";
    if (!FileUtils.test(config_path, FileTest.EXISTS)) {
      Logger.error("Couldn't find config at \"" + config_path + "\"");
      this.application.quit();
      return;
    }

    Executor.execute_async(config_path);
  }
}
