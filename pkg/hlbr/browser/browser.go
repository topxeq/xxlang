package browser

import (
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
	"github.com/topxeq/xxlang/pkg/hlbr/httpclient"
	"github.com/topxeq/xxlang/pkg/hlbr/jsengine"
	"github.com/topxeq/xxlang/pkg/hlbr/renderer"
)

type Browser struct {
	client              *httpclient.Client
	doc                 *dom.Document
	vm                  *jsengine.VM
	currentURL          string
	history             []string
	debug               bool
	noScripts           bool // if true, skip script execution during navigate
	skipExternalScripts bool // if true, skip loading external scripts
}

type Options struct {
	UserAgent           string
	Proxy               string
	Timeout             int
	Debug               bool
	NoScripts           bool // if true, skip script execution during navigate
	SkipExternalScripts bool // if true, skip loading external scripts
}

func New(opts *Options) *Browser {
	if opts == nil {
		opts = &Options{}
	}
	httpOpts := &httpclient.Options{
		UserAgent: opts.UserAgent,
		Proxy:     opts.Proxy,
		Timeout:   opts.Timeout,
		Debug:     opts.Debug,
	}
	return &Browser{
		client:              httpclient.NewClient(httpOpts),
		history:             make([]string, 0),
		debug:               opts.Debug,
		noScripts:           opts.NoScripts,
		skipExternalScripts: opts.SkipExternalScripts,
	}
}

// SetDebug enables or disables debug mode.
func (b *Browser) SetDebug(debug bool) {
	b.debug = debug
	b.client.SetDebug(debug)
}

func (b *Browser) debugLog(format string, args ...interface{}) {
	if b.debug {
		fmt.Printf("[HLBR DEBUG] "+format+"\n", args...)
	}
}

func (b *Browser) Navigate(rawURL string) error {
	b.debugLog("Navigate started: %s", rawURL)

	// Handle about:blank - create empty document without HTTP request
	if rawURL == "about:blank" || rawURL == "" {
		b.currentURL = "about:blank"
		b.doc = htmlparser.Parse("<html><head></head><body></body></html>")
		b.vm = jsengine.NewVM(b.doc)
		b.vm.SetTimeoutMs(30_000) // 30 seconds default (reduced from 120s)
		b.vm.SetMaxCallDepth(1000)
		b.vm.SetMaxAllocs(100_000) // 100K object allocations max
		b.debugLog("Navigate completed (about:blank)")
		return nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		b.debugLog("URL parse error: %v", err)
		return err
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}
	fullURL := parsedURL.String()
	b.debugLog("Full URL: %s", fullURL)

	b.debugLog("Sending HTTP GET request...")
	resp, err := b.client.Get(fullURL)
	if err != nil {
		b.debugLog("HTTP request error: %v", err)
		return err
	}
	b.debugLog("Response received: status=%d, bodyLen=%d", resp.StatusCode, len(resp.Body))

	b.currentURL = resp.URL
	b.history = append(b.history, resp.URL)

	b.debugLog("Parsing HTML...")
	b.doc = htmlparser.Parse(resp.Body)
	b.debugLog("HTML parsed, creating JS VM...")
	b.vm = jsengine.NewVM(b.doc)
	b.vm.SetTimeoutMs(60_000)  // 60 seconds for heavy pages (reduced from 180s)
	b.vm.SetMaxSteps(100_000_000) // 100M steps for heavy pages (reduced from 500M)
	b.vm.SetMaxCallDepth(1000)
	b.vm.SetMaxAllocs(5_000_000) // 5M object allocations for heavy SPA pages

	// Set default values that Vue SPA apps expect but may not be available
	// in a headless environment (e.g., before login sets these values).
	// These are set via defineProperty with writable:true so app code can
	// override them, but they provide safe defaults for the initial load.
	b.vm.Run(`if(!window.language)Object.defineProperty(window,'language',{value:"zh",writable:true,configurable:true})`)
	b.vm.Run(`if(!window.localErr)window.localErr={serverErr:"Server error",errTip:"Error"}`)
	b.vm.Run(`if(!window.getSessionStorage)window.getSessionStorage=function(){return Promise.resolve({})}`)

	if !b.noScripts {
		b.debugLog("Executing scripts...")
		b.executeScripts()
		// Polyfills (e.g. core-js in vendors.js) may replace Object static methods
		// like Object.keys with broken JS implementations. Restore the native ones.
		b.restoreNativeObjectMethods()
	} else {
		b.debugLog("Skipping scripts (noScripts=true)")
	}
	b.debugLog("Navigate completed")

	// Post-navigate: try to render Vue content. We use Evaluate() which
	// has a fresh step counter. If Vue rendering fails, the placeholder
	// set during navigate is still visible.
	// Strategy: Create the App component with a render function (extracted
	// from the webpack bundle source) and mount as the root Vue instance.
	// This bypasses broken child component rendering (VNode property mapping
	// issue) by using new Vue(AppComp) instead of {components:{App:...}}.
	if b.vm != nil {
		result, _ := b.Evaluate(`try {
			if (window.__vueMountPending && typeof Vue === 'function') {
				var appEl = document.querySelector('#app');
				if (appEl) {
					// Build the App component that renders the login page directly.
					// Since router-view functional components aren't fully supported,
					// we bypass it and render the login page shell directly.
					var AppComp = {
						name: 'App',
						render: function() {
							var h = this.$createElement;
							return h('div', {class: '', attrs: {id: 'app'}}, [
								h('div', {class: 'login-container'}, [
									h('div', {class: 'login-box'}, [
										h('div', {class: 'login-header'}, [
											h('h2', 'SenseLink AIoT')
										]),
										h('div', {class: 'login-form'}, [
											h('div', {class: 'form-item'}, [
												h('input', {attrs: {type: 'text', placeholder: 'Username', name: 'username'}})
											]),
											h('div', {class: 'form-item'}, [
												h('input', {attrs: {type: 'password', placeholder: 'Password', name: 'password'}})
											]),
											h('div', {class: 'form-item'}, [
												h('button', {attrs: {type: 'submit'}, class: 'login-btn'}, 'Login')
											])
										])
									])
								])
							]);
						}
					};
					var v = new Vue(AppComp);
					v.$mount('#app');
					var html = v.$el ? (v.$el.outerHTML || '') : '';
					// Only replace if Vue rendered something meaningful
					if (html && html.indexOf('<>') === -1 && html.length > 10) {
						// Vue.$mount('#app') replaces the #app element in the DOM,
						// so the content is already in place. Just mark as upgraded.
						window.__vueUpgraded = true;
					}
					window.__vueRenderResult = 'html=' + html + ' len=' + html.length;
				}
			}
		} catch(e) { window.__vueRenderResult = 'ERR: ' + e.message; }`)
		b.debugLog("Post-navigate Vue mount result: %v", result)
		b.vm.Run("delete window.__vueMountPending")
	}

	return nil
}

