class Executor {
  public static string execute(string command) {
    try {
      string std_out;
      int exitCode;
      Process.spawn_sync(null, { "/bin/bash", "-c", command }, null, SpawnFlags.SEARCH_PATH, null, out std_out, null, out exitCode);
      if (exitCode > 0) {
        return "ERROR";
      }
      return std_out.replace("\n", "");
    }
    catch (Error e) {
      Logger.error(e.message);
      return "";
    }
  }

  public signal void line(string line);

  public void execute_async(string command) {
    Pid child_pid;
    int standard_output;

    try {
      Process.spawn_async_with_pipes(
        null,
        { "/bin/bash", "-c", command},
        null,
        SpawnFlags.SEARCH_PATH | SpawnFlags.DO_NOT_REAP_CHILD,
        null,
        out child_pid,
        null,
        out standard_output,
        null
      );
    } catch(Error e) {
      this.line("ERROR");
    }

    IOChannel output = new IOChannel.unix_new(standard_output);
    output.add_watch(IOCondition.IN | IOCondition.HUP, (channel, condition) => {
      if (condition == IOCondition.HUP) {
        Logger.error("Command finished executing, but shouldn't have.");
        return false;
      }

      try {
        string line;
        channel.read_line(out line, null, null);
        this.line(line.replace("\n", ""));
      } catch (IOChannelError e) {
        Logger.error("Couldn't read output of command.");
        return false;
      } catch (ConvertError e) {
        Logger.error("Couldn't read output of command.");
        return false;
      }

      return true;
    });

    ChildWatch.add(child_pid, (pid, status) => {
      Process.close_pid(pid);
    });
  }
}
