// pkg/stdlib/backup.go
// Backup module for Xxlang - file backup and synchronization functionality.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "backup",
		Exports: map[string]objects.Object{
			// ============================================================
			// Task Creation Functions
			// ============================================================

			// newTask creates a new BackupTask object.
			// Optional: options map with keys: mode, compareStrategy, hashAlgorithm,
			// deleteExtra, conflictPolicy, excludePatterns.
			"newTask": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return Error("newTask takes at most 1 argument (options map)")
				}

				if len(args) == 0 {
					return objects.NewBackupTask()
				}

				// Parse options map
				optsMap, ok := args[0].(*objects.Map)
				if !ok {
					return Error("argument must be a map (options)")
				}

				opts := parseBackupOptions(optsMap)
				return objects.NewBackupTaskWithOptions(opts)
			}),

			// ============================================================
			// Quick Backup Functions
			// ============================================================

			// localToLocal performs a quick local-to-local backup.
			// Arguments: srcDir (string), dstDir (string), opts? (map)
			// Returns: BackupResult
			"localToLocal": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("localToLocal takes at least 2 arguments (srcDir, dstDir)")
				}
				if len(args) > 3 {
					return Error("localToLocal takes at most 3 arguments (srcDir, dstDir, opts)")
				}

				// Parse source directory
				srcDir, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (srcDir)")
				}

				// Parse destination directory
				dstDir, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (dstDir)")
				}

				// Create task with default settings
				task := objects.NewBackupTask()

				// Apply options if provided
				if len(args) == 3 {
					if args[2] != objects.NULL {
						optsMap, ok := args[2].(*objects.Map)
						if !ok {
							return Error("third argument must be a map (opts) or null")
						}
						applyBackupOptions(task, optsMap)
					}
				}

				// Set source and target
				task.SetSourceLocal(srcDir.Value)
				task.SetTargetLocal(dstDir.Value)

				// Execute backup
				return task.Run()
			}),

			// localToRemote performs a quick local-to-remote backup.
			// Arguments: client (SSHClient), localDir (string), remoteDir (string), opts? (map)
			// Returns: BackupResult
			"localToRemote": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 || len(args) > 4 {
					return Error("localToRemote takes 3 or 4 arguments (client, localDir, remoteDir, options?)")
				}
				client, ok := args[0].(*objects.SSHClient)
				if !ok {
					return Error("first argument must be SSHClient")
				}
				localDir, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be string (localDir)")
				}
				remoteDir, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be string (remoteDir)")
				}

				task := objects.NewBackupTask()
				task.SetSourceLocal(localDir.Value)
				task.SetTargetRemote(client, remoteDir.Value)

				if len(args) == 4 {
					if optsMap, ok := args[3].(*objects.Map); ok {
						applyBackupOptions(task, optsMap)
					}
				}

				return task.Run()
			}),

			// remoteToLocal performs a quick remote-to-local backup.
			// Arguments: client (SSHClient), remoteDir (string), localDir (string), opts? (map)
			// Returns: BackupResult
			"remoteToLocal": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 || len(args) > 4 {
					return Error("remoteToLocal takes 3 or 4 arguments (client, remoteDir, localDir, options?)")
				}
				client, ok := args[0].(*objects.SSHClient)
				if !ok {
					return Error("first argument must be SSHClient")
				}
				remoteDir, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be string (remoteDir)")
				}
				localDir, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be string (localDir)")
				}

				task := objects.NewBackupTask()
				task.SetSourceRemote(client, remoteDir.Value)
				task.SetTargetLocal(localDir.Value)

				if len(args) == 4 {
					if optsMap, ok := args[3].(*objects.Map); ok {
						applyBackupOptions(task, optsMap)
					}
				}

				return task.Run()
			}),

			// remoteToRemote performs a quick remote-to-remote backup.
			// Arguments: srcClient (SSHClient), dstClient (SSHClient), srcDir (string), dstDir (string), opts? (map)
			// Returns: BackupResult
			"remoteToRemote": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 4 || len(args) > 5 {
					return Error("remoteToRemote takes 4 or 5 arguments (srcClient, dstClient, srcDir, dstDir, options?)")
				}
				srcClient, ok := args[0].(*objects.SSHClient)
				if !ok {
					return Error("first argument must be SSHClient")
				}
				dstClient, ok := args[1].(*objects.SSHClient)
				if !ok {
					return Error("second argument must be SSHClient")
				}
				srcDir, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be string (srcDir)")
				}
				dstDir, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be string (dstDir)")
				}

				task := objects.NewBackupTask()
				task.SetSourceRemote(srcClient, srcDir.Value)
				task.SetTargetRemote(dstClient, dstDir.Value)

				if len(args) == 5 {
					if optsMap, ok := args[4].(*objects.Map); ok {
						applyBackupOptions(task, optsMap)
					}
				}

				return task.Run()
			}),

			// ============================================================
			// Type Check Functions
			// ============================================================

			// isBackupTask checks if an object is a BackupTask.
			"isBackupTask": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isBackupTask takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.BackupTask)
				return Bool(ok)
			}),

			// isBackupResult checks if an object is a BackupResult.
			"isBackupResult": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isBackupResult takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.BackupResult)
				return Bool(ok)
			}),

			// isBackupProgress checks if an object is a BackupProgress.
			"isBackupProgress": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isBackupProgress takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.BackupProgress)
				return Bool(ok)
			}),
		},
	})
}

