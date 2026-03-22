// pkg/objects/bigfloat.go
// BigFloat represents an arbitrary-precision floating-point value
package objects

import (
	"fmt"
	"hash/fnv"
	"math/big"
)

// DefaultBigFloatPrecision is the default precision for BigFloat values (256 bits)
const DefaultBigFloatPrecision = 256

// BigFloat represents an arbitrary-precision floating-point value
// It wraps Go's math/big.Float for unlimited precision floating-point arithmetic
type BigFloat struct {
	Value *big.Float
}

// NewBigFloat creates a new BigFloat from a big.Float
func NewBigFloat(val *big.Float) *BigFloat {
	if val == nil {
		f := new(big.Float).SetPrec(DefaultBigFloatPrecision)
		return &BigFloat{Value: f}
	}
	// Copy and set precision
	f := new(big.Float).SetPrec(DefaultBigFloatPrecision).Set(val)
	return &BigFloat{Value: f}
}

// NewBigFloatFromFloat64 creates a new BigFloat from a float64
func NewBigFloatFromFloat64(f float64) *BigFloat {
	val := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetFloat64(f)
	return &BigFloat{Value: val}
}

// NewBigFloatFromString creates a new BigFloat from a string representation
func NewBigFloatFromString(s string) (*BigFloat, error) {
	val := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	_, _, err := val.Parse(s, 10)
	if err != nil {
		return nil, fmt.Errorf("invalid big float format: %s", s)
	}
	return &BigFloat{Value: val}, nil
}

// NewBigFloatFromInt64 creates a new BigFloat from an int64
func NewBigFloatFromInt64(n int64) *BigFloat {
	val := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetInt64(n)
	return &BigFloat{Value: val}
}

// NewBigFloatFromInt creates a new BigFloat from an Int object
func NewBigFloatFromInt(i *Int) *BigFloat {
	val := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetInt64(i.Value)
	return &BigFloat{Value: val}
}

// NewBigFloatFromBigInt creates a new BigFloat from a BigInt object
func NewBigFloatFromBigInt(b *BigInt) *BigFloat {
	val := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetInt(b.Value)
	return &BigFloat{Value: val}
}

// Type returns the object type
func (b *BigFloat) Type() ObjectType { return BigFloatType }

// TypeTag returns the type tag for fast type checking
func (b *BigFloat) TypeTag() TypeTag { return TagBigFloat }

// Inspect returns the string representation (without the 'm' suffix)
func (b *BigFloat) Inspect() string {
	return b.Value.Text('g', -1)
}

// ToBool converts the BigFloat to a boolean
func (b *BigFloat) ToBool() *Bool {
	if b.Value.Sign() == 0 {
		return FALSE
	}
	return TRUE
}

// HashKey returns the hash key for map operations
func (b *BigFloat) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(b.Value.Text('g', 20)))
	return HashKey{Type: BigFloatType, Value: h.Sum64()}
}

// Clone creates a copy of the BigFloat
func (b *BigFloat) Clone() *BigFloat {
	val := new(big.Float).SetPrec(DefaultBigFloatPrecision).Set(b.Value)
	return &BigFloat{Value: val}
}

// Cmp compares two BigFloat values
// Returns -1 if b < other, 0 if b == other, 1 if b > other
func (b *BigFloat) Cmp(other *BigFloat) int {
	return b.Value.Cmp(other.Value)
}

// Add returns a new BigFloat that is the sum of b and other
func (b *BigFloat) Add(other *BigFloat) *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.Add(b.Value, other.Value)
	return &BigFloat{Value: result}
}

// AddFloat64 returns a new BigFloat that is the sum of b and a float64
func (b *BigFloat) AddFloat64(f float64) *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	other := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetFloat64(f)
	result.Add(b.Value, other)
	return &BigFloat{Value: result}
}

// Sub returns a new BigFloat that is b minus other
func (b *BigFloat) Sub(other *BigFloat) *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.Sub(b.Value, other.Value)
	return &BigFloat{Value: result}
}

// SubFloat64 returns a new BigFloat that is b minus a float64
func (b *BigFloat) SubFloat64(f float64) *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	other := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetFloat64(f)
	result.Sub(b.Value, other)
	return &BigFloat{Value: result}
}

