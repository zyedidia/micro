package buffer
/*
#cgo CFLAGS: -g
#cgo LDFLAGS: -lclang

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <clang-c/Index.h>

#define recalloc(mem_ptr, count, size) (\
    {\
        void * val;\
        if (NULL == (mem_ptr))\
        {\
            val = calloc((count), (size));\
        }\
        else\
        {\
            val = realloc((mem_ptr), (count) * (size));\
        }\
        val;\
    }\
    )

typedef struct suggestion_struct{
    char * str;
    int kind;
    int str_len;
}sug_t;

typedef struct suggestions_struct{
    sug_t ** sugs;
    int * sug_plen;
	int sug_len;
} sugs_t;

//Create suggestions strings
sugs_t printCodeCompletionSuggestions(CXCodeCompleteResults* results){
    unsigned numResults = results->NumResults;
    sugs_t sugs;
    sugs.sugs = NULL;
    sugs.sug_plen = NULL;
    sugs.sug_len = 0;
    
    CXCompletionResult *result = results->Results;
    for (unsigned i = 0; i < numResults; i++){
        CXCompletionString completionString = result[i].CompletionString;
        unsigned priority = clang_getCompletionPriority(completionString);
        if (priority > 50){
            continue;
        }

        unsigned numChunks = clang_getNumCompletionChunks(completionString);
    
        sugs.sug_plen = (int*)recalloc(sugs.sug_plen, sugs.sug_len+1, sizeof(int));
        sugs.sug_plen[sugs.sug_len] = numChunks;
        sugs.sugs = (sug_t**)recalloc(sugs.sugs, sugs.sug_len+1, sizeof(sug_t*));
        sugs.sugs[sugs.sug_len] = (sug_t*)calloc(numChunks, sizeof(sug_t));

        for (unsigned chunkNumber = 0; chunkNumber < numChunks; chunkNumber++){
            CXString chunk = clang_getCompletionChunkText(completionString, chunkNumber);
            const char * chunk_c = clang_getCString(chunk);
            sugs.sugs[sugs.sug_len][chunkNumber].str_len = strlen(chunk_c);
			sugs.sugs[sugs.sug_len][chunkNumber].str = (char*)calloc(sugs.sugs[sugs.sug_len][chunkNumber].str_len, sizeof(char));
			memcpy(sugs.sugs[sugs.sug_len][chunkNumber].str, chunk_c, sugs.sugs[sugs.sug_len][chunkNumber].str_len);
            clang_disposeString(chunk);

            enum CXCompletionChunkKind kind = clang_getCompletionChunkKind(completionString, chunkNumber);
			sugs.sugs[sugs.sug_len][chunkNumber].kind = kind;

        }
        sugs.sug_len += 1;
    }
    clang_disposeCodeCompleteResults(results);
    return sugs;
}

sugs_t CSuggestions(const char * Buff, const char * Fname, int line, int column){
    sugs_t sug;
    sug.sugs = NULL;
    sug.sug_plen = NULL;
    sug.sug_len = 0;
    CXCodeCompleteResults * res = NULL;

    struct CXUnsavedFile* uf = (struct CXUnsavedFile*)calloc(1, sizeof(struct CXUnsavedFile));
    uf[0].Filename = Fname;
    uf[0].Contents = Buff;
    uf[0].Length = strlen(Buff);
    
    CXIndex idx = clang_createIndex(0, 0);

    CXTranslationUnit u = clang_parseTranslationUnit(idx, uf->Filename, NULL, 0, uf, 1, CXTranslationUnit_None);
    if (NULL != u){ 
    	res = clang_codeCompleteAt(u, uf->Filename, line, column, uf, 1, clang_defaultCodeCompleteOptions());
    	if (NULL != res){
    		sug = printCodeCompletionSuggestions(res);
    	}
    	clang_disposeTranslationUnit(u);
	}
    clang_disposeIndex(idx);
    free(uf);

    return sug;
}

void CFreeSug(sugs_t sug){
    for (int i = 0; i < sug.sug_len; i++){
        for (int j = 0; j < sug.sug_plen[i]; j++){
            free(sug.sugs[i][j].str);
        }
        free(sug.sugs[i]);
    }
    free(sug.sugs);
    free(sug.sug_plen);
}

char * CRetrSug_SugStr(sugs_t sug, int index){
    int strlen = 0;
    for (int i = 0; i < sug.sug_plen[index]; i++){
        strlen += sug.sugs[index][i].str_len;
        if (sug.sugs[index][i].kind == 15){
            strlen++;
        }
    }
    strlen++;
    char * sugstr = (char*)calloc(strlen, sizeof(char));
    for (int i = 0; i < sug.sug_plen[index]; i++){
        memcpy(sugstr + strnlen(sugstr, strlen), sug.sugs[index][i].str, sug.sugs[index][i].str_len);
        if (sug.sugs[index][i].kind == 15){
            strcat(sugstr, " ");
        }
    }
    sugstr[strlen-1] = '\0';
    return sugstr;
}

int CRetrSug_SugsLen(sugs_t sug, int index){
	return sug.sug_plen[index];
}

sug_t CRetrSug_Sug(sugs_t sug, int line, int index){
	return sug.sugs[line][index];
}

int CRetrSug_Kind(sugs_t sug, int line, int index){
    return sug.sugs[line][index].kind;
}

char * CRetrSug_KindStr(sugs_t sug, int line, int index){
    return sug.sugs[line][index].str;
}
*/
import "C"
import (
	"bytes"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"unsafe"
	
	"github.com/zyedidia/micro/v2/internal/util"
)

