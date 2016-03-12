import termbox;
import view;

import std.conv: to;

class StatusLine {
    View view;

    this() { }

    void update() {

    }

    void display() {
        int y = height() - 2;
        string file = view.buf.name;
        if (view.buf.toString != view.buf.savedText) {
            file ~= " +";
        }
        file ~= "  (" ~ to!string(view.cursor.y) ~ "," ~ to!string(view.cursor.x) ~ ")";
        foreach (x; 0 .. width()) {
            if (x >= 1 && x < 1 + file.length) {
                setCell(x, y, cast(uint) file[x - 1], Color.black, Color.blue);
            } else  {
                setCell(x, y, ' ', Color.black, Color.blue);
            }
        }
    }
}