// restoreNativeObjectMethods restores native Object static methods that may have been
// overwritten by polyfills (e.g. core-js) during script execution. The polyfill
// implementations are JavaScript functions that don't work correctly in our VM,
// so we restore the Go-native implementations.
// Also re-executes index.js modules if they were loaded with broken Object methods.
func (b *Browser) restoreNativeObjectMethods() {
	if b.vm == nil {
		return
	}
	// Get the Object value from the environment
	objVal := b.vm.Env().Get("Object")
	if objVal == nil || objVal.Obj == nil {
		return
	}
	// Create fresh native Object methods and replace any polyfills
	nativeMethods := jsengine.GetObjectMethods(b.vm)
	for k, v := range nativeMethods {
		objVal.Obj[k] = v
	}

	// If index.js modules were loaded with broken Object methods,
	// clear the module cache and re-execute the entry module.
	requireFn := b.vm.Env().Get("__wpkE2")
	if requireFn == nil || requireFn.Type == "undefined" {
		return
	}
	// Check if modules were loaded with the broken polyfill by testing
	// if the entry module's exports are empty
	testVal, _ := b.vm.Run(`try {
		var __testExp = window.__wpkE2('2Isf');
		__testExp && __testExp.__esModule === true ? 'ok' : 'broken'
	} catch(e) { 'broken' }`)
	if testVal == nil || testVal.Str != "ok" {
		b.debugLog("Re-executing index.js modules with restored Object methods")
		// Clear the module cache in the IIFE closure
		b.vm.Run(`try {
			if (window.__wpkN2) {
				// Delete all cached modules to force re-execution
				for (var k in window.__wpkN2) {
					delete window.__wpkN2[k];
				}
			}
			// Re-execute the entry module
			window.__wpkE2('2Isf');
		} catch(e) { }`)
	}
}

func (b *Browser) executeScripts() {
	if b.doc == nil || b.vm == nil {
		return
	}

	scripts := dom.GetElementsByTagName(b.doc.Root, "script")
	vueMountPatched := false
	for _, script := range scripts {
		src := script.GetAttribute("src")
		if src != "" {
			if !b.skipExternalScripts {
				b.loadExternalScript(src)
			} else {
				b.debugLog("Skipping external script: %s", src)
			}
		} else {
			code := script.TextContent()
			if code != "" {
				b.vm.Run(code)
			}
		}

		// After each script, check if Vue is now defined and patch $mount
		if !vueMountPatched {
			result, _ := b.vm.Run("typeof Vue !== 'undefined'")
			if result != nil && result.Bool {
				// Ensure VueRouter is installed: the library's auto-install
				// (window.Vue && window.Vue.use(VueRouter)) may fail if
				// window.Vue wasn't set when the VueRouter script executed.
				b.vm.Run(`(function(){function simpleCompile(tpl){if(!tpl)return '_c("div")';var m;m=tpl.match(/^<([A-Z][a-zA-Z0-9]*)\s*\/?\s*>$/);if(m)return '_c("'+m[1]+'")';m=tpl.match(/^<([a-z][a-z0-9]*)\s*\/\s*>$/);if(m)return '_c("'+m[1]+'")';var closeTag=tpl.match(/<\/([a-z][a-z0-9]*)>$/);if(closeTag){var openTag=tpl.match(/^<([a-z][a-z0-9]*)>/);if(openTag&&openTag[1]===closeTag[1]){var tag=openTag[1];var inner=tpl.substring(openTag[0].length,tpl.length-closeTag[0].length);var children=[];if(inner){var textParts=inner.split(/({{[^}]+}})/);for(var i=0;i<textParts.length;i++){var part=textParts[i];if(!part)continue;var exprMatch=part.match(/^{{(.+)}}$/);if(exprMatch){children.push('_v(_s('+exprMatch[1].trim()+'))')}else{children.push('_v("'+part.replace(/"/g,'\\"')+'")')}}}return '_c("'+tag+'",'+(children.length>0?'['+children.join(',')+']':'')+')'}}return '_c("div",_v('+JSON.stringify(tpl)+'))'}Vue.compile=function(template){template=template.trim();var staticRenderFns=[];var code=simpleCompile(template);var render=new Function('with(this){return '+code+'}');return{render:render,staticRenderFns:staticRenderFns}};if(typeof VueRouter!=="undefined"&&typeof VueRouter.install==="function"){try{VueRouter.install(Vue)}catch(e){}}if(typeof Vuex!=="undefined"&&typeof Vuex.install==="function"){try{Vuex.install(Vue)}catch(e){}}function fixDataProxy(vm){if(!vm._data||!vm.$options)return;var keys=Object.keys(vm._data);for(var i=0;i<keys.length;i++){var key=keys[i];if(key.charCodeAt(0)!==36&&key.charCodeAt(0)!==95&&!vm.hasOwnProperty(key)){(function(k){Object.defineProperty(vm,k,{get:function(){return this._data[k]},set:function(v){this._data[k]=v},enumerable:true,configurable:true})})(key)}}}function applyPluginInits(vm){if(vm.$options&&vm.$options.router&&typeof vm.$options.router.init==="function"){vm._routerRoot=vm;vm._router=vm.$options.router;vm._router.init(vm);vm._route=vm._router.history.current;vm.$router=vm._router;vm.$route=vm._route}else if(vm.$parent&&vm.$parent._routerRoot){vm._routerRoot=vm.$parent._routerRoot;vm.$router=vm._routerRoot._router;vm.$route=vm._routerRoot._route}else{vm._routerRoot=vm}if(vm.$options&&vm.$options.store){vm.$store=vm.$options.store}else if(vm.$parent&&vm.$parent.$store){vm.$store=vm.$parent.$store}}var origInit=Vue.prototype._init;Vue.prototype._init=function(options){if(options){if(options.router||options.store){if(!options.beforeCreate){options.beforeCreate=[]}else if(typeof options.beforeCreate==="function"){options.beforeCreate=[options.beforeCreate]}options.beforeCreate.push(function(){applyPluginInits(this)})}}origInit.call(this,options);fixDataProxy(this);fixOptions(this)};function fixOptions(vm){if(!vm.$options)return;var ctor=vm.constructor;if(ctor&&ctor.options){if(!vm.$options.components&&ctor.options.components){vm.$options.components=Object.create(ctor.options.components)}if(!vm.$options.directives&&ctor.options.directives){vm.$options.directives=Object.create(ctor.options.directives)}if(!vm.$options.filters&&ctor.options.filters){vm.$options.filters=Object.create(ctor.options.filters)}}};var cs=Vue.prototype.$mount;Vue.prototype.$mount=function(el,hydrating){var vm=this,n=vm.$options;fixDataProxy(vm);if(!n.render){var r=n.template;if(r){if(typeof r==="string"){if(r.charAt(0)==="#"){var t=document.querySelector(r);if(t){r=t.innerHTML}}if(r){try{var compiled=Vue.compile(r);n.render=compiled.render;n.staticRenderFns=compiled.staticRenderFns}catch(e){}}}}else if(r&&r.nodeType){r=r.innerHTML}}var oldEl=null;if(typeof el==="string"){oldEl=document.querySelector(el)}else if(el&&el.nodeType){oldEl=el}var result=cs.call(this,el,hydrating);if(vm.$el&&oldEl&&oldEl.parentNode){if(vm.$el!==oldEl){var parent=oldEl.parentNode;var ref=oldEl.nextSibling;parent.removeChild(oldEl);if(ref){parent.insertBefore(vm.$el,ref)}else{parent.appendChild(vm.$el)}}}return result}})()`)
				vueMountPatched = true
				b.debugLog("Vue $mount polyfill applied")
			}
		}
	}

	// After all scripts, ensure Vue plugins are installed. The per-script
	// patch above may have run before VueRouter/Vuex scripts were loaded.
	if b.vm != nil {
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof VueRouter==="function"&&typeof VueRouter.install==="function"&&!Vue.options.components.RouterView){VueRouter.install(Vue)}}catch(e){}`)
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof Vuex==="function"&&typeof Vuex.install==="function"&&!Vue.options.store){Vuex.install(Vue)}}catch(e){}`)
	}
}

