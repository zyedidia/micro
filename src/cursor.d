import buffer;

class Cursor {
    Buffer buf;
    int x, y;

    int lastX;

    this(Buffer buf) {
        this.buf = buf;
    }

    @property int loc() {
        int loc;
        foreach (i; 0 .. y) {
            loc += buf.lines[i].count + 1;
        }
        loc += x;
        return loc;
    }

    @property void loc(int value) {
        int loc;
        foreach (y, l; buf.lines) {
            if (loc + l.count+1 > value) {
                this.y = cast(int) y;
                x = value - loc;
                return;
            } else {
                loc += l.count+1;
            }
        }
    }
}
