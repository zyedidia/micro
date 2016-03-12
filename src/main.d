import termbox;
import buffer;
import cursor;
import view;

import std.stdio;
import std.file;

void main(string[] args) {
    string filename = "";
    string fileTxt = "";

    if (args.length > 1) {
        filename = args[1];
        if (!exists(filename)) {
            File file = File(filename, "w");
            file.close();
        } else {
            if (isDir(filename)) {
                writeln(filename, " is a directory");
                return;
            }
            fileTxt = readText(filename);
            if (fileTxt is null) {
                fileTxt = "";
            }
        }
    } else {
        if (stdin.size != 0) {
            foreach (line; stdin.byLine()) {
                fileTxt ~= line ~ "\n";
            }
        }
    }

    Buffer buf = new Buffer(fileTxt, filename);
    init();

    Cursor cursor = new Cursor();
    View v = new View(buf, cursor);

    setInputMode(InputMode.mouse);

    Event e;
    try {
        while (e.key != Key.ctrlQ) {
            clear();

            v.display();

            flush();
            pollEvent(&e);

            v.update(e);
        }
    } catch (object.Error e) {
        shutdown();
        writeln(e);
        return;
    } catch (Exception e) {
        shutdown();
        writeln(e);
        return;
    }

    shutdown();
}