// A Completer is a function that takes a buffer and returns info
// describing what autocompletions should be inserted at the current
// cursor location
// It returns a list of string suggestions which will be inserted at
// the current cursor location if selected as well as a list of
// suggestion names which can be displayed in an autocomplete box or
// other UI element
type Completer func(*Buffer) ([]string, []string)

func (b *Buffer) GetSuggestions() {

}

// Autocomplete starts the autocomplete process
func (b *Buffer) Autocomplete(c Completer) bool {
	b.Completions, b.Suggestions = c(b)
	if len(b.Completions) != len(b.Suggestions) || len(b.Completions) == 0 {
		return false
	}
	b.CurSuggestion = -1
	b.CycleAutocomplete(true)
	return true
}

// CycleAutocomplete moves to the next suggestion
func (b *Buffer) CycleAutocomplete(forward bool) {
	prevSuggestion := b.CurSuggestion

	if forward {
		b.CurSuggestion++
	} else {
		b.CurSuggestion--
	}
	if b.CurSuggestion >= len(b.Suggestions) {
		b.CurSuggestion = 0
	} else if b.CurSuggestion < 0 {
		b.CurSuggestion = len(b.Suggestions) - 1
	}

	c := b.GetActiveCursor()
	start := c.Loc
	end := c.Loc
	if prevSuggestion < len(b.Suggestions) && prevSuggestion >= 0 {
		start = end.Move(-util.CharacterCountInString(b.Completions[prevSuggestion]), b)
	}

	b.Replace(start, end, b.Completions[b.CurSuggestion])
	if len(b.Suggestions) > 1 {
		b.HasSuggestions = true
	}
}

// GetWord gets the most recent word separated by any separator
// (whitespace, punctuation, any non alphanumeric character)
func GetWord(b *Buffer) ([]byte, int) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)

	if c.X == 0 || util.IsWhitespace(b.RuneAt(c.Loc.Move(-1, b))) {
		return []byte{}, -1
	}

	if util.IsNonAlphaNumeric(b.RuneAt(c.Loc.Move(-1, b))) {
		return []byte{}, c.X
	}

	args := bytes.FieldsFunc(l, util.IsNonAlphaNumeric)
	input := args[len(args)-1]
	return input, c.X - util.CharacterCount(input)
}

// GetArg gets the most recent word (separated by ' ' only)
func GetArg(b *Buffer) (string, int) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)

	args := bytes.Split(l, []byte{' '})
	input := string(args[len(args)-1])
	argstart := 0
	for i, a := range args {
		if i == len(args)-1 {
			break
		}
		argstart += util.CharacterCount(a) + 1
	}

	return input, argstart
}

