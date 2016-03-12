import std.string, std.stdio;
import std.algorithm: min, max;
import std.conv: to;
import std.math: floor;

// Rope data structure to store the text in the buffer
class Rope {
    private Rope left;
    private Rope right;
    private string value = null;

    ulong length;

    const int splitLength = 1000;
    const int joinLength = 500;
    const double rebalanceRatio = 1.2;

    this(string str) {
        this.value = str;
        this.length = str.count;

        adjust();
    }

    void adjust() {
        if (value !is null) {
            if (length > splitLength) {
                auto divide = cast(int) floor(length / 2.0);
                left = new Rope(value[0 .. divide]);
                right = new Rope(value[divide .. $]);
                value = null;
            }
        } else {
            if (length < joinLength) {
                value = left.toString() ~ right.toString();
                left = null;
                right = null;
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
            length = value.count;
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
            this.value = to!string(this.value.to!dstring[0 .. position] ~ value.to!dstring ~ this.value.to!dstring[position .. $]);
            length = this.value.count;
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
            left = null;
            right = null;
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

    string substring(ulong start, ulong end) {
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
