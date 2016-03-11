import termbox;
import cursor;
import buffer;
import clipboard;

import std.conv: to;
import std.stdio;

class View {
    int topline;
    int xOffset;

    int width;
    int height;

    Buffer buf;
    Cursor cursor;

    this(Buffer buf, int topline = 0, int width = termbox.width(), int height = termbox.height()-2) {
        this.topline = topline;
        this.width = width;
        this.height = height;

        this.buf = buf;
        this.cursor = new Cursor(buf);
    }

    void cursorUp() {
        if (cursor.y > 0) {
            cursor.y--;
            cursor.x = cursor.lastX;
            if (cursor.x > buf.lines[cursor.y].length) {
                cursor.x = cast(int) buf.lines[cursor.y].length;
            }
        }
    }

    void cursorDown() {
        if (cursor.y < buf.lines.length - 1) {
            cursor.y++;
            cursor.x = cursor.lastX;
            if (cursor.x > buf.lines[cursor.y].length) {
                cursor.x = cast(int) buf.lines[cursor.y].length;
            }
        }
    }

    void cursorRight() {
        if (cursor.x < buf.lines[cursor.y].length) {
            cursor.x++;
            cursor.lastX = cursor.x;
        }
    }

    void cursorLeft() {
        if (cursor.x > 0) {
            cursor.x--;
            cursor.lastX = cursor.x;
        }
    }

    void update(Event e) {
        if (e.key == Key.mouseWheelUp) {
            if (topline > 0) {
                topline--;
            }
        } else if (e.key == Key.mouseWheelDown) {
            if (topline < buf.lines.length - height) {
                topline++;
            }
        } else {
            if (e.key == Key.arrowUp) {
                cursorUp();
            } else if (e.key == Key.arrowDown) {
                cursorDown();
            } else if (e.key == Key.arrowRight) {
                cursorRight();
            } else if (e.key == Key.arrowLeft) {
                cursorLeft();
            } else if (e.key == Key.mouseLeft) {
                cursor.x = e.x - xOffset;
                if (cursor.x < 0) {
                    cursor.x = 0;
                }
                cursor.y = e.y + topline;
                cursor.lastX = cursor.x;
                if (cursor.x > buf.lines[cursor.y].length) {
                    cursor.x = cast(int) buf.lines[cursor.y].length;
                }
            } else if (e.key == Key.ctrl_s) {
                buf.save();
            } else if (e.key == Key.ctrl_v) {
                if (Clipboard.supported) {
                    buf.insert(cursor.loc, Clipboard.read());
                }
            } else {
                if (e.ch != 0) {
                    buf.insert(cursor.loc, to!string(to!char(e.ch)));
                    cursorRight();
                } else if (e.key == Key.space) {
                    buf.insert(cursor.loc, " ");
                    cursorRight();
                } else if (e.key == Key.enter) {
                    buf.insert(cursor.loc, "\n");
                    cursor.loc = cursor.loc + 1;
                } else if (e.key == Key.backspace2) {
                    if (cursor.loc != 0) {
                        cursor.loc = cursor.loc - 1;
                        buf.remove(cursor.loc, cursor.loc + 1);
                    }
                }
            }

            if (cursor.y < topline) {
                topline--;
            }

            if (cursor.y > topline + height-1) {
                topline++;
            }
        }
    }

    void display() {
        int x, y;
        string[] lines;
        if (topline + height > buf.lines.length) {
            lines = buf.lines[topline .. $];
        } else  {
            lines = buf.lines[topline .. topline + height];
        }
        ulong maxLength = to!string(buf.lines.length).length;
        xOffset = cast(int) maxLength + 1;
        foreach (i, line; lines) {
            string lineNum = to!string(i + topline + 1);
            foreach (_; 0 .. maxLength - lineNum.length) {
                setCell(cast(int) x++, cast(int) y, ' ', Color.default_, Color.black);
            }
            foreach (dchar ch; lineNum) {
                setCell(cast(int) x++, cast(int) y, ch, Color.default_, Color.black);
            }
            setCell(cast(int) x++, cast(int) y, ' ', Color.default_ | Attribute.bold, Color.black);

            foreach (dchar ch; line) {
                setCell(cast(int) x++, cast(int) y, ch, Color.default_, Color.default_);
            }
            y++;
            x = 0;
        }

        if (cursor.y - topline < 0 || cursor.y - topline > height) {
            hideCursor();
        } else {
            setCursor(cursor.x + xOffset, cursor.y - topline);
        }
    }
}
