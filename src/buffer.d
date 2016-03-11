import std.math;
import std.stdio;
import std.utf: count;
import std.string;
import std.conv: to;
import std.algorithm: min, max;

class Buffer {
    private string value = null;
    private Buffer left;
    private Buffer right;

    string name = "";
    string savedText;

    ulong count;

    int splitLength = 1000;
    int joinLength = 500;
    double rebalanceRatio = 1.2;

    this() { }

    this(string str, string name = "") {
        this.value = str;
        this.count = str.count;
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
            if (count > splitLength) {
                auto divide = cast(int) floor(count / 2.0);
                left = new Buffer(value[0 .. divide]);
                right = new Buffer(value[divide .. $]);
            }
        } else {
            if (count < joinLength) {
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
            value = to!string(value.to!dstring[0 .. start] ~ value.to!dstring[end .. $]);
            count = value.count;
        } else {
            auto leftStart = min(start, left.count);
            auto leftEnd = min(end, left.count);
            auto rightStart = max(0, min(start - left.count, right.count));
            auto rightEnd = max(0, min(end - left.count, right.count));
            if (leftStart < left.count) {
                left.remove(leftStart, leftEnd);
            }
            if (rightEnd > 0) {
                right.remove(rightStart, rightEnd);
            }
            count = left.count + right.count;
        }

        adjust();
    }

    void insert(ulong position, string value) {
        if (this.value !is null) {
            this.value = to!string(this.value.to!dstring[0 .. position] ~ value.to!dstring ~ this.value.to!dstring[position .. $]);
            count = this.value.count;
        } else {
            if (position < left.count) {
                left.insert(position, value);
                count = left.count + right.count;
            } else {
                right.insert(position - left.count, value);
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
            if (left.count / right.count > rebalanceRatio ||
                right.count / left.count > rebalanceRatio) {
                rebuild();
            } else {
                left.rebalance();
                right.rebalance();
            }
        }
    }

    string substring(ulong start, ulong end = count) {
        if (value !is null) {
            return value[start .. end];
        } else {
            auto leftStart = min(start, left.count);
            auto leftEnd = min(end, left.count);
            auto rightStart = max(0, min(start - left.count, right.count));
            auto rightEnd = max(0, min(end - left.count, right.count));

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
