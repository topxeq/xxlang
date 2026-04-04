// pkg/compiler/compiler.go
package compiler

import (
	"fmt"
	"path"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// SymbolScope represents the scope of a symbol
type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
)

// Symbol represents a named variable in a scope
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable manages symbol definitions and resolution
type SymbolTable struct {
	Outer          *SymbolTable
	Store          map[string]Symbol
	NumDefinitions int
	FreeSymbols    []Symbol
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	s := &SymbolTable{
		Store:       make(map[string]Symbol),
		FreeSymbols: []Symbol{},
	}
	// Define built-in functions
	s.DefineBuiltin(0, "len")
	s.DefineBuiltin(1, "pr")
	s.DefineBuiltin(2, "pln")
	s.DefineBuiltin(3, "typeOf")
	s.DefineBuiltin(4, "substr")
	s.DefineBuiltin(5, "split")
	s.DefineBuiltin(6, "join")
	s.DefineBuiltin(7, "trim")
	s.DefineBuiltin(8, "upper")
	s.DefineBuiltin(9, "lower")
	s.DefineBuiltin(10, "containsStr")
	s.DefineBuiltin(11, "replace")
	s.DefineBuiltin(12, "startsWith")
	s.DefineBuiltin(13, "endsWith")
	s.DefineBuiltin(14, "abs")
	s.DefineBuiltin(15, "floor")
	s.DefineBuiltin(16, "ceil")
	s.DefineBuiltin(17, "sqrt")
	s.DefineBuiltin(18, "pow")
	s.DefineBuiltin(19, "min")
	s.DefineBuiltin(20, "max")
	s.DefineBuiltin(21, "int")
	s.DefineBuiltin(22, "float")
	s.DefineBuiltin(23, "string")
	s.DefineBuiltin(24, "push")
	s.DefineBuiltin(25, "pop")
	s.DefineBuiltin(26, "first")
	s.DefineBuiltin(27, "last")
	s.DefineBuiltin(28, "rest")
	s.DefineBuiltin(29, "concat")
	s.DefineBuiltin(30, "indexOf")
	s.DefineBuiltin(31, "containsArr")
	s.DefineBuiltin(32, "keys")
	s.DefineBuiltin(33, "values")
	s.DefineBuiltin(34, "hasKey")
	s.DefineBuiltin(35, "delete")
	s.DefineBuiltin(36, "range")
	s.DefineBuiltin(37, "sort")
	s.DefineBuiltin(38, "sum")
	s.DefineBuiltin(39, "avg")
	s.DefineBuiltin(40, "reverse")
	s.DefineBuiltin(41, "runCode")
	s.DefineBuiltin(42, "loadPlugin")
	// String utilities
	s.DefineBuiltin(43, "repeat")
	s.DefineBuiltin(44, "lpad")
	s.DefineBuiltin(45, "rpad")
	s.DefineBuiltin(46, "charAt")
	s.DefineBuiltin(47, "trimLeft")
	s.DefineBuiltin(48, "trimRight")
	// Type checking
	s.DefineBuiltin(49, "isEmpty")
	s.DefineBuiltin(50, "isString")
	s.DefineBuiltin(51, "isNumber")
	s.DefineBuiltin(52, "isInt")
	s.DefineBuiltin(53, "isFloat")
	s.DefineBuiltin(54, "isArray")
	s.DefineBuiltin(55, "isMap")
	s.DefineBuiltin(56, "isBool")
	s.DefineBuiltin(57, "isFunction")
	s.DefineBuiltin(58, "isNull")
	// Math utilities
	s.DefineBuiltin(59, "round")
	s.DefineBuiltin(60, "clamp")
	s.DefineBuiltin(61, "sign")
	s.DefineBuiltin(62, "random")
	s.DefineBuiltin(63, "randomInt")
	// Array utilities
	s.DefineBuiltin(64, "unique")
	s.DefineBuiltin(65, "flatten")
	s.DefineBuiltin(66, "without")
	s.DefineBuiltin(67, "take")
	s.DefineBuiltin(68, "drop")
	// Map utilities
	s.DefineBuiltin(69, "merge")
	s.DefineBuiltin(70, "entries")
	// Format
	s.DefineBuiltin(71, "format")
	// Object utilities
	s.DefineBuiltin(72, "copy")
	s.DefineBuiltin(73, "clone")
	s.DefineBuiltin(74, "equals")
	s.DefineBuiltin(75, "defaults")
	// Encoding & Hash
	s.DefineBuiltin(76, "base64Encode")
	s.DefineBuiltin(77, "base64Decode")
	s.DefineBuiltin(78, "hexEncode")
	s.DefineBuiltin(79, "hexDecode")
	s.DefineBuiltin(80, "md5")
	s.DefineBuiltin(81, "sha256")
	// Time & UUID
	s.DefineBuiltin(82, "sleep")
	s.DefineBuiltin(83, "now")
	s.DefineBuiltin(84, "nowMs")
	s.DefineBuiltin(85, "uuid")
	// String enhancement
	s.DefineBuiltin(86, "trimPrefix")
	s.DefineBuiltin(87, "trimSuffix")
	s.DefineBuiltin(88, "count")
	s.DefineBuiltin(89, "isDigit")
	s.DefineBuiltin(90, "isAlpha")
	s.DefineBuiltin(91, "isAlphaNum")
	// Array enhancement
	s.DefineBuiltin(92, "find")
	s.DefineBuiltin(93, "findIndex")
	s.DefineBuiltin(94, "includes")
	s.DefineBuiltin(95, "shuffle")
	s.DefineBuiltin(96, "sample")
	s.DefineBuiltin(97, "chunk")
	// Command line argument utilities
	s.DefineBuiltin(98, "getSwitch")
	s.DefineBuiltin(99, "switchExists")
	// Output utilities
	s.DefineBuiltin(100, "pl")
	s.DefineBuiltin(101, "prf")
	// Validation utilities
	s.DefineBuiltin(102, "checkErr")
	s.DefineBuiltin(103, "checkEmpty")
	// OTP utilities
	s.DefineBuiltin(104, "genOtpCode")
	// Type conversion
	s.DefineBuiltin(105, "toStr")
	s.DefineBuiltin(106, "toJson")
	s.DefineBuiltin(107, "fromJson")
	// Dynamic code
	s.DefineBuiltin(108, "delegate")
	// Array functions (Charlang compatibility)
	s.DefineBuiltin(109, "append")
	s.DefineBuiltin(110, "appendArray")
	s.DefineBuiltin(111, "arrayContains")
	s.DefineBuiltin(112, "removeItems")
	s.DefineBuiltin(113, "bytes")
	s.DefineBuiltin(114, "chars")
	s.DefineBuiltin(115, "plt")
	s.DefineBuiltin(116, "make")
	// BigInt/BigFloat
	s.DefineBuiltin(117, "bigInt")
	s.DefineBuiltin(118, "bigFloat")
	s.DefineBuiltin(119, "isBigInt")
	s.DefineBuiltin(120, "isBigFloat")
	// Chars (Unicode character handling)
	s.DefineBuiltin(121, "toChars")
	s.DefineBuiltin(122, "charLen")
	// HTTP built-in functions (for server mode)
	s.DefineBuiltin(123, "writeResp")
	s.DefineBuiltin(124, "setRespHeader")
	s.DefineBuiltin(125, "addRespHeader")
	s.DefineBuiltin(126, "getReqHeader")
	s.DefineBuiltin(127, "getReqHeaders")
	s.DefineBuiltin(128, "setCookie")
	s.DefineBuiltin(129, "getCookie")
	s.DefineBuiltin(130, "getCookies")
	s.DefineBuiltin(131, "parseForm")
	// parseJSON, getReqBody, getReqBodyBytes moved to http module
	s.DefineBuiltin(132, "status")
	s.DefineBuiltin(133, "redirect")
	s.DefineBuiltin(134, "serveFile")
	// getMimeType moved to http module
	s.DefineBuiltin(135, "setContentType")
	s.DefineBuiltin(136, "queryParam")
	s.DefineBuiltin(137, "queryParams")
	s.DefineBuiltin(138, "formValue")
	s.DefineBuiltin(139, "httpStatusName")
	s.DefineBuiltin(140, "isHttpReq")
	s.DefineBuiltin(141, "isHttpResp")
	s.DefineBuiltin(142, "urlEncode")
	s.DefineBuiltin(143, "urlDecode")
	// WebSocket functions moved to http module
	// Concurrency built-in functions
	s.DefineBuiltin(144, "makeTube")
	s.DefineBuiltin(145, "closeTube")
	s.DefineBuiltin(146, "tubeLen")
	s.DefineBuiltin(147, "tubeCap")
	s.DefineBuiltin(148, "tubeClosed")
	s.DefineBuiltin(149, "tubeSend")
	s.DefineBuiltin(150, "tubeRecv")
	s.DefineBuiltin(151, "tubeTrySend")
	s.DefineBuiltin(152, "tubeTryRecv")
	s.DefineBuiltin(153, "newMutex")
	s.DefineBuiltin(154, "newRWMutex")
	s.DefineBuiltin(155, "newWaitGroup")
	s.DefineBuiltin(156, "newOnce")
	s.DefineBuiltin(157, "newCond")
	s.DefineBuiltin(158, "newAtomic")
	// Context built-in functions
	s.DefineBuiltin(159, "newContext")
	s.DefineBuiltin(160, "contextWithTimeout")
	s.DefineBuiltin(161, "contextWithCancel")
	s.DefineBuiltin(162, "contextWithDeadline")
	s.DefineBuiltin(163, "contextCancel")
	s.DefineBuiltin(164, "contextDone")
	s.DefineBuiltin(165, "contextErr")
	s.DefineBuiltin(166, "contextIsDone")
	s.DefineBuiltin(167, "contextDeadline")
	// HTTP Client built-in functions (getWeb family)
	s.DefineBuiltin(168, "getWeb")
	s.DefineBuiltin(169, "getWebBytes")
	s.DefineBuiltin(170, "getWebObject")
	s.DefineBuiltin(171, "postWeb")
	s.DefineBuiltin(172, "postWebObject")
	s.DefineBuiltin(173, "urlExists")
	s.DefineBuiltin(174, "httpStatus")
	// Reader/Writer built-in functions
	s.DefineBuiltin(175, "getWebReader")
	s.DefineBuiltin(176, "ioCopy")
	s.DefineBuiltin(177, "isReader")
	s.DefineBuiltin(178, "isWriter")
	s.DefineBuiltin(179, "newBytesReader")
	s.DefineBuiltin(180, "newStringReader")
	// Encryption built-in functions (Charlang compatible)
	s.DefineBuiltin(181, "encryptTextByTXTE")
	s.DefineBuiltin(182, "decryptTextByTXTE")
	s.DefineBuiltin(183, "encryptDataByTXDEE")
	s.DefineBuiltin(184, "decryptDataByTXDEE")
	s.DefineBuiltin(185, "encryptTextByTXDEE")
	s.DefineBuiltin(186, "decryptTextByTXDEE")
	s.DefineBuiltin(187, "encryptData")
	s.DefineBuiltin(188, "encryptBytes")
	s.DefineBuiltin(189, "decryptData")
	s.DefineBuiltin(190, "decryptBytes")
	s.DefineBuiltin(191, "encryptText")
	s.DefineBuiltin(192, "encryptStr")
	s.DefineBuiltin(193, "decryptText")
	s.DefineBuiltin(194, "decryptStr")
	s.DefineBuiltin(195, "encryptStream")
	s.DefineBuiltin(196, "decryptStream")
	s.DefineBuiltin(197, "aesEncrypt")
	s.DefineBuiltin(198, "aesDecrypt")
	s.DefineBuiltin(199, "downloadFile")
	// Database built-in functions (String-based - Charlang compatible)
	s.DefineBuiltin(200, "formatSQLValue")
	s.DefineBuiltin(201, "dbConnect")
	s.DefineBuiltin(202, "dbClose")
	s.DefineBuiltin(203, "dbQuery")
	s.DefineBuiltin(204, "dbQueryOrdered")
	s.DefineBuiltin(205, "dbQueryRecs")
	s.DefineBuiltin(206, "dbQueryMap")
	s.DefineBuiltin(207, "dbQueryMapArray")
	s.DefineBuiltin(208, "dbQueryCount")
	s.DefineBuiltin(209, "dbQueryFloat")
	s.DefineBuiltin(210, "dbQueryString")
	s.DefineBuiltin(211, "dbExec")
	// Database built-in functions (Typed - preserve native types)
	s.DefineBuiltin(212, "dbQueryTyped")
	s.DefineBuiltin(213, "dbQueryRowTyped")
	s.DefineBuiltin(214, "dbQueryArrayTyped")
	s.DefineBuiltin(215, "dbQueryValueTyped")
	// OrderedMap built-in functions
	s.DefineBuiltin(216, "isOrderedMap")
	s.DefineBuiltin(217, "newOrderedMap")
	// System command built-in functions
	s.DefineBuiltin(218, "systemCmd")
	s.DefineBuiltin(219, "systemCmdDetached")
	s.DefineBuiltin(220, "systemStart")
	// Test assertion functions moved to testing module
	// s.DefineBuiltin(220, "testByText")        // use testing.byText()
	// s.DefineBuiltin(221, "testByStartsWith")  // use testing.byStartsWith()
	// s.DefineBuiltin(222, "testByEndsWith")    // use testing.byEndsWith()
	// s.DefineBuiltin(223, "testByContains")    // use testing.byContains()
	// s.DefineBuiltin(224, "testByReg")         // use testing.byReg()
	// s.DefineBuiltin(225, "testByRegContains") // use testing.byRegContains()
	// s.DefineBuiltin(226, "dumpVar")           // use debug.dumpVar()
	// s.DefineBuiltin(227, "debugInfo")         // use debug.info()
	// File system built-in functions (Batch 1)
	s.DefineBuiltin(228, "fileExists")
	s.DefineBuiltin(229, "isDir")
	s.DefineBuiltin(230, "loadText")
	s.DefineBuiltin(231, "saveText")
	s.DefineBuiltin(232, "appendText")
	s.DefineBuiltin(233, "copyFile")
	s.DefineBuiltin(234, "renameFile")
	s.DefineBuiltin(235, "removeFile")
	s.DefineBuiltin(236, "removeDir")
	s.DefineBuiltin(237, "getFileList")
	s.DefineBuiltin(238, "joinPath")
	s.DefineBuiltin(239, "getCurDir")
	s.DefineBuiltin(240, "getHomeDir")
	s.DefineBuiltin(241, "getTempDir")
	s.DefineBuiltin(242, "ensureMakeDirs")
	s.DefineBuiltin(243, "getFileExt")
	s.DefineBuiltin(244, "extractFileDir")
	s.DefineBuiltin(245, "extractFileName")
	s.DefineBuiltin(246, "getFileInfo")
	s.DefineBuiltin(247, "loadLines")
	s.DefineBuiltin(248, "getFileAbs")
	s.DefineBuiltin(249, "getFileRel")
	s.DefineBuiltin(250, "isFile")
	s.DefineBuiltin(251, "saveBytes")
	s.DefineBuiltin(252, "loadBytes")
	// File system enhancement (Batch 1b)
	s.DefineBuiltin(498, "dirExists")
	s.DefineBuiltin(499, "pathExists")
	s.DefineBuiltin(500, "copyPath")
	s.DefineBuiltin(501, "moveFile")
	s.DefineBuiltin(502, "getFileSize")
	// Time enhancement built-in functions (Batch 2)
	s.DefineBuiltin(253, "getNowStr")
	s.DefineBuiltin(254, "getNowTimeStamp")
	s.DefineBuiltin(255, "formatTime")
	s.DefineBuiltin(256, "timeToTick")
	s.DefineBuiltin(257, "timeAddSecs")
	s.DefineBuiltin(258, "timeAddDate")
	s.DefineBuiltin(259, "timeBefore")
	s.DefineBuiltin(260, "strToTime")
	s.DefineBuiltin(261, "timeAfter")
	s.DefineBuiltin(262, "timeEqual")
	s.DefineBuiltin(263, "timeDiff")
	s.DefineBuiltin(264, "timeDiffSecs")
	s.DefineBuiltin(265, "parseTime")
	s.DefineBuiltin(266, "isTime")
	// Regex enhancement built-in functions (Batch 3)
	s.DefineBuiltin(267, "regMatch")
	s.DefineBuiltin(268, "regContains")
	s.DefineBuiltin(269, "regFindFirst")
	s.DefineBuiltin(270, "regFindAll")
	s.DefineBuiltin(271, "regFindFirstGroups")
	s.DefineBuiltin(272, "regFindAllGroups")
	s.DefineBuiltin(273, "regReplace")
	s.DefineBuiltin(274, "regSplit")
	s.DefineBuiltin(275, "regCount")
	s.DefineBuiltin(276, "regQuote")
	s.DefineBuiltin(277, "regFindAllIndex")
	// Encoding enhancement built-in functions (Batch 4)
	s.DefineBuiltin(278, "urlEncodeComponent")
	s.DefineBuiltin(279, "urlDecodeComponent")
	s.DefineBuiltin(280, "htmlEncode")
	s.DefineBuiltin(281, "htmlDecode")
	s.DefineBuiltin(282, "sha1")
	s.DefineBuiltin(283, "sha512")
	s.DefineBuiltin(284, "hashStr")
	s.DefineBuiltin(285, "toHex")
	s.DefineBuiltin(286, "unhex")
	s.DefineBuiltin(287, "hexToStr")
	// System/Environment built-in functions (Batch 5)
	s.DefineBuiltin(288, "getEnv")
	s.DefineBuiltin(289, "setEnv")
	s.DefineBuiltin(290, "getOSName")
	s.DefineBuiltin(291, "getOSArch")
	s.DefineBuiltin(292, "getOSArgs")
	s.DefineBuiltin(293, "getAppPath")
	s.DefineBuiltin(294, "getAppDir")
	s.DefineBuiltin(295, "exit")
	s.DefineBuiltin(296, "getSysInfo")
	s.DefineBuiltin(297, "getPid")
	s.DefineBuiltin(298, "getPPid")
	s.DefineBuiltin(299, "hostname")
	// Math enhancement built-in functions (Batch 6)
	// Note: sin, cos, tan, etc. moved to math module. Placeholders kept for index stability.
	s.DefineBuiltin(300, "sin")
	s.DefineBuiltin(301, "cos")
	s.DefineBuiltin(302, "tan")
	s.DefineBuiltin(303, "asin")
	s.DefineBuiltin(304, "acos")
	s.DefineBuiltin(305, "atan")
	s.DefineBuiltin(306, "atan2")
	s.DefineBuiltin(307, "exp")
	s.DefineBuiltin(308, "log")
	s.DefineBuiltin(309, "log10")
	s.DefineBuiltin(310, "log2")
	s.DefineBuiltin(311, "pi")
	s.DefineBuiltin(312, "e")
	s.DefineBuiltin(313, "degToRad")
	s.DefineBuiltin(314, "radToDeg")
	s.DefineBuiltin(315, "adjustFloat")
	s.DefineBuiltin(316, "toKMG")
	s.DefineBuiltin(317, "trunc")
	s.DefineBuiltin(318, "isInf")
	s.DefineBuiltin(319, "isNaN")
	s.DefineBuiltin(320, "isFinite")
	// JSON enhancement built-in functions (Batch 7)
	s.DefineBuiltin(321, "formatJson")
	s.DefineBuiltin(322, "compactJson")
	s.DefineBuiltin(323, "getJsonNodeStr")
	s.DefineBuiltin(324, "getJsonNodeStrs")
	s.DefineBuiltin(325, "strsToJson")
	s.DefineBuiltin(326, "jsonValid")
	s.DefineBuiltin(327, "jsonType")
	// Compression built-in functions (Batch 8)
	s.DefineBuiltin(328, "compressData")
	s.DefineBuiltin(329, "uncompressData")
	s.DefineBuiltin(330, "compressStr")
	s.DefineBuiltin(331, "uncompressStr")
	s.DefineBuiltin(332, "zipPath")
	s.DefineBuiltin(333, "zipPaths")
	s.DefineBuiltin(334, "unzipToPath")
	s.DefineBuiltin(335, "getFileListInZip")
	s.DefineBuiltin(336, "loadBytesInZip")
	s.DefineBuiltin(337, "addFileToZip")
	// Input/Clipboard built-in functions (Batch 9)
	s.DefineBuiltin(338, "getInput")
	s.DefineBuiltin(339, "getInputf")
	s.DefineBuiltin(340, "getChar")
	s.DefineBuiltin(341, "getMultiLineInput")
	s.DefineBuiltin(342, "getPassword")
	s.DefineBuiltin(343, "confirm")
	s.DefineBuiltin(344, "readLine")
	s.DefineBuiltin(345, "getClipText")
	s.DefineBuiltin(346, "setClipText")
	// String enhancement built-in functions (Batch 10)
	s.DefineBuiltin(347, "strContainsIn")
	s.DefineBuiltin(348, "strRuneLen")
	s.DefineBuiltin(349, "strIn")
	s.DefineBuiltin(350, "strGetLastComponent")
	s.DefineBuiltin(351, "strFindDiffPos")
	s.DefineBuiltin(352, "strDiff")
	s.DefineBuiltin(353, "strFindAllSub")
	s.DefineBuiltin(354, "limitStr")
	s.DefineBuiltin(355, "strQuote")
	s.DefineBuiltin(356, "strUnquote")
	s.DefineBuiltin(357, "strToInt")
	s.DefineBuiltin(358, "getTextSimilarity")
	s.DefineBuiltin(359, "fuzzyFind")
	// strRemoveBom moved to strings module - use strings.removeBom(), strings.addBom(), strings.bom()
	s.DefineBuiltin(360, "strRemoveBom") // placeholder for index stability
	// String functions moved to string module
	// s.DefineBuiltin(361, "wordCount")      // use string.wordCount()
	// s.DefineBuiltin(362, "lineCount")      // use string.lineCount()
	// s.DefineBuiltin(363, "reverseStr")     // use string.reverse()
	// s.DefineBuiltin(364, "capitalize")     // use string.capitalize()
	// s.DefineBuiltin(365, "title")          // use string.title()
	// s.DefineBuiltin(366, "swapCase")       // use string.swapCase()
	// s.DefineBuiltin(367, "center")         // use string.center()
	// s.DefineBuiltin(368, "zfill")          // use string.zfill()
	// s.DefineBuiltin(369, "isSpace")        // use string.isSpace()
	// Collection enhancement built-in functions (Batch 11)
	s.DefineBuiltin(370, "mapArray")
	s.DefineBuiltin(371, "filterArray")
	s.DefineBuiltin(372, "reduceArray")
	s.DefineBuiltin(373, "forEach")
	s.DefineBuiltin(374, "flatMap")
	s.DefineBuiltin(375, "every")
	s.DefineBuiltin(376, "some")
	s.DefineBuiltin(377, "groupBy")
	s.DefineBuiltin(378, "partition")
	s.DefineBuiltin(379, "zip")
	s.DefineBuiltin(380, "unzip")
	s.DefineBuiltin(381, "fill")
	s.DefineBuiltin(382, "rangeNum")
	s.DefineBuiltin(383, "intersection")
	s.DefineBuiltin(384, "difference")
	s.DefineBuiltin(385, "union")
	s.DefineBuiltin(386, "countBy")
	s.DefineBuiltin(387, "sortBy")
	// Utility built-in functions (Batch 12)
	s.DefineBuiltin(388, "sprintf")
	s.DefineBuiltin(389, "toBool")
	s.DefineBuiltin(390, "toInt")
	s.DefineBuiltin(391, "toFloat")
	s.DefineBuiltin(392, "isUndefined")
	s.DefineBuiltin(393, "isCallable")
	s.DefineBuiltin(394, "isIterable")
	s.DefineBuiltin(395, "isError")
	s.DefineBuiltin(396, "error")
	s.DefineBuiltin(397, "getErrStr")
	s.DefineBuiltin(398, "isErrStr")
	s.DefineBuiltin(399, "typeCode")
	s.DefineBuiltin(400, "swap")
	s.DefineBuiltin(401, "coalesce")
	s.DefineBuiltin(402, "defaultVal")
	// String processing enhancement built-in functions (Batch 13)
	s.DefineBuiltin(403, "strSplitLines")
	s.DefineBuiltin(404, "strContainsAny")
	s.DefineBuiltin(405, "strIndex")
	s.DefineBuiltin(406, "strLastIndex")
	s.DefineBuiltin(407, "strSplitN")
	s.DefineBuiltin(408, "strPad")
	s.DefineBuiltin(409, "strSub")
	s.DefineBuiltin(410, "intToStr")
	s.DefineBuiltin(411, "floatToStr")
	s.DefineBuiltin(412, "charCode")
	s.DefineBuiltin(413, "charFromCode")
	s.DefineBuiltin(414, "reverseMap")
	s.DefineBuiltin(415, "simpleStrToMap")
	s.DefineBuiltin(416, "mapToStr")
	s.DefineBuiltin(417, "bitNot")
	s.DefineBuiltin(418, "bitAnd")
	s.DefineBuiltin(419, "bitOr")
	s.DefineBuiltin(420, "bitXor")
	s.DefineBuiltin(421, "bitShiftLeft")
	s.DefineBuiltin(422, "bitShiftRight")
	// Check/validate and bytes built-in functions (Batch 14)
	s.DefineBuiltin(423, "isNil")
	s.DefineBuiltin(424, "isNull")
	s.DefineBuiltin(425, "isNilOrEmpty")
	s.DefineBuiltin(426, "isNilOrErr")
	s.DefineBuiltin(427, "isBytes")
	s.DefineBuiltin(428, "isChars")
	s.DefineBuiltin(429, "pass")
	s.DefineBuiltin(430, "errStrf")
	s.DefineBuiltin(431, "errf")
	s.DefineBuiltin(432, "errToEmpty")
	s.DefineBuiltin(433, "sscanf")
	s.DefineBuiltin(434, "bytesStartsWith")
	s.DefineBuiltin(435, "bytesEndsWith")
	s.DefineBuiltin(436, "bytesContains")
	s.DefineBuiltin(437, "bytesIndex")
	s.DefineBuiltin(438, "compareBytes")
	s.DefineBuiltin(439, "compareText")
	// Miscellaneous built-in functions (Batch 15)
	s.DefineBuiltin(440, "getRandomInt")
	s.DefineBuiltin(441, "getRandomFloat")
	s.DefineBuiltin(442, "getRandomStr")
	s.DefineBuiltin(443, "createTempDir")
	s.DefineBuiltin(444, "createTempFile")
	s.DefineBuiltin(445, "changeDir")
	s.DefineBuiltin(446, "lookPath")
	s.DefineBuiltin(447, "joinUrlPath")
	s.DefineBuiltin(448, "parseUrl")
	s.DefineBuiltin(449, "parseQuery")
	s.DefineBuiltin(450, "isHttps")
	s.DefineBuiltin(451, "genToken")
	s.DefineBuiltin(452, "genOtpCode")
	s.DefineBuiltin(453, "checkOtpCode")
	// Unicode/Text processing built-in functions (Batch 16)
	s.DefineBuiltin(454, "toPinYin")
	s.DefineBuiltin(455, "kanaToRomaji")
	s.DefineBuiltin(456, "kanjiToKana")
	s.DefineBuiltin(457, "kanjiToRomaji")
	// JWT built-in functions (Batch 17)
	s.DefineBuiltin(458, "genJwtToken")
	s.DefineBuiltin(459, "parseJwtToken")
	// Task/Scheduling functions removed - use task module instead
	// s.DefineBuiltin(460, "isCronExprValid")
	// s.DefineBuiltin(461, "isCronExprDue")
	// s.DefineBuiltin(462, "runTicker")
	// s.DefineBuiltin(463, "stopTicker")
	// Image processing functions - genQr, scanQr, getImageInfo, resizeImage moved to image module
	// s.DefineBuiltin(464, "genQr")
	// s.DefineBuiltin(465, "scanQr")
	// s.DefineBuiltin(466, "getImageInfo")
	// s.DefineBuiltin(467, "resizeImage")
	s.DefineBuiltin(468, "createImage") // kept as builtin (alias to image.createImage)
	// Network communication functions - newFtpClient, newSshClient moved to ftp/ssh modules
	// s.DefineBuiltin(469, "newFtpClient")
	// s.DefineBuiltin(470, "newSshClient")
	// Excel/XLSX functions - use xlsx.create(), xlsx.open(), csv.read(), csv.write()
	// s.DefineBuiltin(471, "newExcel")
	// s.DefineBuiltin(472, "openExcel")
	// s.DefineBuiltin(484, "readCsv")
	// s.DefineBuiltin(485, "writeCsv")
	// Data format built-in functions (Batch 22)
	// XML functions - use xml.parse(), xml.parseFile(), xml.create()
	// s.DefineBuiltin(486, "parseXml")
	// s.DefineBuiltin(487, "parseXmlFile")
	// s.DefineBuiltin(488, "newXmlDoc")
	// YAML functions - use yaml.parse(), yaml.stringify(), yaml.toJson(), yaml.fromJson()
	// s.DefineBuiltin(489, "parseYaml")
	// s.DefineBuiltin(490, "toYaml")
	// s.DefineBuiltin(491, "yamlToJson")
	// s.DefineBuiltin(492, "jsonToYaml")
	// TOML functions - use toml.parse(), toml.encode(), toml.create(), toml.isValid()
	// s.DefineBuiltin(493, "parseToml")
	// s.DefineBuiltin(494, "toToml")
	// s.DefineBuiltin(495, "newToml")
	// s.DefineBuiltin(496, "tomlValid")
	// Email sending functions - use mail.newClient(), mail.send()
	// s.DefineBuiltin(486, "sendMail")
	// s.DefineBuiltin(487, "newMailClient")
	// Byte-index string functions
	s.DefineBuiltin(488, "byteIndexOf")
	s.DefineBuiltin(489, "byteSubstr")
	s.DefineBuiltin(490, "byteLen")
	// String enhancement (Batch 24)
	s.DefineBuiltin(491, "strCount")
	// Simple encoding (Batch 25)
	s.DefineBuiltin(492, "simpleEncode")
	s.DefineBuiltin(493, "simpleDecode")
	// Time enhancement (Batch 26)
	s.DefineBuiltin(494, "now")
	s.DefineBuiltin(495, "getNowStrCompact")
	s.DefineBuiltin(496, "timeToTimeStamp")
	s.DefineBuiltin(497, "timeStampToTime")
	// Print aliases (Batch 27)
	s.DefineBuiltin(503, "print")   // alias for pr
	s.DefineBuiltin(504, "println") // alias for pln
	s.DefineBuiltin(505, "printf")  // alias for prf
	s.DefineBuiltin(506, "concatBytes")
	s.DefineBuiltin(507, "plv")     // print value with %#v format
	s.DefineBuiltin(508, "spr")     // alias for sprintf
	return s
}

