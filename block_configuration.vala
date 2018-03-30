class BlockConfiguration {
  public Block[] left;
  public Block[] center;
  public Block[] right;

  public BlockConfiguration() {}

  public static BlockConfiguration from_file(string path) {
    var parser = new Json.Parser();
    parser.load_from_file(path);
    var left = parser.get_root().get_object().get_array_member("left");
    var center = parser.get_root().get_object().get_array_member("center");
    var right = parser.get_root().get_object().get_array_member("right");

    var block_configuration = new BlockConfiguration();
    block_configuration.left = new Block[left.get_length()];
    block_configuration.center = new Block[center.get_length()];
    block_configuration.right = new Block[right.get_length()];

    for (var i = 0; i < left.get_length(); i++) {
      var element = left.get_object_element(i);
      block_configuration.left[i] = new Block(element);
    }

    for (var i = 0; i < center.get_length(); i++) {
      var element = center.get_object_element(i);
      block_configuration.center[i] = new Block(element);
    }

    for (var i = 0; i < right.get_length(); i++) {
      var element = right.get_object_element(i);
      block_configuration.right[i] = new Block(element);
    }

    return block_configuration;
  }
}

