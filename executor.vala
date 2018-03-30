class Executor {
  public static string execute(string command) {
    try {
      string std_out;
      int exitCode;
      Process.spawn_sync(null, { "/bin/bash", "-c", command }, null, SpawnFlags.SEARCH_PATH, null, out std_out, null, out exitCode);
      if (exitCode > 0) {
        return "ERROR";
      }
      return std_out.strip();
    }
    catch (Error e) {
      Logger.error(e.message);
      return "";
    }
  }
}
