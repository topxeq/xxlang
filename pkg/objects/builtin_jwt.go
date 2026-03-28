// pkg/objects/builtin_jwt.go
// JWT (JSON Web Token) functions for Xxlang
// Note: genJwtToken and parseJwtToken have been moved to crypto module
package objects

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	// JWT builtins removed - use crypto module instead
	// Builtins["genJwtToken"] = &Builtin{Fn: builtinGenJwtToken}
	// Builtins["parseJwtToken"] = &Builtin{Fn: builtinParseJwtToken}
}

// GenJwtToken generates a JWT token from payload and secret.
// Exported for use by crypto module.
func GenJwtToken(payload *Map, secret []byte, withType, base64Secret bool, expireSeconds int64) (string, error) {
	// Decode base64 secret if needed
	if base64Secret {
		decoded, err := base64.RawURLEncoding.DecodeString(string(secret))
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 secret: %v", err)
		}
		secret = decoded
	}

	// Build claims from payload map
	claims := jwt.MapClaims{}
	for _, pair := range payload.Pairs {
		if key, ok := pair.Key.(*String); ok {
			claims[key.Value] = objectToInterface(pair.Value)
		}
	}

	// Add expiration if specified
	if expireSeconds > 0 {
		claims["exp"] = time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix()
	}

	// Add iat (issued at) if not present
	if _, exists := claims["iat"]; !exists {
		claims["iat"] = time.Now().Unix()
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Set header
	if withType {
		token.Header["typ"] = "JWT"
	}

	// Sign token
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

// ParseJwtToken parses and validates a JWT token with secret.
// Exported for use by crypto module.
func ParseJwtToken(tokenStr string, secret []byte, base64Secret, noValidate bool) (map[string]interface{}, error) {
	// Decode base64 secret if needed
	if base64Secret {
		decoded, err := base64.RawURLEncoding.DecodeString(string(secret))
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 secret: %v", err)
		}
		secret = decoded
	}

	if noValidate {
		// Parse without validation
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			return nil, fmt.Errorf("failed to parse JWT token: %v", err)
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return mapClaimsToGoMap(claims), nil
		}
		return nil, fmt.Errorf("failed to extract claims from token")
	}

	// Parse with validation
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse/validate JWT token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return mapClaimsToGoMap(claims), nil
	}

	return nil, fmt.Errorf("failed to extract claims from token")
}

// mapClaimsToGoMap converts jwt.MapClaims to Go map
func mapClaimsToGoMap(claims jwt.MapClaims) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}
	return result
}

// InterfaceToObject converts Go interface{} to Object (exported for crypto module)
func InterfaceToObject(v interface{}) Object {
	if v == nil {
		return NULL
	}

	switch val := v.(type) {
	case bool:
		return &Bool{Value: val}
	case int:
		return NewInt(int64(val))
	case int64:
		return NewInt(val)
	case float64:
		return NewFloat(val)
	case string:
		return NewString(val)
	case json.Number:
		// Try to parse as int first
		if intVal, err := val.Int64(); err == nil {
			return NewInt(intVal)
		}
		if floatVal, err := val.Float64(); err == nil {
			return NewFloat(floatVal)
		}
		return NewString(string(val))
	case []interface{}:
		elements := make([]Object, len(val))
		for i, elem := range val {
			elements[i] = InterfaceToObject(elem)
		}
		return NewArray(elements)
	case map[string]interface{}:
		result := NewMapWithCapacity(len(val))
		for k, v := range val {
			keyObj := NewString(k)
			hashKey := keyObj.HashKey()
			result.Pairs[hashKey] = MapPair{
				Key:   keyObj,
				Value: InterfaceToObject(v),
			}
		}
		return result
	case time.Time:
		return NewString(val.Format(time.RFC3339))
	default:
		// Try to convert via JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return NewString(fmt.Sprintf("%v", v))
		}
		return NewString(string(jsonBytes))
	}
}
