package jsengine

type Node interface {
	node()
}

type Statement interface {
	Node
	stmt()
}

type Expression interface {
	Node
	expr()
}

type Program struct {
	Statements []Statement
}

func (p *Program) node() {}

type ExpressionStmt struct {
	Expr Expression
}

func (s *ExpressionStmt) node() {}
func (s *ExpressionStmt) stmt() {}

type VarDecl struct {
	Name  string
	Value Expression
	// For destructuring
	IsDestructuring bool
	DestructPattern *DestructPattern
}

// DestructPattern represents a destructuring pattern
type DestructPattern struct {
	IsArray bool
	Elements []DestructElement
}

// DestructElement represents an element in a destructuring pattern
type DestructElement struct {
	Name     string  // variable name
	Default  Expression // default value
	Property string  // for object destructuring, property name (if different from variable name)
}

func (s *VarDecl) node() {}
func (s *VarDecl) stmt() {}

type IfStmt struct {
	Cond Expression
	Body []Statement
	Else []Statement
}

func (s *IfStmt) node() {}
func (s *IfStmt) stmt() {}

type ForStmt struct {
	Init Statement
	Cond Expression
	Post Statement
	Body []Statement
}

func (s *ForStmt) node() {}
func (s *ForStmt) stmt() {}

// ForInStmt represents a for-in loop: for (var key in obj) { body }
type ForInStmt struct {
	VarName string
	Object  Expression
	Body    []Statement
}

func (s *ForInStmt) node() {}
func (s *ForInStmt) stmt() {}

type ForOfStmt struct {
	VarName string
	Object  Expression
	Body    []Statement
}

func (s *ForOfStmt) node() {}
func (s *ForOfStmt) stmt() {}

type WhileStmt struct {
	Cond Expression
	Body []Statement
}

func (s *WhileStmt) node() {}
func (s *WhileStmt) stmt() {}

type ReturnStmt struct {
	Value Expression
}

func (s *ReturnStmt) node() {}
func (s *ReturnStmt) stmt() {}

// YieldStmt represents a yield statement in a generator
type YieldStmt struct {
	Value Expression
}

func (s *YieldStmt) node() {}
func (s *YieldStmt) stmt() {}

type BreakStmt struct{}

func (s *BreakStmt) node() {}
func (s *BreakStmt) stmt() {}

type ContinueStmt struct{}

func (s *ContinueStmt) node() {}
func (s *ContinueStmt) stmt() {}

type BlockStmt struct {
	Statements []Statement
}

func (s *BlockStmt) node() {}
func (s *BlockStmt) stmt() {}

// TryStmt represents a try-catch-finally statement
type TryStmt struct {
	Body     []Statement
	CatchVar string
	Catch    []Statement
	Finally  []Statement
}

func (s *TryStmt) node() {}
func (s *TryStmt) stmt() {}

// ThrowStmt represents a throw statement
type ThrowStmt struct {
	Value Expression
}

func (s *ThrowStmt) node() {}
func (s *ThrowStmt) stmt() {}

type FunctionDecl struct {
	Name         string
	Params       []string
	DefaultVals  map[string]Expression // default values for parameters
	RestParam    string                // rest parameter name, e.g., "...args"
	Body         []Statement
	IsAsync      bool // true for async function
}

func (s *FunctionDecl) node() {}
func (s *FunctionDecl) stmt() {}

// GeneratorDecl represents a generator function declaration: function* name() { yield ... }
type GeneratorDecl struct {
	Name      string
	Params    []string
	RestParam string
	Body      []Statement
}

func (s *GeneratorDecl) node() {}
func (s *GeneratorDecl) stmt() {}

type FunctionExpr struct {
	Name         string
	Params       []string
	DefaultVals  map[string]Expression // default values for parameters
	RestParam    string                // rest parameter name
	Body         []Statement
	IsAsync      bool // true for async function expression
}

func (e *FunctionExpr) node() {}
func (e *FunctionExpr) expr() {}

type Ident struct {
	Name string
}

func (e *Ident) node() {}
func (e *Ident) expr() {}

// UpdateExpr represents increment/decrement expressions: i++, ++i, i--, --i
type UpdateExpr struct {
	Operator string     // "++" or "--"
	Prefix   bool       // true for ++i, false for i++
	Operand  Expression // the identifier being updated
}

func (e *UpdateExpr) node() {}
func (e *UpdateExpr) expr() {}

type NumberLit struct {
	Value float64
}

func (e *NumberLit) node() {}
func (e *NumberLit) expr() {}

type BigIntLit struct {
	Value string // Store as string to handle arbitrary precision
}

func (e *BigIntLit) node() {}
func (e *BigIntLit) expr() {}

type StringLit struct {
	Value string
}

func (e *StringLit) node() {}
func (e *StringLit) expr() {}

// TemplateLit represents a template literal: `Hello ${name}!`
type TemplateLit struct {
	Parts []TemplatePart
}

// TemplatePart represents either a string or an expression in a template
type TemplatePart struct {
	IsExpr bool
	Text   string     // Used when IsExpr is false
	Expr   Expression // Used when IsExpr is true
}

func (e *TemplateLit) node() {}
func (e *TemplateLit) expr() {}

