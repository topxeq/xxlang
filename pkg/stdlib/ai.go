// pkg/stdlib/ai.go
// AI module for Xxlang - OpenAI-compatible API access with Coding Plan support.
package stdlib

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

// aiHTTPClient is the shared HTTP client for AI API requests with a long timeout.
var aiHTTPClient = &http.Client{
	Timeout: 120 * time.Second,
}

// codePlanSystemPrompt is the system prompt template for the codePlan function.
// %s #1 = file contents block, %s #2 = user's task prompt.
const codePlanSystemPrompt = `You are a coding assistant integrated into the Xxlang programming language. Your task is to analyze the provided source files and generate a structured implementation plan.

## Source Files

%s

## Your Task

%s

## Response Format

You MUST respond with a valid JSON object (no markdown, no code fences, just raw JSON) with this exact structure:

{
    "summary": "A brief summary of what changes will be made",
    "steps": [
        {
            "file": "path/to/file.ext",
            "action": "create",
            "description": "What this step does",
            "code": "The complete file content for create/modify actions"
        },
        {
            "file": "path/to/existing.ext",
            "action": "modify",
            "description": "What changes are made",
            "code": "The complete modified file content"
        },
        {
            "file": "path/to/remove.ext",
            "action": "delete",
            "description": "Why this file should be deleted",
            "code": ""
        }
    ]
}

Rules:
- The "action" field must be one of: "create", "modify", "delete".
- For "create" and "modify", the "code" field must contain the COMPLETE file content.
- For "delete", the "code" field should be empty.
- Include ALL files that need to change, even if only small parts change.
- Make sure all code is syntactically correct.
- Provide only the JSON response, no additional text.`

// ============================================================
// Map value extraction helpers
// ============================================================

// aiGetMapString extracts a string value from an Xxlang Map by key.
func aiGetMapString(m *objects.Map, key string) string {
	pair, ok := m.Pairs[objects.NewString(key).HashKey()]
	if !ok {
		return ""
	}
	if s, ok := pair.Value.(*objects.String); ok {
		return s.Value
	}
	return ""
}

// aiGetMapBool extracts a boolean value from an Xxlang Map by key.
func aiGetMapBool(m *objects.Map, key string) bool {
	pair, ok := m.Pairs[objects.NewString(key).HashKey()]
	if !ok {
		return false
	}
	if b, ok := pair.Value.(*objects.Bool); ok {
		return b.Value
	}
	return false
}

// aiGetMapFloat extracts a float64 value from an Xxlang Map by key.
// Also accepts Int values, converting them to float64.
func aiGetMapFloat(m *objects.Map, key string) float64 {
	pair, ok := m.Pairs[objects.NewString(key).HashKey()]
	if !ok {
		return 0
	}
	if f, ok := pair.Value.(*objects.Float); ok {
		return f.Value
	}
	if i, ok := pair.Value.(*objects.Int); ok {
		return float64(i.Value)
	}
	return 0
}

// aiGetMapInt extracts an int64 value from an Xxlang Map by key.
func aiGetMapInt(m *objects.Map, key string) int64 {
	pair, ok := m.Pairs[objects.NewString(key).HashKey()]
	if !ok {
		return 0
	}
	if i, ok := pair.Value.(*objects.Int); ok {
		return i.Value
	}
	return 0
}

// aiGetMapArray extracts an Array value from an Xxlang Map by key.
func aiGetMapArray(m *objects.Map, key string) (*objects.Array, bool) {
	pair, ok := m.Pairs[objects.NewString(key).HashKey()]
	if !ok {
		return nil, false
	}
	arr, ok := pair.Value.(*objects.Array)
	return arr, ok
}

// ============================================================
// AI request helpers
// ============================================================

// buildMessagesArray converts an Xxlang Array of {role, content} maps
// to a Go slice suitable for JSON marshaling into OpenAI messages format.
func buildMessagesArray(arr *objects.Array) ([]map[string]interface{}, error) {
	messages := make([]map[string]interface{}, 0, len(arr.Elements))
	for _, elem := range arr.Elements {
		msgMap, ok := elem.(*objects.Map)
		if !ok {
			return nil, fmt.Errorf("each message must be a map")
		}
		role := aiGetMapString(msgMap, "role")
		content := aiGetMapString(msgMap, "content")
		if role == "" {
			return nil, fmt.Errorf("each message must have a 'role' field")
		}
		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": content,
		})
	}
	return messages, nil
}