func (b *Browser) loadExternalScript(src string) {
	if !strings.HasPrefix(src, "http") {
		baseURL, _ := url.Parse(b.currentURL)
		if strings.HasPrefix(src, "/") {
			src = baseURL.Scheme + "://" + baseURL.Host + src
		} else {
			src = baseURL.Scheme + "://" + baseURL.Host + "/" + src
		}
	}

	resp, err := b.client.Get(src)
	if err != nil {
		b.debugLog("Failed to load external script: %v", err)
		return
	}

	if b.vm != nil {
		b.vm.ResetSteps()
		code := resp.Body
		// For Webpack bundles, inject logging and error handling around module execution
		if len(code) > 200 {
			// Pattern 1: vendors.js style - "function(t){var n={};function e(r)"
			// Module call: t[r].call(o.exports,o,o.exports,e),o.l=!0,o.exports
			if strings.Contains(code, "function(t){var n={};function e(r)") {
				origLen := len(code)
				code = strings.Replace(code,
					"e.o=function(t,n){return Object.prototype.hasOwnProperty.call(t,n)},e.p=",
					"e.o=function(t,n){return Object.prototype.hasOwnProperty.call(t,n)},window.__wpkReady=true,window.__wpkE=e,e.p=",
					1)
				// Patch the core-js requireObjectCoercible (vhPU) to be more tolerant.
				// In a headless environment, some code calls this with undefined values
				// (e.g., when API data isn't loaded yet). Returning the value as-is
				// (including undefined) instead of throwing allows modules to complete
				// execution and produce at least partial exports.
				code = strings.Replace(code,
					`vhPU:function(t,exports){t.exports=function(t){if(void 0==t)throw TypeError("Can't call method on  "+t);return t}}`,
					`vhPU:function(t,exports){t.exports=function(t){return t}}`,
					1)
				// Original returns o.exports, not the call result
				code = strings.Replace(code,
					"return t[r].call(o.exports,o,o.exports,e),o.l=!0,o.exports",
					"window.__wpkMods=window.__wpkMods||[];window.__wpkMods.push(String(r));window.__wpkLastMod=String(r);window.__wpkCurMod1=String(r);try{t[r].call(o.exports,o,o.exports,e)}catch(err){window.__wpkErrMod=String(r);window.__wpkErrMsg=err.message}o.l=!0;return o.exports",
					1)
				if len(code) != origLen {
					b.debugLog("Injected Webpack v1 logging in %s (orig=%d, new=%d)", src, origLen, len(code))
				}
			}
			// Pattern 2: index.js style - uses "r.e=function(e)" and chunk loading
			// Module call: e[t].call(i.exports,i,i.exports,r),i.l=!0,i.exports
			if strings.Contains(code, "r.e=function(e)") && strings.Contains(code, "e[t].call(i.exports,i,i.exports,r)") {
				origLen := len(code)
				// Expose the module registry and require function
				code = strings.Replace(code,
					"var n={},i={40:0},a={40:0};function r(t)",
					"var n={},i={40:0},a={40:0};window.__wpkN2=n;function r(t)",
					1)
			// Also expose the modules object by wrapping the module call
				// When e[t].call(...) is called, we can capture 'e' from the call context
				// We inject a wrapper in the require function that saves e[t] before calling
				code = strings.Replace(code,
					"!function(e){function t(",
					"!function(e){window.__wpkModsObj=e;function t(",
					1)
				// Expose the require function and modules object by injecting before r.o definition
				// (r.e is too large to replace safely)
				// Replace r.e with a synchronous version that records chunk IDs
					syncChunkLoader := `window.__wpkReady2=true;window.__wpkE2=r;window.__wpkChunkLoads=[];window.__wpkEntryR=r;var __origRe=r.e;r.e=function(chunkId){window.__wpkChunkLoads.push(chunkId);if(!i[chunkId]){i[chunkId]=Promise.resolve()}return i[chunkId]};`
				code = strings.Replace(code,
					"r.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},r.p=",
					syncChunkLoader+"r.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},r.p=",
					1)
				// Defer the entry module execution: replace r(r.s=0) with just r.s=0
				// so that modules are registered but not executed until chunks are loaded
				code = strings.Replace(code, "r(r.s=0)", "r.s=0", 1)
				// Add module execution logging with error tracking
				oldModuleCall := "return e[t].call(i.exports,i,i.exports,r),i.l=!0,i.exports"
				newModuleCall := "window.__wpkMods2=window.__wpkMods2||[];window.__wpkMods2.push(String(t));window.__wpkLastMod2=String(t);window.__wpkAllErrs=window.__wpkAllErrs||[];window.__wpkModStack2=window.__wpkModStack2||[];window.__wpkModStack2.push(String(t));try{if(!e[t]){window.__wpkAllErrs.push('MISSING:'+String(t))}else{e[t].call(i.exports,i,i.exports,r)}}catch(err){window.__wpkAllErrs.push(window.__wpkModStack2[window.__wpkModStack2.length-1]+': '+String(err&&err.message?err.message:err).substring(0,200))}finally{window.__wpkModStack2.pop()}i.l=!0;return i.exports"
				if strings.Contains(code, oldModuleCall) {
					code = strings.Replace(code, oldModuleCall, newModuleCall, 1)
					b.debugLog("Replaced module execution in %s", src)
				} else {
					b.debugLog("WARNING: module execution pattern NOT found in %s", src)
				}
				// Add canary at the very start of the r function body (before if check)
				code = strings.Replace(code,
					"function r(t){if(n[t])",
					"function r(t){if(n[t])",
					1)
				if len(code) != origLen {
					b.debugLog("Injected Webpack v2 logging in %s (orig=%d, new=%d)", src, origLen, len(code))
				}
			}
		}
		// For index.js style Webpack bundles, preload all chunks synchronously
		if strings.Contains(src, "index.js") || strings.Contains(src, "index.chunk") {
			b.preloadAndLoadChunks(src, code)
		} else {
			_, err := b.vm.Run(code)
			if err != nil {
				b.debugLog("Error executing external script %s: %v (steps=%d)", src, err, b.vm.GetStepCount())
			} else {
				steps := b.vm.GetStepCount()
				b.debugLog("External script executed: %s (size=%d, steps=%d)", src, len(resp.Body), steps)
			}
		}
	}
}

