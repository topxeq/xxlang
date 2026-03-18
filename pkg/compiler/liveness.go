// pkg/compiler/liveness.go
package compiler

import (
	"github.com/topxeq/xxlang/pkg/parser"
)

// LivenessAnalyzer computes live intervals for variables in a function
type LivenessAnalyzer struct {
	// Map from variable name to its live interval
	intervals map[string]*LiveInterval
	// Current instruction position
	currentPos int
	// Current symbol table
	symbolTable *SymbolTable
}

// NewLivenessAnalyzer creates a new liveness analyzer
func NewLivenessAnalyzer() *LivenessAnalyzer {
	return &LivenessAnalyzer{
		intervals: make(map[string]*LiveInterval),
	}
}

// Analyze performs liveness analysis on a function body
// Returns a map from variable name to its live interval
func (la *LivenessAnalyzer) Analyze(fn *parser.FunctionLiteral, st *SymbolTable) map[string]*LiveInterval {
	la.intervals = make(map[string]*LiveInterval)
	la.currentPos = 0
	la.symbolTable = st

	// First pass: collect all local variables and their definition points
	la.collectDefinitions(fn.Body)

	// Second pass: compute last use points
	la.currentPos = 0
	la.computeLastUse(fn.Body)

	return la.intervals
}

// collectDefinitions finds all variable definitions and records their start positions
func (la *LivenessAnalyzer) collectDefinitions(block *parser.BlockStatement) {
	for _, stmt := range block.Statements {
		la.collectDefinitionInStatement(stmt)
	}
}

// collectDefinitionInStatement processes a statement for variable definitions
func (la *LivenessAnalyzer) collectDefinitionInStatement(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VarStatement:
		// Variable definition
		name := s.Name.Value
		if sym, ok := la.symbolTable.Resolve(name); ok && sym.Scope == LocalScope {
			la.recordDefinition(&sym, la.currentPos)
		}
		la.currentPos++
		// Process the initial value
		if s.Value != nil {
			la.collectUseInExpression(s.Value)
		}

	case *parser.ConstStatement:
		// Constant definition
		name := s.Name.Value
		if sym, ok := la.symbolTable.Resolve(name); ok && sym.Scope == LocalScope {
			la.recordDefinition(&sym, la.currentPos)
		}
		la.currentPos++
		if s.Value != nil {
			la.collectUseInExpression(s.Value)
		}

	case *parser.ReturnStatement:
		la.currentPos++
		if s.ReturnValue != nil {
			la.collectUseInExpression(s.ReturnValue)
		}

	case *parser.ExpressionStatement:
		la.collectDefinitionInExpression(s.Expression)

	case *parser.BlockStatement:
		la.collectDefinitions(s)

	case *parser.IfStatement:
		la.currentPos++ // condition check
		if s.Condition != nil {
			la.collectUseInExpression(s.Condition)
		}
		la.collectDefinitions(s.Consequence)
		if s.Alternative != nil {
			la.collectDefinitions(s.Alternative)
		}

	case *parser.WhileStatement:
		la.currentPos++ // condition check
		if s.Condition != nil {
			la.collectUseInExpression(s.Condition)
		}
		la.collectDefinitions(s.Body)

	case *parser.ForStatement:
		// Init
		if s.Init != nil {
			la.collectDefinitionInStatement(s.Init)
		}
		// Condition
		la.currentPos++
		if s.Condition != nil {
			la.collectUseInExpression(s.Condition)
		}
		// Body
		la.collectDefinitions(s.Body)
		// Update
		if s.Update != nil {
			la.collectDefinitionInStatement(s.Update)
		}

	case *parser.ForInStatement:
		la.currentPos++ // iterator setup
		if s.Key != nil {
			if sym, ok := la.symbolTable.Resolve(s.Key.Value); ok && sym.Scope == LocalScope {
				la.recordDefinition(&sym, la.currentPos)
			}
		}
		if s.Value != nil {
			if sym, ok := la.symbolTable.Resolve(s.Value.Value); ok && sym.Scope == LocalScope {
				la.recordDefinition(&sym, la.currentPos)
			}
		}
		la.collectUseInExpression(s.Iterable)
		la.collectDefinitions(s.Body)

	case *parser.SwitchStatement:
		la.currentPos++ // switch value
		la.collectUseInExpression(s.Expression)
		for _, c := range s.Cases {
			la.currentPos++ // case comparison
			la.collectUseInExpression(c.Expression)
			la.collectDefinitions(c.Consequence)
		}
		if s.Default != nil {
			la.collectDefinitions(s.Default.Consequence)
		}

	case *parser.TryStatement:
		la.collectDefinitions(s.Block)
		if s.Catch != nil {
			la.currentPos++ // catch variable
			la.collectDefinitions(s.Catch.Block)
		}
		if s.Finally != nil {
			la.collectDefinitions(s.Finally.Block)
		}

	case *parser.ThrowStatement:
		la.currentPos++
		la.collectUseInExpression(s.ErrExpr)

	case *parser.ImportStatement:
		la.currentPos++

	case *parser.ExportStatement:
		la.currentPos++
		la.collectDefinitionInStatement(s.Exportable)

	case *parser.ClassStatement:
		la.currentPos++
		for _, m := range s.Methods {
			// Methods are handled separately as functions
			_ = m
		}

	case *parser.BreakStatement, *parser.ContinueStatement:
		la.currentPos++
	}
}

