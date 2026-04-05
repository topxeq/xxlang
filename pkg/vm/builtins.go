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
		// Note: round, random moved to math module
		nil,                           // 59: round removed
		objects.Builtins["clamp"],     // 60
		objects.Builtins["sign"],      // 61
		nil,                           // 62: random removed
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
		objects.Builtins["sleep"], // 82
		objects.Builtins["now"],   // 83
		objects.Builtins["nowMs"], // 84
		objects.Builtins["uuid"],  // 85
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
		objects.Builtins["getSwitch"],    // 98
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
		objects.Builtins["append"],        // 109
		objects.Builtins["appendArray"],   // 110
		objects.Builtins["arrayContains"], // 111
		objects.Builtins["removeItems"],   // 112
		objects.Builtins["bytes"],         // 113
		objects.Builtins["chars"],         // 114
		objects.Builtins["plt"],           // 115
		objects.Builtins["make"],          // 116
		// BigInt/BigFloat
		objects.Builtins["bigInt"],     // 117
		objects.Builtins["bigFloat"],   // 118
		objects.Builtins["isBigInt"],   // 119
		objects.Builtins["isBigFloat"], // 120
		// Chars (Unicode character handling)
		objects.Builtins["toChars"], // 121
		objects.Builtins["charLen"], // 122
		// HTTP built-in functions (for server mode)
		objects.Builtins["writeResp"],     // 123
		objects.Builtins["setRespHeader"], // 124
		objects.Builtins["addRespHeader"], // 125
		objects.Builtins["getReqHeader"],  // 126
		objects.Builtins["getReqHeaders"], // 127
		objects.Builtins["setCookie"],     // 128
		objects.Builtins["getCookie"],     // 129
		objects.Builtins["getCookies"],    // 130
		objects.Builtins["parseForm"],     // 131
		// parseJSON, getReqBody, getReqBodyBytes moved to http module
		objects.Builtins["status"],    // 132
		objects.Builtins["redirect"],  // 133
		objects.Builtins["serveFile"], // 134
		// getMimeType moved to http module
		objects.Builtins["setContentType"], // 135
		objects.Builtins["queryParam"],     // 136
		objects.Builtins["queryParams"],    // 137
		objects.Builtins["formValue"],      // 138
		objects.Builtins["httpStatusName"], // 139
		objects.Builtins["isHttpReq"],      // 140
		objects.Builtins["isHttpResp"],     // 141
		objects.Builtins["urlEncode"],      // 142
		objects.Builtins["urlDecode"],      // 143
		// WebSocket functions moved to http module
		// Concurrency built-in functions
		objects.Builtins["makeTube"],     // 144
		objects.Builtins["closeTube"],    // 145
		objects.Builtins["tubeLen"],      // 146
		objects.Builtins["tubeCap"],      // 147
		objects.Builtins["tubeClosed"],   // 148
		objects.Builtins["tubeSend"],     // 149
		objects.Builtins["tubeRecv"],     // 150
		objects.Builtins["tubeTrySend"],  // 151
		objects.Builtins["tubeTryRecv"],  // 152
		objects.Builtins["newMutex"],     // 153
		objects.Builtins["newRWMutex"],   // 154
		objects.Builtins["newWaitGroup"], // 155
		objects.Builtins["newOnce"],      // 156
		objects.Builtins["newCond"],      // 157
		objects.Builtins["newAtomic"],    // 158
		// Context built-in functions
		objects.Builtins["newContext"],          // 159
		objects.Builtins["contextWithTimeout"],  // 160
		objects.Builtins["contextWithCancel"],   // 161
		objects.Builtins["contextWithDeadline"], // 162
		objects.Builtins["contextCancel"],       // 163
		objects.Builtins["contextDone"],         // 164
		objects.Builtins["contextErr"],          // 165
		objects.Builtins["contextIsDone"],       // 166
		objects.Builtins["contextDeadline"],     // 167
		// HTTP Client built-in functions (getWeb family)
		objects.Builtins["getWeb"],        // 168
		objects.Builtins["getWebBytes"],   // 169
		objects.Builtins["getWebObject"],  // 170
		objects.Builtins["postWeb"],       // 171
		objects.Builtins["postWebObject"], // 172
		objects.Builtins["urlExists"],     // 173
		objects.Builtins["httpStatus"],    // 174
		// Reader/Writer built-in functions
		objects.Builtins["getWebReader"],    // 175
		objects.Builtins["ioCopy"],          // 176
		objects.Builtins["isReader"],        // 177
		objects.Builtins["isWriter"],        // 178
		objects.Builtins["newBytesReader"],  // 179
		objects.Builtins["newStringReader"], // 180
		// Encryption built-in functions (Charlang compatible)
		objects.Builtins["encryptTextByTXTE"],  // 181
		objects.Builtins["decryptTextByTXTE"],  // 182
		objects.Builtins["encryptDataByTXDEE"], // 183
		objects.Builtins["decryptDataByTXDEE"], // 184
		objects.Builtins["encryptTextByTXDEE"], // 185
		objects.Builtins["decryptTextByTXDEE"], // 186
		objects.Builtins["encryptData"],        // 187
		objects.Builtins["encryptBytes"],       // 188
		objects.Builtins["decryptData"],        // 189
		objects.Builtins["decryptBytes"],       // 190
		objects.Builtins["encryptText"],        // 191
		objects.Builtins["encryptStr"],         // 192
		objects.Builtins["decryptText"],        // 193
		objects.Builtins["decryptStr"],         // 194
		objects.Builtins["encryptStream"],      // 195
		objects.Builtins["decryptStream"],      // 196
		objects.Builtins["aesEncrypt"],         // 197
		objects.Builtins["aesDecrypt"],         // 198
		objects.Builtins["downloadFile"],       // 199
		// Database built-in functions (String-based - Charlang compatible)
		objects.Builtins["formatSQLValue"],  // 200
		objects.Builtins["dbConnect"],       // 201
		objects.Builtins["dbClose"],         // 202
		objects.Builtins["dbQuery"],         // 203
		objects.Builtins["dbQueryOrdered"],  // 204
		objects.Builtins["dbQueryRecs"],     // 205
		objects.Builtins["dbQueryMap"],      // 206
		objects.Builtins["dbQueryMapArray"], // 207
		objects.Builtins["dbQueryCount"],    // 208
		objects.Builtins["dbQueryFloat"],    // 209
		objects.Builtins["dbQueryString"],   // 210
		objects.Builtins["dbExec"],          // 211
		// Database built-in functions (Typed - preserve native types)
		objects.Builtins["dbQueryTyped"],      // 212
		objects.Builtins["dbQueryRowTyped"],   // 213
		objects.Builtins["dbQueryArrayTyped"], // 214
		objects.Builtins["dbQueryValueTyped"], // 215
		// OrderedMap built-in functions
		objects.Builtins["isOrderedMap"],  // 216
		objects.Builtins["newOrderedMap"], // 217
		// System command built-in functions
		objects.Builtins["systemCmd"],         // 218
		objects.Builtins["systemCmdDetached"], // 219
		objects.Builtins["systemStart"],       // 220
		// Test assertion functions moved to testing module
		nil, // 221: testByText removed - use testing.byText()
		nil, // 222: testByStartsWith removed - use testing.byStartsWith()
		nil, // 223: testByEndsWith removed - use testing.byEndsWith()
		nil, // 224: testByContains removed - use testing.byContains()
		nil, // 225: testByReg removed - use testing.byReg()
		nil, // 226: testByRegContains removed - use testing.byRegContains()
		nil, // 227: dumpVar removed - use debug.dumpVar()
		// debugInfo removed - use debug.info() (no placeholder needed)
		// File system built-in functions (Batch 1)
		objects.Builtins["fileExists"],      // 228
		objects.Builtins["isDir"],           // 229
		objects.Builtins["loadText"],        // 230
		objects.Builtins["saveText"],        // 231
		objects.Builtins["appendText"],      // 232
		objects.Builtins["copyFile"],        // 233
		objects.Builtins["renameFile"],      // 234
		objects.Builtins["removeFile"],      // 235
		objects.Builtins["removeDir"],       // 236
		objects.Builtins["getFileList"],     // 237
		objects.Builtins["joinPath"],        // 238
		objects.Builtins["getCurDir"],       // 239
		objects.Builtins["getHomeDir"],      // 240
		objects.Builtins["getTempDir"],      // 241
		objects.Builtins["ensureMakeDirs"],  // 242
		objects.Builtins["getFileExt"],      // 243
		objects.Builtins["extractFileDir"],  // 244
		objects.Builtins["extractFileName"], // 245
		objects.Builtins["getFileInfo"],     // 246
		objects.Builtins["loadLines"],       // 247
		objects.Builtins["getFileAbs"],      // 248
		objects.Builtins["getFileRel"],      // 249
		objects.Builtins["isFile"],          // 250
		objects.Builtins["saveBytes"],       // 251
		objects.Builtins["loadBytes"],       // 252
		// Time enhancement built-in functions (Batch 2)
		objects.Builtins["getNowStr"],       // 253
		objects.Builtins["getNowTimeStamp"], // 254
		objects.Builtins["formatTime"],      // 255
		objects.Builtins["timeToTick"],      // 256
		objects.Builtins["timeAddSecs"],     // 257
		objects.Builtins["timeAddDate"],     // 258
		objects.Builtins["timeBefore"],      // 259
		objects.Builtins["strToTime"],       // 260
		objects.Builtins["timeAfter"],       // 261
		objects.Builtins["timeEqual"],       // 262
		objects.Builtins["timeDiff"],        // 263
		objects.Builtins["timeDiffSecs"],    // 264
		objects.Builtins["parseTime"],       // 265
		objects.Builtins["isTime"],          // 266
		// Regex enhancement built-in functions (Batch 3)
		objects.Builtins["regMatch"],           // 267
		objects.Builtins["regContains"],        // 268
		objects.Builtins["regFindFirst"],       // 269
		objects.Builtins["regFindAll"],         // 270
		objects.Builtins["regFindFirstGroups"], // 271
		objects.Builtins["regFindAllGroups"],   // 272
		objects.Builtins["regReplace"],         // 273
		objects.Builtins["regSplit"],           // 274
		objects.Builtins["regCount"],           // 275
		objects.Builtins["regQuote"],           // 276
		objects.Builtins["regFindAllIndex"],    // 277
		// Encoding enhancement built-in functions (Batch 4)
		objects.Builtins["urlEncodeComponent"], // 278
		objects.Builtins["urlDecodeComponent"], // 279
		objects.Builtins["htmlEncode"],         // 280
		objects.Builtins["htmlDecode"],         // 281
		objects.Builtins["sha1"],               // 282
		objects.Builtins["sha512"],             // 283
		objects.Builtins["hashStr"],            // 284
		objects.Builtins["toHex"],              // 285
		objects.Builtins["unhex"],              // 286
		objects.Builtins["hexToStr"],           // 287
		// System/Environment built-in functions (Batch 5)
		objects.Builtins["getEnv"],     // 288
		objects.Builtins["setEnv"],     // 289
		objects.Builtins["getOSName"],  // 290
		objects.Builtins["getOSArch"],  // 291
		objects.Builtins["getOSArgs"],  // 292
		objects.Builtins["getAppPath"], // 293
		objects.Builtins["getAppDir"],  // 294
		objects.Builtins["exit"],       // 295
		objects.Builtins["getSysInfo"], // 296
		objects.Builtins["getPid"],     // 297
		objects.Builtins["getPPid"],    // 298
		objects.Builtins["hostname"],   // 299
		// Math enhancement built-in functions (Batch 6)
		// Note: sin, cos, tan, asin, acos, atan, atan2, exp, log, log10, log2, pi, e, degToRad, radToDeg removed
		// Use math module instead. Placeholders kept for index stability.
		nil,                             // 300: sin removed
		nil,                             // 301: cos removed
		nil,                             // 302: tan removed
		nil,                             // 303: asin removed
		nil,                             // 304: acos removed
		nil,                             // 305: atan removed
		nil,                             // 306: atan2 removed
		nil,                             // 307: exp removed
		nil,                             // 308: log removed
		nil,                             // 309: log10 removed
		nil,                             // 310: log2 removed
		nil,                             // 311: pi removed
		nil,                             // 312: e removed
		nil,                             // 313: degToRad removed
		nil,                             // 314: radToDeg removed
		objects.Builtins["adjustFloat"], // 315
		objects.Builtins["toKMG"],       // 316
		objects.Builtins["trunc"],       // 317
		objects.Builtins["isInf"],       // 318
		objects.Builtins["isNaN"],       // 319
		objects.Builtins["isFinite"],    // 320
		// JSON enhancement built-in functions (Batch 7)
		objects.Builtins["formatJson"],      // 321
		objects.Builtins["compactJson"],     // 322
		objects.Builtins["getJsonNodeStr"],  // 323
		objects.Builtins["getJsonNodeStrs"], // 324
		objects.Builtins["strsToJson"],      // 325
		objects.Builtins["jsonValid"],       // 326
		objects.Builtins["jsonType"],        // 327
		// Compression built-in functions (Batch 8)
		objects.Builtins["compressData"],     // 328
		objects.Builtins["uncompressData"],   // 329
		objects.Builtins["compressStr"],      // 330
		objects.Builtins["uncompressStr"],    // 331
		objects.Builtins["zipPath"],          // 332
		objects.Builtins["zipPaths"],         // 333
		objects.Builtins["unzipToPath"],      // 334
		objects.Builtins["getFileListInZip"], // 335
		objects.Builtins["loadBytesInZip"],   // 336
		objects.Builtins["addFileToZip"],     // 337
		// Input/Clipboard built-in functions (Batch 9)
		objects.Builtins["getInput"],          // 338
		objects.Builtins["getInputf"],         // 339
		objects.Builtins["getChar"],           // 340
		objects.Builtins["getKey"],            // 341
		objects.Builtins["getMultiLineInput"], // 342
		objects.Builtins["getPassword"],       // 343
		objects.Builtins["confirm"],           // 344
		objects.Builtins["readLine"],          // 345
		objects.Builtins["getClipText"],       // 346
		objects.Builtins["setClipText"],       // 347
		// String enhancement built-in functions (Batch 10)
		objects.Builtins["strContainsIn"],       // 349
		objects.Builtins["strRuneLen"],          // 350
		objects.Builtins["strIn"],               // 351
		objects.Builtins["strGetLastComponent"], // 352
		objects.Builtins["strFindDiffPos"],      // 353
		objects.Builtins["strDiff"],             // 354
		objects.Builtins["strFindAllSub"],       // 355
		objects.Builtins["limitStr"],            // 356
		objects.Builtins["strQuote"],            // 357
		objects.Builtins["strUnquote"],          // 358
		objects.Builtins["strToInt"],            // 359
		objects.Builtins["getTextSimilarity"],   // 360
		objects.Builtins["fuzzyFind"],           // 361
		// strRemoveBom moved to strings module
		nil, // 362: strRemoveBom removed - use strings.removeBom()
		// String functions moved to string module
		nil, // 361: wordCount removed - use string.wordCount()
		nil, // 362: lineCount removed - use string.lineCount()
		nil, // 363: reverseStr removed - use string.reverse()
		nil, // 364: capitalize removed - use string.capitalize()
		nil, // 365: title removed - use string.title()
		nil, // 366: swapCase removed - use string.swapCase()
		nil, // 367: center removed - use string.center()
		nil, // 368: zfill removed - use string.zfill()
		nil, // 369: isSpace removed - use string.isSpace()
		// Collection enhancement built-in functions (Batch 11)
		objects.Builtins["mapArray"],     // 370
		objects.Builtins["filterArray"],  // 371
		objects.Builtins["reduceArray"],  // 372
		objects.Builtins["forEach"],      // 373
		objects.Builtins["flatMap"],      // 374
		objects.Builtins["every"],        // 375
		objects.Builtins["some"],         // 376
		objects.Builtins["groupBy"],      // 377
		objects.Builtins["partition"],    // 378
		objects.Builtins["zip"],          // 379
		objects.Builtins["unzip"],        // 380
		objects.Builtins["fill"],         // 381
		objects.Builtins["rangeNum"],     // 382
		objects.Builtins["intersection"], // 383
		objects.Builtins["difference"],   // 384
		objects.Builtins["union"],        // 385
		objects.Builtins["countBy"],      // 386
		objects.Builtins["sortBy"],       // 387
		// Utility built-in functions (Batch 12)
		objects.Builtins["sprintf"],     // 388
		objects.Builtins["toBool"],      // 389
		objects.Builtins["toInt"],       // 390
		objects.Builtins["toFloat"],     // 391
		objects.Builtins["isUndefined"], // 392
		objects.Builtins["isCallable"],  // 393
		objects.Builtins["isIterable"],  // 394
		objects.Builtins["isError"],     // 395
		objects.Builtins["error"],       // 396
		objects.Builtins["getErrStr"],   // 397
		objects.Builtins["isErrStr"],    // 398
		objects.Builtins["typeCode"],    // 399
		objects.Builtins["swap"],        // 400
		objects.Builtins["coalesce"],    // 401
		objects.Builtins["defaultVal"],  // 402
		// String processing enhancement built-in functions (Batch 13)
		objects.Builtins["strSplitLines"],  // 403
		objects.Builtins["strContainsAny"], // 404
		objects.Builtins["strIndex"],       // 405
		objects.Builtins["strLastIndex"],   // 406
		objects.Builtins["strSplitN"],      // 407
		objects.Builtins["strPad"],         // 408
		objects.Builtins["strSub"],         // 409
		objects.Builtins["intToStr"],       // 410
		objects.Builtins["floatToStr"],     // 411
		objects.Builtins["charCode"],       // 412
		objects.Builtins["charFromCode"],   // 413
		objects.Builtins["reverseMap"],     // 414
		objects.Builtins["simpleStrToMap"], // 415
		objects.Builtins["mapToStr"],       // 416
		objects.Builtins["bitNot"],         // 417
		objects.Builtins["bitAnd"],         // 418
		objects.Builtins["bitOr"],          // 419
		objects.Builtins["bitXor"],         // 420
		objects.Builtins["bitShiftLeft"],   // 421
		objects.Builtins["bitShiftRight"],  // 422
		// Check/validate and bytes built-in functions (Batch 14)
		objects.Builtins["isNil"],           // 423
		objects.Builtins["isNull"],          // 424
		objects.Builtins["isNilOrEmpty"],    // 425
		objects.Builtins["isNilOrErr"],      // 426
		objects.Builtins["isBytes"],         // 427
		objects.Builtins["isChars"],         // 428
		objects.Builtins["pass"],            // 429
		objects.Builtins["errStrf"],         // 430
		objects.Builtins["errf"],            // 431
		objects.Builtins["errToEmpty"],      // 432
		objects.Builtins["sscanf"],          // 433
		objects.Builtins["bytesStartsWith"], // 434
		objects.Builtins["bytesEndsWith"],   // 435
		objects.Builtins["bytesContains"],   // 436
		objects.Builtins["bytesIndex"],      // 437
		objects.Builtins["compareBytes"],    // 438
		objects.Builtins["compareText"],     // 439
		// Miscellaneous built-in functions (Batch 15)
		objects.Builtins["getRandomInt"],   // 440
		objects.Builtins["getRandomFloat"], // 441
		objects.Builtins["getRandomStr"],   // 442
		objects.Builtins["createTempDir"],  // 443
		objects.Builtins["createTempFile"], // 444
		objects.Builtins["changeDir"],      // 445
		objects.Builtins["lookPath"],       // 446
		objects.Builtins["joinUrlPath"],    // 447
		objects.Builtins["parseUrl"],       // 448
		objects.Builtins["parseQuery"],     // 449
		objects.Builtins["isHttps"],        // 450
		objects.Builtins["genToken"],       // 451
		objects.Builtins["genOtpCode"],     // 452
		objects.Builtins["checkOtpCode"],   // 453
		// Unicode/Text processing built-in functions (Batch 16)
		// Note: toPinYin, kanaToRomaji, kanjiToKana, kanjiToRomaji moved to locale module
		nil, // 454: toPinYin removed
		nil, // 455: kanaToRomaji removed
		nil, // 456: kanjiToKana removed
		nil, // 457: kanjiToRomaji removed
		// JWT built-in functions (Batch 17)
		// Note: genJwtToken, parseJwtToken moved to crypto module
		nil, // 458: genJwtToken removed
		nil, // 459: parseJwtToken removed
		// Task/Scheduling built-in functions (Batch 18)
		// Note: isCronExprValid, isCronExprDue, runTicker, stopTicker moved to task module
		nil, // 460: isCronExprValid removed
		nil, // 461: isCronExprDue removed
		nil, // 462: runTicker removed
		nil, // 463: stopTicker removed
		// Image processing built-in functions (Batch 19)
		// Note: genQr, scanQr, getImageInfo, resizeImage moved to image module
		// createImage is kept as a builtin (alias to image.createImage)
		nil,                             // 464: genQr removed
		nil,                             // 465: scanQr removed
		nil,                             // 466: getImageInfo removed
		nil,                             // 467: resizeImage removed
		objects.Builtins["createImage"], // 468: kept as builtin (alias)
		// Network communication built-in functions (Batch 20)
		// Note: newFtpClient, newSshClient removed - use ftp.connect() and ssh.connect() instead
		nil, // 469: newFtpClient removed
		nil, // 470: newSshClient removed
		// Excel/XLSX functions - use xlsx.create(), xlsx.open(), csv.read(), csv.write()
		nil, // 471: newExcel removed
		nil, // 472: openExcel removed
		nil, // 473: readCsv removed
		nil, // 474: writeCsv removed
		// Data format built-in functions (Batch 22)
		// XML functions - use xml.parse(), xml.parseFile(), xml.create()
		nil, // 475: parseXml removed
		nil, // 476: parseXmlFile removed
		nil, // 477: newXmlDoc removed
		// YAML functions - use yaml.parse(), yaml.stringify(), yaml.toJson(), yaml.fromJson()
		nil, // 478: parseYaml removed
		nil, // 479: toYaml removed
		nil, // 480: yamlToJson removed
		nil, // 481: jsonToYaml removed
		// TOML functions - use toml.parse(), toml.encode(), toml.create(), toml.isValid()
		nil, // 482: parseToml removed
		nil, // 483: toToml removed
		nil, // 484: newToml removed
		nil, // 485: tomlValid removed
		// Email sending functions - use mail.newClient(), mail.send()
		nil, // 486: sendMail removed
		nil, // 487: newMailClient removed
		// Byte-index string functions (Batch 23)
		objects.Builtins["byteIndexOf"], // 488
		objects.Builtins["byteSubstr"],  // 489
		objects.Builtins["byteLen"],     // 490
		// String enhancement (Batch 24)
		objects.Builtins["strCount"], // 491
		// Simple encoding (Batch 25)
		objects.Builtins["simpleEncode"], // 492
		objects.Builtins["simpleDecode"], // 493
		// Time enhancement (Batch 26)
		objects.Builtins["now"],              // 494
		objects.Builtins["getNowStrCompact"], // 495
		objects.Builtins["timeToTimeStamp"],  // 496
		objects.Builtins["timeStampToTime"],  // 497
		// File system enhancement (Batch 27)
		objects.Builtins["dirExists"],   // 498
		objects.Builtins["pathExists"],  // 499
		objects.Builtins["copyPath"],    // 500
		objects.Builtins["moveFile"],    // 501
		objects.Builtins["getFileSize"], // 502
		// Print aliases (Batch 27)
		objects.Builtins["print"],   // 503: alias for pr
		objects.Builtins["println"], // 504: alias for pln
		objects.Builtins["printf"],  // 505: alias for prf
		objects.Builtins["concatBytes"], // 506
		objects.Builtins["plv"],     // 507: print value with %#v format
		objects.Builtins["spr"],     // 508: alias for sprintf
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
