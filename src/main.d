import termbox;
import view;
import buffer;
import cursor;
import statusline;

import std.regex;
import core.exception;
import std.conv;
import std.file;
import std.range;
import std.string;
import std.stdio;

void main(string[] args) {
    if (args.length < 2) {
        return;
    }
    string filename = args[1];

    string fileSrc = readText(filename);

    init();

    Buffer buffer = new Buffer(fileSrc, filename);
    View v = new View(buffer);
    StatusLine sl = new StatusLine();
    sl.view = v;

    setInputMode(InputMode.mouse);

    Event e;
    try {
        while (e.key != Key.esc) {
            clear();

            v.display();
            sl.display();

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
