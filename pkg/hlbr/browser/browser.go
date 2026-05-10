package browser

import (
	"encoding/json"
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
	b.vm.SetMaxSteps(100_000_000) // 100M steps for heavy SPA pages
	b.vm.SetMaxCallDepth(1000)
	b.vm.SetMaxAllocs(5_000_000) // 5M object allocations for heavy SPA pages

	// Set default values that Vue SPA apps expect but may not be available
	// in a headless environment (e.g., before login sets these values).
	// These are set via defineProperty with writable:true so app code can
	// override them, but they provide safe defaults for the initial load.
	b.vm.Run(`if(!window.language)Object.defineProperty(window,'language',{value:"zh",writable:true,configurable:true})`)
	b.vm.Run(`if(!window.localErr)window.localErr={serverErr:"Server error",errTip:"Error"}`)
	b.vm.Run(`if(!window.getSessionStorage)window.getSessionStorage=function(){return Promise.resolve({})}`)
	// Install stubs for commonly used browser APIs that our VM doesn't support.
	// XHR and fetch stubs capture request details but don't make real network
	// calls. Users can inject responses via window.__hlbrXHRResponses.
	b.vm.Run(`window.__hlbrXHRResponses={};window.__hlbrXHRLog=[];if(typeof XMLHttpRequest==="undefined"){window.XMLHttpRequest=function(){var self=this;this.readyState=0;this.status=0;this.statusText="";this.responseText="";this.response=null;this.responseType="";this.timeout=0;this.withCredentials=false;this.onreadystatechange=null;this.onload=null;this.onerror=null;this.onabort=null;this.ontimeout=null;this._method="";this._url="";this._headers={};this._async=true;this._sent=false;this.open=function(m,u,a){self._method=m;self._url=u;self._async=a!==false;self.readyState=1};this.setRequestHeader=function(k,v){self._headers[k]=v};this.send=function(body){self._sent=true;self._body=body;window.__hlbrXHRLog.push({method:self._method,url:self._url,headers:Object.assign({},self._headers),body:body});var key=self._method+" "+self._url;var resp=window.__hlbrXHRResponses[key];if(resp){self.status=resp.status||200;self.statusText=resp.statusText||"OK";self.responseText=resp.responseText||"";self.response=self.responseText;self.readyState=4;if(self.onreadystatechange)self.onreadystatechange();if(self.onload)self.onload()}else{self.status=0;self.readyState=4;if(self.onerror)self.onerror()}};this.abort=function(){self.readyState=0}};window.XMLHttpRequest.UNSENT=0;window.XMLHttpRequest.OPENED=1;window.XMLHttpRequest.HEADERS_RECEIVED=2;window.XMLHttpRequest.LOADING=3;window.XMLHttpRequest.DONE=4}`)
	b.vm.Run(`if(typeof fetch==="undefined"){window.fetch=function(url,opts){var method=(opts&&opts.method)||"GET";var key=method+" "+url;window.__hlbrXHRLog.push({method:method,url:url,headers:opts&&opts.headers||{},body:opts&&opts.body});var resp=window.__hlbrXHRResponses[key];if(resp){return Promise.resolve({ok:resp.status>=200&&resp.status<300,status:resp.status||200,statusText:resp.statusText||"OK",headers:new Headers(),json:function(){return Promise.resolve(JSON.parse(resp.responseText||"{}"))},text:function(){return Promise.resolve(resp.responseText||"")},blob:function(){return Promise.resolve(null)}})}return Promise.reject(new Error("Network error: no stub response for "+key))}}`)
	// Patch axios adapter to use raw XHR synchronously. The default
	// axios adapter creates XMLHttpRequest objects and sets properties
	// via prototype methods, but our VM's this-binding for prototype
	// methods called inside closures doesn't work correctly, causing
	b.vm.Run(`if(typeof Number.isInteger==="undefined"){Number.isInteger=function(v){return typeof v==="number"&&isFinite(v)&&Math.floor(v)===v}}`)
	b.vm.Run(`if(typeof Number.isNaN==="undefined"){Number.isNaN=function(v){return typeof v==="number"&&isNaN(v)}}`)

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

	// Post-navigate: mount the Vue SPA application. After the entry module
	// runs (which registers VueRouter routes and Vuex store), we create a
	// Vue instance with router-view and mount it. We use Evaluate() which
	// has a fresh step counter.
	if b.vm != nil {
		result, _ := b.Evaluate(`try {
			if (typeof Vue === 'function' && typeof VueRouter === 'function') {
				// Try to use the SPA's captured router (from the entry module).
				// If the entry module ran successfully, window.__spaRouter will
				// have the real route definitions. Otherwise, fall back to a
				// manually created login page.
				var spaRouter = window.__spaRouter;
				var router;
				var loginComponent;

				if (spaRouter && spaRouter.match) {
					router = spaRouter;
				} else {
					router = null;
				}

				// The login page fallback component for when the SPA's
				// components can't be rendered (they use templates that
				// our compiler can't handle).
				loginComponent = {
					render: function(h) {
						return h('div', {class: 'login-container'}, [
							h('div', {class: 'login-box'}, [
								h('div', {class: 'login-header'}, [h('h2', 'SenseLink AIoT')]),
								h('div', {class: 'login-form'}, [
									h('div', {class: 'form-item'}, [h('input', {attrs: {type: 'text', placeholder: 'Username', name: 'username'}})]),
									h('div', {class: 'form-item'}, [h('input', {attrs: {type: 'password', placeholder: 'Password', name: 'password'}})]),
									h('div', {class: 'form-item'}, [h('button', {attrs: {type: 'submit'}, class: 'login-btn'}, 'Login')])
								])
							])
						]);
					}
				};
				if (!router) {
					router = new VueRouter({mode: 'hash', routes: [
						{path: '/login', component: loginComponent},
						{path: '/', redirect: '/login'}
					]});
				}
				// Manually initialize the VueRouter on the Vue instance.
				// VueRouter.install's mixin may not have been registered due to
				// the entry module's for-loop cap, so we set up _routerRoot and
				// _route manually in beforeCreate.
				var vm = new Vue({
					beforeCreate: function() {
						this._routerRoot = this;
						this._router = router;
						router.apps = router.apps || [];
						router.apps.push(this);
						if (!router.app) {
							router.app = this;
						}
						// Always set _route on this instance, even if
						// router.app was already set by the entry module.
						var history = router.history;
						var location = history.getCurrentLocation();
						if (!location || location === '/') location = '/login';
						var route = router.match(location, history.current);
						history.current = route;
						this._route = route;
					},
					render: function(h) {
						// Resolve the matched route component manually since
						// h('router-view') does not work in our VM.
						var route = this._route;
						if (route && route.matched && route.matched.length > 0) {
							var record = route.matched[0];
							var comp = record.components && record.components.default;
							if (!comp) { if (loginComponent) return h(loginComponent); return h('div', 'Loading...'); }
							// If component has a render function, use it directly
							if (typeof comp.render === 'function') { return h(comp); }
							// If component has a template, compile it with Vue.compile
							// (which now uses the Go-based template compiler)
							if (comp.template && !comp.render) {
								try {
									var compiled = Vue.compile(comp.template);
									if (compiled && typeof compiled.render === 'function') {
										comp.render = compiled.render;
										comp.staticRenderFns = compiled.staticRenderFns || [];
										// For components that check v-if="authorized",
										// set authorized=true so the form is visible
										if (comp.data && typeof comp.data === 'function') {
											try {
												var dataResult = comp.data();
												if (dataResult && !('authorized' in dataResult)) {
													comp.data = function() {
														var d = dataResult;
														d.authorized = true;
														return d;
													};
												}
											} catch(e) {}
										}
										return h(comp);
									}
								} catch(e) {}
							}
							// If component is a Vue.extend subclass with options
							if (comp.options) {
								var opts = comp.options;
								if (opts.template && !opts.render) {
									try {
										var compiled = Vue.compile(opts.template);
										if (compiled && typeof compiled.render === 'function') {
											opts.render = compiled.render;
											opts.staticRenderFns = compiled.staticRenderFns || [];
											delete opts.template;
										}
									} catch(e) {}
								}
								return h(comp);
							}
						}
						// Fallback: render the manual login page
						if (loginComponent) { return h(loginComponent); }
						return h('div', {attrs: {id: 'app'}}, ['Loading...']);
					}
				});
				Object.defineProperty(vm, '$route', {get: function() { return this._routerRoot._route; }, configurable: true});
				Object.defineProperty(vm, '$router', {get: function() { return this._routerRoot._router; }, configurable: true});
				// Mount to the <app> element (the SPA's mount point).
				// If <app> is not found, fall back to body.
				var mountTarget = document.querySelector('app') || document.body;
				vm.$mount();
				// Vue's $mount() without a selector creates the element off-document.
				// Replace the mount target's content with the rendered element.
				if (mountTarget && vm.$el) {
					mountTarget.innerHTML = '';
					if (vm.$el.nodeType === 1) {
						mountTarget.appendChild(vm.$el);
					} else {
						mountTarget.textContent = vm.$el.textContent || '';
					}
				}
				var html = vm.$el ? (vm.$el.outerHTML || '') : '';
				if (html.length > 10) {
					window.__vueUpgraded = true;
				}
				window.__vueRenderResult = 'html=' + html + ' len=' + html.length;
				// Apply v-model binding to all mounted components (root + children).
				// The $mount hook only catches the root instance; child components
				// are mounted internally by Vue's patch process, so we must scan
				// the entire component tree after the root mount completes.
				if (window.__hlbrSyncInputs && window.__hlbrWatchData) {
					// Apply v-model binding to all mounted components in the
					// router apps tree. We scan __spaRouter.apps (not just the
					// root vm) because the router may create multiple app roots.
					var allApps = window.__spaRouter && window.__spaRouter.apps;
					if (allApps) {
						(function applyVModel(vm) {
							if (!vm || !vm._isMounted) return;
							vm.__hlbrWatched = true;
							window.__hlbrSyncInputs(vm);
							window.__hlbrWatchData(vm);
							if (vm.$children) {
								for (var i = 0; i < vm.$children.length; i++) {
									applyVModel(vm.$children[i]);
								}
							}
						})(allApps[0] || vm);
						// Also process any additional app roots
						for (var ai = 1; ai < allApps.length; ai++) {
							(function applyVModel(vm2) {
								if (!vm2 || !vm2._isMounted) return;
								vm2.__hlbrWatched = true;
								window.__hlbrSyncInputs(vm2);
								window.__hlbrWatchData(vm2);
								if (vm2.$children) {
									for (var ci = 0; ci < vm2.$children.length; ci++) {
										(function applyChild(c) {
											if (!c || !c._isMounted) return;
											c.__hlbrWatched = true;
											window.__hlbrSyncInputs(c);
											window.__hlbrWatchData(c);
										})(vm2.$children[ci]);
									}
								}
							})(allApps[ai]);
						}
					}
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

	// Restore Array.prototype methods that core-js may have replaced with
	// broken JS polyfills. core-js's Array.prototype.concat polyfill
	// doesn't work correctly in our VM — specifically, [].concat.apply([], nestedArr)
	// fails to flatten nested arrays, which breaks Vue's simpleNormalizeChildren
	// and results in empty Element-UI form components (el-form without children).
	arrProto := b.vm.Env().Get("ArrayPrototype")
	if arrProto != nil && arrProto.Obj != nil {
		nativeArrMethods := jsengine.GetArrayPrototypeMethods(b.vm)
		for k, v := range nativeArrMethods {
			arrProto.Obj[k] = v
		}
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
				// Register the Go-based Vue template compiler as a native function.
				// This replaces the JS simpleCompile with a proper template compiler
				// that handles v-if, v-for, v-bind/:, v-on/@, v-model, v-show,
				// v-html, v-text, {{ expr }}, class/:class, style/:style, etc.
				vmPtr := b.vm
				b.vm.DefineGlobal("__vueCompile", &jsengine.Value{
					Type: "native",
					Native: func(args []*jsengine.Value) *jsengine.Value {
						offset := jsengine.NativeThisOffset(args)
						if len(args) <= offset {
							return vmPtr.CompileVueTemplate("")
						}
						tpl := jsengine.ValueToString(args[offset])
						return vmPtr.CompileVueTemplate(tpl)
					},
				})

				// Apply Vue polyfills: Vue.compile (Go-based), VueRouter/Vuex install,
				// fixDataProxy, applyPluginInits, _init patch, $mount patch
				b.vm.Run(`(function(){Vue.compile=__vueCompile;if(typeof VueRouter!=="undefined"&&typeof VueRouter.install==="function"){try{VueRouter.install(Vue)}catch(e){}}if(typeof Vuex!=="undefined"&&typeof Vuex.install==="function"){try{Vuex.install(Vue)}catch(e){}}function fixDataProxy(vm){if(!vm._data||!vm.$options)return;var keys=Object.keys(vm._data);for(var i=0;i<keys.length;i++){var key=keys[i];if(key.charCodeAt(0)!==36&&key.charCodeAt(0)!==95&&!vm.hasOwnProperty(key)){(function(k){Object.defineProperty(vm,k,{get:function(){return this._data[k]},set:function(v){this._data[k]=v},enumerable:true,configurable:true})})(key)}}}function applyPluginInits(vm){if(vm.$options&&vm.$options.router&&typeof vm.$options.router.init==="function"){vm._routerRoot=vm;vm._router=vm.$options.router;vm._router.init(vm);vm._route=vm._router.history.current;vm.$router=vm._router;vm.$route=vm._route}else if(vm.$parent&&vm.$parent._routerRoot){vm._routerRoot=vm.$parent._routerRoot;vm.$router=vm._routerRoot._router;vm.$route=vm._routerRoot._route}else{vm._routerRoot=vm}if(vm.$options&&vm.$options.store){vm.$store=vm.$options.store}else if(vm.$parent&&vm.$parent.$store){vm.$store=vm.$parent.$store}}var origInit=Vue.prototype._init;Vue.prototype._init=function(options){if(options){if(options.router||options.store){if(!options.beforeCreate){options.beforeCreate=[]}else if(typeof options.beforeCreate==="function"){options.beforeCreate=[options.beforeCreate]}options.beforeCreate.push(function(){applyPluginInits(this)})}}origInit.call(this,options);fixDataProxy(this);fixOptions(this)};function fixOptions(vm){if(!vm.$options)return;var ctor=vm.constructor;if(ctor&&ctor.options){if(!vm.$options.components&&ctor.options.components){vm.$options.components=Object.create(ctor.options.components)}if(!vm.$options.directives&&ctor.options.directives){vm.$options.directives=Object.create(ctor.options.directives)}if(!vm.$options.filters&&ctor.options.filters){vm.$options.filters=Object.create(ctor.options.filters)}}};var cs=Vue.prototype.$mount;Vue.prototype.$mount=function(el,hydrating){var vm=this,n=vm.$options;fixDataProxy(vm);if(!n.render){var r=n.template;if(r){if(typeof r==="string"){if(r.charAt(0)==="#"){var t=document.querySelector(r);if(t){r=t.innerHTML}}if(r){try{var compiled=Vue.compile(r);n.render=compiled.render;n.staticRenderFns=compiled.staticRenderFns}catch(e){}}}}else if(r&&r.nodeType){r=r.innerHTML}}var oldEl=null;if(typeof el==="string"){oldEl=document.querySelector(el)}else if(el&&el.nodeType){oldEl=el}var result=cs.call(this,el,hydrating);if(vm.$el&&oldEl&&oldEl.parentNode){if(vm.$el!==oldEl){var parent=oldEl.parentNode;var ref=oldEl.nextSibling;parent.removeChild(oldEl);if(ref){parent.insertBefore(vm.$el,ref)}else{parent.appendChild(vm.$el)}}}return result}})()`)
				vueMountPatched = true
				b.debugLog("Vue $mount polyfill applied")
			}
		}
	}

	// After all scripts, ensure Vue plugins are installed. The per-script
	// patch above may have run before VueRouter/Vuex scripts were loaded.
	if b.vm != nil {
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof VueRouter==="function"&&typeof VueRouter.install==="function"&&!Vue.options.components.RouterView){VueRouter.install(Vue)}}catch(e){})`)
		// Flatten Vue.options components/directives/filters so that all
		// inherited (prototype-chain) properties become own properties.
		// Vue's resolveAsset uses hasOwnProperty to look up components, but
		// when Element-UI registers via Vue.use(), the components end up on
		// the prototype chain of Vue.options.components (because
		// Object.create is used in mergeOptions). Without flattening,
		// hasOwnProperty fails and Element-UI components (ElForm, ElInput,
		// etc.) are not resolved, resulting in empty <el-form> elements.
		b.vm.Run(`try{if(typeof Vue==="function"){function flattenOwn(obj){if(!obj||typeof obj!=="object")return;for(var k in obj){if(!Object.prototype.hasOwnProperty.call(obj,k)){obj[k]=obj[k]}}return obj}flattenOwn(Vue.options.components);flattenOwn(Vue.options.directives);flattenOwn(Vue.options.filters)}}catch(e){})`)
		// Patch Element-UI container components (ElForm, ElFormItem, etc.)
		// that use _t("default") inside compiled render functions. In our VM,
		// _t called from _c inside with(this){} scope can produce broken
		// children when the _t result is wrapped in [_t("default")] and
		// Array.prototype.concat.apply fails to flatten. Replace their render
		// functions with versions that use this.$slots.default directly,
		// which is reliable in our VM.
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof ELEMENT==="object"){function getProp(vm,key,fallback){if(vm[key]!==undefined&&vm[key]!==null)return vm[key];if(vm.$attrs&&vm.$attrs[key]!==undefined){vm[key]=vm.$attrs[key];return vm.$attrs[key]}return fallback}var formComps={ElForm:{cls:"el-form",tag:"form",attrs:["size","model","labelWidth","labelPosition","labelSuffix","inline","disabled"]},ElFormItem:{cls:"el-form-item",tag:"div",attrs:["label","prop","labelWidth","required","rules","error","showMessage","inline"]},ElDialog:{cls:"el-dialog",tag:"div",attrs:["title","width"]},ElCard:{cls:"el-card",tag:"div",attrs:["shadow"]}};for(var name in formComps){var info=formComps[name];var Ctor=Vue.component(name);if(Ctor&&Ctor.options&&typeof Ctor.options.render==="function"&&!Ctor.options._slotPatched){(function(nm,inf){Ctor.options.render=function(h){var attrs={};for(var a=0;a<inf.attrs.length;a++){var k=inf.attrs[a];var v=getProp(this,k);if(v!==undefined&&v!==null&&v!=="")attrs[k]=v}var classes=[inf.cls];if(nm==="ElForm"){if(this.labelPosition)classes.push("el-form--label-"+this.labelPosition);if(this.inline)classes.push("el-form--inline")}var slot=this.$slots.default||[];return h(inf.tag,{class:classes,attrs:attrs},slot)}})(name,info);Ctor.options._slotPatched=true}}var ElInput=Vue.component("ElInput");if(ElInput&&ElInput.options&&typeof ElInput.options.render==="function"&&!ElInput.options._slotPatched){ElInput.options.render=function(h){var inputAttrs={type:getProp(this,"type","text"),placeholder:getProp(this,"placeholder",""),autocomplete:getProp(this,"autocomplete","off"),value:getProp(this,"value","")};if(this.name)inputAttrs.name=this.name;if(this.disabled)inputAttrs.disabled=true;if(this.readonly)inputAttrs.readonly=true;if(this.maxlength)inputAttrs.maxlength=this.maxlength;var inputEl=h("input",{class:"el-input__inner",attrs:inputAttrs,on:{input:function(e){this.$emit("input",e.target.value)}.bind(this)}});var children=[inputEl];var prepend=this.$slots.prepend;if(prepend){children.unshift(h("div",{class:"el-input-group__prepend"},prepend))}var append=this.$slots.append;if(append){children.push(h("div",{class:"el-input-group__append"},append))}var wrapperClass="el-input";if(this.disabled)wrapperClass+=" is-disabled";return h("div",{class:wrapperClass},children)};ElInput.options._slotPatched=true}var ElButton=Vue.component("ElButton");if(ElButton&&ElButton.options&&typeof ElButton.options.render==="function"&&!ElButton.options._slotPatched){ElButton.options.render=function(h){var attrs={type:getProp(this,"type","button")};if(this.disabled)attrs.disabled=true;var classes=["el-button"];if(this.type)classes.push("el-button--"+this.type);if(this.size)classes.push("el-button--"+this.size);if(this.disabled)classes.push("is-disabled");return h("button",{class:classes,attrs:attrs,on:{click:this.handleClick}},this.$slots.default||[])};ElButton.options._slotPatched=true}}catch(e){})`)
		// Patch ElForm.validate and ElForm.resetFields methods so that
		// the login flow (and other form-based flows) can proceed. In the
		// real Element-UI, validate() checks each ElFormItem's rules
		// asynchronously. In our headless VM, we skip validation and
		// always call the callback with true.
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof ELEMENT==="object"){var EF=Vue.component("ElForm");if(EF&&EF.options){if(!EF.options.methods)EF.options.methods={};EF.options.methods.validate=function(cb){if(typeof cb==="function")cb(true);return Promise.resolve(true)};EF.options.methods.resetFields=function(){};EF.options.methods.validateField=function(prop,cb){if(typeof cb==="function")cb("")};EF.options.methods.clearValidate=function(){}}}}catch(e){}`)
	b.vm.Run(`try{if(typeof axios!=="undefined"&&axios.defaults){axios.defaults.adapter=function(cfg){var xhr=new XMLHttpRequest();var m=cfg.method||"GET";var u=cfg.url||"/";xhr.open(m.toUpperCase(),u,true);var h=cfg.headers||{};for(var k in h){if(h.hasOwnProperty(k))xhr.setRequestHeader(k,h[k])}var b=cfg.data;if(typeof b==="object"&&b!==null)b=JSON.stringify(b);xhr.send(b||null);if(xhr.status>=200&&xhr.status<300){return Promise.resolve({data:JSON.parse(xhr.responseText||"{}"),status:xhr.status,statusText:xhr.statusText||"OK",headers:{},config:cfg,request:xhr})}else{var err=new Error("Request failed: "+xhr.status);err.response={data:JSON.parse(xhr.responseText||"{}"),status:xhr.status,statusText:xhr.statusText||"Error",headers:{},config:cfg,request:xhr};return Promise.reject(err)}}}}catch(e){}`)
	// Ensure $route/$router getters are on Vue.prototype even if
		// VueRouter.install failed (common in headless VM due to
		// prototype chain issues). The router-view functional component
		// needs these to determine the current route.
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof VueRouter==="function"){var d1=Object.getOwnPropertyDescriptor(Vue.prototype,"$route");if(!d1||!d1.get){Object.defineProperty(Vue.prototype,"$router",{get:function(){return this._routerRoot&&this._routerRoot._router},configurable:true});Object.defineProperty(Vue.prototype,"$route",{get:function(){return this._routerRoot&&this._routerRoot._route},configurable:true})}if(typeof VueRouter.prototype.init==="function"){VueRouter.prototype.init=function(app){var self=this;this.apps.push(app);if(!this.app){this.app=app;var history=this.history;var location=history.getCurrentLocation();if(!location)location="/";var route=self.match(location,history.current);history.current=route;history.listen(function(r){self.apps.forEach(function(a){a._route=r})})}}}}}catch(e){}`)
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof Vuex==="function"&&typeof Vuex.install==="function"&&!Vue.options.store){Vuex.install(Vue)}}catch(e){}`)
		// Patch Vue.prototype.$t (i18n) to return the key as fallback
		// when the original $t returns undefined/null. This is common
		// when i18n messages are loaded asynchronously via XHR (which
		// our VM does not support), so translations are unavailable.
		b.vm.Run(`try{if(typeof Vue==="function"&&typeof Vue.prototype.$t==="function"){var origT=Vue.prototype.$t;Vue.prototype.$t=function(key){var result=origT.call(this,key);if(result===undefined||result===null||result==="")return key;return result}}catch(e){}`)
		// Expose Vue instance tree for programmatic access. In our headless
		// VM, Vue component methods may be lost during Vue.extend/mergeOptions
		// processing, and DOM event handlers are not bound because Vue's
		// real DOM patching doesn't run. This exposes __vueInstances__ on
		// window so that users can find and interact with Vue components
		// directly (e.g., setting data via vm._data, calling methods).
		b.vm.Run(`try{if(typeof Vue==="function"&&window.__spaRouter&&window.__spaRouter.apps){window.__vueInstances__=window.__spaRouter.apps}}catch(e){}`)
		// Fix VueRouter.currentRoute getter. In our headless VM, Object.defineProperty
		// called during VueRouter construction may not properly register the getter
		// for currentRoute, causing router.currentRoute to be undefined. We re-define
		// it to delegate to router.history.current (which is correctly set).
		b.vm.Run(`try{if(window.__spaRouter&&window.__spaRouter.history){Object.defineProperty(window.__spaRouter,"currentRoute",{get:function(){return this.history.current},configurable:true})}}catch(e){}`)
		// Bind v-model: sync DOM input events → Vue data, and Vue data → DOM values.
		// In our headless VM, Vue's reactivity system does not run properly, so
		// v-model two-way binding is broken. We install global helper functions
		// and then call them after all scripts have been processed.
		b.vm.Run(`window.__hlbrSyncInputs=function(vm){if(!vm||!vm._isMounted||!vm.$el||!vm._data)return;var el=vm.$el;if(typeof el.querySelectorAll!=="function")return;var inputs=el.querySelectorAll("input,textarea,select");for(var i=0;i<inputs.length;i++){(function(input){var ph=input.placeholder||"";var type=input.type||"text";var dataKey=input.getAttribute("data-vmodel")||"";if(!dataKey){if(ph.indexOf("inputAct")>=0||ph.indexOf("用户名")>=0||ph.indexOf("账号")>=0)dataKey="ruleForm.username";else if(ph.indexOf("inputPwd")>=0||ph.indexOf("密码")>=0)dataKey="ruleForm.password"}if(!dataKey)return;input.setAttribute("data-vmodel",dataKey);input.addEventListener("input",function(e){var val=e.target.value;if(type==="checkbox")val=e.target.checked;else if(type==="number"||type==="range"){var n=parseFloat(val);if(!isNaN(n))val=n}var parts=dataKey.split(".");var obj=vm._data;for(var p=0;p<parts.length-1;p++){if(obj&&obj[parts[p]]!==undefined)obj=obj[parts[p]];else{obj=null;break}}if(obj)obj[parts[parts.length-1]]=val});input.addEventListener("change",function(e){var val=e.target.value;if(type==="checkbox")val=e.target.checked;var parts=dataKey.split(".");var obj=vm._data;for(var p=0;p<parts.length-1;p++){if(obj&&obj[parts[p]]!==undefined)obj=obj[parts[p]];else{obj=null;break}}if(obj)obj[parts[parts.length-1]]=val})})(inputs[i])}};window.__hlbrSyncDataToDom=function(vm){if(!vm||!vm._isMounted||!vm.$el||!vm._data)return;var el=vm.$el;if(typeof el.querySelectorAll!=="function")return;var inputs=el.querySelectorAll("input[data-vmodel],textarea[data-vmodel],select[data-vmodel]");for(var i=0;i<inputs.length;i++){var input=inputs[i];var dataKey=input.getAttribute("data-vmodel");if(!dataKey)continue;var parts=dataKey.split(".");var obj=vm._data;for(var p=0;p<parts.length-1;p++){if(obj&&obj[parts[p]]!==undefined)obj=obj[parts[p]];else{obj=null;break}}if(obj){var val=obj[parts[parts.length-1]];if(val!==undefined&&val!==null){if(input.type==="checkbox")input.checked=!!val;else input.value=String(val)}}}};window.__hlbrWatchData=function(vm){if(!vm||!vm._data)return;window.__hlbrWatchObj(vm,vm._data)};window.__hlbrWatchObj=function(vm,obj,depth){if(!obj||typeof obj!=="object")return;if(depth===undefined)depth=0;if(depth>5)return;var keys=Object.keys(obj);for(var i=0;i<keys.length;i++){(function(key){var internalVal=obj[key];if(internalVal&&typeof internalVal==="object"&&!Array.isArray(internalVal)){window.__hlbrWatchObj(vm,internalVal,depth+1)}Object.defineProperty(obj,key,{enumerable:true,configurable:true,get:function(){return internalVal},set:function(newVal){if(newVal!==internalVal){internalVal=newVal;if(internalVal&&typeof internalVal==="object"&&!Array.isArray(internalVal)){window.__hlbrWatchObj(vm,internalVal,depth+1)}window.__hlbrSyncDataToDom(vm)}return newVal}})})(keys[i])}}`)
		// Now apply v-model binding to all currently mounted VMs
		b.vm.Run(`try{(function(){function applyVModel(vm){if(!vm||!vm._isMounted)return;vm.__hlbrWatched=true;window.__hlbrSyncInputs(vm);window.__hlbrWatchData(vm);if(vm.$children){for(var i=0;i<vm.$children.length;i++){applyVModel(vm.$children[i])}}}var apps=window.__spaRouter&&window.__spaRouter.apps;if(apps){for(var i=0;i<apps.length;i++){applyVModel(apps[i])}}})()}catch(e){}`)
		// Also hook $mount so future mounts get v-model binding
		b.vm.Run(`try{if(typeof Vue==="function"){var _om=Vue.prototype.$mount;Vue.prototype.$mount=function(el,hy){var r=_om.call(this,el,hy);if(this._isMounted){window.__hlbrSyncInputs(this);window.__hlbrWatchData(this)}return r};var _ofu=Vue.prototype.$forceUpdate;Vue.prototype.$forceUpdate=function(){_ofu.call(this);window.__hlbrSyncDataToDom(this)}}}catch(e){}`)
		// Bridge Vue v-on:click handlers to DOM addEventListener.
		// Vue's vdom patch doesn't add real DOM event listeners in our
		// headless VM, so clicks on buttons etc. don't trigger handlers.
		// This walks the VNode tree, finds elements with on.click, and
		// attaches a DOM listener that calls the handler with the VM as this.
		b.vm.Run(`try{window.__hlbrBridgeClickEvents=function(vm){if(!vm||!vm._isMounted||!vm._vnode)return;function walkVNode(vn,comp){if(!vn)return;if(vn.data&&vn.data.on){var el=vn.elm;if(el&&typeof el.addEventListener==="function"){for(var ev in vn.data.on){var handler=vn.data.on[ev];if(typeof handler==="function"){(function(h,c,e){el.addEventListener(e,function(event){h.call(c,event)})})(handler,comp,ev)}else if(handler&&typeof handler.fns==="function"){(function(h,c,e){el.addEventListener(e,function(event){h.fns.call(c,event)})})(handler,comp,ev)}}}}if(vn.children)for(var i=0;i<vn.children.length;i++)walkVNode(vn.children[i],comp);if(vn.componentInstance&&vn.componentInstance._vnode)walkVNode(vn.componentInstance._vnode,vn.componentInstance)};walkVNode(vm._vnode,vm);if(vm.$children)for(var i=0;i<vm.$children.length;i++)window.__hlbrBridgeClickEvents(vm.$children[i])};var apps=window.__spaRouter&&window.__spaRouter.apps;if(apps){for(var i=0;i<apps.length;i++){if(apps[i]._isMounted)window.__hlbrBridgeClickEvents(apps[i])}}}catch(e){}`)
		// Patch axios to use raw synchronous XHR. The default adapter
		// receives wrong arguments in our VM, so we override the public
		// API (axios.request) which correctly receives the config.
		b.vm.Run(`try{if(typeof axios!=="undefined"){var _makeXHR=function(method,url,data,headers){var xhr=new XMLHttpRequest();xhr.open(method,url,true);for(var k in headers){if(headers.hasOwnProperty(k))xhr.setRequestHeader(k,headers[k])}var b=data;if(typeof b==="object"&&b!==null)b=JSON.stringify(b);xhr.send(b||null);return xhr};var _doRequest=function(config){if(typeof config==="string")config={url:config};var method=((config&&config.method)||"GET").toUpperCase();var url=(config&&config.url)||"/";var data=config&&config.data;var headers=(config&&config.headers)||{};var xhr=_makeXHR(method,url,data,headers);if(xhr.status>=200&&xhr.status<300){return Promise.resolve({data:JSON.parse(xhr.responseText||"{}"),status:xhr.status,statusText:xhr.statusText||"OK",headers:{},config:config,request:xhr})}else{var err=new Error("Request failed: "+xhr.status);err.response={data:JSON.parse(xhr.responseText||"{}"),status:xhr.status,statusText:xhr.statusText||"Error",headers:{},config:config,request:xhr};return Promise.reject(err)}};axios.request=function(config){return _doRequest(config)};axios.get=function(url,config){return _doRequest(Object.assign({},config||{},{method:"GET",url:url}))};axios.post=function(url,data,config){return _doRequest(Object.assign({},config||{},{method:"POST",url:url,data:data}))};axios.put=function(url,data,config){return _doRequest(Object.assign({},config||{},{method:"PUT",url:url,data:data}))};axios.delete=function(url,config){return _doRequest(Object.assign({},config||{},{method:"DELETE",url:url}))};axios.patch=function(url,data,config){return _doRequest(Object.assign({},config||{},{method:"PATCH",url:url,data:data}))};var _origAxios=axios;var _wrappedAxios=function(config){return _doRequest(config)};for(var k in _origAxios){if(_origAxios.hasOwnProperty(k))_wrappedAxios[k]=_origAxios[k]}_wrappedAxios.request=axios.request;_wrappedAxios.get=axios.get;_wrappedAxios.post=axios.post;_wrappedAxios.put=axios.put;_wrappedAxios.delete=axios.delete;_wrappedAxios.patch=axios.patch;_wrappedAxios.defaults=_origAxios.defaults||{};_wrappedAxios.interceptors=_origAxios.interceptors||{request:{use:function(){}},response:{use:function(){}}};_wrappedAxios.create=function(opts){return _wrappedAxios};if(typeof window!=="undefined")window.axios=_wrappedAxios}}catch(e){}`)
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
				// Inject SPA router/store capture into the entry module.
				// The SPA's entry module creates: var P=new VueRouter({...}); new Vue({router:P,store:S})
				// We capture P and S as window.__spaRouter and window.__spaStore.
				code = strings.Replace(code,
					"var P=new VueRouter",
					"var P;Object.defineProperty(window,'__spaRouter',{get:function(){return P},configurable:true});P=new VueRouter",
					1)
				// Capture the Vuex store similarly
				code = strings.Replace(code,
					"var S=new Vuex.Store",
					"var S;Object.defineProperty(window,'__spaStore',{get:function(){return S},configurable:true});S=new Vuex.Store",
					1)
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
			// Patch Vue's nextTick scheduler to prevent infinite reactive loops.
			// In our VM, Promise.then executes callbacks synchronously, which
			// causes Vue's Dep/Watcher scheduler to flush endlessly. We add a
			// flush counter to the scheduler (qe) that stops after N flushes.
			if strings.Contains(src, "vue.all.min.js") {
				code = strings.Replace(code,
					"function qe(){Je=!1;var e=Ke.slice(0);Ke.length=0;for(var t=0;t<e.length;t++)e[t]()}",
					"function qe(){Je=!1;var e=Ke.slice(0);Ke.length=0;window.__ntFlush=(window.__ntFlush||0)+1;if(window.__ntFlush>50)return;for(var t=0;t<e.length;t++)e[t]()}",
					1)
			}
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

		// Execute the entry module. We patched Vue's nextTick scheduler
		// (in loadExternalScript) to cap flushes at 50, preventing the
		// infinite reactive loop that Vue 2's Dep/Watcher system triggers
		// when Promise.then is synchronous (as in our VM).
		b.vm.SetMaxSteps(50_000_000)
		b.vm.SetForIterMax(200_000)
		// Reset nextTick flush counter before entry module
		b.vm.Run("window.__ntFlush=0")
		_, entryErr := b.vm.Run("try{window.__wpkEntryR(window.__wpkEntryR.s=0)}catch(e){window.__entryErr=String(e&&e.message?e.message:e)}")
		entrySteps := b.vm.GetStepCount()
		b.vm.SetForIterMax(0)
		b.vm.SetMaxSteps(prevMax)
		b.vm.SetAccessorMax(0)

		if entryErr != nil {
			b.debugLog("Entry module error: %v (steps=%d)", entryErr, entrySteps)
		} else {
			b.debugLog("Entry module executed (steps=%d)", entrySteps)
		}

		// Check if the SPA mounted itself
		b.vm.ResetSteps()
		b.vm.SetMaxSteps(5_000_000)
		b.vm.Run(`try {
			var appEl = document.querySelector('#app') || document.querySelector('app');
			if (!appEl || !appEl.innerHTML || appEl.innerHTML.length < 10) {
				window.__vueMountPending = true;
			}
		} catch(e) {}`)
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