// collectDefinitionInExpression processes an expression for variable definitions
func (la *LivenessAnalyzer) collectDefinitionInExpression(expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.AssignmentExpression:
		// Assignment might define a variable
		if ident, ok := e.Left.(*parser.Identifier); ok {
			if sym, ok := la.symbolTable.Resolve(ident.Value); ok && sym.Scope == LocalScope {
				la.recordDefinition(&sym, la.currentPos)
			}
		}
		la.currentPos++
		la.collectUseInExpression(e.Value)

	case *parser.CompoundAssignmentExpression:
		if ident, ok := e.Left.(*parser.Identifier); ok {
			if sym, ok := la.symbolTable.Resolve(ident.Value); ok && sym.Scope == LocalScope {
				la.recordDefinition(&sym, la.currentPos)
			}
		}
		la.currentPos++
		la.collectUseInExpression(e.Right)

	case *parser.InfixExpression:
		la.collectDefinitionInExpression(e.Left)
		la.currentPos++
		la.collectDefinitionInExpression(e.Right)

	case *parser.PrefixExpression:
		la.currentPos++
		la.collectDefinitionInExpression(e.Right)

	case *parser.PostfixExpression:
		if ident, ok := e.Left.(*parser.Identifier); ok {
			if sym, ok := la.symbolTable.Resolve(ident.Value); ok && sym.Scope == LocalScope {
				la.recordDefinition(&sym, la.currentPos)
			}
		}
		la.currentPos++

	case *parser.CallExpression:
		la.collectDefinitionInExpression(e.Function)
		for _, arg := range e.Arguments {
			la.currentPos++
			la.collectDefinitionInExpression(arg)
		}
		la.currentPos++ // call instruction

	case *parser.IndexExpression:
		la.collectDefinitionInExpression(e.Left)
		la.currentPos++
		la.collectDefinitionInExpression(e.Index)

	case *parser.DotExpression:
		la.currentPos++
		la.collectDefinitionInExpression(e.Object)

	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			la.currentPos++
			la.collectDefinitionInExpression(elem)
		}
		la.currentPos++ // array creation

	case *parser.MapLiteral:
		for key, val := range e.Pairs {
			la.currentPos++
			la.collectDefinitionInExpression(key)
			la.currentPos++
			la.collectDefinitionInExpression(val)
		}
		la.currentPos++ // map creation

	case *parser.FunctionLiteral:
		// Nested function - skip for now (handled separately)

	case *parser.TernaryExpression:
		la.collectDefinitionInExpression(e.Condition)
		la.currentPos++
		la.collectDefinitionInExpression(e.Consequent)
		la.currentPos++
		la.collectDefinitionInExpression(e.Alternative)

	case *parser.SpreadExpression:
		la.collectDefinitionInExpression(e.Expression)

	case *parser.NewExpression:
		la.currentPos++
		la.collectDefinitionInExpression(e.Class)
		for _, arg := range e.Arguments {
			la.currentPos++
			la.collectDefinitionInExpression(arg)
		}

	default:
		la.currentPos++
	}
}

