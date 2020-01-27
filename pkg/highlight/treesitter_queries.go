package highlight

// https://github.com/tree-sitter/tree-sitter-go/blob/1de485aa40e3923afaa2b87acbd53c7c5f49aa7f/queries/highlights.scm
const goQuery = `
; Function calls

(call_expression
  function: (identifier) @function)

(call_expression
  function: (selector_expression
    field: (field_identifier) @function.method))

; Function definitions

(function_declaration
  name: (identifier) @function)

(method_declaration
  name: (field_identifier) @function.method)

; Identifiers

(type_identifier) @type
(field_identifier) @property
(identifier) @variable

; Operators

"--" @operator
"-" @operator
"-=" @operator
":=" @operator
"!" @operator
"!=" @operator
"..." @operator
"*" @operator
"*" @operator
"*=" @operator
"/" @operator
"/=" @operator
"&" @operator
"&&" @operator
"&=" @operator
"%" @operator
"%=" @operator
"^" @operator
"^=" @operator
"+" @operator
"++" @operator
"+=" @operator
"<-" @operator
"<" @operator
"<<" @operator
"<<=" @operator
"<=" @operator
"=" @operator
"==" @operator
">" @operator
">=" @operator
">>" @operator
">>=" @operator
"|" @operator
"|=" @operator
"||" @operator

; Keywords

"break" @keyword
"case" @keyword
"chan" @keyword
"const" @keyword
"continue" @keyword
"default" @keyword
"defer" @keyword
"else" @keyword
"fallthrough" @keyword
"for" @keyword
"func" @keyword
"go" @keyword
"goto" @keyword
"if" @keyword
"import" @keyword
"interface" @keyword
"map" @keyword
"package" @keyword
"range" @keyword
"return" @keyword
"select" @keyword
"struct" @keyword
"switch" @keyword
"type" @keyword
"var" @keyword

; Literals

(interpreted_string_literal) @string
(raw_string_literal) @string
(rune_literal) @string
(escape_sequence) @escape

(int_literal) @number
(float_literal) @number
(imaginary_literal) @number

(true) @constant.builtin
(false) @constant.builtin
(nil) @constant.builtin

(comment) @comment
`

// https://github.com/tree-sitter/tree-sitter-rust/blob/1c37782a5528979a22991f8ed50dd3d3e423ac92/queries/highlights.scm
const rustQuery = `
; Identifier conventions

; Assume all-caps names are constants
((identifier) @constant
 (match? @constant "^[A-Z][A-Z\\d_]+$'"))

; Assume that uppercase names in paths are types
((scoped_identifier
  path: (identifier) @type)
 (match? @type "^[A-Z]"))
((scoped_identifier
  path: (scoped_identifier
    name: (identifier) @type))
 (match? @type "^[A-Z]"))

; Assume other uppercase names are enum constructors
((identifier) @constructor
 (match? @constructor "^[A-Z]"))

; Function calls

(call_expression
  function: (identifier) @function)
(call_expression
  function: (field_expression
    field: (field_identifier) @function.method))
(call_expression
  function: (scoped_identifier
    "::"
    name: (identifier) @function))

(generic_function
  function: (identifier) @function)
(generic_function
  function: (scoped_identifier
    name: (identifier) @function))
(generic_function
  function: (field_expression
    field: (field_identifier) @function.method))

(macro_invocation
  macro: (identifier) @function.macro
  "!" @function.macro)

; Function definitions

(function_item (identifier) @function)
(function_signature_item (identifier) @function)

; Other identifiers

(type_identifier) @type
(primitive_type) @type.builtin
(field_identifier) @property

(line_comment) @comment
(block_comment) @comment

"(" @punctuation.bracket
")" @punctuation.bracket
"[" @punctuation.bracket
"]" @punctuation.bracket

(type_arguments
  "<" @punctuation.bracket
  ">" @punctuation.bracket)
(type_parameters
  "<" @punctuation.bracket
  ">" @punctuation.bracket)

"::" @punctuation.delimiter
"." @punctuation.delimiter
";" @punctuation.delimiter

(parameter (identifier) @variable.parameter)

(lifetime (identifier) @label)

"break" @keyword
"const" @keyword
"continue" @keyword
"default" @keyword
"dyn" @keyword
"else" @keyword
"enum" @keyword
"extern" @keyword
"fn" @keyword
"for" @keyword
"if" @keyword
"impl" @keyword
"in" @keyword
"let" @keyword
"let" @keyword
"loop" @keyword
"macro_rules!" @keyword
"match" @keyword
"mod" @keyword
"move" @keyword
"pub" @keyword
"ref" @keyword
"return" @keyword
"static" @keyword
"struct" @keyword
"trait" @keyword
"type" @keyword
"union" @keyword
"unsafe" @keyword
"use" @keyword
"where" @keyword
"while" @keyword
(mutable_specifier) @keyword
(use_list (self) @keyword)
(scoped_use_list (self) @keyword)
(scoped_identifier (self) @keyword)
(super) @keyword

(self) @variable.builtin

(char_literal) @string
(string_literal) @string
(raw_string_literal) @string

(boolean_literal) @constant.builtin
(integer_literal) @constant.builtin
(float_literal) @constant.builtin

(escape_sequence) @escape

(attribute_item) @attribute
(inner_attribute_item) @attribute

"as" @operator
"*" @operator
"&" @operator
"'" @operator
`