// FileComplete autocompletes filenames
func FileComplete(b *Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := GetArg(b)

	sep := string(os.PathSeparator)
	dirs := strings.Split(input, sep)

	var files []os.FileInfo
	var err error
	if len(dirs) > 1 {
		directories := strings.Join(dirs[:len(dirs)-1], sep) + sep

		directories, _ = util.ReplaceHome(directories)
		files, err = ioutil.ReadDir(directories)
	} else {
		files, err = ioutil.ReadDir(".")
	}

	if err != nil {
		return nil, nil
	}

	var suggestions []string
	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			name += sep
		}
		if strings.HasPrefix(name, dirs[len(dirs)-1]) {
			suggestions = append(suggestions, name)
		}
	}

	sort.Strings(suggestions)
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		var complete string
		if len(dirs) > 1 {
			complete = strings.Join(dirs[:len(dirs)-1], sep) + sep + suggestions[i]
		} else {
			complete = suggestions[i]
		}
		completions[i] = util.SliceEndStr(complete, c.X-argstart)
	}

	return completions, suggestions
}

// BufferComplete autocompletes based on previous words in the buffer
func BufferComplete(b *Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := GetWord(b)

	if argstart == -1 {
		return []string{}, []string{}
	}

	inputLen := util.CharacterCount(input)

	suggestionsSet := make(map[string]struct{})

	var suggestions []string
	for i := c.Y; i >= 0; i-- {
		l := b.LineBytes(i)
		words := bytes.FieldsFunc(l, util.IsNonAlphaNumeric)
		for _, w := range words {
			if bytes.HasPrefix(w, input) && util.CharacterCount(w) > inputLen {
				strw := string(w)
				if _, ok := suggestionsSet[strw]; !ok {
					suggestionsSet[strw] = struct{}{}
					suggestions = append(suggestions, strw)
				}
			}
		}
	}
	for i := c.Y + 1; i < b.LinesNum(); i++ {
		l := b.LineBytes(i)
		words := bytes.FieldsFunc(l, util.IsNonAlphaNumeric)
		for _, w := range words {
			if bytes.HasPrefix(w, input) && util.CharacterCount(w) > inputLen {
				strw := string(w)
				if _, ok := suggestionsSet[strw]; !ok {
					suggestionsSet[strw] = struct{}{}
					suggestions = append(suggestions, strw)
				}
			}
		}
	}
	if len(suggestions) > 1 {
		suggestions = append(suggestions, string(input))
	}

	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}

	return completions, suggestions
}

// BufferCompleteClang autocompletes based on what Clang autocomplete gives us
func BufferCompleteClang(b *Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := GetWord(b)

	if argstart == -1 {
		return []string{}, []string{}
	}

	inputLen := util.CharacterCount(input)
	var suggestions []string
	suggestionsSet := make(map[string]string)
	var buff string
	for i := 0; i < b.LinesNum(); i++ {
		buff += b.Line(i)
		buff += "\r\n"
	}

	cbuff := C.CString(buff)
	cfnme := C.CString(b.GetName())
	
	suggestionsStruc := C.CSuggestions(cbuff, cfnme, C.int(c.Y+1), C.int(c.X+1))
	C.free(unsafe.Pointer(cbuff))
	C.free(unsafe.Pointer(cfnme))

	for i := C.int(0); i < suggestionsStruc.sug_len; i++ {
		cStr := C.CRetrSug_SugStr(suggestionsStruc, i)
		Str := C.GoString(cStr)
		C.free(unsafe.Pointer(cStr))

		completion := ""
		for j := C.int(0); j < C.CRetrSug_SugsLen(suggestionsStruc, i); j++ {
			if C.CRetrSug_Kind(suggestionsStruc, i, j) != 15 &&
				C.CRetrSug_Kind(suggestionsStruc, i, j) != 14 &&
				C.CRetrSug_Kind(suggestionsStruc, i, j) != 6 &&
				C.CRetrSug_Kind(suggestionsStruc, i, j) != 7 &&
				C.CRetrSug_Kind(suggestionsStruc, i, j) != 3 {
				completion += C.GoString(C.CRetrSug_KindStr(suggestionsStruc, i, j))
			}
		}
		compbytes := []byte(completion)
		if bytes.HasPrefix([]byte(compbytes), input) && util.CharacterCount(compbytes) > inputLen {
			suggestions = append(suggestions, Str)
			suggestionsSet[Str] = completion
		}
	}
	C.CFreeSug(suggestionsStruc)

	if len(suggestions) > 1 {
		suggestions = append(suggestions, string(input))
		suggestionsSet[string(input)] = ""
	}
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestionsSet[suggestions[i]], c.X-argstart)
	}

	return completions, suggestions
}