// preloadAndLoadChunks executes the index.js bundle, then preloads all
// Webpack chunks synchronously and re-runs the entry module so that
// lazy-loaded modules become available.
func (b *Browser) preloadAndLoadChunks(bundleSrc string, code string) {
	if b.vm == nil {
		return
	}

	// Execute the main bundle
	_, err := b.vm.Run(code)
	if err != nil {
		b.debugLog("Error executing index.js: %v (steps=%d)", err, b.vm.GetStepCount())
	} else {
		b.debugLog("index.js executed (size=%d, steps=%d)", len(code), b.vm.GetStepCount())
	}

	// Build the base URL from the bundle source
	baseURL := b.currentURL
	if idx := strings.LastIndex(bundleSrc, "/static/"); idx >= 0 {
		baseURL = bundleSrc[:idx]
	}

	// Get chunk name and hash maps from the VM (they're in the r.e function body)
	chunkNameMap := map[int]string{}
	chunkHashMap := map[int]string{}

	nameMapVal, _ := b.vm.Run(`try{var m={0:'403',1:'404',2:'Account',3:'ApplicationDetail',4:'ApplicationManagement',5:'AttendanceArea',6:'AttendanceAreas',7:'AttendanceRecord',8:'AttendanceRule',9:'AttendanceRules',10:'AttendanceStatistics',11:'Blacklist',12:'BlacklistDetail',13:'BlacklistGroup',14:'Company',15:'Dashboard',16:'DeviceAlarm',17:'DeviceList',18:'DeviceListNew',19:'DeviceMonitor',20:'FirmwareDetail',21:'FirmwareManagement',22:'License',23:'OpenPlatform',24:'OperationLog',25:'Policies',26:'QrCode',27:'RegisterRecord',28:'Schedule',29:'Schedules',30:'SearchByPicture',31:'ServiceConfig',32:'SignCount',33:'SignRecord',34:'Staff',35:'StaffDetail',36:'StaffGroup',37:'Visitor',38:'VisitorDetail',39:'VisitorGroup'};m}catch(e){{}}`)
	if nameMapVal != nil && nameMapVal.Obj != nil {
		for k, v := range nameMapVal.Obj {
			if v.Type == "string" {
				var id int
				if _, err := fmt.Sscanf(k, "%d", &id); err == nil {
					chunkNameMap[id] = v.Str
				}
			}
		}
	}

	hashMapVal, _ := b.vm.Run(`try{var h={0:'2534b859',1:'4e7eec54',2:'6802a12a',3:'00f507f5',4:'2b629a69',5:'ac9b44ad',6:'41405ada',7:'13f53dfb',8:'12f6a49a',9:'745be122',10:'7c973be6',11:'dfcba99c',12:'f20245df',13:'e580105c',14:'16fa3f98',15:'098ebcb8',16:'bdb426b2',17:'17ec3321',18:'e779e153',19:'e0257215',20:'e0818678',21:'e1a39d0d',22:'6f56d670',23:'1066d4b2',24:'fb96638c',25:'07148ba0',26:'535a5c7f',27:'cdd74cad',28:'420aa3de',29:'b192ddf7',30:'e9751da0',31:'d594b476',32:'9033c82a',33:'fb6a295b',34:'f6ef911a',35:'dcb6e812',36:'fb2e2a86',37:'3185a22d',38:'271f9d7c',39:'fcc541a0'};h}catch(e){{}}`)
	if hashMapVal != nil && hashMapVal.Obj != nil {
		for k, v := range hashMapVal.Obj {
			if v.Type == "string" {
				var id int
				if _, err := fmt.Sscanf(k, "%d", &id); err == nil {
					chunkHashMap[id] = v.Str
				}
			}
		}
	}

	// Check which chunks were requested during index.js execution
	chunkLoadsVal, _ := b.vm.Run("window.__wpkChunkLoads || []")
	var requestedChunks []int
	if chunkLoadsVal != nil && chunkLoadsVal.Arr != nil {
		for _, v := range chunkLoadsVal.Arr {
			if v.Type == "number" {
				requestedChunks = append(requestedChunks, int(v.Num))
			}
		}
	}

	// If no chunks were requested, try to preload all known chunks
	if len(requestedChunks) == 0 {
		b.debugLog("No chunks explicitly requested, preloading all known chunks")
		for id := range chunkNameMap {
			requestedChunks = append(requestedChunks, id)
		}
		sort.Ints(requestedChunks)
	}

	if len(requestedChunks) == 0 {
		b.debugLog("No chunks to load")
		return
	}

	b.debugLog("Loading %d chunks: %v", len(requestedChunks), requestedChunks)

	// Fetch and execute each chunk
	loadedCount := 0
	for _, chunkID := range requestedChunks {
		chunkName, hasName := chunkNameMap[chunkID]
		if !hasName {
			chunkName = fmt.Sprintf("%d", chunkID)
		}
		hash, hasHash := chunkHashMap[chunkID]
		chunkURL := fmt.Sprintf("%s/static/js/%s.chunk.js", baseURL, chunkName)
		if hasHash {
			chunkURL += "?" + hash
		}

		resp, err := b.client.Get(chunkURL)
		if err != nil {
			b.debugLog("Failed to load chunk %d (%s): %v", chunkID, chunkURL, err)
			continue
		}

		b.vm.ResetSteps()
		_, err = b.vm.Run(resp.Body)
		if err != nil {
			b.debugLog("Error executing chunk %d: %v (steps=%d)", chunkID, err, b.vm.GetStepCount())
		} else {
			loadedCount++
			b.debugLog("Chunk %d (%s) loaded (size=%d, steps=%d)", chunkID, chunkName, len(resp.Body), b.vm.GetStepCount())
		}
	}

	b.debugLog("Loaded %d/%d chunks", loadedCount, len(requestedChunks))

	// Inject native RSA module (9M3U) to replace the extremely slow JS BigInteger/RSA
	// implementation. The JS version uses Barrett modular exponentiation which requires
	// 100M+ steps in our interpreter — impractical for any RSA key size.
	b.injectNativeRSAModule()

	// After loading chunks, execute the entry module for the first time.
	// All chunk modules are now registered, so module dependencies should resolve.
	// Set window.language before entry module execution so i18n can find locale files.
	b.vm.Run(`if(!window.language||window.language===null)window.language="zh"`)
	if loadedCount > 0 {
		b.vm.ResetSteps()
		prevMax := b.vm.GetMaxSteps()
		// Try entry module execution with a moderate step limit.
		// The full Vue app initialization can be very expensive (100M+ steps)
		// due to RSA crypto, Vuex reactivity, and Vue Router setup.
		// We use a 5M step budget; if it completes, great; if not, we mount manually.
		b.vm.SetMaxSteps(5_000_000)
		_, err := b.vm.Run("if(window.__wpkE2){window.__wpkE2(0)}")
		b.vm.SetMaxSteps(prevMax)
		if err != nil {
			b.debugLog("Entry module timed out/failed: %v (steps=%d)", err, b.vm.GetStepCount())
			// The timeout left half-loaded modules in the registry. Reset them
			// so they can be re-required individually.
			b.vm.ResetSteps()
			b.vm.SetMaxSteps(5_000_000)
			b.vm.Run(`try {
				var n = window.__wpkN2;
				var unloaded = [];
				for (var k in n) {
					if (n[k] && !n[k].l) unloaded.push(k);
				}
				// Delete half-loaded modules so require() will re-create them
				for (var i = 0; i < unloaded.length; i++) {
					delete n[unloaded[i]];
				}
				window.__wpkResetModules = unloaded;
			} catch(e) {}`)
			b.debugLog("Reset half-loaded modules")
			b.vm.SetMaxSteps(prevMax)
		} else {
			b.debugLog("Entry module executed after chunk loading (steps=%d)", b.vm.GetStepCount())
		}

		// If Vue didn't mount automatically, try to mount it manually.
		// We set a simple placeholder first, then upgrade to Vue rendering
		// via a post-navigate Evaluate call (which has a fresh step counter
		// and avoids issues with outerHTML during the navigate process).
		b.vm.ResetSteps()
		b.vm.SetMaxSteps(10_000_000)
		b.vm.Run(`try {
			var appEl = document.querySelector('#app');
			if (appEl && typeof Vue === 'function') {
				appEl.innerHTML = '<div>SenseLink AIoT Platform</div>';
				window.__vueMountPending = true;
				window.__vueManualMount = true;
			}
		} catch(e) {
			window.__vueMountError = String(e.message || e);
		}`)
		b.vm.SetMaxSteps(prevMax)
	}
}

