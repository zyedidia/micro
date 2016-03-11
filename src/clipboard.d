import std.process: execute, spawnProcess, pipe;

class Clipboard {
    static bool supported;
    version(OSX) {
        static bool init() {
            return supported = true;
        }

        static void write(string txt) {
            auto p = pipe();
            p.writeEnd.write(txt);
            spawnProcess("pbcopy", p.readEnd());
        }

        static string read() {
            return execute("pbpaste").output;
        }
    }

    version(linux) {
        import std.exception: collectException;
        string[] copyCmd;
        string[] pasteCmd;

        static bool init() {
            if (collectException(execute(["xsel", "-h"]))) {
                if (collectException(execute(["xclip", "-h"]))) {
                    return supported = false;
                } else  {
                    copyCmd = ["xclip", "-in", "-selection", "clipboard"];
                    pasteCmd = ["xclip", "-out", "-selection", "clipboard"];
                    return supported = true;
                }
            } else {
                copyCmd = ["xsel", "--input", "--clipboard"];
                pasteCmd = ["xsel", "--output", "--clipboard"];
                return supported = true;
            }
        }

        static void write(string txt) {
            auto p = pipe();
            p.writeEnd.write(txt);
            spawnProcess(copyCmd, p.readEnd());
        }

        static string read() {
            return execute(pasteCmd).output;
        }
    }

    version(Windows) {
        // No windows support yet
    }
}

unittest {
    string text = "æêáóìëæêî";
    assert(Clipboard.init());
    Clipboard.write(text);
    assert(Clipboard.read() == text);
}
