struct AddMenuOptions {
  string block_name;
  string text;
  string command;

  public void remove_nulls() {
    if (block_name == null) {
      block_name = "";
    }
    if (text == null) {
      text = "";
    }
    if (command == null) {
      command = "";
    }
  }
}