// Mul returns a new BigFloat that is b multiplied by other
func (b *BigFloat) Mul(other *BigFloat) *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.Mul(b.Value, other.Value)
	return &BigFloat{Value: result}
}

// MulFloat64 returns a new BigFloat that is b multiplied by a float64
func (b *BigFloat) MulFloat64(f float64) *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	other := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetFloat64(f)
	result.Mul(b.Value, other)
	return &BigFloat{Value: result}
}

// Div returns a new BigFloat that is b divided by other
// Returns nil if other is zero
func (b *BigFloat) Div(other *BigFloat) *BigFloat {
	if other.Value.Sign() == 0 {
		return nil
	}
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.Quo(b.Value, other.Value)
	return &BigFloat{Value: result}
}

// DivFloat64 returns a new BigFloat that is b divided by a float64
// Returns nil if f is zero
func (b *BigFloat) DivFloat64(f float64) *BigFloat {
	if f == 0 {
		return nil
	}
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	other := new(big.Float).SetPrec(DefaultBigFloatPrecision).SetFloat64(f)
	result.Quo(b.Value, other)
	return &BigFloat{Value: result}
}

// Neg returns a new BigFloat that is the negation of b
func (b *BigFloat) Neg() *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.Neg(b.Value)
	return &BigFloat{Value: result}
}

// Abs returns a new BigFloat that is the absolute value of b
func (b *BigFloat) Abs() *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.Abs(b.Value)
	return &BigFloat{Value: result}
}

// Floor returns the greatest integer value less than or equal to b
func (b *BigFloat) Floor() *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.SetMode(big.ToNegativeInf).Set(b.Value)
	return &BigFloat{Value: result}
}

// Ceil returns the least integer value greater than or equal to b
func (b *BigFloat) Ceil() *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.SetMode(big.ToPositiveInf).Set(b.Value)
	return &BigFloat{Value: result}
}

// Round returns the nearest integer value to b
func (b *BigFloat) Round() *BigFloat {
	result := new(big.Float).SetPrec(DefaultBigFloatPrecision)
	result.SetMode(big.ToNearestEven).Set(b.Value)
	return &BigFloat{Value: result}
}

// ToFloat64 converts the BigFloat to float64
// May lose precision for very large or high-precision values
func (b *BigFloat) ToFloat64() (float64, big.Accuracy) {
	return b.Value.Float64()
}

// ToInt64 converts the BigFloat to int64 (truncates toward zero)
// Returns false if the value overflows int64
func (b *BigFloat) ToInt64() (int64, bool) {
	i := new(big.Int)
	b.Value.Int(i)
	if !i.IsInt64() {
		return 0, false
	}
	return i.Int64(), true
}

// ToBigInt converts the BigFloat to a BigInt (truncates toward zero)
func (b *BigFloat) ToBigInt() *BigInt {
	i := new(big.Int)
	b.Value.Int(i)
	return &BigInt{Value: i}
}

// Sign returns the sign of the BigFloat
// Returns -1 for negative, 0 for zero, 1 for positive
func (b *BigFloat) Sign() int {
	return b.Value.Sign()
}

// IsZero returns true if the BigFloat is zero
func (b *BigFloat) IsZero() bool {
	return b.Value.Sign() == 0
}

// IsNegative returns true if the BigFloat is negative
func (b *BigFloat) IsNegative() bool {
	return b.Value.Sign() < 0
}

// IsPositive returns true if the BigFloat is positive
func (b *BigFloat) IsPositive() bool {
	return b.Value.Sign() > 0
}

// Precision returns the precision of the BigFloat in bits
func (b *BigFloat) Precision() uint {
	return b.Value.Prec()
}

// SetPrecision returns a new BigFloat with the specified precision
func (b *BigFloat) SetPrecision(prec uint) *BigFloat {
	result := new(big.Float).SetPrec(prec).Set(b.Value)
	return &BigFloat{Value: result}
}

// String returns the string representation
func (b *BigFloat) String() string {
	return b.Value.Text('g', -1)
}

// FormatBigFloat formats a BigFloat value for display with the 'm' suffix
func FormatBigFloat(b *BigFloat) string {
	return b.Value.Text('g', -1) + "m"
}
