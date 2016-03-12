import rope;

import std.string, std.stdio;

class Buffer {
    private Rope text;

    string path;
    string savedText;

    private string value;

    string[] lines;

    this(string txt, string path) {
        text = new Rope(txt);
        this.path = path;
        update();
    }

    void save() {
        saveAs(path);
    }

    void saveAs(string filename) {
        string bufTxt = text.toString();
        File f = File(filename, "w");
        f.write(bufTxt);
        f.close();
        savedText = bufTxt;
    }

    override
    string toString() {
        return value;
    }

    void update() {
        value = text.toString();
        if (value == "") {
            lines = [""];
        } else {
            lines = value.split("\n");
        }
    }

    @property ulong length() {
        return text.length;
    }

    void remove(ulong start, ulong end) {
        text.remove(start, end);
        update();
    }
    void insert(ulong position, string value) {
        text.insert(position, value);
        update();
    }
    string substring(ulong start, ulong end = -1) {
        if (end == -1) {
            update();
            return text.substring(start, text.length);
        } else {
            update();
            return text.substring(start, end);
        }
    }
    char charAt(ulong pos) {
        update();
        return text.charAt(pos);
    }
}
