import termbox;
import view;

class StatusLine {
    View view;

    this(View v) {
        this.view = v;
    }

    void display() {
        int y = view.height;
        string file = view.buf.path;
        if (file == "") {
            file = "untitled";
        }
        if (view.buf.toString != view.buf.savedText) {
            file ~= " +";
        }
        file ~= "  (" ~ to!string(view.cursor.y + 1) ~ "," ~ to!string(view.cursor.x + 1) ~ ")";
        foreach (x; 0 .. view.width) {
            if (x >= 1 && x < 1 + file.length) {
                setCell(x, y, cast(uint) file[x - 1], Color.black, Color.blue);
            } else  {
                setCell(x, y, ' ', Color.black, Color.blue);
            }
        }
    }
}