// NewEnclosedSymbolTable creates a new symbol table with an outer scope
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Define adds a new symbol to the symbol table
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.NumDefinitions}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.Store[name] = symbol
	s.NumDefinitions++
	return symbol
}

// Resolve finds a symbol in the symbol table or outer scopes
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := s.Store[name]
	// If we found a builtin, check if outer scope has a non-builtin with the same name
	// (local variables should shadow builtins)
	if ok && symbol.Scope == BuiltinScope && s.Outer != nil {
		outerSymbol, outerOk := s.Outer.Resolve(name)
		if outerOk && outerSymbol.Scope != BuiltinScope {
			// Found a non-builtin in outer scope, use that instead
			if outerSymbol.Scope == GlobalScope {
				return outerSymbol, true
			}
			// Add to free symbols if not already there
			free := Symbol{Name: name, Scope: FreeScope, Index: len(s.FreeSymbols)}
			s.FreeSymbols = append(s.FreeSymbols, outerSymbol)
			s.Store[name] = free
			return free, true
		}
	}
	if !ok && s.Outer != nil {
		symbol, ok = s.Outer.Resolve(name)
		if !ok {
			return symbol, false
		}

		// If we find a symbol in outer scope that's not global or builtin,
		// it becomes a free variable
		if symbol.Scope == GlobalScope || symbol.Scope == BuiltinScope {
			return symbol, true
		}

		// Add to free symbols if not already there
		free := Symbol{Name: name, Scope: FreeScope, Index: len(s.FreeSymbols)}
		s.FreeSymbols = append(s.FreeSymbols, symbol)
		s.Store[name] = free
		return free, true
	}
	return symbol, ok
}

