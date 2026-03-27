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
		objects.Builtins["plt"],           // 114
		objects.Builtins["make"],          // 115
		// BigInt/BigFloat
		objects.Builtins["bigInt"],     // 116
		objects.Builtins["bigFloat"],   // 117
		objects.Builtins["isBigInt"],   // 118
		objects.Builtins["isBigFloat"], // 119
		// Chars (Unicode character handling)
		objects.Builtins["toChars"], // 120
		objects.Builtins["charLen"], // 121
		// HTTP built-in functions (for server mode)
		objects.Builtins["writeResp"],       // 122
		objects.Builtins["setRespHeader"],   // 123
		objects.Builtins["addRespHeader"],   // 124
		objects.Builtins["getReqHeader"],    // 125
		objects.Builtins["getReqHeaders"],   // 126
		objects.Builtins["setCookie"],       // 127
		objects.Builtins["getCookie"],       // 128
		objects.Builtins["getCookies"],      // 129
		objects.Builtins["parseForm"],       // 130
		objects.Builtins["parseJSON"],       // 131
		objects.Builtins["getReqBody"],      // 132
		objects.Builtins["getReqBodyBytes"], // 133
		objects.Builtins["status"],          // 134
		objects.Builtins["redirect"],        // 135
		objects.Builtins["serveFile"],       // 136
		objects.Builtins["getMimeType"],     // 137
		objects.Builtins["setContentType"],  // 138
		objects.Builtins["queryParam"],      // 139
		objects.Builtins["queryParams"],     // 140
		objects.Builtins["formValue"],       // 141
		objects.Builtins["httpStatusName"],  // 142
		objects.Builtins["isHttpReq"],       // 143
		objects.Builtins["isHttpResp"],      // 144
		objects.Builtins["urlEncode"],       // 145
		objects.Builtins["urlDecode"],       // 146
		// WebSocket built-in functions
		objects.Builtins["webSocket"],    // 147
		objects.Builtins["wsReadMsg"],    // 148
		objects.Builtins["wsSendText"],   // 149
		objects.Builtins["wsSendBinary"], // 150
		objects.Builtins["wsSendClose"],  // 151
		objects.Builtins["wsClose"],      // 152
		objects.Builtins["isWebSocket"],  // 153
		// Concurrency built-in functions
		objects.Builtins["makeTube"],     // 154
		objects.Builtins["closeTube"],    // 155
		objects.Builtins["tubeLen"],      // 156
		objects.Builtins["tubeCap"],      // 157
		objects.Builtins["tubeClosed"],   // 158
		objects.Builtins["tubeSend"],     // 159
		objects.Builtins["tubeRecv"],     // 160
		objects.Builtins["tubeTrySend"],  // 161
		objects.Builtins["tubeTryRecv"],  // 162
		objects.Builtins["newMutex"],     // 163
		objects.Builtins["newRWMutex"],   // 164
		objects.Builtins["newWaitGroup"], // 165
		objects.Builtins["newOnce"],      // 166
		objects.Builtins["newCond"],      // 167
		objects.Builtins["newAtomic"],    // 168
		// Context built-in functions
		objects.Builtins["newContext"],          // 169
		objects.Builtins["contextWithTimeout"],  // 170
		objects.Builtins["contextWithCancel"],   // 171
		objects.Builtins["contextWithDeadline"], // 172
		objects.Builtins["contextCancel"],       // 173
		objects.Builtins["contextDone"],         // 174
		objects.Builtins["contextErr"],          // 175
		objects.Builtins["contextIsDone"],       // 176
		objects.Builtins["contextDeadline"],     // 177
		// HTTP Client built-in functions (getWeb family)
		objects.Builtins["getWeb"],        // 178
		objects.Builtins["getWebBytes"],   // 179
		objects.Builtins["getWebObject"],  // 180
		objects.Builtins["postWeb"],       // 181
		objects.Builtins["postWebObject"], // 182
		objects.Builtins["urlExists"],     // 183
		objects.Builtins["httpStatus"],    // 184
		// Reader/Writer built-in functions
		objects.Builtins["getWebReader"],    // 185
		objects.Builtins["ioCopy"],          // 186
		objects.Builtins["isReader"],        // 187
		objects.Builtins["isWriter"],        // 188
		objects.Builtins["newBytesReader"],  // 189
		objects.Builtins["newStringReader"], // 190
		// Encryption built-in functions (Charlang compatible)
		objects.Builtins["encryptTextByTXTE"],  // 191
		objects.Builtins["decryptTextByTXTE"],  // 192
		objects.Builtins["encryptDataByTXDEE"], // 193
		objects.Builtins["decryptDataByTXDEE"], // 194
		objects.Builtins["encryptTextByTXDEE"], // 195
		objects.Builtins["decryptTextByTXDEE"], // 196
		objects.Builtins["encryptData"],        // 197
		objects.Builtins["encryptBytes"],       // 198
		objects.Builtins["decryptData"],        // 199
		objects.Builtins["decryptBytes"],       // 200
		objects.Builtins["encryptText"],        // 201
		objects.Builtins["encryptStr"],         // 202
		objects.Builtins["decryptText"],        // 203
		objects.Builtins["decryptStr"],         // 204
		objects.Builtins["encryptStream"],      // 205
		objects.Builtins["decryptStream"],      // 206
		objects.Builtins["aesEncrypt"],         // 207
		objects.Builtins["aesDecrypt"],         // 208
		objects.Builtins["downloadFile"],       // 209
		// Database built-in functions (String-based - Charlang compatible)
		objects.Builtins["formatSQLValue"],  // 210
		objects.Builtins["dbConnect"],       // 211
		objects.Builtins["dbClose"],         // 212
		objects.Builtins["dbQuery"],         // 213
		objects.Builtins["dbQueryOrdered"],  // 214
		objects.Builtins["dbQueryRecs"],     // 215
		objects.Builtins["dbQueryMap"],      // 216
		objects.Builtins["dbQueryMapArray"], // 217
		objects.Builtins["dbQueryCount"],    // 218
		objects.Builtins["dbQueryFloat"],    // 219
		objects.Builtins["dbQueryString"],   // 220
		objects.Builtins["dbExec"],          // 221
		// Database built-in functions (Typed - preserve native types)
		objects.Builtins["dbQueryTyped"],      // 222
		objects.Builtins["dbQueryRowTyped"],   // 223
		objects.Builtins["dbQueryArrayTyped"], // 224
		objects.Builtins["dbQueryValueTyped"], // 225
		// OrderedMap built-in functions
		objects.Builtins["isOrderedMap"],  // 226
		objects.Builtins["newOrderedMap"], // 227
		// System command built-in functions
		objects.Builtins["systemCmd"],         // 228
		objects.Builtins["systemCmdDetached"], // 229
		objects.Builtins["systemStart"],       // 230
		// Test assertion built-in functions
		objects.Builtins["testByText"],        // 231
		objects.Builtins["testByStartsWith"],  // 232
		objects.Builtins["testByEndsWith"],    // 233
		objects.Builtins["testByContains"],    // 234
		objects.Builtins["testByReg"],         // 235
		objects.Builtins["testByRegContains"], // 236
		objects.Builtins["dumpVar"],           // 237
		objects.Builtins["debugInfo"],         // 238
		// File system built-in functions (Batch 1)
		objects.Builtins["fileExists"],      // 239
		objects.Builtins["isDir"],           // 240
		objects.Builtins["loadText"],        // 241
		objects.Builtins["saveText"],        // 242
		objects.Builtins["appendText"],      // 243
		objects.Builtins["copyFile"],        // 244
		objects.Builtins["renameFile"],      // 245
		objects.Builtins["removeFile"],      // 246
		objects.Builtins["removeDir"],       // 247
		objects.Builtins["getFileList"],     // 248
		objects.Builtins["joinPath"],        // 249
		objects.Builtins["getCurDir"],       // 250
		objects.Builtins["getHomeDir"],      // 251
		objects.Builtins["getTempDir"],      // 252
		objects.Builtins["ensureMakeDirs"],  // 253
		objects.Builtins["getFileExt"],      // 254
		objects.Builtins["extractFileDir"],  // 255
		objects.Builtins["extractFileName"], // 256
		objects.Builtins["getFileInfo"],     // 257
		objects.Builtins["loadLines"],       // 258
		objects.Builtins["getFileAbs"],      // 259
		objects.Builtins["getFileRel"],      // 260
		objects.Builtins["isFile"],          // 261
		// Time enhancement built-in functions (Batch 2)
		objects.Builtins["getNowStr"],       // 262
		objects.Builtins["getNowTimeStamp"], // 263
		objects.Builtins["formatTime"],      // 264
		objects.Builtins["timeToTick"],      // 265
		objects.Builtins["timeAddSecs"],     // 266
		objects.Builtins["timeAddDate"],     // 267
		objects.Builtins["timeBefore"],      // 268
		objects.Builtins["strToTime"],       // 269
		objects.Builtins["timeAfter"],       // 270
		objects.Builtins["timeEqual"],       // 271
		objects.Builtins["timeDiff"],        // 272
		objects.Builtins["timeDiffSecs"],    // 273
		objects.Builtins["parseTime"],       // 274
		objects.Builtins["isTime"],          // 275
		// Regex enhancement built-in functions (Batch 3)
		objects.Builtins["regMatch"],           // 276
		objects.Builtins["regContains"],        // 277
		objects.Builtins["regFindFirst"],       // 278
		objects.Builtins["regFindAll"],         // 279
		objects.Builtins["regFindFirstGroups"], // 280
		objects.Builtins["regFindAllGroups"],   // 281
		objects.Builtins["regReplace"],         // 282
		objects.Builtins["regSplit"],           // 283
		objects.Builtins["regCount"],           // 284
		objects.Builtins["regQuote"],           // 285
		objects.Builtins["regFindAllIndex"],    // 286
		// Encoding enhancement built-in functions (Batch 4)
		objects.Builtins["urlEncodeComponent"], // 287
		objects.Builtins["urlDecodeComponent"], // 288
		objects.Builtins["htmlEncode"],         // 289
		objects.Builtins["htmlDecode"],         // 290
		objects.Builtins["sha1"],               // 291
		objects.Builtins["sha512"],             // 292
		objects.Builtins["hashStr"],            // 293
		objects.Builtins["toHex"],              // 294
		objects.Builtins["unhex"],              // 295
		objects.Builtins["hexToStr"],           // 296
		// System/Environment built-in functions (Batch 5)
		objects.Builtins["getEnv"],     // 297
		objects.Builtins["setEnv"],     // 298
		objects.Builtins["getOSName"],  // 299
		objects.Builtins["getOSArch"],  // 300
		objects.Builtins["getOSArgs"],  // 301
		objects.Builtins["getAppPath"], // 302
		objects.Builtins["getAppDir"],  // 303
		objects.Builtins["exit"],       // 304
		objects.Builtins["getSysInfo"], // 305
		objects.Builtins["getPid"],     // 306
		objects.Builtins["getPPid"],    // 307
		objects.Builtins["hostname"],   // 308
		// Math enhancement built-in functions (Batch 6)
		objects.Builtins["sin"],         // 309
		objects.Builtins["cos"],         // 310
		objects.Builtins["tan"],         // 311
		objects.Builtins["asin"],        // 312
		objects.Builtins["acos"],        // 313
		objects.Builtins["atan"],        // 314
		objects.Builtins["atan2"],       // 315
		objects.Builtins["exp"],         // 316
		objects.Builtins["log"],         // 317
		objects.Builtins["log10"],       // 318
		objects.Builtins["log2"],        // 319
		objects.Builtins["pi"],          // 320
		objects.Builtins["e"],           // 321
		objects.Builtins["degToRad"],    // 322
		objects.Builtins["radToDeg"],    // 323
		objects.Builtins["adjustFloat"], // 324
		objects.Builtins["toKMG"],       // 325
		objects.Builtins["trunc"],       // 326
		objects.Builtins["isInf"],       // 327
		objects.Builtins["isNaN"],       // 328
		objects.Builtins["isFinite"],    // 329
		// JSON enhancement built-in functions (Batch 7)
		objects.Builtins["formatJson"],      // 330
		objects.Builtins["compactJson"],     // 331
		objects.Builtins["getJsonNodeStr"],  // 332
		objects.Builtins["getJsonNodeStrs"], // 333
		objects.Builtins["strsToJson"],      // 334
		objects.Builtins["jsonValid"],       // 335
		objects.Builtins["jsonType"],        // 336
		// Compression built-in functions (Batch 8)
		objects.Builtins["compressData"],     // 337
		objects.Builtins["uncompressData"],   // 338
		objects.Builtins["compressStr"],      // 339
		objects.Builtins["uncompressStr"],    // 340
		objects.Builtins["zipPath"],          // 341
		objects.Builtins["zipPaths"],         // 342
		objects.Builtins["unzipToPath"],      // 343
		objects.Builtins["getFileListInZip"], // 344
		objects.Builtins["loadBytesInZip"],   // 345
		objects.Builtins["addFileToZip"],     // 346
		// Input/Clipboard built-in functions (Batch 9)
		objects.Builtins["getInput"],          // 347
		objects.Builtins["getInputf"],         // 348
		objects.Builtins["getChar"],           // 349
		objects.Builtins["getMultiLineInput"], // 350
		objects.Builtins["getPassword"],       // 351
		objects.Builtins["confirm"],           // 352
		objects.Builtins["readLine"],          // 353
		objects.Builtins["getClipText"],       // 354
		objects.Builtins["setClipText"],       // 355
		// String enhancement built-in functions (Batch 10)
		objects.Builtins["strContainsIn"],       // 356
		objects.Builtins["strRuneLen"],          // 357
		objects.Builtins["strIn"],               // 358
		objects.Builtins["strGetLastComponent"], // 359
		objects.Builtins["strFindDiffPos"],      // 360
		objects.Builtins["strDiff"],             // 361
		objects.Builtins["strFindAllSub"],       // 362
		objects.Builtins["limitStr"],            // 363
		objects.Builtins["strQuote"],            // 364
		objects.Builtins["strUnquote"],          // 365
		objects.Builtins["strToInt"],            // 366
		objects.Builtins["getTextSimilarity"],   // 367
		objects.Builtins["fuzzyFind"],           // 368
		objects.Builtins["strRemoveBom"],        // 369
		objects.Builtins["wordCount"],           // 370
		objects.Builtins["lineCount"],           // 371
		objects.Builtins["reverseStr"],          // 372
		objects.Builtins["capitalize"],          // 373
		objects.Builtins["title"],               // 374
		objects.Builtins["swapCase"],            // 375
		objects.Builtins["center"],              // 376
		objects.Builtins["zfill"],               // 377
		objects.Builtins["isSpace"],             // 378
		// Collection enhancement built-in functions (Batch 11)
		objects.Builtins["mapArray"],     // 379
		objects.Builtins["filterArray"],  // 380
		objects.Builtins["reduceArray"],  // 381
		objects.Builtins["forEach"],      // 382
		objects.Builtins["flatMap"],      // 383
		objects.Builtins["every"],        // 384
		objects.Builtins["some"],         // 385
		objects.Builtins["groupBy"],      // 386
		objects.Builtins["partition"],    // 387
		objects.Builtins["zip"],          // 388
		objects.Builtins["unzip"],        // 389
		objects.Builtins["fill"],         // 390
		objects.Builtins["rangeNum"],     // 391
		objects.Builtins["intersection"], // 392
		objects.Builtins["difference"],   // 393
		objects.Builtins["union"],        // 394
		objects.Builtins["countBy"],      // 395
		objects.Builtins["sortBy"],       // 396
		// Utility built-in functions (Batch 12)
		objects.Builtins["sprintf"],     // 397
		objects.Builtins["toBool"],      // 398
		objects.Builtins["toInt"],       // 399
		objects.Builtins["toFloat"],     // 400
		objects.Builtins["isUndefined"], // 401
		objects.Builtins["isCallable"],  // 402
		objects.Builtins["isIterable"],  // 403
		objects.Builtins["isError"],     // 404
		objects.Builtins["error"],       // 405
		objects.Builtins["getErrStr"],   // 406
		objects.Builtins["isErrStr"],    // 407
		objects.Builtins["typeCode"],    // 408
		objects.Builtins["clone"],       // 409
		objects.Builtins["swap"],        // 410
		objects.Builtins["coalesce"],    // 411
		objects.Builtins["defaultVal"],  // 412
		// String processing enhancement built-in functions (Batch 13)
		objects.Builtins["strSplitLines"],  // 413
		objects.Builtins["strContainsAny"], // 414
		objects.Builtins["strIndex"],       // 415
		objects.Builtins["strLastIndex"],   // 416
		objects.Builtins["strSplitN"],      // 417
		objects.Builtins["strPad"],         // 418
		objects.Builtins["strSub"],         // 419
		objects.Builtins["intToStr"],       // 420
		objects.Builtins["floatToStr"],     // 421
		objects.Builtins["charCode"],       // 422
		objects.Builtins["charFromCode"],   // 423
		objects.Builtins["reverseMap"],     // 424
		objects.Builtins["simpleStrToMap"], // 425
		objects.Builtins["mapToStr"],       // 426
		objects.Builtins["bitNot"],         // 427
		objects.Builtins["bitAnd"],         // 428
		objects.Builtins["bitOr"],          // 429
		objects.Builtins["bitXor"],         // 430
		objects.Builtins["bitShiftLeft"],   // 431
		objects.Builtins["bitShiftRight"],  // 432
		// Check/validate and bytes built-in functions (Batch 14)
		objects.Builtins["isNil"],           // 433
		objects.Builtins["isNull"],          // 434
		objects.Builtins["isNilOrEmpty"],    // 435
		objects.Builtins["isNilOrErr"],      // 436
		objects.Builtins["isBytes"],         // 437
		objects.Builtins["isChars"],         // 438
		objects.Builtins["pass"],            // 439
		objects.Builtins["errStrf"],         // 440
		objects.Builtins["errf"],            // 441
		objects.Builtins["errToEmpty"],      // 442
		objects.Builtins["sscanf"],          // 443
		objects.Builtins["bytesStartsWith"], // 444
		objects.Builtins["bytesEndsWith"],   // 445
		objects.Builtins["bytesContains"],   // 446
		objects.Builtins["bytesIndex"],      // 447
		objects.Builtins["compareBytes"],    // 448
		objects.Builtins["compareText"],     // 449
		// Miscellaneous built-in functions (Batch 15)
		objects.Builtins["getRandomInt"],   // 450
		objects.Builtins["getRandomFloat"], // 451
		objects.Builtins["getRandomStr"],   // 452
		objects.Builtins["createTempDir"],  // 453
		objects.Builtins["createTempFile"], // 454
		objects.Builtins["changeDir"],      // 455
		objects.Builtins["lookPath"],       // 456
		objects.Builtins["joinUrlPath"],    // 457
		objects.Builtins["parseUrl"],       // 458
		objects.Builtins["parseQuery"],     // 459
		objects.Builtins["isHttps"],        // 460
		objects.Builtins["genToken"],       // 461
		objects.Builtins["genOtpCode"],     // 462
		objects.Builtins["checkOtpCode"],   // 463
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