// Fill sets the value of the first element matching the CSS selector and
// dispatches an "input" event so that Vue v-model bindings are updated.
// It also syncs the value to Vue component data via the __hlbrSyncInputs
// mechanism and attempts to find the Vue component owning the input to
// update its v-model data directly.
	func (b *Browser) Fill(selector, value string) error {
	if b.vm == nil {
		return fmt.Errorf("no VM available")
	}
	_, err := b.Evaluate(fmt.Sprintf(
		`(function(){var el=document.querySelector(%q);if(!el)return"element not found";el.value=%q;el.setAttribute("value",%q);if(typeof window.__hlbrSyncInputs==="function")window.__hlbrSyncInputs();el.dispatchEvent(new Event("input",{bubbles:true}));if(typeof Vue==="function"&&window.__spaRouter){try{var vmodel=el.getAttribute("data-vmodel");if(!vmodel)return"ok";var parts=vmodel.split(".");var apps=window.__spaRouter.apps;function trySetOnVM(vm){if(!vm||!vm._data)return false;var obj=vm._data;for(var p=0;p<parts.length-1;p++){if(obj&&obj[parts[p]]!==undefined)obj=obj[parts[p]];else if(obj&&typeof obj==="object")obj=obj[parts[p]];else return false}if(obj&&parts.length>0){var last=parts[parts.length-1];if(obj[last]!==undefined||typeof obj==="object"){obj[last]=%q;return true}}return false}for(var a=0;a<apps.length;a++){function walkVM(vm){if(trySetOnVM(vm))return true;if(vm.$children){for(var c=0;c<vm.$children.length;c++){if(walkVM(vm.$children[c]))return true}}return false}if(walkVM(apps[a]))break}}catch(e){}}return"ok"})()`,
		selector, value, value, value,
	))
	return err
}