// DefineBuiltin adds a built-in function symbol
func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Scope: BuiltinScope, Index: index}
	s.Store[name] = symbol
	return symbol
}

// CompilationScope represents a compilation scope
type CompilationScope struct {
	instructions []byte
}

// InlineableFuncInfo stores information about an inlineable function
type InlineableFuncInfo struct {
	ConstIndex int // Index in constants pool
	NumParams  int
	Body       []byte
}

// Bytecode represents compiled bytecode
type Bytecode struct {
	Instructions      []byte
	Constants         []objects.Object
	SourceMap         *SourceMap                  // Maps instruction positions to source locations
	InlineableGlobals map[int]*InlineableFuncInfo // Global index -> inlineable function info
}

// CompiledFunction represents a compiled function
type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
	NumRegs       int      // Maximum register used (for frame initialization)
	FreeVariables []Symbol // Free variables captured from outer scope
	Variadic      bool     // True if function has a variadic parameter

	// Inlining support
	IsInlineable bool   // True if function body is a single return expression
	InlineBody   []byte // Inlined bytecode (without return)
}

// Type returns the object type
func (cf *CompiledFunction) Type() objects.ObjectType { return objects.FunctionType }

// TypeTag returns the type tag for fast type checking
func (cf *CompiledFunction) TypeTag() objects.TypeTag { return objects.TagFunction }

