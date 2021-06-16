#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <clang-c/Index.h>
#include <string>
#include <memory>

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

typedef struct suggestions_struct{
	char ** sug_str;
	int ** sug_kind;
	int * sug_str_len;
	int sug_len;
} suggestions_t;

//Printf to string
template<typename ... Args> std::string string_format(const std::string& format, Args ... args){
    //Extra space for '\0'
    int size_s = std::snprintf(nullptr, 0, format.c_str(), args ...) + 1;
    if(0 >= size_s){
        return std::string();
    }
    auto size = static_cast<size_t>(size_s);
    auto buf = std::make_unique<char[]>(size);
    std::snprintf(buf.get(), size, format.c_str(), args ...);
    //We don't want the '\0' inside
    return std::string(buf.get(), buf.get() + size - 1);
}

//Create suggestions strings
suggestions_t printCodeCompletionSuggestions(CXCodeCompleteResults* results){
    unsigned numResults = results->NumResults;
    suggestions_t suggestions;
    memset(&suggestions, 0, sizeof(suggestions_t));

    CXCompletionResult *result = results->Results;
    for (unsigned i = 0; i < numResults; i++){
        CXCompletionString completionString = result[i].CompletionString;
        unsigned priority = clang_getCompletionPriority(completionString);
        if (priority > 50){
            continue;
        }

        unsigned numChunks = clang_getNumCompletionChunks(completionString);
        std::string chunk_str = std::string();
        
        suggestions.sug_str = (char**)recalloc(suggestions.sug_str, i+1, sizeof(char*));
        suggestions.sug_kind = (int**)recalloc(suggestions.sug_kind, i+1, sizeof(int*));
        suggestions.sug_str_len = (int*)recalloc(suggestions.sug_str_len, i+1, sizeof(int));

        for (unsigned chunkNumber = 0; chunkNumber < numChunks; chunkNumber++){
            CXString chunk = clang_getCompletionChunkText(completionString, chunkNumber);
            enum CXCompletionChunkKind kind = clang_getCompletionChunkKind(completionString, chunkNumber);

            chunk_str += string_format("%s ", chunk);
            suggestions.sug_kind[i][chunkNumber] = kind;
            
            clang_disposeString(chunk);
        }
        char * sug_str = (char*)calloc(chunk_str.length(), sizeof(char));
        memcpy(sug_str, (char*)chunk_str.c_str(), chunk_str.length());
        suggestions.sug_str[i] = sug_str;
        suggestions.sug_str_len[i] = (int)chunk_str.length();
        suggestions.sug_len = i + 1;
    }
    clang_disposeCodeCompleteResults(results);
    return suggestions;
}

suggestions_t FindErrors(char * Buff, char * Fname, unsigned line, unsigned column){
    CXIndex idx = clang_createIndex(0, 0);
    CXUnsavedFile* uf = new CXUnsavedFile;
    uf[0].Filename = Fname;
    uf[0].Contents = Buff;
    uf[0].Length = strlen(Buff);

    CXTranslationUnit u = clang_parseTranslationUnit(idx, uf->Filename, NULL, 0, uf, 1, CXTranslationUnit_None);
    CXCodeCompleteResults* res = clang_codeCompleteAt(u, uf->Filename, line, column, uf, 1, clang_defaultCodeCompleteOptions());

    suggestions_t suggestions = printCodeCompletionSuggestions(res);
    
    clang_disposeTranslationUnit(u);
    clang_disposeIndex(idx);

    return suggestions;
}
