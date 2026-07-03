// pkg/objects/builtin_list.go
// Single source of truth for builtin function names and indices.
// Add new builtins here, then both compiler and VM will pick them up automatically.
// Indices are auto-assigned from list position — no manual index management needed.
package objects

import "sync"

// BuiltinRegistry lists all builtin function names.
// The list position determines the auto-assigned index (0, 1, 2, ...).
// To add a new builtin:
//   1. Implement the function in builtin.go (or a split file)
//   2. Append the name to this list
//   3. Done - indices are auto-assigned at init time
var BuiltinRegistry = []string{
	// Core
	"len",
	"pr",
	"pln",
	"typeOf",
	// String utilities
	"substr",
	"split",
	"join",
	"trim",
	"upper",
	"lower",
	"containsStr",
	"replace",
	"startsWith",
	"endsWith",
	// Math
	"abs",
	"floor",
	"ceil",
	"sqrt",
	"pow",
	"min",
	"max",
	// Type conversion
	"int",
	"float",
	"string",
	// Array utilities
	"push",
	"pop",
	"first",
	"last",
	"rest",
	"concat",
	"indexOf",
	"containsArr",
	// Map utilities
	"keys",
	"values",
	"hasKey",
	"delete",
	// Collection utilities
	"range",
	"sort",
	"sum",
	"avg",
	"reverse",
	// Dynamic code
	"runCode",
	"loadPlugin",
	// String enhancement
	"repeat",
	"lpad",
	"rpad",
	"padLeft",
	"padRight",
	"charAt",
	"trimLeft",
	"trimRight",
	// Type checking
	"isEmpty",
	"isString",
	"isNumber",
	"isInt",
	"isFloat",
	"isArray",
	"isMap",
	"isBool",
	"isFunction",
	"isNull",
	// Math enhancement
	"clamp",
	"sign",
	"randomInt",
	// Array enhancement
	"unique",
	"flatten",
	"without",
	"take",
	"drop",
	// Map enhancement
	"merge",
	"entries",
	// Format
	"format",
	// Object utilities
	"copy",
	"clone",
	"equals",
	"defaults",
	// Encoding & Hash
	"base64Encode",
	"base64Decode",
	"hexEncode",
	"hexDecode",
	"md5",
	"sha256",
	// Time & UUID
	"sleep",
	"sleepSec",
	"now",
	"nowMs",
	"uuid",
	// String enhancement (Batch 2)
	"trimPrefix",
	"trimSuffix",
	"count",
	"isDigit",
	"isAlpha",
	"isAlphaNum",
	// Array enhancement (Batch 2)
	"find",
	"findIndex",
	"includes",
	"shuffle",
	"sample",
	"chunk",
	// Command line argument utilities
	"getSwitch",
	"switchExists",
	"getParam",
	// Output utilities
	"pl",
	"prf",
	// Validation utilities
	"checkErr",
	"checkEmpty",
	// OTP utilities
	"genOtpCode",
	"checkOtpCode",
	// Type conversion (Batch 2)
	"toStr",
	"toJson",
	"fromJson",
	// Dynamic code (Batch 2)
	"delegate",
	// Array functions (Charlang compatibility)
	"append",
	"appendArray",
	"appendList",
	"appendSlice",
	"arrayContains",
	"removeItems",
	"bytes",
	"chars",
	"plt",
	"make",
	// BigInt/BigFloat
	"bigInt",
	"bigFloat",
	"isBigInt",
	"isBigFloat",
	// Unicode character handling
	"toChars",
	"charLen",
	// HTTP server built-in functions
	"httpStatusName",
	"isHttpReq",
	"isHttpResp",
	"urlEncode",
	"urlDecode",
	// HTTP response/request builtins (declared in HttpBuiltins, registered in init)
	"writeResp",
	"setRespHeader",
	"addRespHeader",
	"setContentType",
	"status",
	"getReqHeader",
	"getReqHeaders",
	"setCookie",
	"getCookie",
	"getCookies",
	"redirect",
	"serveFile",
	"queryParam",
	"queryParams",
	"formValue",
	"parseForm",
	// Concurrency built-in functions
	"makeTube",
	"closeTube",
	"tubeLen",
	"tubeCap",
	"tubeClosed",
	"tubeSend",
	"tubeRecv",
	"tubeTrySend",
	"tubeTryRecv",
	"newMutex",
	"newRWMutex",
	"newWaitGroup",
	"newOnce",
	"newCond",
	"newAtomic",
	// Context built-in functions
	"newContext",
	"contextWithTimeout",
	"contextWithCancel",
	"contextWithDeadline",
	"contextCancel",
	"contextDone",
	"contextErr",
	"contextIsDone",
	"contextDeadline",
	// HTTP Client built-in functions (getWeb family)
	"getWeb",
	"getWebBytes",
	"getWebObject",
	"postWeb",
	"postWebObject",
	"urlExists",
	"httpStatus",
	// Reader/Writer built-in functions
	"getWebReader",
	"ioCopy",
	"isReader",
	"isWriter",
	"newBytesReader",
	"newStringReader",
	// Encryption built-in functions (Charlang compatible)
	"encryptTextByTXTE",
	"decryptTextByTXTE",
	"encryptDataByTXDEE",
	"decryptDataByTXDEE",
	"encryptTextByTXDEE",
	"decryptTextByTXDEE",
	"encryptData",
	"encryptBytes",
	"decryptData",
	"decryptBytes",
	"encryptText",
	"encryptStr",
	"decryptText",
	"decryptStr",
	"encryptStream",
	"decryptStream",
	"aesEncrypt",
	"aesDecrypt",
	"downloadFile",
	// Database built-in functions (String-based - Charlang compatible)
	"formatSQLValue",
	"dbConnect",
	"dbClose",
	"dbQuery",
	"dbQueryOrdered",
	"dbQueryRecs",
	"dbQueryMap",
	"dbQueryMapArray",
	"dbQueryCount",
	"dbQueryFloat",
	"dbQueryString",
	"dbExec",
	// Database built-in functions (Typed - preserve native types)
	"dbQueryTyped",
	"dbQueryRowTyped",
	"dbQueryArrayTyped",
	"dbQueryValueTyped",
	// OrderedMap built-in functions
	"isOrderedMap",
	"newOrderedMap",
	// System command built-in functions
	"systemCmd",
	"systemCmdDetached",
	"systemStart",
	// File system built-in functions
	"fileExists",
	"isDir",
	"loadText",
	"saveText",
	"appendText",
	"copyFile",
	"renameFile",
	"removeFile",
	"removeDir",
	"getFileList",
	"joinPath",
	"getCurDir",
	"getHomeDir",
	"getTempDir",
	"ensureMakeDirs",
	"getFileExt",
	"extractFileDir",
	"extractFileName",
	"getFileInfo",
	"loadLines",
	"getFileAbs",
	"getFileRel",
	"isFile",
	"saveBytes",
	"loadBytes",
	// Time enhancement built-in functions
	"getNowStr",
	"getNowTimeStamp",
	"formatTime",
	"timeToTick",
	"timeAddSecs",
	"timeAddDate",
	"timeBefore",
	"strToTime",
	"timeAfter",
	"timeEqual",
	"timeDiff",
	"timeDiffSecs",
	"parseTime",
	"isTime",
	// Regex built-in functions
	"regMatch",
	"regContains",
	"regFindFirst",
	"regFindAll",
	"regFindFirstGroups",
	"regFindAllGroups",
	"regReplace",
	"regSplit",
	"regCount",
	"regQuote",
	"regFindAllIndex",
	// Encoding enhancement built-in functions
	"urlEncodeComponent",
	"urlDecodeComponent",
	"htmlEncode",
	"htmlDecode",
	"sha1",
	"sha512",
	"hashStr",
	"toHex",
	"unhex",
	"hexToStr",
	// System/Environment built-in functions
	"getEnv",
	"setEnv",
	"getOSName",
	"getOSArch",
	"getOSArgs",
	"getAppPath",
	"getAppDir",
	"exit",
	"getSysInfo",
	"getPid",
	"getPPid",
	"hostname",
	// Float utilities
	"adjustFloat",
	"toKMG",
	"trunc",
	"isInf",
	"isNaN",
	"isFinite",
	// JSON enhancement built-in functions
	"formatJson",
	"compactJson",
	"getJsonNodeStr",
	"getJsonNodeStrs",
	"strsToJson",
	"jsonValid",
	"jsonType",
	"jsonPath",
	// Compression built-in functions
	"compressData",
	"uncompressData",
	"compressStr",
	"uncompressStr",
	"zipPath",
	"zipPaths",
	"unzipToPath",
	"getFileListInZip",
	"loadBytesInZip",
	"addFileToZip",
	// Input/Clipboard built-in functions
	"getInput",
	"getInputf",
	"getChar",
	"getKey",
	"getMultiLineInput",
	"getPassword",
	"confirm",
	"readLine",
	"getClipText",
	"setClipText",
	// String enhancement built-in functions (Batch 10)
	"strContainsIn",
	"strRuneLen",
	"strIn",
	"strGetLastComponent",
	"strFindDiffPos",
	"strDiff",
	"strFindAllSub",
	"limitStr",
	"strQuote",
	"strUnquote",
	"strToInt",
	"strReverse",
	"getTextSimilarity",
	"fuzzyFind",
	// Collection enhancement built-in functions
	"mapArray",
	"filterArray",
	"reduceArray",
	"forEach",
	"flatMap",
	"every",
	"some",
	"groupBy",
	"partition",
	"zip",
	"unzip",
	"fill",
	"rangeNum",
	"intersection",
	"difference",
	"union",
	"countBy",
	"sortBy",
	// Utility built-in functions
	"sprintf",
	"toBool",
	"toInt",
	"toFloat",
	"isUndefined",
	"isCallable",
	"isIterable",
	"isError",
	"error",
	"getErrStr",
	"isErrStr",
	"typeCode",
	"swap",
	"coalesce",
	"defaultVal",
	// String processing enhancement built-in functions
	"strSplitLines",
	"strContainsAny",
	"strIndex",
	"strLastIndex",
	"strSplitN",
	"strPad",
	"strSub",
	"intToStr",
	"floatToStr",
	"charCode",
	"charFromCode",
	"reverseMap",
	"simpleStrToMap",
	"mapToStr",
	// Bitwise operations
	"bitNot",
	"bitAnd",
	"bitOr",
	"bitXor",
	"bitShiftLeft",
	"bitShiftRight",
	// Check/validate and bytes built-in functions
	"isNil",
	"isNilOrEmpty",
	"isNilOrErr",
	"isUndef",
	"isErr",
	"isBytes",
	"isChars",
	"pass",
	"errStrf",
	"errf",
	"errToEmpty",
	"sscanf",
	"bytesStartsWith",
	"bytesEndsWith",
	"bytesContains",
	"bytesIndex",
	"compareBytes",
	"compareText",
	// Miscellaneous built-in functions
	"getRandomInt",
	"getRandomFloat",
	"getRandomStr",
	"createTempDir",
	"createTempFile",
	"changeDir",
	"lookPath",
	"joinUrlPath",
	"parseUrl",
	"parseQuery",
	"isHttps",
	"genToken",
	// Image
	"createImage",
	// Byte-index string functions
	"byteIndexOf",
	"byteSubstr",
	"byteLen",
	// String enhancement (Batch 24)
	"strCount",
	// Simple encoding
	"simpleEncode",
	"simpleDecode",
	// Time enhancement (Batch 26)
	"getNowStrCompact",
	"timeToTimeStamp",
	"timeStampToTime",
	// File system enhancement
	"dirExists",
	"pathExists",
	"copyPath",
	"moveFile",
	"getFileSize",
	// Print aliases
	"print",
	"println",
	"printf",
	"concatBytes",
	"plv",
	"spr",
}