// ============================================================
// Helper Functions
// ============================================================

// parseBackupOptions extracts backup options from a Map object.
func parseBackupOptions(m *objects.Map) map[string]interface{} {
	opts := make(map[string]interface{})

	// Helper function to get string value from map
	getString := func(key string) (string, bool) {
		keyObj := objects.NewString(key)
		if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
			if s, ok := pair.Value.(*objects.String); ok {
				return s.Value, true
			}
		}
		return "", false
	}

	// Helper function to get bool value from map
	getBool := func(key string) (bool, bool) {
		keyObj := objects.NewString(key)
		if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
			if b, ok := pair.Value.(*objects.Bool); ok {
				return b.Value, true
			}
		}
		return false, false
	}

	// Helper function to get array of strings from map
	getStringArray := func(key string) ([]string, bool) {
		keyObj := objects.NewString(key)
		if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
			if arr, ok := pair.Value.(*objects.Array); ok {
				result := []string{}
				for _, elem := range arr.Elements {
					if s, ok := elem.(*objects.String); ok {
						result = append(result, s.Value)
					}
				}
				return result, len(result) > 0
			}
		}
		return nil, false
	}

	// Extract mode
	if mode, ok := getString("mode"); ok {
		opts["mode"] = mode
	}

	// Extract compareStrategy
	if strategy, ok := getString("compareStrategy"); ok {
		opts["compareStrategy"] = strategy
	}

	// Extract hashAlgorithm
	if algo, ok := getString("hashAlgorithm"); ok {
		opts["hashAlgorithm"] = algo
	}

	// Extract deleteExtra
	if del, ok := getBool("deleteExtra"); ok {
		opts["deleteExtra"] = del
	}

	// Extract conflictPolicy
	if policy, ok := getString("conflictPolicy"); ok {
		opts["conflictPolicy"] = policy
	}

	// Extract excludePatterns
	if patterns, ok := getStringArray("excludePatterns"); ok {
		opts["excludePatterns"] = patterns
	}

	return opts
}

// applyBackupOptions applies options from a Map to a BackupTask.
func applyBackupOptions(task *objects.BackupTask, m *objects.Map) {
	// Helper function to get string value from map
	getString := func(key string) (string, bool) {
		keyObj := objects.NewString(key)
		if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
			if s, ok := pair.Value.(*objects.String); ok {
				return s.Value, true
			}
		}
		return "", false
	}

	// Helper function to get bool value from map
	getBool := func(key string) (bool, bool) {
		keyObj := objects.NewString(key)
		if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
			if b, ok := pair.Value.(*objects.Bool); ok {
				return b.Value, true
			}
		}
		return false, false
	}

	// Helper function to get array of strings from map
	getStringArray := func(key string) ([]string, bool) {
		keyObj := objects.NewString(key)
		if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
			if arr, ok := pair.Value.(*objects.Array); ok {
				result := []string{}
				for _, elem := range arr.Elements {
					if s, ok := elem.(*objects.String); ok {
						result = append(result, s.Value)
					}
				}
				return result, true
			}
		}
		return nil, false
	}

	// Apply mode
	if mode, ok := getString("mode"); ok {
		task.SetMode(mode)
	}

	// Apply compareStrategy
	if strategy, ok := getString("compareStrategy"); ok {
		task.SetCompareStrategy(strategy)
	}

	// Apply hashAlgorithm
	if algo, ok := getString("hashAlgorithm"); ok {
		task.SetHashAlgorithm(algo)
	}

	// Apply deleteExtra
	if del, ok := getBool("deleteExtra"); ok {
		task.SetDeleteExtra(del)
	}

	// Apply conflictPolicy
	if policy, ok := getString("conflictPolicy"); ok {
		task.SetConflictPolicy(policy)
	}

	// Apply excludePatterns
	if patterns, ok := getStringArray("excludePatterns"); ok {
		task.SetExcludePatterns(patterns)
	}
}