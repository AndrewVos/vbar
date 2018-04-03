struct AddBlockOptions {
  BlockLocation location;
  string name;
  string text;
  string command;
  string tail_command;
  string click_command;
  double interval;

  public void remove_nulls() {
    if (name == null) {
      name = "";
    }
    if (text == null) {
      text = "";
    }
    if (command == null) {
      command = "";
    }
    if (tail_command == null) {
      tail_command = "";
    }
    if (click_command == null) {
      click_command = "";
    }
  }
}
