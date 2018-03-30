public class VbarApplication : Gtk.Application {
  public VbarApplication() {
    Object(application_id: "com.andrewvos.vbar", flags: ApplicationFlags.FLAGS_NONE);
  }

  protected override void activate() {
    this.hold();
    var window = new VbarWindow();
    window.show_all();
  }

  public static int main(string[] args) {
    var application = new VbarApplication();
    return application.run(args);
  }
}
