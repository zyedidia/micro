string emptyString(int size) {
    string str;
    foreach (i; 0 .. size) {
        str ~= " ";
    }
    return str;
}
