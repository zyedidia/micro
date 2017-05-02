var temp = 1
// single line comment TODO:

/*

Multi line comment TODO TODO: FIXME FIXME: "string here" XXX: XXX
more test here
*/

// preprocessor 

#define 
#if
#endif
#elseif
#else
#selector
#line
#file
#column 

func (_ test: Int)

 __LINE__  
 __FILE__
 __FILE__
 __FUNCTION__
 
 ifnot
 c
 
func test(_ test: Int)  -> String
{
	print("Test string here")
}

// c types
CBool CShort CInt
// number
5656
// float number
12345.6466
//hex number

// binaray number

// keywords

// Declaration Keywords
associatedtype	class	deinit	enum	extension	fileprivate
func	import	init	inout	internal	let	open
operator	private	protocol	public	static	struct
subscript	typealias	var
    
// Statements Keywords
break	case	continue	default	defer	do		else	fallthrough
for		guard	if			inif	repeat	return	switch	where		while
 //  keyword.reserved
associativity	convenience	dynamic	didSet	final	get	infix	indirect	lazy
left	mutating	none	nonmutating		override	postfix	
precedence	prefix	Protocol	required
right	set	Type	unowned	weak	willSet
    
// Expression and types
as	as!	as?	Any	catch	is	rethrows	super	self	throw
throws		try	try!	try?	Optional
    
abs	advance	alignof	alignofValue	anyGenerator	assert	assertionFailure
bridgeFromObjectiveC	bridgeFromObjectiveCUnconditional	bridgeToObjectiveC
bridgeToObjectiveCUnconditional	c	contains	count	countElements
countLeadingZeros	debugPrint	debugPrintln	distance	dropFirst
dropLast	dump	encodeBitsAsWords enumerate equal fatalError filter	find
getBridgedObjectiveCType	getVaList	indices	insertionSort
isBridgedToObjectiveC|isBridgedVerbatimToObjectiveC|isUniquelyReferenced|isUniquelyReferencedNonObjC

join	lexicographicalCompare	map	max	maxElement	min	minElement	numericCast	overlaps	partition|posix
precondition|preconditionFailure|print|println|quickSort|readLine|reduce|reflect
reinterpretCast!reverse|roundUpToAlignment|sizeof|sizeofValue|sort|split|startsWith|stride
strideof|strideofValue|swap|toString|transcode|underestimateCount|unsafeAddressOf|unsafeBitCast
unsafeDowncast|unsafeUnwrap|unsafeReflect|withExtendedLifetime|withObjectAtPlusZero|withUnsafePointer
withUnsafePointerToObject|withUnsafeMutablePointer|withUnsafeMutablePointers|withUnsafePointer
withUnsafePointers withVaList	zip

// Meta
@warn_unused_result
@exported
@lazy
@noescape
@NSCopying
@NSManaged
@objc|@convention|@required
@noreturn|@IBAction|@IBDesignable|@IBInspectable|@IBOutlet|@infix|@prefix|@postfix|@autoclosure|@testable|@available
@nonobjc|@NSApplicationMain|@UIApplicationMain
   
// Constant
true	false	nil	
    
// Storage Types
Int8	
Int16
Int32
Int64
Int
UInt
String
Bit
Bool
Character
Double
Optional
Float
Any 
Range
AnyObject

// interpolation string with 
"test 345$£%^&& * \\ \0 \\ \" \'  \n \t  (test) \(test456) "
let message = "\(multiplier) times 2.5 is \(Double(multiplier) * 2.5) "

// unicode string
    let dollarSign = "\u{24}"        // $,  Unicode scalar U+0024
    let blackHeart = "\u{2665}"      // ♥,  Unicode scalar U+2665
    let sparklingHeart = "\u{1F496}" // 💖, Unicode scalar U+1F496

    let combinedEAcute: Character = "\u{65}\u{301}"

    
    
        let unusualMenagerie = "Koala 🐨, Snail 🐌, Penguin 🐧, Dromedary 🐪"
        print("unusualMenagerie has \(unusualMenagerie.characters.count) characters")
        // Prints "unusualMenagerie has 40 characters"
    
    