// collectUseInExpression records variable uses in an expression
func (la *LivenessAnalyzer) collectUseInExpression(expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		if sym, ok := la.symbolTable.Resolve(e.Value); ok && sym.Scope == LocalScope {
			la.recordUse(&sym, la.currentPos)
		}

	case *parser.AssignmentExpression:
		la.collectUseInExpression(e.Value)
		// Left is being defined, not used (unless it's a compound assignment)
		if ident, ok := e.Left.(*parser.Identifier); ok {
			if sym, ok := la.symbolTable.Resolve(ident.Value); ok && sym.Scope == LocalScope {
				la.recordUse(&sym, la.currentPos)
			}
		}

	case *parser.CompoundAssignmentExpression:
		la.collectUseInExpression(e.Left)
		la.collectUseInExpression(e.Right)

	case *parser.InfixExpression:
		la.collectUseInExpression(e.Left)
		la.collectUseInExpression(e.Right)

	case *parser.PrefixExpression:
		la.collectUseInExpression(e.Right)

	case *parser.PostfixExpression:
		la.collectUseInExpression(e.Left)

	case *parser.CallExpression:
		la.collectUseInExpression(e.Function)
		for _, arg := range e.Arguments {
			la.collectUseInExpression(arg)
		}

	case *parser.IndexExpression:
		la.collectUseInExpression(e.Left)
		la.collectUseInExpression(e.Index)

	case *parser.DotExpression:
		la.collectUseInExpression(e.Object)

	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			la.collectUseInExpression(elem)
		}

	case *parser.MapLiteral:
		for key, val := range e.Pairs {
			la.collectUseInExpression(key)
			la.collectUseInExpression(val)
		}

	case *parser.FunctionLiteral:
		// Nested function body

	case *parser.TernaryExpression:
		la.collectUseInExpression(e.Condition)
		la.collectUseInExpression(e.Consequent)
		la.collectUseInExpression(e.Alternative)

	case *parser.SpreadExpression:
		la.collectUseInExpression(e.Expression)

	case *parser.NewExpression:
		la.collectUseInExpression(e.Class)
		for _, arg := range e.Arguments {
			la.collectUseInExpression(arg)
		}
	}
}

// computeLastUse finds the last use point for each variable
func (la *LivenessAnalyzer) computeLastUse(block *parser.BlockStatement) {
	for _, stmt := range block.Statements {
		la.computeLastUseInStatement(stmt)
	}
}