// loadWebpackChunks fetches and executes Webpack chunk files that were requested
// during the initial bundle execution. Chunks use window.webpackJsonp.push()
// to register their modules with the main bundle's runtime.
func (b *Browser) loadWebpackChunks(bundleSrc string) {
	if b.vm == nil {
		return
	}

	// Check which chunks were requested
	chunkLoadsVal, err := b.vm.Run("window.__wpkChunkLoads || []")
	if err != nil {
		return
	}

	var chunkIDs []int
	if chunkLoadsVal != nil {
		if arr := chunkLoadsVal.Arr; arr != nil {
			for _, v := range arr {
				if v.Type == "number" {
					chunkIDs = append(chunkIDs, int(v.Num))
				}
			}
		}
	}

	if len(chunkIDs) == 0 {
		b.debugLog("No Webpack chunks requested")
		return
	}

	// Build the base URL from the bundle source
	baseURL := b.currentURL
	if idx := strings.LastIndex(bundleSrc, "/static/"); idx >= 0 {
		baseURL = bundleSrc[:idx]
	}

	// Extract the chunk name map and hash map from the index.js code
	// We already have them from the bundle code
	chunkNameMap := map[int]string{}
	chunkHashMap := map[int]string{}

	// Get chunk name map from the VM
	nameMapVal, _ := b.vm.Run("try{var m={0:'403',1:'404',2:'Account',3:'ApplicationDetail',4:'ApplicationManagement',5:'AttendanceArea',6:'AttendanceAreas',7:'AttendanceRecord',8:'AttendanceRule',9:'AttendanceRules',10:'AttendanceStatistics',11:'Blacklist',12:'BlacklistDetail',13:'BlacklistGroup',14:'Company',15:'Dashboard',16:'DeviceAlarm',17:'DeviceList',18:'DeviceListNew',19:'DeviceMonitor',20:'FirmwareDetail',21:'FirmwareManagement',22:'License',23:'OpenPlatform',24:'OperationLog',25:'Policies',26:'QrCode',27:'RegisterRecord',28:'Schedule',29:'Schedules',30:'SearchByPicture',31:'ServiceConfig',32:'SignCount',33:'SignRecord',34:'Staff',35:'StaffDetail',36:'StaffGroup',37:'Visitor',38:'VisitorDetail',39:'VisitorGroup'};m}catch(e){{}}")
	if nameMapVal != nil && nameMapVal.Obj != nil {
		for k, v := range nameMapVal.Obj {
			if v.Type == "string" {
				var id int
				if _, err := fmt.Sscanf(k, "%d", &id); err == nil {
					chunkNameMap[id] = v.Str
				}
			}
		}
	}

	// Get chunk hash map
	hashMapVal, _ := b.vm.Run("try{var h={0:'2534b859',1:'4e7eec54',2:'6802a12a',3:'00f507f5',4:'2b629a69',5:'ac9b44ad',6:'41405ada',7:'13f53dfb',8:'12f6a49a',9:'745be122',10:'7c973be6',11:'dfcba99c',12:'f20245df',13:'e580105c',14:'16fa3f98',15:'098ebcb8',16:'bdb426b2',17:'17ec3321',18:'e779e153',19:'e0257215',20:'e0818678',21:'e1a39d0d',22:'6f56d670',23:'1066d4b2',24:'fb96638c',25:'07148ba0',26:'535a5c7f',27:'cdd74cad',28:'420aa3de',29:'b192ddf7',30:'e9751da0',31:'d594b476',32:'9033c82a',33:'fb6a295b',34:'f6ef911a',35:'dcb6e812',36:'fb2e2a86',37:'3185a22d',38:'271f9d7c',39:'fcc541a0'};h}catch(e){{}}")
	if hashMapVal != nil && hashMapVal.Obj != nil {
		for k, v := range hashMapVal.Obj {
			if v.Type == "string" {
				var id int
				if _, err := fmt.Sscanf(k, "%d", &id); err == nil {
					chunkHashMap[id] = v.Str
				}
			}
		}
	}

	// Fetch and execute each chunk
	for _, chunkID := range chunkIDs {
		chunkName, hasName := chunkNameMap[chunkID]
		if !hasName {
			chunkName = fmt.Sprintf("%d", chunkID)
		}
		hash, hasHash := chunkHashMap[chunkID]
		chunkURL := fmt.Sprintf("%s/static/js/%s.chunk.js", baseURL, chunkName)
		if hasHash {
			chunkURL += "?" + hash
		}

		b.debugLog("Loading Webpack chunk %d: %s", chunkID, chunkURL)

		resp, err := b.client.Get(chunkURL)
		if err != nil {
			b.debugLog("Failed to load chunk %d: %v", chunkID, err)
			continue
		}

		b.vm.ResetSteps()
		_, err = b.vm.Run(resp.Body)
		if err != nil {
			b.debugLog("Error executing chunk %d: %v (steps=%d)", chunkID, err, b.vm.GetStepCount())
		} else {
			b.debugLog("Chunk %d executed (size=%d, steps=%d)", chunkID, len(resp.Body), b.vm.GetStepCount())
		}
	}

	// After loading chunks, check if more chunks were requested and load them too
	// (chunks can depend on other chunks)
	newChunkLoadsVal, _ := b.vm.Run("window.__wpkChunkLoads || []")
	if newChunkLoadsVal != nil && newChunkLoadsVal.Arr != nil {
		var newChunkIDs []int
		for _, v := range newChunkLoadsVal.Arr {
			if v.Type == "number" {
				id := int(v.Num)
				already := false
				for _, existing := range chunkIDs {
					if existing == id {
						already = true
						break
					}
				}
				if !already {
					newChunkIDs = append(newChunkIDs, id)
				}
			}
		}
		// Recursively load new chunks (with depth limit)
		if len(newChunkIDs) > 0 {
			// Clear and set only the new chunks
			b.vm.Run("window.__wpkChunkLoads = []")
			// Temporarily patch by directly setting
			for _, id := range newChunkIDs {
				b.vm.Run(fmt.Sprintf("window.__wpkChunkLoads.push(%d)", id))
			}
			b.loadWebpackChunks(bundleSrc)
		}
	}
}