// doAIRequest constructs and sends an HTTP request to an OpenAI-compatible API.
func doAIRequest(baseUrl, apiKey, endpoint, method string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := strings.TrimRight(baseUrl, "/") + endpoint
	req, err := http.NewRequest(method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return aiHTTPClient.Do(req)
}

// parseAIResponse reads the full HTTP response body and unmarshals JSON.
func parseAIResponse(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	return result, nil
}

// handleStreamResponse reads an SSE stream from the API, prints content deltas
// in real-time, and returns the assembled response.
func handleStreamResponse(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var contentBuilder strings.Builder
	var model string
	var finishReason string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and SSE comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Only process data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// End of stream
		if data == "[DONE]" {
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // skip malformed chunks
		}

		// Extract model name from first chunk
		if m, ok := chunk["model"].(string); ok && m != "" {
			model = m
		}

		// Extract content delta
		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}

		if delta, ok := choice["delta"].(map[string]interface{}); ok {
			if content, ok := delta["content"].(string); ok {
				fmt.Print(content)
				contentBuilder.WriteString(content)
			}
		}

		if fr, ok := choice["finish_reason"].(string); ok && fr != "" {
			finishReason = fr
		}
	}

	fmt.Println() // final newline after streaming output

	// Build response in standard format
	result := map[string]interface{}{
		"content":      contentBuilder.String(),
		"model":        model,
		"finishReason": finishReason,
	}

	return result, nil
}

// extractChatResult converts a parsed API response map to an Xxlang OrderedMap.
func extractChatResult(data map[string]interface{}) objects.Object {
	result := objects.NewOrderedMap()

	// Extract content from choices[0].message.content
	content := ""
	finishReason := ""
	modelName := ""
	usage := map[string]interface{}{}

	if m, ok := data["model"].(string); ok {
		modelName = m
	}

	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if c, ok := msg["content"].(string); ok {
					content = c
				}
			}
			if fr, ok := choice["finish_reason"].(string); ok {
				finishReason = fr
			}
		}
	}

	if u, ok := data["usage"].(map[string]interface{}); ok {
		usage = u
	}

	result.Set(objects.NewString("content"), objects.NewString(content))
	result.Set(objects.NewString("model"), objects.NewString(modelName))
	result.Set(objects.NewString("finishReason"), objects.NewString(finishReason))

	// Build usage map
	usageMap := objects.NewOrderedMap()
	if pt, ok := usage["prompt_tokens"]; ok {
		usageMap.Set(objects.NewString("promptTokens"), objects.GoValueToObject(pt))
	}
	if ct, ok := usage["completion_tokens"]; ok {
		usageMap.Set(objects.NewString("completionTokens"), objects.GoValueToObject(ct))
	}
	if tt, ok := usage["total_tokens"]; ok {
		usageMap.Set(objects.NewString("totalTokens"), objects.GoValueToObject(tt))
	}
	result.Set(objects.NewString("usage"), usageMap)

	return result
}