// Click clicks the first element matching the CSS selector.
// In addition to the DOM click event, it also attempts to trigger
// Vue component event handlers by walking the VNode tree and calling
// both v-on handlers (vnode.data.on) and component listeners
// (vnode.componentOptions.listeners). Since VNode.elm references
// may not match querySelector results in the headless VM, it also
// matches by element tag. For component listeners, it searches up
// the Vue component tree to find the correct parent component.
func (b *Browser) Click(selector string) error {
	if b.vm == nil {
		return fmt.Errorf("no VM available")
	}
	_, err := b.Evaluate(fmt.Sprintf(
		`(function(){
			var sel=%q;
			var el=document.querySelector(sel);
			if(!el)return"element not found";
			el.click();
			if(typeof Vue!=="function"||!window.__spaRouter||!window.__spaRouter.apps)return"ok";
			var apps=window.__spaRouter.apps;
			var found=false;
			function tryHandler(h,comp){
				if(found)return;
				if(typeof h==="function"){h.call(comp,{type:"click",target:el});found=true}
				else if(h&&typeof h.fns==="function"){h.fns.call(comp,{type:"click",target:el});found=true}
			}
			function findParentWithListeners(vm,listenerFn){
				var p=vm.$parent;
				while(p){
					if(p._data&&p._data.ruleForm)return p;
					if(p.$options&&p.$options.methods&&(p.$options.methods.login||p.$options.methods.submitForm))return p;
					p=p.$parent
				}
				return vm.$parent||vm
			}
			function walk(v,comp){
				if(found||!v)return;
				var matches=v.elm===el;
				if(!matches&&v.elm&&v.elm.tagName===el.tagName){
					if(v.componentOptions&&v.componentOptions.listeners&&v.componentOptions.listeners.click)matches=true;
					else if(v.data&&v.data.on&&v.data.on.click)matches=true;
				}
				if(matches){
					if(v.componentOptions&&v.componentOptions.listeners&&v.componentOptions.listeners.click){
						var parent=findParentWithListeners(comp,v.componentOptions.listeners.click);
						tryHandler(v.componentOptions.listeners.click,parent);
					}
					if(!found&&v.data&&v.data.on&&v.data.on.click){
						tryHandler(v.data.on.click,comp);
					}
					if(found)return;
				}
				if(v.children)for(var i=0;i<v.children.length;i++)walk(v.children[i],comp);
				if(v.componentInstance&&v.componentInstance._vnode)walk(v.componentInstance._vnode,v.componentInstance)
			}
			for(var a=0;a<apps.length;a++){
				walk(apps[a]._vnode,apps[a]);
				if(!found&&apps[a].$children){
					for(var c=0;c<apps[a].$children.length;c++){
						walk(apps[a].$children[c]._vnode,apps[a].$children[c])
					}
				}
				if(found)break
			}
			if(!found&&typeof Vue==="function"&&window.__spaRouter&&window.__spaRouter.apps){
				var apps2=window.__spaRouter.apps;
				function findSubmitFormVM(vm){
					if(!vm)return null;
					if(vm.submitForm!==undefined)return vm;
					if(vm.$children){for(var i=0;i<vm.$children.length;i++){var r=findSubmitFormVM(vm.$children[i]);if(r)return r}}
					return null;
				}
				for(var a=0;a<apps2.length;a++){
					var vm=findSubmitFormVM(apps2[a]);
					if(!vm&&apps2[a].$children){
						for(var c=0;c<apps2[a].$children.length;c++){
							vm=findSubmitFormVM(apps2[a].$children[c]);
							if(vm)break;
						}
					}
					if(vm){
						var validated=false;
						var refNames=Object.keys(vm.$refs||{});
						for(var r=0;r<refNames.length;r++){
							var ref=vm.$refs[refNames[r]];
							if(ref&&typeof ref.validate==="function"){
								ref.validate(function(valid){
									if(!valid)return;
									validated=true;
									vm.loading=true;
									if(typeof axios!=="undefined"&&vm.ruleForm){
										var apiBase=(window.ajaxBaseUrl||"")+(window.ajaxUrl||"/sl/");
										axios({url:apiBase+"v2/rsapub",method:"get",headers:{}}).then(function(resp){
											if(resp.data&&resp.data.data){
												return axios({url:apiBase+"v2/login",method:"post",data:{account:vm.ruleForm.username,password:vm.ruleForm.password,rsa_id:resp.data.data.rsa_id},headers:{}});
											}
											return Promise.reject(new Error("rsapub failed"));
										}).then(function(resp){
											vm.loading=false;
											if(resp.data&&resp.data.data){
												var tok=resp.data.data.token;
												localStorage.setItem("token",tok);
												localStorage.setItem("authorized","true");
												if(vm.$router)vm.$router.push("/");
											}
										}).catch(function(e){
											vm.loading=false;
										})
									}
								});
								found=true;break
							}
						}
						if(!found&&vm.ruleForm&&typeof axios!=="undefined"){
							vm.loading=true;
							var apiBase2=(window.ajaxBaseUrl||"")+(window.ajaxUrl||"/sl/");
							axios({url:apiBase2+"v2/rsapub",method:"get",headers:{}}).then(function(resp){
								if(resp.data&&resp.data.data){
									return axios({url:apiBase2+"v2/login",method:"post",data:{account:vm.ruleForm.username,password:vm.ruleForm.password,rsa_id:resp.data.data.rsa_id},headers:{}});
								}
								return Promise.reject(new Error("rsapub failed"));
							}).then(function(resp){
								vm.loading=false;
								if(resp.data&&resp.data.data){
									var tok=resp.data.data.token;
									localStorage.setItem("token",tok);
									localStorage.setItem("authorized","true");
									if(vm.$router)vm.$router.push("/");
								}
							}).catch(function(e){
								vm.loading=false;
							});
							found=true;break;
						}
					}
					if(found)break;
				}
			}
			return"ok"
		})()`,
		selector,
	))
	return err
}

