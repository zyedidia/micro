import std.math;
import std.stdio;
import std.utf;
import std.string;
import std.conv: to;
import std.algorithm;

class Buffer {
    private string value = null;
    private Buffer left;
    private Buffer right;

    string name = "";
    string savedText;

    ulong length;

    int splitLength = 1000;
    int joinLength = 500;
    double rebalanceRatio = 1.2;

    this() { }

    this(string str, string name = "") {
        this.value = str;
        this.length = str.length;
        this.name = name;
        this.savedText = str;

        left = new Buffer();
        right = new Buffer();
        left.value = "";
        right.value = "";

        adjust();
    }

    void save(string filename = null) {
        if (filename is null) {
            filename = name;
        }
        if (filename != "") {
            string bufSrc = this.toString();
            File f = File(filename, "w");
            f.write(bufSrc);
            f.close();
            savedText = bufSrc;
        }
    }

    @property string[] lines() {
        string str = this.toString();
        if (str == "") {
            return [""];
        } else {
            return str.split("\n");
        }
    }

    void adjust() {
        if (value !is null) {
            if (length > splitLength) {
                auto divide = cast(int) floor(length / 2.0);
                left = new Buffer(value[0 .. divide]);
                right = new Buffer(value[divide .. $]);
            }
        } else {
            if (length < joinLength) {
                value = left.toString() ~ right.toString();
            }
        }
    }

    override
    string toString() {
        if (value !is null) {
            return value;
        } else {
            return left.toString ~ right.toString();
        }
    }

    void remove(ulong start, ulong end) {
        if (value !is null) {
            value = value[0 .. start] ~ value[end .. $];
            length = value.length;
        } else {
            auto leftStart = min(start, left.length);
            auto leftEnd = min(end, left.length);
            auto rightStart = max(0, min(start - left.length, right.length));
            auto rightEnd = max(0, min(end - left.length, right.length));
            if (leftStart < left.length) {
                left.remove(leftStart, leftEnd);
            }
            if (rightEnd > 0) {
                right.remove(rightStart, rightEnd);
            }
            length = left.length + right.length;
        }

        adjust();
    }

    void insert(ulong position, string value) {
        if (this.value !is null) {
            this.value = this.value[0 .. position] ~ value ~ this.value[position .. $];
            length = this.value.length;
        } else {
            if (position < left.length) {
                left.insert(position, value);
                length = left.length + right.length;
            } else {
                right.insert(position - left.length, value);
            }
        }

        adjust();
    }

    void rebuild() {
        if (value is null) {
            value = left.toString() ~ right.toString();
            adjust();
        }
    }

    void rebalance() {
        if (value is null) {
            if (left.length / right.length > rebalanceRatio ||
                right.length / left.length > rebalanceRatio) {
                rebuild();
            } else {
                left.rebalance();
                right.rebalance();
            }
        }
    }

    string substring(ulong start, ulong end = length) {
        if (value !is null) {
            return value[start .. end];
        } else {
            auto leftStart = min(start, left.length);
            auto leftEnd = min(end, left.length);
            auto rightStart = max(0, min(start - left.length, right.length));
            auto rightEnd = max(0, min(end - left.length, right.length));

            if (leftStart != leftEnd) {
                if (rightStart != rightEnd) {
                    return left.substring(leftStart, leftEnd) ~ right.substring(rightStart, rightEnd);
                } else {
                    return left.substring(leftStart, leftEnd);
                }
            } else {
                if (rightStart != rightEnd) {
                    return right.substring(rightStart, rightEnd);
                } else {
                    return "";
                }
            }
        }
    }

    char charAt(ulong pos) {
        return to!char(substring(pos, pos + 1));
    }
}