// Inspect returns the string representation
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%d]", len(cf.Instructions))
}

// ToBool converts to boolean
func (cf *CompiledFunction) ToBool() *objects.Bool { return objects.TRUE }

// HashKey returns the hash key
func (cf *CompiledFunction) HashKey() objects.HashKey {
	return objects.HashKey{Type: objects.FunctionType, Value: 0}
}

// EmittedInstruction represents the last emitted instruction
type EmittedInstruction struct {
	Opcode   Opcode
	Position int
}

// Compiler transforms AST into bytecode
type Compiler struct {
	constants []objects.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction

	// Source mapping
	sourceMap  *SourceMap
	sourceFile string
	sourceCode string

	// Optimization context tracking
	safeArrayAccess map[string]bool // Track which array accesses are safe (bounds-checked by loop)

	// Inlining support
	inlineableGlobals map[int]*InlineableFuncInfo // Global index -> inlineable function info

	// Optimization options
	options OptimizationFlags

	// Loop context for break/continue support
	loopContexts []loopContext
}

// loopContext tracks break/continue positions within a loop
type loopContext struct {
	continuePos   int   // Position to jump to for continue (set after body is compiled)
	breakPos      []int // Positions of break jumps to patch
	continueJumps []int // Positions of continue jumps to patch (for for-loops)
}

// New creates a new compiler with default optimizations
func New() *Compiler {
	return &Compiler{
		constants:         []objects.Object{},
		symbolTable:       NewSymbolTable(),
		scopes:            []CompilationScope{{instructions: []byte{}}},
		scopeIndex:        0,
		sourceMap:         NewSourceMap(),
		safeArrayAccess:   make(map[string]bool),
		inlineableGlobals: make(map[int]*InlineableFuncInfo),
		options:           DefaultOptimizations(),
	}
}

// NewWithOptions creates a new compiler with custom optimization settings
func NewWithOptions(opts OptimizationFlags) *Compiler {
	return &Compiler{
		constants:         []objects.Object{},
		symbolTable:       NewSymbolTable(),
		scopes:            []CompilationScope{{instructions: []byte{}}},
		scopeIndex:        0,
		sourceMap:         NewSourceMap(),
		safeArrayAccess:   make(map[string]bool),
		inlineableGlobals: make(map[int]*InlineableFuncInfo),
		options:           opts,
	}
}

// NewWithState creates a new compiler with existing state
func NewWithState(s *SymbolTable, constants []objects.Object) *Compiler {
	return &Compiler{
		constants:         constants,
		symbolTable:       s,
		scopes:            []CompilationScope{{instructions: []byte{}}},
		scopeIndex:        0,
		sourceMap:         NewSourceMap(),
		safeArrayAccess:   make(map[string]bool),
		inlineableGlobals: make(map[int]*InlineableFuncInfo),
		options:           DefaultOptimizations(),
	}
}

// DefineGlobal defines a global variable in the symbol table
// This is used by runCode to pre-define arguments before compilation
func (c *Compiler) DefineGlobal(name string) Symbol {
	return c.symbolTable.Define(name)
}

// ResolveSymbol resolves a symbol by name
// This is used by runCode to find argument indices after compilation
func (c *Compiler) ResolveSymbol(name string) (Symbol, bool) {
	return c.symbolTable.Resolve(name)
}

// SetSource sets the source file path and code for error reporting
func (c *Compiler) SetSource(path string, code string) {
	c.sourceFile = path
	c.sourceCode = code
	c.sourceMap.SetSourceFile(path, code)
}

