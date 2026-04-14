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

type FunctionDecl struct {
	Name   string
	Params []string
	Body   []Statement
}

func (s *FunctionDecl) node() {}
func (s *FunctionDecl) stmt() {}

type FunctionExpr struct {
	Name   string
	Params []string
	Body   []Statement
}

func (e *FunctionExpr) node() {}
func (e *FunctionExpr) expr() {}

type Ident struct {
	Name string
}

func (e *Ident) node() {}
func (e *Ident) expr() {}

type NumberLit struct {
	Value float64
}

func (e *NumberLit) node() {}
func (e *NumberLit) expr() {}

type StringLit struct {
	Value string
}

func (e *StringLit) node() {}
func (e *StringLit) expr() {}

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

type ObjectLit struct {
	Properties []ObjectProperty
}

type ObjectProperty struct {
	Key   string
	Value Expression
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
