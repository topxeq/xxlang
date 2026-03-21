(module
  ;; Plugin metadata
  (func $plugin_name (export "plugin_name") (param $ptr i32) (result i32)
    (i32.store8 (local.get $ptr) (i32.const 110))  ;; 'n'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 1)) (i32.const 111))  ;; 'o'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 2)) (i32.const 114))  ;; 'r'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 3)) (i32.const 101))  ;; 'e'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 4)) (i32.const 115))  ;; 's'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 5)) (i32.const 117))  ;; 'u'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 6)) (i32.const 108))  ;; 'l'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 7)) (i32.const 116))  ;; 't'
    (i32.const 8)
  )

  (func $plugin_version (export "plugin_version") (param $ptr i32) (result i32)
    (i32.store8 (local.get $ptr) (i32.const 49))  ;; '1'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 1)) (i32.const 46))  ;; '.'
    (i32.store8 (i32.add (local.get $ptr) (i32.const 2)) (i32.const 48))  ;; '0'
    (i32.const 3)
  )

  ;; Function with two arguments but NO return value - this tests the edge case
  (func $call_no_result (export "call_no_result") (param $a i64) (param $b i64)
    ;; This function intentionally returns nothing
    ;; It just exists to test the "no result" error path
    nop
  )

  ;; Memory for the plugin
  (memory (export "memory") 1)
)