// Compile compiles an AST node into bytecode
func (c *Compiler) Compile(node parser.Node) error {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *parser.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}
		c.emit(OpPop)

	case *parser.VarStatement:
		// Pre-define the variable if the value is a function literal
		// This allows recursive functions like: var f = func() { f() }
		if fn, ok := node.Value.(*parser.FunctionLiteral); ok {
			// Define the variable first so the function body can reference it
			symbol := c.symbolTable.Define(node.Name.Value)
			// Set the function name for better error messages
			fn.Name = node.Name.Value
			// Compile the function literal
			if err := c.Compile(fn); err != nil {
				return err
			}
			// Store the function in the variable
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			}
		} else {
			if err := c.Compile(node.Value); err != nil {
				return err
			}
			symbol := c.symbolTable.Define(node.Name.Value)
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			}
		}

	case *parser.ShortVarStatement:
		// Short variable declaration (:=) - same semantics as var but simpler syntax
		// Pre-define the variable if the value is a function literal
		if fn, ok := node.Value.(*parser.FunctionLiteral); ok {
			symbol := c.symbolTable.Define(node.Name.Value)
			fn.Name = node.Name.Value
			if err := c.Compile(fn); err != nil {
				return err
			}
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			}
		} else {
			if err := c.Compile(node.Value); err != nil {
				return err
			}
			symbol := c.symbolTable.Define(node.Name.Value)
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			}
		}

	case *parser.ConstStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		switch symbol.Scope {
		case GlobalScope:
			c.emit(OpSetGlobal, symbol.Index)
		case LocalScope:
			c.emit(OpSetLocal, symbol.Index)
		}

	case *parser.ReturnStatement:
		if node.ReturnValue != nil {
			// Check for tail call optimization
			if callExpr, ok := node.ReturnValue.(*parser.CallExpression); ok {
				// This is a tail call - compile as tail call
				if err := c.compileTailCall(callExpr); err != nil {
					return err
				}
				return nil
			}
			if err := c.Compile(node.ReturnValue); err != nil {
				return err
			}
		} else {
			c.emit(OpNull)
		}
		c.emit(OpReturn)

	case *parser.BlockStatement:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *parser.IfStatement:
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		// Jump to else/end if false
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Compile consequence
		if err := c.Compile(node.Consequence); err != nil {
			return err
		}

		// Remove the last OpPop if the block has a return value
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Jump over alternative
		jumpPos := c.emit(OpJump, 9999)

		// Fix jump to else/end position
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative != nil {
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
			// Remove the last OpPop if the block has a return value
			if c.lastInstruction.Opcode == OpPop {
				c.removeLastInstruction()
			}
		} else {
			c.emit(OpNull)
		}

		// Fix jump over alternative
		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

		// Add pop for if expression
		c.emit(OpPop)

	case *parser.SwitchStatement:
		// Compile switch expression (value on stack)
		if err := c.Compile(node.Expression); err != nil {
			return err
		}

		// Track positions for patching jumps to end
		var endJumps []int

		// Track positions for "jump to next case" to patch later
		var caseJumpPositions []int

		// Compile each case
		for i, caseStmt := range node.Cases {
			// Duplicate switch value for comparison
			c.emit(OpDup)

			// Compile case expression
			if err := c.Compile(caseStmt.Expression); err != nil {
				return err
			}

			// Compare: switchValue == caseValue
			c.emit(OpEqual)

			// Jump to next case/default if not matched
			jumpNotMatchedPos := c.emit(OpJumpIfFalse, 9999)

			// Match found: pop the duplicated switch value
			c.emit(OpPop)

			// Compile case body
			if err := c.Compile(caseStmt.Consequence); err != nil {
				return err
			}

			// Remove trailing pop if present
			if c.lastInstruction.Opcode == OpPop {
				c.removeLastInstruction()
			}

			// Jump to end of switch
			endJumpPos := c.emit(OpJump, 9999)
			endJumps = append(endJumps, endJumpPos)

			// Patch "not matched" jump to next case
			nextPos := len(c.currentInstructions())
			c.changeOperand(jumpNotMatchedPos, nextPos)

			// Store position in case we need to chain cases
			caseJumpPositions = append(caseJumpPositions, jumpNotMatchedPos)

			// If this is the last case and there's no default,
			// we need to pop the switch value
			if i == len(node.Cases)-1 && node.Default == nil {
				// This is where the "not matched" jump lands
				// Pop the switch value
				c.emit(OpPop)
			}
		}

		// Compile default case if exists
		if node.Default != nil {
			// At this point, switch value is still on stack (from failed case matches)
			// Pop it since we don't need it for default
			c.emit(OpPop)

			// Compile default body
			if err := c.Compile(node.Default.Consequence); err != nil {
				return err
			}

			// Remove trailing pop if present
			if c.lastInstruction.Opcode == OpPop {
				c.removeLastInstruction()
			}
		}

		// Patch all "jump to end" positions
		endPos := len(c.currentInstructions())
		for _, pos := range endJumps {
			c.changeOperand(pos, endPos)
		}

		// Add pop for switch expression
		c.emit(OpPop)

	case *parser.WhileStatement:
		// Save position for loop start
		loopStart := len(c.currentInstructions())

		// Push loop context for break/continue tracking
		c.pushLoopContext(loopStart)

		// Compile condition
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		// Jump if false
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Remove the last OpPop from body
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Jump back to loop start
		c.emit(OpJump, loopStart)

		// Fix jump position
		afterBodyPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterBodyPos)

		// Pop loop context and patch break positions
		c.popLoopContext(afterBodyPos)

	case *parser.ForStatement:
		// for (init; condition; update) { body }
		// Compile init
		if node.Init != nil {
			if err := c.Compile(node.Init); err != nil {
				return err
			}
		}

		// Save position for loop start
		loopStart := len(c.currentInstructions())

		// Push loop context (initially continue goes to condition, will update later)
		c.pushLoopContext(loopStart)

		// Compile condition (if none, use true)
		if node.Condition != nil {
			if err := c.Compile(node.Condition); err != nil {
				return err
			}
		} else {
			c.emit(OpTrue)
		}

		// Jump if false
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Analyze loop for safe array access patterns
		// Pattern: for i := 0; i < len(arr); i++ { ... arr[i] ... }
		c.analyzeLoopSafety(node)

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Remove the last OpPop from body
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Record update position for continue (continue should jump to update)
		if node.Update != nil {
			updatePos := len(c.currentInstructions())
			// Patch all continue jumps to point to update
			c.patchContinueJumps(updatePos)
			// Update the continuePos for any continue statements in the update itself
			if len(c.loopContexts) > 0 {
				c.loopContexts[len(c.loopContexts)-1].continuePos = updatePos
			}

			if err := c.Compile(node.Update); err != nil {
				return err
			}
			// Remove the result of update statement
			if c.lastInstruction.Opcode == OpPop {
				c.removeLastInstruction()
			}
		}

		// Jump back to loop start
		c.emit(OpJump, loopStart)

		// Fix jump position
		afterBodyPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterBodyPos)

		// Pop loop context and patch break positions
		c.popLoopContext(afterBodyPos)

		// Clear safe array access tracking after loop
		c.safeArrayAccess = make(map[string]bool)

	case *parser.ForInStatement:
		// for (key, value in iterable) { body }
		// for (value in iterable) { body }

		// Compile iterable
		if err := c.Compile(node.Iterable); err != nil {
			return err
		}

		// Initialize index to 0
		indexConst := c.addConstant(objects.NewInt(0))
		c.emit(OpConstant, indexConst)

		// Initialize iterator to null
		c.emit(OpNull)

		// Loop start
		loopStart := len(c.currentInstructions())

		// Push loop context for break/continue tracking
		c.pushLoopContext(loopStart)

		// Jump if finished (when iterator is null after iteration)
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Duplicate current iterator state
		c.emit(OpDup)

		// Set value variable (or key if only one variable)
		if node.Value != nil {
			symbol := c.symbolTable.Define(node.Value.Value)
			c.emit(OpSetGlobal, symbol.Index)
		}

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Remove the last OpPop from body
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Jump back to loop start
		c.emit(OpJump, loopStart)

		// Fix jump position
		afterBodyPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterBodyPos)

		// Pop loop context and patch break positions
		c.popLoopContext(afterBodyPos)

	case *parser.BreakStatement:
		// Emit jump to after loop (will be patched when loop ends)
		breakPos := c.emit(OpJump, 9999)
		c.addBreakPos(breakPos)

	case *parser.ContinueStatement:
		// Emit jump to continue target (will be patched for for-loops)
		// For while loops, use the current continuePos directly
		continuePos := c.currentLoopContinuePos()
		if continuePos >= 0 {
			jumpPos := c.emit(OpJump, continuePos)
			// Also track for possible repatching (for for-loops with update)
			c.addContinueJump(jumpPos)
		}

	case *parser.IntegerLiteral:
		integer := objects.NewInt(node.Value)
		c.emit(OpConstant, c.addConstant(integer))

	case *parser.FloatLiteral:
		float := &objects.Float{Value: node.Value}
		c.emit(OpConstant, c.addConstant(float))

	case *parser.BigIntLiteral:
		bigInt, err := objects.NewBigIntFromString(node.Value)
		if err != nil {
			return fmt.Errorf("line %d:%d: could not parse %q as big int: %v",
				node.Token.Line, node.Token.Column, node.Value, err)
		}
		c.emit(OpConstant, c.addConstant(bigInt))

	case *parser.BigFloatLiteral:
		bigFloat, err := objects.NewBigFloatFromString(node.Value)
		if err != nil {
			return fmt.Errorf("line %d:%d: could not parse %q as big float: %v",
				node.Token.Line, node.Token.Column, node.Value, err)
		}
		c.emit(OpConstant, c.addConstant(bigFloat))

	case *parser.StringLiteral:
		str := objects.InternString(node.Value)
		c.emit(OpConstant, c.addConstant(str))

	case *parser.BooleanLiteral:
		if node.Value {
			c.emit(OpTrue)
		} else {
			c.emit(OpFalse)
		}

	case *parser.NullLiteral:
		c.emit(OpNull)

	case *parser.ArrayLiteral:
		for _, el := range node.Elements {
			if err := c.Compile(el); err != nil {
				return err
			}
		}
		c.emit(OpArray, len(node.Elements))

	case *parser.MapLiteral:
		// Sort keys for deterministic order
		keys := make([]parser.Expression, 0, len(node.Pairs))
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		for _, k := range keys {
			if err := c.Compile(k); err != nil {
				return err
			}
			if err := c.Compile(node.Pairs[k]); err != nil {
				return err
			}
		}
		c.emit(OpMap, len(node.Pairs))

	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		switch symbol.Scope {
		case GlobalScope:
			c.emit(OpGetGlobal, symbol.Index)
		case LocalScope:
			c.emit(OpGetLocal, symbol.Index)
		case BuiltinScope:
			c.emit(OpBuiltin, symbol.Index)
		case FreeScope:
			c.emit(OpGetFree, symbol.Index)
		}

	case *parser.PrefixExpression:
		switch node.Operator {
		case "++", "--":
			// Prefix increment/decrement: ++x or --x
			// Returns the new value (unlike postfix which returns old value)
			switch right := node.Right.(type) {
			case *parser.Identifier:
				symbol, ok := c.symbolTable.Resolve(right.Value)
				if !ok {
					return fmt.Errorf("undefined variable %s", right.Value)
				}

				// Get current value
				switch symbol.Scope {
				case GlobalScope:
					c.emit(OpGetGlobal, symbol.Index)
				case LocalScope:
					c.emit(OpGetLocal, symbol.Index)
				case FreeScope:
					c.emit(OpGetFree, symbol.Index)
				}

				// Add 1 or subtract 1
				one := c.addConstant(objects.NewInt(1))
				c.emit(OpConstant, one)

				switch node.Operator {
				case "++":
					c.emit(OpAdd)
				case "--":
					c.emit(OpSub)
				}

				// Store result (OpSetGlobal pushes the value back)
				switch symbol.Scope {
				case GlobalScope:
					c.emit(OpSetGlobal, symbol.Index)
				case LocalScope:
					c.emit(OpSetLocal, symbol.Index)
				case FreeScope:
					c.emit(OpSetFree, symbol.Index)
				}
				// Result is already on stack from OpSetGlobal

			default:
				return fmt.Errorf("prefix %s operator not supported for type: %T", node.Operator, right)
			}

		default:
			// Handle other prefix operators: - and !
			if err := c.Compile(node.Right); err != nil {
				return err
			}

			switch node.Operator {
			case "-":
				c.emit(OpNeg)
			case "!":
				c.emit(OpNot)
			default:
				return fmt.Errorf("unknown operator %s", node.Operator)
			}
		}

	case *parser.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(OpAdd)
		case "-":
			c.emit(OpSub)
		case "*":
			c.emit(OpMul)
		case "/":
			c.emit(OpDiv)
		case "%":
			c.emit(OpMod)
		case "==":
			c.emit(OpEqual)
		case "!=":
			c.emit(OpNotEqual)
		case "<":
			c.emit(OpLess)
		case ">":
			c.emit(OpGreater)
		case "<=":
			c.emit(OpLessEqual)
		case ">=":
			c.emit(OpGreaterEqual)
		case "&&":
			c.emit(OpAnd)
		case "||":
			c.emit(OpOr)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *parser.IndexExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}

		// Check if this is a safe array access (bounds-checked by loop)
		// Pattern: arr[i] where i is the loop variable and arr is the array being iterated
		if arr, ok := node.Left.(*parser.Identifier); ok {
			if idx, ok := node.Index.(*parser.Identifier); ok {
				if c.isArrayAccessSafe(arr.Value, idx.Value) {
					c.emit(OpIndexSafe)
				} else {
					c.emit(OpIndex)
				}
			} else {
				c.emit(OpIndex)
			}
		} else {
			c.emit(OpIndex)
		}

	case *parser.SliceExpression:
		// Compile the left expression (array or string)
		if err := c.Compile(node.Left); err != nil {
			return err
		}

		// Compile start index (or push null if nil)
		if node.Start != nil {
			if err := c.Compile(node.Start); err != nil {
				return err
			}
		} else {
			c.emit(OpNull)
		}

		// Compile end index (or push null if nil)
		if node.End != nil {
			if err := c.Compile(node.End); err != nil {
				return err
			}
		} else {
			c.emit(OpNull)
		}

		c.emit(OpSlice)

	case *parser.AssignmentExpression:
		// Compile the value first
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		// Handle different left-hand side types
		switch left := node.Left.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpSetFree, symbol.Index)
			}
		case *parser.IndexExpression:
			// For arr[i] = value, we need stack to be: [arr, index, value]
			// But we already compiled value, which is on stack
			// We need to compile arr and index, then swap to get correct order
			if err := c.Compile(left.Left); err != nil {
				return err
			}
			if err := c.Compile(left.Index); err != nil {
				return err
			}
			// Stack is now: [value, arr, index]
			// We need: [arr, index, value]
			// OpSetIndex pops: value, index, left
			// So we need to reorder: pop value, push arr, push index, push value
			// Actually, let's change the VM to handle the order correctly
			c.emit(OpSetIndex)
		case *parser.DotExpression:
			// For obj.field = value, compile object then emit OpSetField
			if err := c.Compile(left.Object); err != nil {
				return err
			}
			// Stack is now: [value, object]
			// Add field name constant
			nameIdx := c.addConstant(objects.InternString(left.Property.Value))
			c.emit(OpSetField, nameIdx)
		default:
			return fmt.Errorf("cannot assign to %T", left)
		}

	case *parser.FunctionLiteral:
		// If named function, define the name first for recursion support
		var funcSymbol Symbol
		if node.Name != "" {
			funcSymbol = c.symbolTable.Define(node.Name)
		}

		// Enter function scope
		c.enterScope()

		// Define parameters as local variables
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Ensure function ends with return
		if c.lastInstruction.Opcode != OpReturn {
			c.emit(OpNull)
			c.emit(OpReturn)
		}

		// Leave function scope and get compiled function
		compiledFn := c.leaveScope()
		compiledFn.NumParameters = len(node.Parameters)

		// Add function to constants
		fnIndex := c.addConstant(compiledFn)

		// Emit OpClosure with function index and number of free variables
		// For each free variable, emit code to push its value onto the stack
		for _, freeVar := range compiledFn.FreeVariables {
			// Push the value of each free variable onto the stack
			// This will be captured in the closure
			switch freeVar.Scope {
			case GlobalScope:
				c.emit(OpGetGlobal, freeVar.Index)
			case LocalScope:
				c.emit(OpGetLocal, freeVar.Index)
			case FreeScope:
				// Free variable in outer function - get it from outer's free vars
				c.emit(OpGetFree, freeVar.Index)
			}
		}

		// Emit closure instruction
		c.emit(OpClosure, fnIndex, len(compiledFn.FreeVariables))

		// If named function, bind it to its name
		if node.Name != "" {
			switch funcSymbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, funcSymbol.Index)
				// Track inlineable functions for optimization
				if compiledFn.IsInlineable && len(compiledFn.FreeVariables) == 0 {
					c.inlineableGlobals[funcSymbol.Index] = &InlineableFuncInfo{
						ConstIndex: fnIndex,
						NumParams:  compiledFn.NumParameters,
						Body:       compiledFn.InlineBody,
					}
				}
			case LocalScope:
				c.emit(OpSetLocal, funcSymbol.Index)
			}
		}

	case *parser.CallExpression:
		// Check if this is a method call (obj.method())
		if dot, ok := node.Function.(*parser.DotExpression); ok {
			// Compile the object
			if err := c.Compile(dot.Object); err != nil {
				return err
			}
			// Get the method name
			nameConst := c.addConstant(objects.InternString(dot.Property.Value))
			c.emit(OpGetMethod, nameConst)
			// Compile arguments
			for _, arg := range node.Arguments {
				if err := c.Compile(arg); err != nil {
					return err
				}
			}
			// Call method with this binding
			c.emit(OpCallMethod, len(node.Arguments))
		} else {
			// Regular function call
			// Note: Function inlining is disabled due to complexity with closure semantics
			// Compile function
			if err := c.Compile(node.Function); err != nil {
				return err
			}

			// Compile arguments
			for _, arg := range node.Arguments {
				if err := c.Compile(arg); err != nil {
					return err
				}
			}

			c.emit(OpCall, len(node.Arguments))
		}

	case *parser.DotExpression:
		// Compile object
		if err := c.Compile(node.Object); err != nil {
			return err
		}
		// Add property name to constants
		nameConst := c.addConstant(objects.InternString(node.Property.Value))
		c.emit(OpGetMethod, nameConst)

	case *parser.TernaryExpression:
		// condition ? consequent : alternative
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		jumpFalsePos := c.emit(OpJumpIfFalse, 9999)

		if err := c.Compile(node.Consequent); err != nil {
			return err
		}

		jumpEndPos := c.emit(OpJump, 9999)

		afterConsequentPos := len(c.currentInstructions())
		c.changeOperand(jumpFalsePos, afterConsequentPos)

		if err := c.Compile(node.Alternative); err != nil {
			return err
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpEndPos, afterAlternativePos)

	case *parser.CompoundAssignmentExpression:
		// x += 1 is equivalent to x = x + 1
		// Get the current value
		switch left := node.Left.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}

			// Get current value
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpGetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpGetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpGetFree, symbol.Index)
			}

			// Compile right side
			if err := c.Compile(node.Right); err != nil {
				return err
			}

			// Apply operation
			switch node.Operator {
			case "+=":
				c.emit(OpAdd)
			case "-=":
				c.emit(OpSub)
			case "*=":
				c.emit(OpMul)
			case "/=":
				c.emit(OpDiv)
			case "%=":
				c.emit(OpMod)
			}

			// Store result
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpSetFree, symbol.Index)
			}
		}

	case *parser.PostfixExpression:
		// x++ or x--
		switch left := node.Left.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}

			// Get current value
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpGetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpGetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpGetFree, symbol.Index)
			}

			// Add 1 or subtract 1
			one := c.addConstant(objects.NewInt(1))
			c.emit(OpConstant, one)

			switch node.Operator {
			case "++":
				c.emit(OpAdd)
			case "--":
				c.emit(OpSub)
			}

			// Store result (OpSetGlobal pushes the value back)
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpSetFree, symbol.Index)
			}
			// Result is the new value on stack from OpSetGlobal
			// For postfix, we need to return old value, so we decrement/increment back
			one2 := c.addConstant(objects.NewInt(1))
			c.emit(OpConstant, one2)
			switch node.Operator {
			case "++":
				c.emit(OpSub) // new - 1 = old
			case "--":
				c.emit(OpAdd) // new + 1 = old
			}
		}

	case *parser.ImportStatement:
		return c.compileImportStatement(node)

	case *parser.ExportStatement:
		return c.compileExportStatement(node)

	case *parser.ClassStatement:
		return c.compileClassStatement(node)

	case *parser.NewExpression:
		return c.compileNewExpression(node)

	case *parser.ThisExpression:
		return c.compileThisExpression(node)

	case *parser.SuperCallExpression:
		return c.compileSuperCallExpression(node)

	case *parser.TryStatement:
		return c.compileTryStatement(node)

	case *parser.ThrowStatement:
		return c.compileThrowStatement(node)

	default:
		return fmt.Errorf("unknown node type: %T", node)
	}

	return nil
}

