string emptyString(int size) {
    string str;
    foreach (i; 0 .. size) {
        str ~= " ";
    }
    return str;
}

int numOccurences(string str, char c) {
    int n;
    foreach (letter; str) {
        if (letter == c) {
            n++;
        }
    }
    return n;
}