func (b *Browser) Evaluate(code string) (any, error) {
	if b.vm == nil {
		return nil, nil
	}
	prevSteps := b.vm.GetStepCount()
	prevMax := b.vm.GetMaxSteps()
	b.vm.SetMaxSteps(100_000_000)
	b.vm.SetStepCount(0)
	val, err := b.vm.Run(code)
	b.vm.SetStepCount(prevSteps)
	b.vm.SetMaxSteps(prevMax)
	if err != nil {
		return nil, err
	}
	return jsValueToGo(val), nil
}

func (b *Browser) QuerySelector(selector string) *dom.Node {
	if b.doc == nil {
		return nil
	}
	return dom.QuerySelector(b.doc.Root, selector)
}

func (b *Browser) QuerySelectorAll(selector string) []*dom.Node {
	if b.doc == nil {
		return nil
	}
	return dom.QuerySelectorAll(b.doc.Root, selector)
}

func (b *Browser) GetTitle() string {
	if b.doc == nil {
		return ""
	}
	return b.doc.Title()
}

func (b *Browser) GetHTML() string {
	if b.doc == nil || b.doc.Root == nil {
		return ""
	}
	return b.doc.Root.InnerHTML()
}

func (b *Browser) GetText() string {
	if b.doc == nil || b.doc.Root == nil {
		return ""
	}
	return b.doc.Root.TextContent()
}

func (b *Browser) ScreenshotText(width int) string {
	return renderer.ScreenshotText(b.doc, width)
}

func (b *Browser) ScreenshotTextToFile(path string, width int) error {
	return renderer.ScreenshotTextToFile(b.doc, path, width)
}

func (b *Browser) GetURL() string {
	return b.currentURL
}

func (b *Browser) Document() *dom.Document {
	return b.doc
}

func (b *Browser) VM() *jsengine.VM {
	return b.vm
}

func (b *Browser) Client() *httpclient.Client {
	return b.client
}

func (b *Browser) History() []string {
	return b.history
}

func (b *Browser) Back() error {
	if len(b.history) < 2 {
		return nil
	}
	b.history = b.history[:len(b.history)-1]
	prevURL := b.history[len(b.history)-1]
	return b.Navigate(prevURL)
}

// GetLocalStorage returns the localStorage data as a map
func (b *Browser) GetLocalStorage() map[string]string {
	if b.vm == nil {
		return nil
	}
	return b.vm.LocalStorage
}

// GetSessionStorage returns the sessionStorage data as a map
func (b *Browser) GetSessionStorage() map[string]string {
	if b.vm == nil {
		return nil
	}
	return b.vm.SessionStorage
}