// compileTailCall compiles a tail call expression
// A tail call is a function call that is the last operation in a function
// Instead of creating a new call frame, we reuse the current one
func (c *Compiler) compileTailCall(node *parser.CallExpression) error {
	// Compile function
	if err := c.Compile(node.Function); err != nil {
		return err
	}

	// Compile arguments
	for _, arg := range node.Arguments {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	// Emit tail call instruction instead of OpCall + OpReturn
	c.emit(OpTailCall, len(node.Arguments))
	return nil
}

// compileImportStatement compiles an import statement
func (c *Compiler) compileImportStatement(node *parser.ImportStatement) error {
	// Load the module path constant
	pathIdx := c.addConstant(objects.InternString(node.Path.Value))

	// Emit OpLoadModule to load the module and push it onto the stack
	c.emit(OpLoadModule, pathIdx)

	// Handle different import styles
	if node.Name != nil {
		// Default import: import math from "./math"
		// The module (default export) is on the stack, store it in global
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(OpSetGlobal, symbol.Index)
		c.emit(OpPop) // Pop the pushed-back value from OpSetGlobal
	} else if node.Alias != nil {
		// Namespace import: import * as math from "./math"
		// The module object is on the stack, store it in global
		symbol := c.symbolTable.Define(node.Alias.Value)
		c.emit(OpSetGlobal, symbol.Index)
		c.emit(OpPop) // Pop the pushed-back value from OpSetGlobal
	} else if len(node.Names) > 0 {
		// Destructuring import: import { add, sub } from "./math"
		// The module object is on the stack
		for _, name := range node.Names {
			// Duplicate module reference for each name
			c.emit(OpDup)
			// Get the export by name
			nameIdx := c.addConstant(objects.InternString(name.Value))
			c.emit(OpGetExport, nameIdx)
			// Store in global
			symbol := c.symbolTable.Define(name.Value)
			c.emit(OpSetGlobal, symbol.Index)
			// Pop the pushed-back value from OpSetGlobal
			c.emit(OpPop)
		}
		// Pop the original module reference
		c.emit(OpPop)
	} else {
		// Simple import: import "time" or import "./math"
		// Auto-bind module name extracted from path
		path := node.Path.Value

		// Extract module name from path
		// Examples: "time" -> "time", "./math" -> "math", "math" -> "math"
		moduleName := extractModuleName(path)

		if moduleName != "" {
			// Store the module as a global variable
			symbol := c.symbolTable.Define(moduleName)
			c.emit(OpSetGlobal, symbol.Index)
			c.emit(OpPop) // Pop the pushed-back value from OpSetGlobal
		} else {
			// Can't extract name, just load for side effects
			c.emit(OpPop)
		}
	}

	return nil
}

// compileExportStatement compiles an export statement
func (c *Compiler) compileExportStatement(node *parser.ExportStatement) error {
	// Handle different export types
	switch stmt := node.Exportable.(type) {
	case *parser.VarStatement:
		// Compile the value expression
		if err := c.Compile(stmt.Value); err != nil {
			return err
		}
		// Define the variable in the symbol table
		symbol := c.symbolTable.Define(stmt.Name.Value)
		// Store in global (OpSetGlobal pushes the value back)
		c.emit(OpSetGlobal, symbol.Index)
		// Export the variable (value is already on stack from OpSetGlobal)
		nameIdx := c.addConstant(objects.InternString(stmt.Name.Value))
		c.emit(OpSetExport, nameIdx)
		// Pop the pushed-back value from OpSetExport
		c.emit(OpPop)

	case *parser.ConstStatement:
		// Compile the value expression
		if err := c.Compile(stmt.Value); err != nil {
			return err
		}
		// Define the constant in the symbol table
		symbol := c.symbolTable.Define(stmt.Name.Value)
		// Store in global (OpSetGlobal pushes the value back)
		c.emit(OpSetGlobal, symbol.Index)
		// Export the constant (value is already on stack from OpSetGlobal)
		nameIdx := c.addConstant(objects.InternString(stmt.Name.Value))
		c.emit(OpSetExport, nameIdx)
		// Pop the pushed-back value from OpSetExport
		c.emit(OpPop)

	case *parser.ExpressionStatement:
		// Handle function exports: export func add(a, b) { ... }
		if fn, ok := stmt.Expression.(*parser.FunctionLiteral); ok {
			if fn.Name == "" {
				return fmt.Errorf("exported function must have a name")
			}
			// Compile the function - this will emit OpClosure and OpSetGlobal
			if err := c.Compile(fn); err != nil {
				return err
			}
			// The function is now stored in the global by FunctionLiteral compilation
			// OpSetGlobal pushed the value back, so it's on stack
			// Export the function (value is already on stack from OpSetGlobal)
			nameIdx := c.addConstant(objects.InternString(fn.Name))
			c.emit(OpSetExport, nameIdx)
			// Pop the pushed-back value from OpSetExport
			c.emit(OpPop)
			return nil
		}
		return fmt.Errorf("unsupported export expression type: %T", stmt.Expression)

	default:
		return fmt.Errorf("unsupported export type: %T", stmt)
	}

	return nil
}

// Bytecode returns the compiled bytecode
func (c *Compiler) Bytecode() *Bytecode {
	bytecode := &Bytecode{
		Instructions:      c.currentInstructions(),
		Constants:         c.constants,
		SourceMap:         c.sourceMap,
		InlineableGlobals: c.inlineableGlobals,
	}

	// Apply optimizations if enabled
	if c.options.BytecodeOptimizer {
		optimizer := NewOptimizerWithFlags(bytecode, c.options)
		return optimizer.Optimize()
	}

	return bytecode
}

// emit adds an instruction to the bytecode
func (c *Compiler) emit(op Opcode, operands ...int) int {
	ins := Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

// addInstruction adds an instruction to the current scope
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	return posNewInstruction
}

// setLastInstruction updates the last instruction tracking
func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	previous := c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	c.previousInstruction = previous
}