type BoolLit struct {
	Value bool
}

func (e *BoolLit) node() {}
func (e *BoolLit) expr() {}

type NullLit struct{}

func (e *NullLit) node() {}
func (e *NullLit) expr() {}

type UndefinedLit struct{}

func (e *UndefinedLit) node() {}
func (e *UndefinedLit) expr() {}

type BinaryExpr struct {
	Left  Expression
	Op    string
	Right Expression
}

func (e *BinaryExpr) node() {}
func (e *BinaryExpr) expr() {}

type UnaryExpr struct {
	Op   string
	Expr Expression
}

func (e *UnaryExpr) node() {}
func (e *UnaryExpr) expr() {}

type AssignExpr struct {
	Left  Expression
	Op    string
	Right Expression
}

func (e *AssignExpr) node() {}
func (e *AssignExpr) expr() {}

type CallExpr struct {
	Callee Expression
	Args   []Expression
}

func (e *CallExpr) node() {}
func (e *CallExpr) expr() {}

type MemberExpr struct {
	Object   Expression
	Property Expression
	Computed bool
}

func (e *MemberExpr) node() {}
func (e *MemberExpr) expr() {}

type ArrayLit struct {
	Elements []Expression
}

func (e *ArrayLit) node() {}
func (e *ArrayLit) expr() {}

// SpreadExpr represents spread operator: ...arr or ...obj
type SpreadExpr struct {
	Argument Expression
}

func (e *SpreadExpr) node() {}
func (e *SpreadExpr) expr() {}

type ObjectLit struct {
	Properties []ObjectProperty
}

type ObjectProperty struct {
	Key       string
	Value     Expression
	Spread    bool // true if this is a spread property: ...obj
	Computed  bool // true if key is computed: [expr]
	KeyExpr   Expression // for computed property names
	Shorthand bool // true for shorthand: { x } instead of { x: x }
}

func (e *ObjectLit) node() {}
func (e *ObjectLit) expr() {}

type TernaryExpr struct {
	Cond  Expression
	True  Expression
	False Expression
}

func (e *TernaryExpr) node() {}
func (e *TernaryExpr) expr() {}

type ThisExpr struct{}

func (e *ThisExpr) node() {}
func (e *ThisExpr) expr() {}

// SuperExpr represents super() call in constructor
type SuperExpr struct {
	Args []Expression
}

func (e *SuperExpr) node() {}
func (e *SuperExpr) expr() {}

type NewExpr struct {
	Callee Expression
	Args   []Expression
}

func (e *NewExpr) node() {}
func (e *NewExpr) expr() {}

type TypeOfExpr struct {
	Expr Expression
}

func (e *TypeOfExpr) node() {}
func (e *TypeOfExpr) expr() {}

// ArrowFunctionExpr represents an arrow function: (params) => body
type ArrowFunctionExpr struct {
	Params    []string
	RestParam string // rest parameter name
	Body      []Statement
	// Expression is true if body is a single expression (no braces)
	Expression bool
	Expr       Expression // Used when Expression is true
	IsAsync    bool       // true for async arrow function
}

func (e *ArrowFunctionExpr) node() {}
func (e *ArrowFunctionExpr) expr() {}

// AwaitExpr represents an await expression: await promise
type AwaitExpr struct {
	Expr Expression
}

func (e *AwaitExpr) node() {}
func (e *AwaitExpr) expr() {}

// OptionalChainExpr represents optional chaining: obj?.prop or obj?.[key] or obj?.()
type OptionalChainExpr struct {
	Object   Expression
	Property Expression
	Computed bool // true for obj?.[key], false for obj?.prop
}

func (e *OptionalChainExpr) node() {}
func (e *OptionalChainExpr) expr() {}

// NullishCoalescingExpr represents nullish coalescing: a ?? b
type NullishCoalescingExpr struct {
	Left  Expression
	Right Expression
}

func (e *NullishCoalescingExpr) node() {}
func (e *NullishCoalescingExpr) expr() {}

// InstanceofExpr represents instanceof: obj instanceof Constructor
type InstanceofExpr struct {
	Object      Expression
	Constructor Expression
}

func (e *InstanceofExpr) node() {}
func (e *InstanceofExpr) expr() {}

// InExpr represents the in operator: prop in obj
type InExpr struct {
	Property Expression
	Object   Expression
}

func (e *InExpr) node() {}
func (e *InExpr) expr() {}

// ClassDecl represents a class declaration: class Name { ... }
type ClassDecl struct {
	Name       string
	SuperClass string // extends clause, empty if none
	Body       []ClassElement
}

func (s *ClassDecl) node() {}
func (s *ClassDecl) stmt() {}

// ClassExpr represents a class expression: class { ... } or class Name { ... }
type ClassExpr struct {
	Name       string
	SuperClass string
	Body       []ClassElement
}

func (e *ClassExpr) node() {}
func (e *ClassExpr) expr() {}

// ClassElement represents a member of a class body
type ClassElement struct {
	Type       string // "method", "static", "get", "set"
	Name       string
	Params     []string
	Body       []Statement
	IsStatic   bool
	IsGetter   bool
	IsSetter   bool
}