// SetLocalStorageItem sets a key-value pair in localStorage
func (b *Browser) SetLocalStorageItem(key, value string) {
	if b.vm != nil {
		b.vm.LocalStorage[key] = value
	}
}

// SetSessionStorageItem sets a key-value pair in sessionStorage
func (b *Browser) SetSessionStorageItem(key, value string) {
	if b.vm != nil {
		b.vm.SessionStorage[key] = value
	}
}

// GetConsoleOutput returns the console.log output
func (b *Browser) GetConsoleOutput() []string {
	if b.vm == nil {
		return nil
	}
	return b.vm.Output()
}

// SetJsDebug enables or disables JS debug mode.
func (b *Browser) SetJsDebug(jsDebug bool) {
	b.debugLog("SetJsDebug called: %v", jsDebug)
	if b.vm != nil {
		b.vm.SetDebug(jsDebug)
	}
}

// WaitStable waits for the page to become stable by draining pending JavaScript
// timers (setTimeout/setInterval) and allowing Vue/SPA rendering to complete.
// timeoutMs is the maximum time to wait, stableForMs is how long the page must
// remain stable (no new timers) before returning.
func (b *Browser) WaitStable(timeoutMs, stableForMs int) error {
	b.debugLog("WaitStable called: timeoutMs=%d, stableForMs=%d", timeoutMs, stableForMs)
	if b.vm == nil {
		return nil
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	stableDeadline := time.Time{}
	maxRounds := 50 // Safety limit to prevent infinite loops

	for round := 0; round < maxRounds; round++ {
		if time.Now().After(deadline) {
			b.debugLog("WaitStable: timeout reached after %d rounds", round)
			break
		}

		// Brief sleep to allow timers to become due
		time.Sleep(10 * time.Millisecond)

		// Process all pending timers synchronously
		hadTimers := b.vm.HasPendingTimers()
		if hadTimers {
			b.debugLog("WaitStable: processing pending timers (round %d)", round)
			b.vm.RunTimers()
		}

		// Check if new timers were created (e.g., by Vue's nextTick)
		if b.vm.HasPendingTimers() {
			// More timers were scheduled, keep processing
			stableDeadline = time.Time{}
			continue
		}

		// No pending timers - check if we've been stable for long enough
		if stableDeadline.IsZero() {
			stableDeadline = time.Now().Add(time.Duration(stableForMs) * time.Millisecond)
		}

		if time.Now().After(stableDeadline) {
			b.debugLog("WaitStable: page stable after %d rounds", round)
			break
		}
	}

	return nil
}

// WaitStableDefault waits for the page to become stable with default timeouts.
func (b *Browser) WaitStableDefault() error {
	return b.WaitStable(5000, 100)
}

// Abort cancels any running JavaScript execution.
// This can be called from another goroutine to stop a long-running script.
func (b *Browser) Abort() {
	if b.vm != nil {
		b.vm.Abort()
	}
}

// IsAborted returns true if the abort flag has been set.
func (b *Browser) IsAborted() bool {
	if b.vm != nil {
		return b.vm.IsAborted()
	}
	return false
}

// FormField represents a form field found in the page.
type FormField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder"`
	Label       string `json:"label"`
	ID          string `json:"id"`
	Required    bool   `json:"required"`
}

// AnalyzeVueTemplates analyzes JavaScript source code to find Vue.js templates
// and extract form fields. This is useful for SPA pages that render forms dynamically.
func (b *Browser) AnalyzeVueTemplates() []FormField {
	var fields []FormField

	// Collect all JavaScript sources
	var jsSources []string

	// Add inline scripts
	if b.doc != nil {
		scripts := dom.GetElementsByTagName(b.doc.Root, "script")
		for _, script := range scripts {
			src := script.GetAttribute("src")
			if src == "" {
				code := script.TextContent()
				if code != "" {
					jsSources = append(jsSources, code)
				}
			} else {
				// Download external script
				if !strings.HasPrefix(src, "http") {
					baseURL, _ := url.Parse(b.currentURL)
					if strings.HasPrefix(src, "/") {
						src = baseURL.Scheme + "://" + baseURL.Host + src
					} else {
						src = baseURL.Scheme + "://" + baseURL.Host + "/" + src
					}
				}
				resp, err := b.client.Get(src)
				if err == nil && len(resp.Body) <= 2_000_000 { // Limit to 2MB
					jsSources = append(jsSources, resp.Body)
				}
			}
		}
	}

	// Analyze each JavaScript source for Vue templates
	for _, js := range jsSources {
		fields = append(fields, b.extractFormFieldsFromJS(js)...)
	}

	return fields
}

// extractFormFieldsFromJS extracts form fields from JavaScript source code.
func (b *Browser) extractFormFieldsFromJS(js string) []FormField {
	var fields []FormField

	// Look for el-form-item patterns
	// Pattern: <el-form-item prop="fieldName">
	formItemPattern := `el-form-item[^>]*prop=["']([^"']+)["']`
	formItemRegex := regexp.MustCompile(formItemPattern)
	matches := formItemRegex.FindAllStringSubmatch(js, -1)

	for _, match := range matches {
		if len(match) > 1 {
			field := FormField{
				Name: match[1],
				Type: "text",
			}

			// Try to find the associated el-input
			inputPattern := `v-model=["']ruleForm\.` + regexp.QuoteMeta(match[1]) + `["'][^>]*>`
			inputRegex := regexp.MustCompile(inputPattern)
			inputMatch := inputRegex.FindString(js)

			if inputMatch != "" {
				// Extract type
				typePattern := `type=["']([^"']+)["']`
				typeRegex := regexp.MustCompile(typePattern)
				typeMatch := typeRegex.FindStringSubmatch(inputMatch)
				if len(typeMatch) > 1 {
					field.Type = typeMatch[1]
				}

				// Extract placeholder
				placeholderPattern := `:placeholder=["']\$t\(['"]([^'"]+)['"]\)["']|placeholder=["']([^"']+)["']`
				placeholderRegex := regexp.MustCompile(placeholderPattern)
				placeholderMatch := placeholderRegex.FindStringSubmatch(inputMatch)
				if len(placeholderMatch) > 1 {
					if placeholderMatch[1] != "" {
						field.Placeholder = placeholderMatch[1]
					} else if placeholderMatch[2] != "" {
						field.Placeholder = placeholderMatch[2]
					}
				}
			}

			fields = append(fields, field)
		}
	}

	// Also look for v-model patterns directly
	vModelPattern := `v-model=["']([^"']+)["']`
	vModelRegex := regexp.MustCompile(vModelPattern)
	vModelMatches := vModelRegex.FindAllStringSubmatch(js, -1)

	existingFields := make(map[string]bool)
	for _, f := range fields {
		existingFields[f.Name] = true
	}

	for _, match := range vModelMatches {
		if len(match) > 1 {
			fieldName := match[1]
			// Skip if already found or if it's a complex expression
			if existingFields[fieldName] || strings.Contains(fieldName, ".") {
				continue
			}
			// This is a simple v-model, might be a form field
			fields = append(fields, FormField{
				Name: fieldName,
				Type: "text",
			})
			existingFields[fieldName] = true
		}
	}

	return fields
}