// lastInstructionIs returns true if the last instruction matches
func (c *Compiler) lastInstructionIs(op Opcode) bool {
	return c.lastInstruction.Opcode == op
}

// removeLastInstruction removes the last instruction
func (c *Compiler) removeLastInstruction() {
	ins := c.currentInstructions()
	c.scopes[c.scopeIndex].instructions = ins[:len(ins)-1]
	c.lastInstruction = c.previousInstruction
}

// replaceInstruction replaces an instruction at a position
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

// changeOperand changes the operand of an instruction
func (c *Compiler) changeOperand(pos int, operand int) {
	op := Opcode(c.currentInstructions()[pos])
	newInstruction := Make(op, operand)
	c.replaceInstruction(pos, newInstruction)
}

// changeOperands changes multiple operands of an instruction
func (c *Compiler) changeOperands(pos int, operands ...int) {
	op := Opcode(c.currentInstructions()[pos])
	newInstruction := Make(op, operands...)
	c.replaceInstruction(pos, newInstruction)
}

// currentInstructions returns the current scope's instructions
func (c *Compiler) currentInstructions() []byte {
	return c.scopes[c.scopeIndex].instructions
}

// addConstant adds a constant to the constant pool
func (c *Compiler) addConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// enterScope enters a new compilation scope
func (c *Compiler) enterScope() {
	scope := CompilationScope{instructions: []byte{}}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope leaves the current scope and returns the compiled function
func (c *Compiler) leaveScope() *CompiledFunction {
	instructions := c.currentInstructions()

	// Get the number of local variables in this scope
	numLocals := c.symbolTable.NumDefinitions

	// Capture free variables before leaving scope
	freeVars := make([]Symbol, len(c.symbolTable.FreeSymbols))
	copy(freeVars, c.symbolTable.FreeSymbols)

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	// Analyze if function is inlineable (single return expression)
	isInlineable, inlineBody := c.analyzeInlineable(instructions)

	return &CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		FreeVariables: freeVars,
		IsInlineable:  isInlineable,
		InlineBody:    inlineBody,
	}
}

// analyzeInlineable checks if a function body is a single return expression
// that can be inlined at call sites
// Returns true and the inlinable body (without return) if the function is inlineable
func (c *Compiler) analyzeInlineable(instructions []byte) (bool, []byte) {
	if len(instructions) < 2 {
		return false, nil
	}

	// Function is inlineable if it ends with a single value return
	// and doesn't contain side effects (function calls, assignments)
	// or control flow (jumps, multiple returns)

	// Check if ends with OpReturn preceded by value-producing instructions
	lastIP := len(instructions) - 1
	if Opcode(instructions[lastIP]) != OpReturn {
		return false, nil
	}

	// Simple heuristic: if function body is just a few instructions without
	// side effects, it's inlineable
	// Count instructions and check for side effects
	sideEffectOpcodes := map[Opcode]bool{
		OpSetGlobal:   true,
		OpSetLocal:    true,
		OpSetFree:     true,
		OpSetIndex:    true,
		OpSetField:    true,
		OpCall:        true,
		OpTailCall:    true,
		OpCallMethod:  true,
		OpPop:         true,
		OpJump:        true,
		OpJumpIfFalse: true,
		OpJumpIfTrue:  true,
		OpReturn:      true, // Multiple returns not allowed
		OpBreak:       true,
		OpContinue:    true,
	}

	numInstructions := 0
	i := 0
	for i < len(instructions)-1 { // Exclude final return
		op := Opcode(instructions[i])

		if sideEffectOpcodes[op] {
			return false, nil
		}

		// Count instruction
		numInstructions++
		i++

		// Skip operands
		def, err := Lookup(byte(op))
		if err != nil {
			return false, nil
		}
		for _, w := range def.OperandWidths {
			i += w
		}
	}

	// Only inline simple functions (up to 15 instructions)
	// to avoid code bloat. Increased from 10 to 15 to allow more expressions.
	if numInstructions > 15 {
		return false, nil
	}

	// Return body without the OpReturn
	return true, instructions[:lastIP]
}

// compileMethod compiles a method without binding it to a global variable.
// Methods are stored in the class's methods map, not as standalone globals.
func (c *Compiler) compileMethod(node *parser.FunctionLiteral) error {
	// Enter function scope
	c.enterScope()

	// Reserve local slot 0 for 'this' (set by VM when calling method)
	// Define a dummy symbol at index 0 for 'this'
	c.symbolTable.Define("this")

	// Define parameters as local variables (starting at index 1)
	for _, p := range node.Parameters {
		c.symbolTable.Define(p.Value)
	}

	// Compile body
	if err := c.Compile(node.Body); err != nil {
		return err
	}

	// Ensure function ends with return
	if c.lastInstruction.Opcode != OpReturn {
		c.emit(OpNull)
		c.emit(OpReturn)
	}

	// Leave function scope and get compiled function
	compiledFn := c.leaveScope()
	compiledFn.NumParameters = len(node.Parameters)

	// Add function to constants
	fnIndex := c.addConstant(compiledFn)

	// Emit OpClosure with function index and number of free variables
	// For each free variable, emit code to push its value onto the stack
	for _, freeVar := range compiledFn.FreeVariables {
		// Push the value of each free variable onto the stack
		// This will be captured in the closure
		switch freeVar.Scope {
		case GlobalScope:
			c.emit(OpGetGlobal, freeVar.Index)
		case LocalScope:
			c.emit(OpGetLocal, freeVar.Index)
		case FreeScope:
			// Free variable in outer function - get it from outer's free vars
			c.emit(OpGetFree, freeVar.Index)
		}
	}

	// Emit closure instruction
	c.emit(OpClosure, fnIndex, len(compiledFn.FreeVariables))

	// Note: We do NOT bind method names to globals like we do for named functions
	// Methods are only accessible through the class's methods map

	return nil
}

// compileClassStatement compiles a class declaration
func (c *Compiler) compileClassStatement(node *parser.ClassStatement) error {
	// Compile superclass (push null if none)
	if node.SuperClass != nil {
		symbol, ok := c.symbolTable.Resolve(node.SuperClass.Value)
		if !ok {
			return fmt.Errorf("undefined superclass: %s", node.SuperClass.Value)
		}
		c.emit(OpGetGlobal, symbol.Index)
	} else {
		c.emit(OpNull)
	}

	// Compile default fields as key-value pairs
	for _, field := range node.Fields {
		// Key
		nameIdx := c.addConstant(objects.InternString(field.Name.Value))
		c.emit(OpConstant, nameIdx)
		// Value
		if err := c.Compile(field.Value); err != nil {
			return err
		}
	}
	// Create fields map from key-value pairs
	c.emit(OpMap, len(node.Fields))

	// Compile methods as key-value pairs
	for _, method := range node.Methods {
		// Key (method name)
		nameIdx := c.addConstant(objects.InternString(method.Name))
		c.emit(OpConstant, nameIdx)
		// Compile method as function (with 'this' at local 0)
		if err := c.compileMethod(method); err != nil {
			return err
		}
	}
	// Create methods map from key-value pairs
	c.emit(OpMap, len(node.Methods))

	// Create class
	nameIdx := c.addConstant(objects.InternString(node.Name.Value))
	c.emit(OpClass, nameIdx)

	// Store class in global
	symbol := c.symbolTable.Define(node.Name.Value)
	c.emit(OpSetGlobal, symbol.Index)

	return nil
}