// BuiltinIndexMap maps builtin name to auto-assigned index.
// Built lazily on first access via sync.Once to ensure all init() functions
// in other builtin_*.go files have completed before we build the index.
var BuiltinIndexMap map[string]int

// BuiltinFuncArray is the index-to-function lookup array for O(1) dispatch.
// Built lazily on first access via sync.Once.
var BuiltinFuncArray []*Builtin

var builtinOnce sync.Once

// ensureBuiltinIndex ensures the index maps are built.
// Called lazily on first access to avoid init-order issues with split builtin files.
func ensureBuiltinIndex() {
	builtinOnce.Do(func() {
		BuiltinIndexMap = make(map[string]int, len(BuiltinRegistry))
		BuiltinFuncArray = make([]*Builtin, len(BuiltinRegistry))
		for i, name := range BuiltinRegistry {
			BuiltinIndexMap[name] = i
			BuiltinFuncArray[i] = Builtins[name]
		}
	})
}

// GetBuiltinByName returns a builtin function by name.
// Returns nil if no builtin with the given name exists.
func GetBuiltinByName(name string) *Builtin {
	b, ok := Builtins[name]
	if !ok {
		return nil
	}
	return b
}

// GetBuiltinByIndex returns a builtin function by auto-assigned index.
// Returns nil if the index is out of range.
func GetBuiltinByIndex(index int) *Builtin {
	ensureBuiltinIndex()
	if index < 0 || index >= len(BuiltinFuncArray) {
		return nil
	}
	return BuiltinFuncArray[index]
}

// BuiltinNameExists returns true if a builtin with the given name is registered.
func BuiltinNameExists(name string) bool {
	_, ok := Builtins[name]
	return ok
}

// BuiltinIndex returns the auto-assigned index for a builtin name.
// Returns -1 if the name is not found.
func BuiltinIndex(name string) int {
	ensureBuiltinIndex()
	idx, ok := BuiltinIndexMap[name]
	if !ok {
		return -1
	}
	return idx
}