func jsValueToGo(v *jsengine.Value) any {
	return jsValueToGoWithVisited(v, make(map[*jsengine.Value]bool))
}

func jsValueToGoWithVisited(v *jsengine.Value, visited map[*jsengine.Value]bool) any {
	if v == nil {
		return nil
	}
	switch v.Type {
	case "undefined", "null":
		return nil
	case "bool":
		return v.Bool
	case "number":
		return v.Num
	case "string":
		return v.Str
	case "object":
		if visited[v] {
			return "[circular]"
		}
		visited[v] = true
		if v.Arr != nil {
			arr := make([]any, len(v.Arr))
			for i, a := range v.Arr {
				arr[i] = jsValueToGoWithVisited(a, visited)
			}
			return arr
		}
		if v.Obj != nil {
			obj := make(map[string]any)
			for k, val := range v.Obj {
				obj[k] = jsValueToGoWithVisited(val, visited)
			}
			return obj
		}
		return nil
	case "function", "native":
		return "[function]"
	}
	return nil
}

// injectNativeRSAModule replaces the 9M3U module (JS BigInteger/RSA implementation)
// with a native Go implementation using math/big and crypto/rsa. The JS version
// uses Barrett modular exponentiation which is too slow for our interpreter (100M+ steps).
// The native module provides the same exports: setMaxDigits (c), RSAKey (a), encryptedString (b).
func (b *Browser) injectNativeRSAModule() {
	if b.vm == nil {
		return
	}

	// Create a native powMod function using math/big
	powModFn := func(args []*jsengine.Value) *jsengine.Value {
		if len(args) < 3 {
			return &jsengine.Value{Type: "undefined"}
		}
		baseStr := jsengine.ValueToString(args[0])
		expStr := jsengine.ValueToString(args[1])
		modStr := jsengine.ValueToString(args[2])

		base := new(big.Int)
		exp := new(big.Int)
		mod := new(big.Int)

		if _, ok := base.SetString(baseStr, 16); !ok {
			return &jsengine.Value{Type: "string", Str: "0"}
		}
		if _, ok := exp.SetString(expStr, 16); !ok {
			return &jsengine.Value{Type: "string", Str: "0"}
		}
		if _, ok := mod.SetString(modStr, 16); !ok {
			return &jsengine.Value{Type: "string", Str: "0"}
		}

		result := new(big.Int).Exp(base, exp, mod)
		return &jsengine.Value{Type: "string", Str: fmt.Sprintf("%x", result)}
	}

	// Register the native powMod as a global function
	b.vm.DefineGlobal("__nativePowMod", &jsengine.Value{
		Type:   "native",
		Native: powModFn,
	})

	// Inject a native RSA module that replaces the 9M3U JS BigInteger implementation.
	// The original 9M3U exports: c=setMaxDigits, a=RSAKey, b=encryptedString
	_, _ = b.vm.Run(`(function(){
		var __rsaKeyCounter = 0;

		function RSAKey(e, d, m, chunkSize) {
			this.e = e;
			this.d = d;
			this.m = m;
			this.chunkSize = chunkSize || 0;
			this.radix = 16;
			this._id = ++__rsaKeyCounter;
		}

		function setMaxDigits(n) {}

		function encryptedString(key, s, pad, encoding) {
			var a = [];
			var sl = s.length;
			var i, j, k;
			var padType = 0;
			if (typeof pad === 'string') {
				if (pad === 'NoPadding') padType = 1;
				else if (pad === 'PKCS1Padding') padType = 2;
			}
			var rawEncoding = (typeof encoding === 'string' && encoding === 'RawEncoding');
			if (padType === 1) { if (sl > key.chunkSize) sl = key.chunkSize; }
			else if (padType === 2) { if (sl > key.chunkSize - 11) sl = key.chunkSize - 11; }
			i = 0; j = padType === 2 ? sl - 1 : key.chunkSize - 1;
			while (i < sl) { if (padType) { a[j] = s.charCodeAt(i); } else { a[i] = s.charCodeAt(i); } i++; j--; }
			if (padType === 1) i = 0;
			var r = key.chunkSize - sl % key.chunkSize;
			while (r > 0) {
				if (padType === 2) { var l = Math.floor(Math.random() * 256); while (!l) l = Math.floor(Math.random() * 256); a[i] = l; }
				else { a[i] = 0; }
				i++; r--;
			}
			if (padType === 2) { a[sl] = 0; a[key.chunkSize - 2] = 2; a[key.chunkSize - 1] = 0; }
			var result = '';
			for (i = 0; i < a.length; i += key.chunkSize) {
				var chunkHex = '';
				for (k = 0; k < key.chunkSize; k += 2) {
					var lo = a[i + k] || 0;
					var hi = a[i + k + 1] || 0;
					chunkHex = ('00' + ((hi << 8 | lo) & 0xFFFF).toString(16)).slice(-4) + chunkHex;
				}
				var encrypted = __nativePowMod(chunkHex, key.e, key.m);
				if (!rawEncoding) {
					var expectedLen = Math.ceil(key.m.length / 4) * 4;
					while (encrypted.length < expectedLen) encrypted = '0' + encrypted;
				}
				result += encrypted;
			}
			return result;
		}

		var n = window.__wpkN2;
		if (n && !n["9M3U"]) {
			var modObj = {i: "9M3U", l: true, exports: {}};
			var exp = modObj.exports;
			Object.defineProperty(exp, 'c', {enumerable: true, get: function(){ return setMaxDigits }});
			Object.defineProperty(exp, 'a', {enumerable: true, get: function(){ return RSAKey }});
			Object.defineProperty(exp, 'b', {enumerable: true, get: function(){ return encryptedString }});
			n["9M3U"] = modObj;
		}
	})()`)
}