// compileNewExpression compiles a new expression
func (c *Compiler) compileNewExpression(node *parser.NewExpression) error {
	// Get class
	symbol, ok := c.symbolTable.Resolve(node.Class.String())
	if !ok {
		return fmt.Errorf("undefined class: %s", node.Class.String())
	}
	c.emit(OpGetGlobal, symbol.Index)

	// Compile arguments
	for _, arg := range node.Arguments {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	// Create instance
	c.emit(OpNew, len(node.Arguments))

	return nil
}

// compileThisExpression compiles this expression
func (c *Compiler) compileThisExpression(node *parser.ThisExpression) error {
	// Push current instance reference onto stack
	// This is handled at VM level - we emit a special marker
	c.emit(OpGetLocal, 0) // this is always first local in method context
	return nil
}

// compileSuperCallExpression compiles a super.method() call
func (c *Compiler) compileSuperCallExpression(node *parser.SuperCallExpression) error {
	// Push this first (as the instance for OpCallMethod)
	c.emit(OpGetLocal, 0)

	// Get super method
	nameIdx := c.addConstant(objects.InternString(node.Method))
	c.emit(OpSuper, nameIdx)

	// Compile arguments (not including this)
	for _, arg := range node.Args {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	// Call method with OpCallMethod (which handles this separately)
	c.emit(OpCallMethod, len(node.Args))

	return nil
}

// getCompiledFunction retrieves a compiled function by symbol index
// This is used for inlining analysis
func (c *Compiler) getCompiledFunction(symbol Symbol) (*CompiledFunction, bool) {
	// For local variables, we can't easily get the function at compile time
	// For globals, check if the symbol is defined and retrieve from constants
	if symbol.Scope == GlobalScope {
		// Look up in constants if it's a compiled function
		if symbol.Index < len(c.constants) {
			if fn, ok := c.constants[symbol.Index].(*CompiledFunction); ok {
				return fn, true
			}
		}
	}
	return nil, false
}

// inlineFunction inlines a function body at the call site
func (c *Compiler) inlineFunction(fn *CompiledFunction, args []parser.Expression, symbolIndex int) error {
	// For inlineable functions, we need to:
	// 1. Evaluate arguments
	// 2. Map parameter references to argument values
	// 3. Emit inlined body with adjusted local references

	// Compile all arguments first
	argLocals := make([]int, len(args))
	for i, arg := range args {
		if err := c.Compile(arg); err != nil {
			return err
		}
		// Store argument in a temporary local
		tempLocal := c.symbolTable.NumDefinitions + i
		c.emit(OpSetLocal, tempLocal)
		argLocals[i] = tempLocal
	}

	// Emit inlined body
	// The inline body expects parameters to be in locals 0, 1, 2, ...
	// We need to adjust these to reference the argument locals

	// For simplicity, we emit a simpler inlining strategy:
	// Just emit the inline body and let the function's locals be used
	// This works for pure functions without side effects

	for i := 0; i < len(fn.InlineBody); {
		op := Opcode(fn.InlineBody[i])

		// Check if this is a local access that needs remapping
		if op == OpGetLocal {
			// Remap parameter index to argument local
			paramIndex := int(fn.InlineBody[i+1])
			if paramIndex < len(argLocals) {
				// Emit GetLocal for the argument
				c.emit(OpGetLocal, argLocals[paramIndex])
			} else {
				// Non-parameter local, emit as-is
				c.emit(OpGetLocal, paramIndex)
			}
			i += 2
		} else {
			// Copy instruction as-is
			def, err := Lookup(byte(op))
			if err != nil {
				i++
				continue
			}
			instrLen := 1
			operands := make([]int, len(def.OperandWidths))
			for j, w := range def.OperandWidths {
				switch w {
				case 1:
					operands[j] = int(fn.InlineBody[i+1])
				case 2:
					operands[j] = int(fn.InlineBody[i+1])<<8 | int(fn.InlineBody[i+2])
				}
				instrLen += w
			}
			c.emit(op, operands...)
			i += instrLen
		}
	}

	return nil
}

// analyzeLoopSafety detects safe array access patterns in loops
// Pattern: for i := 0; i < len(arr); i++ { ... arr[i] ... }
func (c *Compiler) analyzeLoopSafety(node *parser.ForStatement) {
	// Check for pattern: for i := 0; i < len(arr); i++
	// where i is an identifier and arr is an identifier

	// 1. Check init: i := 0
	var loopVar string
	var arrayVar string

	if init, ok := node.Init.(*parser.VarStatement); ok {
		if _, ok := init.Value.(*parser.IntegerLiteral); ok {
			loopVar = init.Name.Value
		}
	}

	// 2. Check condition: i < len(arr)
	if node.Condition != nil {
		if less, ok := node.Condition.(*parser.InfixExpression); ok && less.Operator == "<" {
			if left, ok := less.Left.(*parser.Identifier); ok && left.Value == loopVar {
				// Check if right side is len(arrayVar)
				if call, ok := less.Right.(*parser.CallExpression); ok {
					if fn, ok := call.Function.(*parser.Identifier); ok && fn.Value == "len" {
						if len(call.Arguments) == 1 {
							if arg, ok := call.Arguments[0].(*parser.Identifier); ok {
								arrayVar = arg.Value
							}
						}
					}
				}
			}
		}
	}

	// 3. Check update: i++ (could be ExpressionStatement containing PostfixExpression)
	if node.Update != nil {
		// Check for ExpressionStatement containing PostfixExpression
		if exprStmt, ok := node.Update.(*parser.ExpressionStatement); ok {
			if postfix, ok := exprStmt.Expression.(*parser.PostfixExpression); ok && postfix.Operator == "++" {
				if left, ok := postfix.Left.(*parser.Identifier); ok && left.Value == loopVar {
					// Pattern matched! Mark arrayVar[loopVar] as safe
					if loopVar != "" && arrayVar != "" {
						c.safeArrayAccess[fmt.Sprintf("%s[%s]", arrayVar, loopVar)] = true
					}
				}
			}
		}
	}
}

// isArrayAccessSafe checks if an array access is safe (bounds-checked by loop context)
func (c *Compiler) isArrayAccessSafe(arrayName, indexName string) bool {
	key := fmt.Sprintf("%s[%s]", arrayName, indexName)
	return c.safeArrayAccess[key]
}

// pushLoopContext starts a new loop context for break/continue tracking
func (c *Compiler) pushLoopContext(continuePos int) {
	c.loopContexts = append(c.loopContexts, loopContext{
		continuePos:   continuePos,
		breakPos:      []int{},
		continueJumps: []int{},
	})
}

// popLoopContext ends the current loop context and patches all break jumps
func (c *Compiler) popLoopContext(afterLoopPos int) {
	if len(c.loopContexts) == 0 {
		return
	}
	ctx := c.loopContexts[len(c.loopContexts)-1]
	// Patch all break positions to jump to after the loop
	for _, pos := range ctx.breakPos {
		c.changeOperand(pos, afterLoopPos)
	}
	c.loopContexts = c.loopContexts[:len(c.loopContexts)-1]
}

// currentLoopContinuePos returns the continue position of the current loop
func (c *Compiler) currentLoopContinuePos() int {
	if len(c.loopContexts) == 0 {
		return -1
	}
	return c.loopContexts[len(c.loopContexts)-1].continuePos
}

// addBreakPos records a break position to be patched later
func (c *Compiler) addBreakPos(pos int) {
	if len(c.loopContexts) == 0 {
		return
	}
	c.loopContexts[len(c.loopContexts)-1].breakPos = append(c.loopContexts[len(c.loopContexts)-1].breakPos, pos)
}

// addContinueJump records a continue jump position to be patched later
func (c *Compiler) addContinueJump(pos int) {
	if len(c.loopContexts) == 0 {
		return
	}
	c.loopContexts[len(c.loopContexts)-1].continueJumps = append(c.loopContexts[len(c.loopContexts)-1].continueJumps, pos)
}

// patchContinueJumps patches all continue jumps to jump to the specified position
func (c *Compiler) patchContinueJumps(pos int) {
	if len(c.loopContexts) == 0 {
		return
	}
	ctx := &c.loopContexts[len(c.loopContexts)-1]
	for _, jumpPos := range ctx.continueJumps {
		c.changeOperand(jumpPos, pos)
	}
	ctx.continueJumps = []int{} // Clear after patching
}

// compileTryStatement compiles a try-catch-finally statement
func (c *Compiler) compileTryStatement(node *parser.TryStatement) error {
	// Push exception handler with placeholder addresses (will be patched)
	pushHandlerPos := c.emit(OpPushHandler, 9999, 9999) // catchAddr, finallyAddr

	// Compile try block
	if err := c.Compile(node.Block); err != nil {
		return err
	}

	// Remove trailing pop if present
	if c.lastInstruction.Opcode == OpPop {
		c.removeLastInstruction()
	}

	// Pop handler after successful try block
	c.emit(OpPopHandler)

	// After try block completes normally, we need to:
	// - If there's a finally: fall through to finally (no jump needed)
	// - If there's no finally but catch: jump past catch
	var jumpPastCatchPos int = -1
	if node.Catch != nil && node.Finally == nil {
		// Only have catch, no finally - need to jump past catch after try
		jumpPastCatchPos = c.emit(OpJump, 9999)
	}

	// Record catch address (0 if no catch)
	catchAddr := 0
	if node.Catch != nil {
		catchAddr = len(c.currentInstructions())

		// The exception value is on the stack, bind it to the variable
		symbol := c.symbolTable.Define(node.Catch.Exception.Value)
		switch symbol.Scope {
		case GlobalScope:
			c.emit(OpSetGlobal, symbol.Index)
		case LocalScope:
			c.emit(OpSetLocal, symbol.Index)
		}

		// Pop the exception value that was pushed back by SetGlobal/SetLocal
		c.emit(OpPop)

		// Compile catch body
		if err := c.Compile(node.Catch.Block); err != nil {
			return err
		}

		// Remove trailing pop if present
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// After catch, fall through to the end (or to finally if present)
		// No jump needed here - we'll naturally fall through to finally/end
	}

	// Record finally address (0 if no finally)
	finallyAddr := 0
	if node.Finally != nil {
		finallyAddr = len(c.currentInstructions())

		// Compile finally block
		if err := c.Compile(node.Finally.Block); err != nil {
			return err
		}

		// Remove trailing pop if present
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}
	}

	// Patch push handler with catch and finally addresses
	c.changeOperands(pushHandlerPos, catchAddr, finallyAddr)

	// Patch jump past catch (if we emitted one)
	if jumpPastCatchPos >= 0 {
		// Jump past catch (to end, which is after finally if present)
		endPos := len(c.currentInstructions())
		c.changeOperand(jumpPastCatchPos, endPos)
	}

	// Add pop for try statement
	c.emit(OpPop)

	return nil
}

// compileThrowStatement compiles a throw statement
func (c *Compiler) compileThrowStatement(node *parser.ThrowStatement) error {
	// Compile the expression to throw (if present)
	if node.ErrExpr != nil {
		if err := c.Compile(node.ErrExpr); err != nil {
			return err
		}
	} else {
		// Throw null if no expression
		c.emit(OpNull)
	}

	c.emit(OpThrow)

	return nil
}

// extractModuleName extracts a module name from a path for auto-binding.
// Examples: "time" -> "time", "./math" -> "math", "math" -> "math"
// Returns empty string if a valid name cannot be extracted.
func extractModuleName(modulePath string) string {
	// Clean the path
	modulePath = strings.TrimSpace(modulePath)

	// Get the base name
	base := path.Base(modulePath)

	// Remove common prefixes
	// "time" -> "time"
	// "plugin/mysql" -> "mysql"
	if base == "." || base == ".." || base == "" {
		return ""
	}

	// Remove file extension if present (e.g., ".xxl")
	if idx := strings.LastIndex(base, "."); idx > 0 {
		base = base[:idx]
	}

	// Validate that it's a valid identifier
	if len(base) == 0 {
		return ""
	}

	// Check first character is letter or underscore
	first := base[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return ""
	}

	// Check rest are alphanumeric or underscore
	for i := 1; i < len(base); i++ {
		c := base[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return ""
		}
	}

	return base
}
