// pkg/vm/builtins.go
// Builtin function support for the VM
package vm

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// getBuiltin returns a builtin function by index
func getBuiltin(index int) *objects.Builtin {
	builtins := []*objects.Builtin{
		objects.Builtins["len"],         // 0
		objects.Builtins["pr"],          // 1
		objects.Builtins["pln"],         // 2
		objects.Builtins["typeOf"],      // 3
		objects.Builtins["substr"],      // 4
		objects.Builtins["split"],       // 5
		objects.Builtins["join"],        // 6
		objects.Builtins["trim"],        // 7
		objects.Builtins["upper"],       // 8
		objects.Builtins["lower"],       // 9
		objects.Builtins["containsStr"], // 10
		objects.Builtins["replace"],     // 11
		objects.Builtins["startsWith"],  // 12
		objects.Builtins["endsWith"],    // 13
		objects.Builtins["abs"],         // 14
		objects.Builtins["floor"],       // 15
		objects.Builtins["ceil"],        // 16
		objects.Builtins["sqrt"],        // 17
		objects.Builtins["pow"],         // 18
		objects.Builtins["min"],         // 19
		objects.Builtins["max"],         // 20
		objects.Builtins["int"],         // 21
		objects.Builtins["float"],       // 22
		objects.Builtins["string"],      // 23
		objects.Builtins["push"],        // 24
		objects.Builtins["pop"],         // 25
		objects.Builtins["first"],       // 26
		objects.Builtins["last"],        // 27
		objects.Builtins["rest"],        // 28
		objects.Builtins["concat"],      // 29
		objects.Builtins["indexOf"],     // 30
		objects.Builtins["containsArr"], // 31
		objects.Builtins["keys"],        // 32
		objects.Builtins["values"],      // 33
		objects.Builtins["hasKey"],      // 34
		objects.Builtins["delete"],      // 35
		objects.Builtins["range"],       // 36
		objects.Builtins["sort"],        // 37
		objects.Builtins["sum"],         // 38
		objects.Builtins["avg"],         // 39
		objects.Builtins["reverse"],     // 40
		objects.Builtins["runCode"],     // 41
		objects.Builtins["loadPlugin"],  // 42
		// String utilities
		objects.Builtins["repeat"],    // 43
		objects.Builtins["lpad"],      // 44
		objects.Builtins["rpad"],      // 45
		objects.Builtins["charAt"],    // 46
		objects.Builtins["trimLeft"],  // 47
		objects.Builtins["trimRight"], // 48
		// Type checking
		objects.Builtins["isEmpty"],    // 49
		objects.Builtins["isString"],   // 50
		objects.Builtins["isNumber"],   // 51
		objects.Builtins["isInt"],      // 52
		objects.Builtins["isFloat"],    // 53
		objects.Builtins["isArray"],    // 54
		objects.Builtins["isMap"],      // 55
		objects.Builtins["isBool"],     // 56
		objects.Builtins["isFunction"], // 57
		objects.Builtins["isNull"],     // 58
		// Math utilities
		objects.Builtins["round"],     // 59
		objects.Builtins["clamp"],     // 60
		objects.Builtins["sign"],      // 61
		objects.Builtins["random"],    // 62
		objects.Builtins["randomInt"], // 63
		// Array utilities
		objects.Builtins["unique"],  // 64
		objects.Builtins["flatten"], // 65
		objects.Builtins["without"], // 66
		objects.Builtins["take"],    // 67
		objects.Builtins["drop"],    // 68
		// Map utilities
		objects.Builtins["merge"],   // 69
		objects.Builtins["entries"], // 70
		// Format
		objects.Builtins["format"], // 71
		// Object utilities
		objects.Builtins["copy"],     // 72
		objects.Builtins["clone"],    // 73
		objects.Builtins["equals"],   // 74
		objects.Builtins["defaults"], // 75
		// Encoding & Hash
		objects.Builtins["base64Encode"], // 76
		objects.Builtins["base64Decode"], // 77
		objects.Builtins["hexEncode"],    // 78
		objects.Builtins["hexDecode"],    // 79
		objects.Builtins["md5"],          // 80
		objects.Builtins["sha256"],       // 81
		// Time & UUID
		objects.Builtins["sleep"],  // 82
		objects.Builtins["now"],    // 83
		objects.Builtins["nowMs"],  // 84
		objects.Builtins["uuid"],   // 85
		// String enhancement
		objects.Builtins["trimPrefix"], // 86
		objects.Builtins["trimSuffix"], // 87
		objects.Builtins["count"],      // 88
		objects.Builtins["isDigit"],    // 89
		objects.Builtins["isAlpha"],    // 90
		objects.Builtins["isAlphaNum"], // 91
		// Array enhancement
		objects.Builtins["find"],      // 92
		objects.Builtins["findIndex"], // 93
		objects.Builtins["includes"],  // 94
		objects.Builtins["shuffle"],   // 95
		objects.Builtins["sample"],    // 96
		objects.Builtins["chunk"],     // 97
		// Command line argument utilities
		objects.Builtins["getSwitch"],   // 98
		objects.Builtins["switchExists"], // 99
		// Output utilities
		objects.Builtins["pl"],  // 100
		objects.Builtins["prf"], // 101
		// Validation utilities
		objects.Builtins["checkErr"],   // 102
		objects.Builtins["checkEmpty"], // 103
		// OTP utilities
		objects.Builtins["genOtpCode"], // 104
		// Type conversion
		objects.Builtins["toStr"],    // 105
		objects.Builtins["toJson"],   // 106
		objects.Builtins["fromJson"], // 107
		// Dynamic code
		objects.Builtins["delegate"], // 108
		// Array functions (Charlang compatibility)
		objects.Builtins["append"],       // 109
		objects.Builtins["appendArray"],   // 110
		objects.Builtins["arrayContains"], // 111
		objects.Builtins["removeItems"],   // 112
		objects.Builtins["bytes"],         // 113
		objects.Builtins["plt"],           // 114
		objects.Builtins["make"],          // 115
		// BigInt/BigFloat
		objects.Builtins["bigInt"],     // 116
		objects.Builtins["bigFloat"],   // 117
		objects.Builtins["isBigInt"],   // 118
		objects.Builtins["isBigFloat"], // 119
		// Chars (Unicode character handling)
		objects.Builtins["toChars"],  // 120
		objects.Builtins["charLen"],  // 121
		// HTTP built-in functions (for server mode)
		objects.Builtins["writeResp"],      // 122
		objects.Builtins["setRespHeader"],  // 123
		objects.Builtins["addRespHeader"],  // 124
		objects.Builtins["getReqHeader"],   // 125
		objects.Builtins["getReqHeaders"],  // 126
		objects.Builtins["setCookie"],      // 127
		objects.Builtins["getCookie"],      // 128
		objects.Builtins["getCookies"],     // 129
		objects.Builtins["parseForm"],      // 130
		objects.Builtins["parseJSON"],      // 131
		objects.Builtins["getReqBody"],     // 132
		objects.Builtins["getReqBodyBytes"], // 133
		objects.Builtins["status"],         // 134
		objects.Builtins["redirect"],       // 135
		objects.Builtins["serveFile"],      // 136
		objects.Builtins["getMimeType"],    // 137
		objects.Builtins["setContentType"], // 138
		objects.Builtins["queryParam"],     // 139
		objects.Builtins["queryParams"],    // 140
		objects.Builtins["formValue"],      // 141
		objects.Builtins["httpStatusName"], // 142
		objects.Builtins["isHttpReq"],      // 143
		objects.Builtins["isHttpResp"],     // 144
		objects.Builtins["urlEncode"],      // 145
		objects.Builtins["urlDecode"],      // 146
		// WebSocket built-in functions
		objects.Builtins["webSocket"],      // 147
		objects.Builtins["wsReadMsg"],      // 148
		objects.Builtins["wsSendText"],     // 149
		objects.Builtins["wsSendBinary"],   // 150
		objects.Builtins["wsSendClose"],    // 151
		objects.Builtins["wsClose"],        // 152
		objects.Builtins["isWebSocket"],    // 153
		// Concurrency built-in functions
		objects.Builtins["makeTube"],    // 154
		objects.Builtins["closeTube"],   // 155
		objects.Builtins["tubeLen"],     // 156
		objects.Builtins["tubeCap"],     // 157
		objects.Builtins["tubeClosed"],  // 158
		objects.Builtins["tubeSend"],    // 159
		objects.Builtins["tubeRecv"],    // 160
		objects.Builtins["tubeTrySend"], // 161
		objects.Builtins["tubeTryRecv"], // 162
		objects.Builtins["newMutex"],    // 163
		objects.Builtins["newRWMutex"],  // 164
		objects.Builtins["newWaitGroup"], // 165
		objects.Builtins["newOnce"],     // 166
		objects.Builtins["newCond"],     // 167
		objects.Builtins["newAtomic"],   // 168
		// Context built-in functions
		objects.Builtins["newContext"],        // 169
		objects.Builtins["contextWithTimeout"], // 170
		objects.Builtins["contextWithCancel"],  // 171
		objects.Builtins["contextWithDeadline"], // 172
		objects.Builtins["contextCancel"],     // 173
		objects.Builtins["contextDone"],       // 174
		objects.Builtins["contextErr"],        // 175
		objects.Builtins["contextIsDone"],     // 176
		objects.Builtins["contextDeadline"],   // 177
	}

	if index < 0 || index >= len(builtins) {
		return nil
	}
	return builtins[index]
}

// GetBuiltinByIndex returns a builtin function by index (exported for JIT)
func GetBuiltinByIndex(index int) *objects.Builtin {
	return getBuiltin(index)
}
