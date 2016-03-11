import termbox;
import view;

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
        foreach (x; 0 .. width()) {
            if (x >= 1 && x < 1 + file.length) {
                setCell(x, y, cast(uint) file[x - 1], Color.white, Color.blue);
            } else  {
                setCell(x, y, ' ', Color.white, Color.blue);
            }
        }
    }
}
