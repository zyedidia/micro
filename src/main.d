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
        fileTxt = readText(filename);
    }

    init();

    Buffer buf = new Buffer(fileTxt, filename);
    Cursor cursor = new Cursor();
    View v = new View(buf, cursor);

    setInputMode(InputMode.mouse);

    Event e;
    try {
        while (e.key != Key.ctrlQ) {
            clear();

            v.display();
            cursor.display();

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