// computeLastUseInStatement processes a statement for variable last uses
func (la *LivenessAnalyzer) computeLastUseInStatement(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VarStatement:
		la.currentPos++
		if s.Value != nil {
			la.updateLastUseInExpression(s.Value)
		}

	case *parser.ConstStatement:
		la.currentPos++
		if s.Value != nil {
			la.updateLastUseInExpression(s.Value)
		}

	case *parser.ReturnStatement:
		la.currentPos++
		if s.ReturnValue != nil {
			la.updateLastUseInExpression(s.ReturnValue)
		}

	case *parser.ExpressionStatement:
		la.computeLastUseInExpression(s.Expression)

	case *parser.BlockStatement:
		la.computeLastUse(s)

	case *parser.IfStatement:
		la.currentPos++
		if s.Condition != nil {
			la.updateLastUseInExpression(s.Condition)
		}
		la.computeLastUse(s.Consequence)
		if s.Alternative != nil {
			la.computeLastUse(s.Alternative)
		}

	case *parser.WhileStatement:
		la.currentPos++
		if s.Condition != nil {
			la.updateLastUseInExpression(s.Condition)
		}
		la.computeLastUse(s.Body)

	case *parser.ForStatement:
		if s.Init != nil {
			la.computeLastUseInStatement(s.Init)
		}
		la.currentPos++
		if s.Condition != nil {
			la.updateLastUseInExpression(s.Condition)
		}
		la.computeLastUse(s.Body)
		if s.Update != nil {
			la.computeLastUseInStatement(s.Update)
		}

	case *parser.ForInStatement:
		la.currentPos++
		la.updateLastUseInExpression(s.Iterable)
		la.computeLastUse(s.Body)

	case *parser.SwitchStatement:
		la.currentPos++
		la.updateLastUseInExpression(s.Expression)
		for _, c := range s.Cases {
			la.currentPos++
			la.updateLastUseInExpression(c.Expression)
			la.computeLastUse(c.Consequence)
		}
		if s.Default != nil {
			la.computeLastUse(s.Default.Consequence)
		}

	case *parser.TryStatement:
		la.computeLastUse(s.Block)
		if s.Catch != nil {
			la.currentPos++
			la.computeLastUse(s.Catch.Block)
		}
		if s.Finally != nil {
			la.computeLastUse(s.Finally.Block)
		}

	case *parser.ThrowStatement:
		la.currentPos++
		la.updateLastUseInExpression(s.ErrExpr)

	case *parser.ImportStatement:
		la.currentPos++

	case *parser.ExportStatement:
		la.currentPos++
		la.computeLastUseInStatement(s.Exportable)

	case *parser.BreakStatement, *parser.ContinueStatement:
		la.currentPos++
	}
}

// computeLastUseInExpression processes an expression for variable last uses
func (la *LivenessAnalyzer) computeLastUseInExpression(expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		if sym, ok := la.symbolTable.Resolve(e.Value); ok && sym.Scope == LocalScope {
			la.recordUse(&sym, la.currentPos)
		}
		la.currentPos++

	case *parser.AssignmentExpression:
		la.computeLastUseInExpression(e.Left)
		la.computeLastUseInExpression(e.Value)

	case *parser.CompoundAssignmentExpression:
		la.computeLastUseInExpression(e.Left)
		la.currentPos++
		la.computeLastUseInExpression(e.Right)

	case *parser.InfixExpression:
		la.computeLastUseInExpression(e.Left)
		la.currentPos++
		la.computeLastUseInExpression(e.Right)

	case *parser.PrefixExpression:
		la.currentPos++
		la.computeLastUseInExpression(e.Right)

	case *parser.PostfixExpression:
		la.computeLastUseInExpression(e.Left)
		la.currentPos++

	case *parser.CallExpression:
		la.computeLastUseInExpression(e.Function)
		for _, arg := range e.Arguments {
			la.currentPos++
			la.computeLastUseInExpression(arg)
		}
		la.currentPos++

	case *parser.IndexExpression:
		la.computeLastUseInExpression(e.Left)
		la.currentPos++
		la.computeLastUseInExpression(e.Index)

	case *parser.DotExpression:
		la.currentPos++
		la.computeLastUseInExpression(e.Object)

	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			la.currentPos++
			la.computeLastUseInExpression(elem)
		}
		la.currentPos++

	case *parser.MapLiteral:
		for key, val := range e.Pairs {
			la.currentPos++
			la.computeLastUseInExpression(key)
			la.currentPos++
			la.computeLastUseInExpression(val)
		}
		la.currentPos++

	case *parser.FunctionLiteral:
		// Nested function

	case *parser.TernaryExpression:
		la.computeLastUseInExpression(e.Condition)
		la.currentPos++
		la.computeLastUseInExpression(e.Consequent)
		la.currentPos++
		la.computeLastUseInExpression(e.Alternative)

	case *parser.SpreadExpression:
		la.computeLastUseInExpression(e.Expression)

	case *parser.NewExpression:
		la.currentPos++
		la.computeLastUseInExpression(e.Class)
		for _, arg := range e.Arguments {
			la.currentPos++
			la.computeLastUseInExpression(arg)
		}

	default:
		la.currentPos++
	}
}

