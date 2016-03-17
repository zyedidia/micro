import termbox;

class Cursor {
    int x, y;
    int lastX;

    uint selectionStart;
    uint selectionEnd;

    this() {}

    this(int x, int y) {
        this.x = x;
        this.y = y;
    }

    void hide() {
        x = y = -1;
    }
}