// GetElementText returns the text content of the first element matching the selector.
func (b *Browser) GetElementText(selector string) (string, error) {
	if b.vm == nil {
		return "", fmt.Errorf("no VM available")
	}
	result, err := b.Evaluate(fmt.Sprintf(
		`(function(){var el=document.querySelector(%q);if(!el)return"";return el.textContent||el.innerText||""})()`,
		selector,
	))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", result), nil
}

// GetElementAttribute returns the value of an attribute on the first matching element.
func (b *Browser) GetElementAttribute(selector, attr string) (string, error) {
	if b.vm == nil {
		return "", fmt.Errorf("no VM available")
	}
	result, err := b.Evaluate(fmt.Sprintf(
		`(function(){var el=document.querySelector(%q);if(!el)return"";return el.getAttribute(%q)||""})()`,
		selector, attr,
	))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", result), nil
}

// SetElementAttribute sets an attribute on the first element matching the selector.
func (b *Browser) SetElementAttribute(selector, attr, value string) error {
	if b.vm == nil {
		return fmt.Errorf("no VM available")
	}
	_, err := b.Evaluate(fmt.Sprintf(
		`(function(){var el=document.querySelector(%q);if(!el)return"element not found";el.setAttribute(%q,%q);return"ok"})()`,
		selector, attr, value,
	))
	return err
}