// updateLastUseInExpression updates the last use position in an expression
func (la *LivenessAnalyzer) updateLastUseInExpression(expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		if sym, ok := la.symbolTable.Resolve(e.Value); ok && sym.Scope == LocalScope {
			la.recordUse(&sym, la.currentPos)
		}

	case *parser.AssignmentExpression:
		la.updateLastUseInExpression(e.Left)
		la.updateLastUseInExpression(e.Value)

	case *parser.CompoundAssignmentExpression:
		la.updateLastUseInExpression(e.Left)
		la.updateLastUseInExpression(e.Right)

	case *parser.InfixExpression:
		la.updateLastUseInExpression(e.Left)
		la.updateLastUseInExpression(e.Right)

	case *parser.PrefixExpression:
		la.updateLastUseInExpression(e.Right)

	case *parser.PostfixExpression:
		la.updateLastUseInExpression(e.Left)

	case *parser.CallExpression:
		la.updateLastUseInExpression(e.Function)
		for _, arg := range e.Arguments {
			la.updateLastUseInExpression(arg)
		}

	case *parser.IndexExpression:
		la.updateLastUseInExpression(e.Left)
		la.updateLastUseInExpression(e.Index)

	case *parser.DotExpression:
		la.updateLastUseInExpression(e.Object)

	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			la.updateLastUseInExpression(elem)
		}

	case *parser.MapLiteral:
		for key, val := range e.Pairs {
			la.updateLastUseInExpression(key)
			la.updateLastUseInExpression(val)
		}

	case *parser.FunctionLiteral:
		// Nested function

	case *parser.TernaryExpression:
		la.updateLastUseInExpression(e.Condition)
		la.updateLastUseInExpression(e.Consequent)
		la.updateLastUseInExpression(e.Alternative)

	case *parser.SpreadExpression:
		la.updateLastUseInExpression(e.Expression)

	case *parser.NewExpression:
		la.updateLastUseInExpression(e.Class)
		for _, arg := range e.Arguments {
			la.updateLastUseInExpression(arg)
		}
	}
}

// recordDefinition records the definition point of a variable
func (la *LivenessAnalyzer) recordDefinition(sym *Symbol, pos int) {
	if interval, exists := la.intervals[sym.Name]; exists {
		// Update existing interval
		if pos < interval.Start {
			interval.Start = pos
		}
	} else {
		// Create new interval
		la.intervals[sym.Name] = &LiveInterval{
			Var:   sym,
			Start: pos,
			End:   pos,
			Reg:   -1,
		}
	}
}

// recordUse records a use of a variable and extends its live range
func (la *LivenessAnalyzer) recordUse(sym *Symbol, pos int) {
	if interval, exists := la.intervals[sym.Name]; exists {
		// Extend the live range
		if pos > interval.End {
			interval.End = pos
		}
	} else {
		// Variable used before definition (parameter or error)
		la.intervals[sym.Name] = &LiveInterval{
			Var:   sym,
			Start: 0, // Parameters start at 0
			End:   pos,
			Reg:   -1,
		}
	}
}

// GetIntervals returns all live intervals as a slice
func (la *LivenessAnalyzer) GetIntervals() []*LiveInterval {
	intervals := make([]*LiveInterval, 0, len(la.intervals))
	for _, interval := range la.intervals {
		intervals = append(intervals, interval)
	}
	return intervals
}
