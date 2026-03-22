// pkg/objects/bigint.go
// BigInt represents an arbitrary-precision integer value
package objects

import (
	"fmt"
	"hash/fnv"
	"math/big"
)

// BigInt represents an arbitrary-precision integer value
// It wraps Go's math/big.Int for unlimited precision integer arithmetic
type BigInt struct {
	Value *big.Int
}

// NewBigInt creates a new BigInt from a big.Int
func NewBigInt(val *big.Int) *BigInt {
	if val == nil {
		return &BigInt{Value: big.NewInt(0)}
	}
	return &BigInt{Value: new(big.Int).Set(val)}
}

// NewBigIntFromInt64 creates a new BigInt from an int64
func NewBigIntFromInt64(n int64) *BigInt {
	return &BigInt{Value: big.NewInt(n)}
}

// NewBigIntFromString creates a new BigInt from a string representation
// The string can be in decimal, hex (0x prefix), octal (0 prefix), or binary (0b prefix)
func NewBigIntFromString(s string) (*BigInt, error) {
	val := new(big.Int)
	_, success := val.SetString(s, 0)
	if !success {
		return nil, fmt.Errorf("invalid big integer format: %s", s)
	}
	return &BigInt{Value: val}, nil
}

// NewBigIntFromInt creates a new BigInt from an Int object
func NewBigIntFromInt(i *Int) *BigInt {
	return &BigInt{Value: big.NewInt(i.Value)}
}

// Type returns the object type
func (b *BigInt) Type() ObjectType { return BigIntType }

// TypeTag returns the type tag for fast type checking
func (b *BigInt) TypeTag() TypeTag { return TagBigInt }

// Inspect returns the string representation (without the 'n' suffix)
func (b *BigInt) Inspect() string {
	return b.Value.String()
}

// ToBool converts the BigInt to a boolean
func (b *BigInt) ToBool() *Bool {
	if b.Value.Sign() == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (b *BigInt) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(b.Value.String()))
	return HashKey{Type: BigIntType, Value: h.Sum64()}
}

// Clone creates a copy of the BigInt
func (b *BigInt) Clone() *BigInt {
	return &BigInt{Value: new(big.Int).Set(b.Value)}
}

// Cmp compares two BigInt values
// Returns -1 if b < other, 0 if b == other, 1 if b > other
func (b *BigInt) Cmp(other *BigInt) int {
	return b.Value.Cmp(other.Value)
}

// Add returns a new BigInt that is the sum of b and other
func (b *BigInt) Add(other *BigInt) *BigInt {
	result := new(big.Int)
	result.Add(b.Value, other.Value)
	return &BigInt{Value: result}
}

// AddInt returns a new BigInt that is the sum of b and an int64
func (b *BigInt) AddInt(n int64) *BigInt {
	result := new(big.Int)
	result.Add(b.Value, big.NewInt(n))
	return &BigInt{Value: result}
}

// Sub returns a new BigInt that is b minus other
func (b *BigInt) Sub(other *BigInt) *BigInt {
	result := new(big.Int)
	result.Sub(b.Value, other.Value)
	return &BigInt{Value: result}
}

// SubInt returns a new BigInt that is b minus an int64
func (b *BigInt) SubInt(n int64) *BigInt {
	result := new(big.Int)
	result.Sub(b.Value, big.NewInt(n))
	return &BigInt{Value: result}
}

// Mul returns a new BigInt that is b multiplied by other
func (b *BigInt) Mul(other *BigInt) *BigInt {
	result := new(big.Int)
	result.Mul(b.Value, other.Value)
	return &BigInt{Value: result}
}

// MulInt returns a new BigInt that is b multiplied by an int64
func (b *BigInt) MulInt(n int64) *BigInt {
	result := new(big.Int)
	result.Mul(b.Value, big.NewInt(n))
	return &BigInt{Value: result}
}

// Div returns a new BigInt that is b divided by other
// Returns nil if other is zero
func (b *BigInt) Div(other *BigInt) *BigInt {
	if other.Value.Sign() == 0 {
		return nil
	}
	result := new(big.Int)
	result.Quo(b.Value, other.Value)
	return &BigInt{Value: result}
}

// DivInt returns a new BigInt that is b divided by an int64
// Returns nil if n is zero
func (b *BigInt) DivInt(n int64) *BigInt {
	if n == 0 {
		return nil
	}
	result := new(big.Int)
	result.Quo(b.Value, big.NewInt(n))
	return &BigInt{Value: result}
}

// Mod returns a new BigInt that is b mod other
// Returns nil if other is zero
func (b *BigInt) Mod(other *BigInt) *BigInt {
	if other.Value.Sign() == 0 {
		return nil
	}
	result := new(big.Int)
	result.Mod(b.Value, other.Value)
	return &BigInt{Value: result}
}

// ModInt returns a new BigInt that is b mod an int64
// Returns nil if n is zero
func (b *BigInt) ModInt(n int64) *BigInt {
	if n == 0 {
		return nil
	}
	result := new(big.Int)
	result.Mod(b.Value, big.NewInt(n))
	return &BigInt{Value: result}
}

// Neg returns a new BigInt that is the negation of b
func (b *BigInt) Neg() *BigInt {
	result := new(big.Int)
	result.Neg(b.Value)
	return &BigInt{Value: result}
}

// Abs returns a new BigInt that is the absolute value of b
func (b *BigInt) Abs() *BigInt {
	result := new(big.Int)
	result.Abs(b.Value)
	return &BigInt{Value: result}
}

// ToInt64 converts the BigInt to int64
// Returns false if the value overflows int64
func (b *BigInt) ToInt64() (int64, bool) {
	if !b.Value.IsInt64() {
		return 0, false
	}
	return b.Value.Int64(), true
}

// ToFloat64 converts the BigInt to float64
// May lose precision for very large values
func (b *BigInt) ToFloat64() float64 {
	f := new(big.Float).SetInt(b.Value)
	result, _ := f.Float64()
	return result
}

// ToBigFloat converts the BigInt to a BigFloat
func (b *BigInt) ToBigFloat() *BigFloat {
	f := new(big.Float).SetInt(b.Value)
	return &BigFloat{Value: f}
}

// BitLen returns the length of the absolute value in bits
func (b *BigInt) BitLen() int {
	return b.Value.BitLen()
}

// Sign returns the sign of the BigInt
// Returns -1 for negative, 0 for zero, 1 for positive
func (b *BigInt) Sign() int {
	return b.Value.Sign()
}

// IsZero returns true if the BigInt is zero
func (b *BigInt) IsZero() bool {
	return b.Value.Sign() == 0
}

// IsNegative returns true if the BigInt is negative
func (b *BigInt) IsNegative() bool {
	return b.Value.Sign() < 0
}

// IsPositive returns true if the BigInt is positive
func (b *BigInt) IsPositive() bool {
	return b.Value.Sign() > 0
}

// String returns the string representation
func (b *BigInt) String() string {
	return b.Value.String()
}

// FormatBigInt formats a BigInt value for display with the 'n' suffix
func FormatBigInt(b *BigInt) string {
	return b.Value.String() + "n"
}