// extractJSONFromResponse attempts to extract JSON from an LLM response string.
// Handles raw JSON and JSON wrapped in markdown code fences.
func extractJSONFromResponse(content string) (map[string]interface{}, error) {
	content = strings.TrimSpace(content)

	// Try direct parse first
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return result, nil
	}

	// Try extracting from markdown code fences
	// Look for ```json ... ``` or ``` ... ```
	for _, delimiter := range []string{"```json", "```"} {
		if idx := strings.Index(content, delimiter); idx != -1 {
			start := idx + len(delimiter)
			end := strings.Index(content[start:], "```")
			if end != -1 {
				extracted := strings.TrimSpace(content[start : start+end])
				if err := json.Unmarshal([]byte(extracted), &result); err == nil {
					return result, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("failed to parse JSON from response")
}

// buildFileContentsBlock constructs the file contents section for codePlan prompt.
func buildFileContentsBlock(files []string) (string, error) {
	var sb strings.Builder
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file '%s': %w", path, err)
		}
		sb.WriteString(fmt.Sprintf("--- FILE: %s ---\n", path))
		sb.WriteString(string(data))
		sb.WriteString("\n--- END FILE ---\n\n")
	}
	return sb.String(), nil
}

// ============================================================
// Module registration
// ============================================================

func init() {
	Register(&Module{
		Name: "ai",
		Exports: map[string]objects.Object{

			// chat sends a chat completion request to an OpenAI-compatible API.
			// Config map: model, messages, baseUrl, apiKey, temperature, maxTokens, stream.
			"chat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("ai.chat() requires 1 argument (config map)")
				}
				config, ok := args[0].(*objects.Map)
				if !ok {
					return Error("ai.chat() argument must be a map")
				}

				// Extract config fields
				model := aiGetMapString(config, "model")
				if model == "" {
					model = "gpt-3.5-turbo"
				}
				baseUrl := aiGetMapString(config, "baseUrl")
				if baseUrl == "" {
					baseUrl = "https://api.openai.com/v1"
				}
				apiKey := aiGetMapString(config, "apiKey")
				temperature := aiGetMapFloat(config, "temperature")
				maxTokens := aiGetMapInt(config, "maxTokens")
				stream := aiGetMapBool(config, "stream")

				// Build messages array
				messagesArr, ok := aiGetMapArray(config, "messages")
				if !ok {
					return Error("ai.chat() requires 'messages' field (array of {role, content})")
				}
				messages, err := buildMessagesArray(messagesArr)
				if err != nil {
					return Error("ai.chat(): " + err.Error())
				}

				// Build request body
				reqBody := map[string]interface{}{
					"model":    model,
					"messages": messages,
				}
				if temperature > 0 {
					reqBody["temperature"] = temperature
				}
				if maxTokens > 0 {
					reqBody["max_tokens"] = maxTokens
				}
				if stream {
					reqBody["stream"] = true
				}

				// Send request
				resp, err := doAIRequest(baseUrl, apiKey, "/chat/completions", "POST", reqBody)
				if err != nil {
					return Error("ai.chat() request failed: " + err.Error())
				}

				// Handle response
				if stream {
					data, err := handleStreamResponse(resp)
					if err != nil {
						return Error("ai.chat() stream error: " + err.Error())
					}
					return extractChatResult(data)
				}

				data, err := parseAIResponse(resp)
				if err != nil {
					return Error("ai.chat() response error: " + err.Error())
				}
				return extractChatResult(data)
			}),

			// complete sends a simple text completion request.
			// Config map: prompt, model, baseUrl, apiKey, temperature, maxTokens, stream.
			"complete": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("ai.complete() requires 1 argument (config map)")
				}
				config, ok := args[0].(*objects.Map)
				if !ok {
					return Error("ai.complete() argument must be a map")
				}

				prompt := aiGetMapString(config, "prompt")
				if prompt == "" {
					return Error("ai.complete() requires 'prompt' field")
				}

				// Build a messages array from the prompt
				model := aiGetMapString(config, "model")
				if model == "" {
					model = "gpt-3.5-turbo"
				}
				baseUrl := aiGetMapString(config, "baseUrl")
				if baseUrl == "" {
					baseUrl = "https://api.openai.com/v1"
				}
				apiKey := aiGetMapString(config, "apiKey")
				temperature := aiGetMapFloat(config, "temperature")
				maxTokens := aiGetMapInt(config, "maxTokens")
				stream := aiGetMapBool(config, "stream")

				messages := []map[string]interface{}{
					{"role": "user", "content": prompt},
				}

				reqBody := map[string]interface{}{
					"model":    model,
					"messages": messages,
				}
				if temperature > 0 {
					reqBody["temperature"] = temperature
				}
				if maxTokens > 0 {
					reqBody["max_tokens"] = maxTokens
				}
				if stream {
					reqBody["stream"] = true
				}

				resp, err := doAIRequest(baseUrl, apiKey, "/chat/completions", "POST", reqBody)
				if err != nil {
					return Error("ai.complete() request failed: " + err.Error())
				}

				if stream {
					data, err := handleStreamResponse(resp)
					if err != nil {
						return Error("ai.complete() stream error: " + err.Error())
					}
					return extractChatResult(data)
				}

				data, err := parseAIResponse(resp)
				if err != nil {
					return Error("ai.complete() response error: " + err.Error())
				}
				return extractChatResult(data)
			}),

			// embed sends a text embedding request.
			// Config map: input, model, baseUrl, apiKey.
			"embed": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("ai.embed() requires 1 argument (config map)")
				}
				config, ok := args[0].(*objects.Map)
				if !ok {
					return Error("ai.embed() argument must be a map")
				}

				inputStr := aiGetMapString(config, "input")
				if inputStr == "" {
					return Error("ai.embed() requires 'input' field")
				}

				model := aiGetMapString(config, "model")
				if model == "" {
					model = "text-embedding-ada-002"
				}
				baseUrl := aiGetMapString(config, "baseUrl")
				if baseUrl == "" {
					baseUrl = "https://api.openai.com/v1"
				}
				apiKey := aiGetMapString(config, "apiKey")

				reqBody := map[string]interface{}{
					"model": model,
					"input": inputStr,
				}

				resp, err := doAIRequest(baseUrl, apiKey, "/embeddings", "POST", reqBody)
				if err != nil {
					return Error("ai.embed() request failed: " + err.Error())
				}

				data, err := parseAIResponse(resp)
				if err != nil {
					return Error("ai.embed() response error: " + err.Error())
				}

				// Extract embedding from response: data[0].embedding
				result := objects.NewOrderedMap()
				result.Set(objects.NewString("model"), objects.NewString(model))

				if dataArr, ok := data["data"].([]interface{}); ok && len(dataArr) > 0 {
					if firstItem, ok := dataArr[0].(map[string]interface{}); ok {
						if embedding, ok := firstItem["embedding"].([]interface{}); ok {
							elements := make([]objects.Object, len(embedding))
							for i, v := range embedding {
								if f, ok := v.(float64); ok {
									elements[i] = objects.NewFloat(f)
								} else {
									elements[i] = objects.NewFloat(0)
								}
							}
							result.Set(objects.NewString("embedding"), objects.NewArray(elements))
						}
					}
				}

				// Extract usage
				if usage, ok := data["usage"].(map[string]interface{}); ok {
					usageMap := objects.NewOrderedMap()
					if pt, ok := usage["prompt_tokens"]; ok {
						usageMap.Set(objects.NewString("promptTokens"), objects.GoValueToObject(pt))
					}
					if tt, ok := usage["total_tokens"]; ok {
						usageMap.Set(objects.NewString("totalTokens"), objects.GoValueToObject(tt))
					}
					result.Set(objects.NewString("usage"), usageMap)
				}

				return result
			}),

			// codePlan analyzes source files and generates an implementation plan using AI.
			// Config map: files (array of paths), prompt (instruction), model, baseUrl, apiKey.
			"codePlan": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("ai.codePlan() requires 1 argument (config map)")
				}
				config, ok := args[0].(*objects.Map)
				if !ok {
					return Error("ai.codePlan() argument must be a map")
				}

				// Extract file paths
				filesArr, ok := aiGetMapArray(config, "files")
				if !ok || len(filesArr.Elements) == 0 {
					return Error("ai.codePlan() requires 'files' field (array of file paths)")
				}

				prompt := aiGetMapString(config, "prompt")
				if prompt == "" {
					return Error("ai.codePlan() requires 'prompt' field")
				}

				var filePaths []string
				for _, elem := range filesArr.Elements {
					s, ok := elem.(*objects.String)
					if !ok {
						return Error("ai.codePlan() file paths must be strings")
					}
					filePaths = append(filePaths, s.Value)
				}

				// Read all files and build the prompt
				filesBlock, err := buildFileContentsBlock(filePaths)
				if err != nil {
					return Error("ai.codePlan(): " + err.Error())
				}

				systemPrompt := fmt.Sprintf(codePlanSystemPrompt, filesBlock, prompt)

				// API config
				model := aiGetMapString(config, "model")
				if model == "" {
					model = "gpt-3.5-turbo"
				}
				baseUrl := aiGetMapString(config, "baseUrl")
				if baseUrl == "" {
					baseUrl = "https://api.openai.com/v1"
				}
				apiKey := aiGetMapString(config, "apiKey")

				messages := []map[string]interface{}{
					{"role": "system", "content": systemPrompt},
					{"role": "user", "content": prompt},
				}

				reqBody := map[string]interface{}{
					"model":    model,
					"messages": messages,
					"temperature": 0.3, // Lower temperature for more structured output
				}

				resp, err := doAIRequest(baseUrl, apiKey, "/chat/completions", "POST", reqBody)
				if err != nil {
					return Error("ai.codePlan() request failed: " + err.Error())
				}

				data, err := parseAIResponse(resp)
				if err != nil {
					return Error("ai.codePlan() response error: " + err.Error())
				}

				// Extract content from the chat response
				content := ""
				if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if msg, ok := choice["message"].(map[string]interface{}); ok {
							if c, ok := msg["content"].(string); ok {
								content = c
							}
						}
					}
				}

				if content == "" {
					return Error("ai.codePlan() received empty response from AI")
				}

				// Parse the JSON plan from the response
				planData, err := extractJSONFromResponse(content)
				if err != nil {
					return Error("ai.codePlan() failed to parse plan JSON: " + err.Error())
				}

				// Build the result as an Xxlang object
				result := objects.GoValueToObject(planData)
				return result
			}),

			// applyPlan applies a code plan (from codePlan) by creating, modifying, or deleting files.
			// Argument: the plan map returned by codePlan.
			"applyPlan": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("ai.applyPlan() requires 1 argument (plan map)")
				}
				planMap, ok := args[0].(*objects.Map)
				if !ok {
					// Also accept OrderedMap
					if om, ok := args[0].(*objects.OrderedMap); ok {
						// Convert OrderedMap to regular Map for extraction
						planMap = om.ToMap()
					} else {
						return Error("ai.applyPlan() argument must be a map")
					}
				}

				// Extract steps array
				stepsObj, ok := aiGetMapArray(planMap, "steps")
				if !ok {
					return Error("ai.applyPlan() plan must have 'steps' array")
				}

				applied := 0
				var errors []string

				for i, elem := range stepsObj.Elements {
					step, ok := elem.(*objects.Map)
					if !ok {
						errors = append(errors, fmt.Sprintf("step %d: not a map", i))
						continue
					}

					file := aiGetMapString(step, "file")
					action := aiGetMapString(step, "action")
					code := aiGetMapString(step, "code")

					if file == "" {
						errors = append(errors, fmt.Sprintf("step %d: missing 'file' field", i))
						continue
					}
					if action == "" {
						errors = append(errors, fmt.Sprintf("step %d: missing 'action' field", i))
						continue
					}

					switch action {
					case "create", "modify":
						dir := filepath.Dir(file)
						if dir != "." && dir != "" {
							if err := os.MkdirAll(dir, 0755); err != nil {
								errors = append(errors, fmt.Sprintf("step %d: failed to create directory '%s': %s", i, dir, err))
								continue
							}
						}
						if err := os.WriteFile(file, []byte(code), 0644); err != nil {
							errors = append(errors, fmt.Sprintf("step %d: failed to write file '%s': %s", i, file, err))
							continue
						}
						applied++
						fmt.Printf("  [%s] %s\n", strings.ToUpper(action), file)

					case "delete":
						if err := os.Remove(file); err != nil {
							errors = append(errors, fmt.Sprintf("step %d: failed to delete file '%s': %s", i, file, err))
							continue
						}
						applied++
						fmt.Printf("  [DELETE] %s\n", file)

					default:
						errors = append(errors, fmt.Sprintf("step %d: unknown action '%s'", i, action))
					}
				}

				// Build result
				result := objects.NewOrderedMap()
				result.Set(objects.NewString("applied"), objects.NewInt(int64(applied)))

				errorElements := make([]objects.Object, len(errors))
				for i, e := range errors {
					errorElements[i] = objects.NewString(e)
				}
				result.Set(objects.NewString("errors"), objects.NewArray(errorElements))

				return result
			}),
		},
	})
}