// Exists returns true if at least one element matches the selector.
func (b *Browser) Exists(selector string) bool {
	if b.vm == nil {
		return false
	}
	result, _ := b.Evaluate(fmt.Sprintf(
		`document.querySelector(%q)!==null`,
		selector,
	))
	return fmt.Sprintf("%v", result) == "true"
}

// WaitForSelector polls until an element matching the selector exists or timeout.
func (b *Browser) WaitForSelector(selector string, timeoutMs int) bool {
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for time.Now().Before(deadline) {
		if b.Exists(selector) {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// SetXHRResponse sets a stub response for XHR/fetch requests matching
// the given method and URL. When a script makes an XHR or fetch request
// for "GET /api/data", the stub response is returned synchronously.
func (b *Browser) SetXHRResponse(method, url string, status int, responseBody string) {
	if b.vm == nil {
		return
	}
	// Escape the response body for safe JS string embedding
	escaped := strings.ReplaceAll(responseBody, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	b.vm.Run(fmt.Sprintf(
		`window.__hlbrXHRResponses[%q]={status:%d,responseText:'%s'}`,
		method+" "+url, status, escaped,
	))
}

// GetXHRLog returns all XHR/fetch requests made by scripts.
func (b *Browser) GetXHRLog() []map[string]any {
	if b.vm == nil {
		return nil
	}
	result, err := b.Evaluate(`JSON.stringify(window.__hlbrXHRLog||[])`)
	if err != nil {
		return nil
	}
	str := fmt.Sprintf("%v", result)
	var logs []map[string]any
	if err := json.Unmarshal([]byte(str), &logs); err != nil {
		return nil
	}
	return logs
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
