// pkg/objects/builtin_list.go
// Single source of truth for builtin function indices.
// Add new builtins here, then both compiler and VM will pick them up automatically.
package objects

// BuiltinRegistry lists all builtin function names in index order.
// Empty string "" = nil placeholder (removed builtin, index preserved for stability).
// To add a new builtin:
//   1. Implement the function in builtin.go (or a split file)
//   2. Append the name to this list
//   3. Done - no need to edit compiler or VM files
var BuiltinRegistry = []string{
	"len", // 0
	"pr", // 1
	"pln", // 2
	"typeOf", // 3
	"substr", // 4
	"split", // 5
	"join", // 6
	"trim", // 7
	"upper", // 8
	"lower", // 9
	"containsStr", // 10
	"replace", // 11
	"startsWith", // 12
	"endsWith", // 13
	"abs", // 14
	"floor", // 15
	"ceil", // 16
	"sqrt", // 17
	"pow", // 18
	"min", // 19
	"max", // 20
	"int", // 21
	"float", // 22
	"string", // 23
	"push", // 24
	"pop", // 25
	"first", // 26
	"last", // 27
	"rest", // 28
	"concat", // 29
	"indexOf", // 30
	"containsArr", // 31
	"keys", // 32
	"values", // 33
	"hasKey", // 34
	"delete", // 35
	"range", // 36
	"sort", // 37
	"sum", // 38
	"avg", // 39
	"reverse", // 40
	"runCode", // 41
	"loadPlugin", // 42
	// String utilities
	"repeat", // 43
	"lpad", // 44
	"rpad", // 45
	"charAt", // 46
	"trimLeft", // 47
	"trimRight", // 48
	// Type checking
	"isEmpty", // 49
	"isString", // 50
	"isNumber", // 51
	"isInt", // 52
	"isFloat", // 53
	"isArray", // 54
	"isMap", // 55
	"isBool", // 56
	"isFunction", // 57
	"isNull", // 58
	// 59: round removed
	"",   // 59: 59: round removed
	"clamp", // 60
	"sign", // 61
	// 62: random removed
	"",   // 62: 62: random removed
	"randomInt", // 63
	// Array utilities
	"unique", // 64
	"flatten", // 65
	"without", // 66
	"take", // 67
	"drop", // 68
	// Map utilities
	"merge", // 69
	"entries", // 70
	// Format
	"format", // 71
	// Object utilities
	"copy", // 72
	"clone", // 73
	"equals", // 74
	"defaults", // 75
	// Encoding & Hash
	"base64Encode", // 76
	"base64Decode", // 77
	"hexEncode", // 78
	"hexDecode", // 79
	"md5", // 80
	"sha256", // 81
	// Time & UUID
	"sleep", // 82
	"sleepSec", // 83
	"now", // 84
	"nowMs", // 85
	"uuid", // 86
	// String enhancement
	"trimPrefix", // 87
	"trimSuffix", // 88
	"count", // 89
	"isDigit", // 90
	"isAlpha", // 91
	"isAlphaNum", // 92
	// Array enhancement
	"find", // 93
	"findIndex", // 94
	"includes", // 95
	"shuffle", // 96
	"sample", // 97
	"chunk", // 98
	// Command line argument utilities
	"getSwitch", // 99
	"switchExists", // 100
	"getParam", // 101 - get positional parameter by index from args array
	// Output utilities
	"pl", // 102
	"prf", // 103
	// Validation utilities
	"checkErr", // 103
	"checkEmpty", // 104
	// OTP utilities
	"genOtpCode", // 105
	// Type conversion
	"toStr", // 106
	"toJson", // 107
	"fromJson", // 108
	// Dynamic code
	"delegate", // 109
	// Array functions (Charlang compatibility)
	"append", // 110
	"appendArray", // 111
	"arrayContains", // 112
	"removeItems", // 113
	"bytes", // 114
	"chars", // 115
	"plt", // 116
	"make", // 117
	// BigInt/BigFloat
	"bigInt", // 118
	"bigFloat", // 119
	"isBigInt", // 120
	"isBigFloat", // 121
	// Chars (Unicode character handling)
	"toChars", // 122
	"charLen", // 123
	// HTTP built-in functions (for server mode)
	"writeResp", // 124
	"setRespHeader", // 125
	"addRespHeader", // 126
	"getReqHeader", // 127
	"getReqHeaders", // 128
	"setCookie", // 129
	"getCookie", // 130
	"getCookies", // 131
	"parseForm", // 132
	// parseJSON, getReqBody, getReqBodyBytes moved to http module
	"status", // 133
	"redirect", // 134
	"serveFile", // 135
	// getMimeType moved to http module
	"setContentType", // 136
	"queryParam", // 137
	"queryParams", // 138
	"formValue", // 139
	"httpStatusName", // 140
	"isHttpReq", // 141
	"isHttpResp", // 142
	"urlEncode", // 143
	"urlDecode", // 144
	// Concurrency built-in functions
	"makeTube", // 145
	"closeTube", // 146
	"tubeLen", // 147
	"tubeCap", // 148
	"tubeClosed", // 149
	"tubeSend", // 150
	"tubeRecv", // 151
	"tubeTrySend", // 152
	"tubeTryRecv", // 153
	"newMutex", // 154
	"newRWMutex", // 155
	"newWaitGroup", // 156
	"newOnce", // 157
	"newCond", // 158
	"newAtomic", // 159
	// Context built-in functions
	"newContext", // 160
	"contextWithTimeout", // 161
	"contextWithCancel", // 162
	"contextWithDeadline", // 163
	"contextCancel", // 164
	"contextDone", // 165
	"contextErr", // 166
	"contextIsDone", // 167
	"contextDeadline", // 168
	// HTTP Client built-in functions (getWeb family)
	"getWeb", // 169
	"getWebBytes", // 170
	"getWebObject", // 171
	"postWeb", // 172
	"postWebObject", // 173
	"urlExists", // 174
	"httpStatus", // 175
	// Reader/Writer built-in functions
	"getWebReader", // 176
	"ioCopy", // 177
	"isReader", // 178
	"isWriter", // 179
	"newBytesReader", // 180
	"newStringReader", // 181
	// Encryption built-in functions (Charlang compatible)
	"encryptTextByTXTE", // 182
	"decryptTextByTXTE", // 183
	"encryptDataByTXDEE", // 184
	"decryptDataByTXDEE", // 185
	"encryptTextByTXDEE", // 186
	"decryptTextByTXDEE", // 187
	"encryptData", // 188
	"encryptBytes", // 189
	"decryptData", // 190
	"decryptBytes", // 191
	"encryptText", // 192
	"encryptStr", // 193
	"decryptText", // 194
	"decryptStr", // 195
	"encryptStream", // 196
	"decryptStream", // 197
	"aesEncrypt", // 198
	"aesDecrypt", // 199
	"downloadFile", // 200
	// Database built-in functions (String-based - Charlang compatible)
	"formatSQLValue", // 201
	"dbConnect", // 202
	"dbClose", // 203
	"dbQuery", // 204
	"dbQueryOrdered", // 205
	"dbQueryRecs", // 206
	"dbQueryMap", // 207
	"dbQueryMapArray", // 208
	"dbQueryCount", // 209
	"dbQueryFloat", // 210
	"dbQueryString", // 211
	"dbExec", // 212
	// Database built-in functions (Typed - preserve native types)
	"dbQueryTyped", // 213
	"dbQueryRowTyped", // 214
	"dbQueryArrayTyped", // 215
	"dbQueryValueTyped", // 216
	// OrderedMap built-in functions
	"isOrderedMap", // 217
	"newOrderedMap", // 218
	// System command built-in functions
	"systemCmd", // 219
	"systemCmdDetached", // 220
	"systemStart", // 221
	// 221: testByText removed - use testing.byText()
	"",   // 222: 221: testByText removed - use testing.byText()
	// 222: testByStartsWith removed - use testing.byStartsWith()
	"",   // 223: 222: testByStartsWith removed - use testing.byStartsWith()
	// 223: testByEndsWith removed - use testing.byEndsWith()
	"",   // 224: 223: testByEndsWith removed - use testing.byEndsWith()
	// 224: testByContains removed - use testing.byContains()
	"",   // 225: 224: testByContains removed - use testing.byContains()
	// 225: testByReg removed - use testing.byReg()
	"",   // 226: 225: testByReg removed - use testing.byReg()
	// 226: testByRegContains removed - use testing.byRegContains()
	"",   // 227: 226: testByRegContains removed - use testing.byRegContains()
	// 227: dumpVar removed - use debug.dumpVar()
	"",   // 228: 227: dumpVar removed - use debug.dumpVar()
	// File system built-in functions (Batch 1)
	"fileExists", // 229
	"isDir", // 230
	"loadText", // 231
	"saveText", // 232
	"appendText", // 233
	"copyFile", // 234
	"renameFile", // 235
	"removeFile", // 236
	"removeDir", // 237
	"getFileList", // 238
	"joinPath", // 239
	"getCurDir", // 240
	"getHomeDir", // 241
	"getTempDir", // 242
	"ensureMakeDirs", // 243
	"getFileExt", // 244
	"extractFileDir", // 245
	"extractFileName", // 246
	"getFileInfo", // 247
	"loadLines", // 248
	"getFileAbs", // 249
	"getFileRel", // 250
	"isFile", // 251
	"saveBytes", // 252
	"loadBytes", // 253
	// Time enhancement built-in functions (Batch 2)
	"getNowStr", // 254
	"getNowTimeStamp", // 255
	"formatTime", // 256
	"timeToTick", // 257
	"timeAddSecs", // 258
	"timeAddDate", // 259
	"timeBefore", // 260
	"strToTime", // 261
	"timeAfter", // 262
	"timeEqual", // 263
	"timeDiff", // 264
	"timeDiffSecs", // 265
	"parseTime", // 266
	"isTime", // 267
	// Regex enhancement built-in functions (Batch 3)
	"regMatch", // 268
	"regContains", // 269
	"regFindFirst", // 270
	"regFindAll", // 271
	"regFindFirstGroups", // 272
	"regFindAllGroups", // 273
	"regReplace", // 274
	"regSplit", // 275
	"regCount", // 276
	"regQuote", // 277
	"regFindAllIndex", // 278
	// Encoding enhancement built-in functions (Batch 4)
	"urlEncodeComponent", // 279
	"urlDecodeComponent", // 280
	"htmlEncode", // 281
	"htmlDecode", // 282
	"sha1", // 283
	"sha512", // 284
	"hashStr", // 285
	"toHex", // 286
	"unhex", // 287
	"hexToStr", // 288
	// System/Environment built-in functions (Batch 5)
	"getEnv", // 289
	"setEnv", // 290
	"getOSName", // 291
	"getOSArch", // 292
	"getOSArgs", // 293
	"getAppPath", // 294
	"getAppDir", // 295
	"exit", // 296
	"getSysInfo", // 297
	"getPid", // 298
	"getPPid", // 299
	"hostname", // 300
	// 300: sin removed
	"",   // 301: 300: sin removed
	// 301: cos removed
	"",   // 302: 301: cos removed
	// 302: tan removed
	"",   // 303: 302: tan removed
	// 303: asin removed
	"",   // 304: 303: asin removed
	// 304: acos removed
	"",   // 305: 304: acos removed
	// 305: atan removed
	"",   // 306: 305: atan removed
	// 306: atan2 removed
	"",   // 307: 306: atan2 removed
	// 307: exp removed
	"",   // 308: 307: exp removed
	// 308: log removed
	"",   // 309: 308: log removed
	// 309: log10 removed
	"",   // 310: 309: log10 removed
	// 310: log2 removed
	"",   // 311: 310: log2 removed
	// 311: pi removed
	"",   // 312: 311: pi removed
	// 312: e removed
	"",   // 313: 312: e removed
	// 313: degToRad removed
	"",   // 314: 313: degToRad removed
	// 314: radToDeg removed
	"",   // 315: 314: radToDeg removed
	"adjustFloat", // 316
	"toKMG", // 317
	"trunc", // 318
	"isInf", // 319
	"isNaN", // 320
	"isFinite", // 321
	// JSON enhancement built-in functions (Batch 7)
	"formatJson", // 322
	"compactJson", // 323
	"getJsonNodeStr", // 324
	"getJsonNodeStrs", // 325
	"strsToJson", // 326
	"jsonValid", // 327
	"jsonType", // 328
	// Compression built-in functions (Batch 8)
	"compressData", // 329
	"uncompressData", // 330
	"compressStr", // 331
	"uncompressStr", // 332
	"zipPath", // 333
	"zipPaths", // 334
	"unzipToPath", // 335
	"getFileListInZip", // 336
	"loadBytesInZip", // 337
	"addFileToZip", // 338
	// Input/Clipboard built-in functions (Batch 9)
	"getInput", // 339
	"getInputf", // 340
	"getChar", // 341
	"getKey", // 342
	"getMultiLineInput", // 343
	"getPassword", // 344
	"confirm", // 345
	"readLine", // 346
	"getClipText", // 347
	"setClipText", // 348
	// 348: placeholder for index stability (skipped in compiler)
	"",   // 349: 348: placeholder for index stability (skipped in compiler)
	// String enhancement built-in functions (Batch 10)
	"strContainsIn", // 350
	"strRuneLen", // 351
	"strIn", // 352
	"strGetLastComponent", // 353
	"strFindDiffPos", // 354
	"strDiff", // 355
	"strFindAllSub", // 356
	"limitStr", // 357
	"strQuote", // 358
	"strUnquote", // 359
	"strToInt", // 360
	"getTextSimilarity", // 361
	"fuzzyFind", // 362
	// 362: strRemoveBom removed - use strings.removeBom()
	"",   // 363: 362: strRemoveBom removed - use strings.removeBom()
	// 363: reverseStr removed - use string.reverse()
	"",   // 364: 363: reverseStr removed - use string.reverse()
	// 364: capitalize removed - use string.capitalize()
	"",   // 365: 364: capitalize removed - use string.capitalize()
	// 365: title removed - use string.title()
	"",   // 366: 365: title removed - use string.title()
	// 366: swapCase removed - use string.swapCase()
	"",   // 367: 366: swapCase removed - use string.swapCase()
	// 367: center removed - use string.center()
	"",   // 368: 367: center removed - use string.center()
	// 368: zfill removed - use string.zfill()
	"",   // 369: 368: zfill removed - use string.zfill()
	// 369: isSpace removed - use string.isSpace()
	"",   // 370: 369: isSpace removed - use string.isSpace()
	// Collection enhancement built-in functions (Batch 11)
	"mapArray", // 371
	"filterArray", // 372
	"reduceArray", // 373
	"forEach", // 374
	"flatMap", // 375
	"every", // 376
	"some", // 377
	"groupBy", // 378
	"partition", // 379
	"zip", // 380
	"unzip", // 381
	"fill", // 382
	"rangeNum", // 383
	"intersection", // 384
	"difference", // 385
	"union", // 386
	"countBy", // 387
	"sortBy", // 388
	// Utility built-in functions (Batch 12)
	"sprintf", // 389
	"toBool", // 390
	"toInt", // 391
	"toFloat", // 392
	"isUndefined", // 393
	"isCallable", // 394
	"isIterable", // 395
	"isError", // 396
	"error", // 397
	"getErrStr", // 398
	"isErrStr", // 399
	"typeCode", // 400
	"swap", // 401
	"coalesce", // 402
	"defaultVal", // 403
	// String processing enhancement built-in functions (Batch 13)
	"strSplitLines", // 404
	"strContainsAny", // 405
	"strIndex", // 406
	"strLastIndex", // 407
	"strSplitN", // 408
	"strPad", // 409
	"strSub", // 410
	"intToStr", // 411
	"floatToStr", // 412
	"charCode", // 413
	"charFromCode", // 414
	"reverseMap", // 415
	"simpleStrToMap", // 416
	"mapToStr", // 417
	"bitNot", // 418
	"bitAnd", // 419
	"bitOr", // 420
	"bitXor", // 421
	"bitShiftLeft", // 422
	"bitShiftRight", // 423
	// Check/validate and bytes built-in functions (Batch 14)
	"isNil", // 424
	"isNull", // 425
	"isNilOrEmpty", // 426
	"isNilOrErr", // 427
	"isBytes", // 428
	"isChars", // 429
	"pass", // 430
	"errStrf", // 431
	"errf", // 432
	"errToEmpty", // 433
	"sscanf", // 434
	"bytesStartsWith", // 435
	"bytesEndsWith", // 436
	"bytesContains", // 437
	"bytesIndex", // 438
	"compareBytes", // 439
	"compareText", // 440
	// Miscellaneous built-in functions (Batch 15)
	"getRandomInt", // 441
	"getRandomFloat", // 442
	"getRandomStr", // 443
	"createTempDir", // 444
	"createTempFile", // 445
	"changeDir", // 446
	"lookPath", // 447
	"joinUrlPath", // 448
	"parseUrl", // 449
	"parseQuery", // 450
	"isHttps", // 451
	"genToken", // 452
	"genOtpCode", // 453
	"checkOtpCode", // 454
	// 454: toPinYin removed
	"",   // 455: 454: toPinYin removed
	// 455: kanaToRomaji removed
	"",   // 456: 455: kanaToRomaji removed
	// 456: kanjiToKana removed
	"",   // 457: 456: kanjiToKana removed
	// 457: kanjiToRomaji removed
	"",   // 458: 457: kanjiToRomaji removed
	// 458: genJwtToken removed
	"",   // 459: 458: genJwtToken removed
	// 459: parseJwtToken removed
	"",   // 460: 459: parseJwtToken removed
	// 460: isCronExprValid removed
	"",   // 461: 460: isCronExprValid removed
	// 461: isCronExprDue removed
	"",   // 462: 461: isCronExprDue removed
	// 462: runTicker removed
	"",   // 463: 462: runTicker removed
	// 463: stopTicker removed
	"",   // 464: 463: stopTicker removed
	// 464: genQr removed
	"",   // 465: 464: genQr removed
	// 465: scanQr removed
	"",   // 466: 465: scanQr removed
	// 466: getImageInfo removed
	"",   // 467: 466: getImageInfo removed
	// 467: resizeImage removed
	"",   // 468: 467: resizeImage removed
	"createImage", // 469
	// 469: placeholder (not defined in compiler)
	"",   // 470: 469: placeholder (not defined in compiler)
	// 470: placeholder
	"",   // 471: 470: placeholder
	// 471: placeholder
	"",   // 472: 471: placeholder
	// 472: placeholder
	"",   // 473: 472: placeholder
	// 473: placeholder
	"",   // 474: 473: placeholder
	// 474: placeholder
	"",   // 475: 474: placeholder
	// 475: placeholder
	"",   // 476: 475: placeholder
	// 476: placeholder
	"",   // 477: 476: placeholder
	// 477: placeholder
	"",   // 478: 477: placeholder
	// 478: placeholder
	"",   // 479: 478: placeholder
	// 479: placeholder
	"",   // 480: 479: placeholder
	// 480: placeholder
	"",   // 481: 480: placeholder
	// 481: placeholder
	"",   // 482: 481: placeholder
	// 482: placeholder
	"",   // 483: 482: placeholder
	// 483: placeholder
	"",   // 484: 483: placeholder
	// 484: placeholder
	"",   // 485: 484: placeholder
	// 485: placeholder
	"",   // 486: 485: placeholder
	// 486: placeholder
	"",   // 487: 486: placeholder
	// 487: placeholder
	"",   // 488: 487: placeholder
	// Byte-index string functions (Batch 23)
	"byteIndexOf", // 489
	"byteSubstr", // 490
	"byteLen", // 491
	// String enhancement (Batch 24)
	"strCount", // 492
	// Simple encoding (Batch 25)
	"simpleEncode", // 493
	"simpleDecode", // 494
	// Time enhancement (Batch 26)
	"now", // 495
	"getNowStrCompact", // 496
	"timeToTimeStamp", // 497
	"timeStampToTime", // 498
	// File system enhancement (Batch 27)
	"dirExists", // 499
	"pathExists", // 500
	"copyPath", // 501
	"moveFile", // 502
	"getFileSize", // 503
	// Print aliases (Batch 27)
	"print", // 504
	"println", // 505
	"printf", // 506
	"concatBytes", // 507
	"plv", // 508
	"spr", // 509
}
