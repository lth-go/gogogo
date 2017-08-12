package parse

// Expr provides all of interfaces for expression.
type Expr interface {
	Pos
	expr()
}

// ExprImpl provide commonly implementations for Expr.
type ExprImpl struct {
	PosImpl // ExprImpl provide Pos() function.
}

// expr provide restraint interface.
func (e *ExprImpl) expr() {}

// NumberExpr provide Number expression.
type NumberExpr struct {
	ExprImpl
	Lit string
}

func (e *NumberExpr) expr() {
	print("* NumberExpr: ", e.Lit, "\n")
}

// StringExpr provide String expression.
type StringExpr struct {
	ExprImpl
	Lit string
}

func (e *StringExpr) expr() {
	print("* StringExpr: ", e.Lit, "\n")
}

// IdentExpr provide identity expression.
type IdentExpr struct {
	ExprImpl
	Lit string
}

func (e *IdentExpr) expr() {
	print("* IdentExpr: ", e.Lit, "\n")
}

// UnaryExpr provide unary minus expression. ex: -1, ^1, ~1.
type UnaryExpr struct {
	ExprImpl
	Operator string
	Expr     Expr
}

func (e *UnaryExpr) expr() {
	print("* UnaryExpr: ", e.Operator, "\n")
	e.Expr.expr()
}

// ParenExpr provide parent block expression.
type ParenExpr struct {
	ExprImpl
	SubExpr Expr
}

func (e *ParenExpr) expr() {
	print("* ParenExpr: \n")
	e.SubExpr.expr()
}

// BinOpExpr provide binary operator expression.
type BinOpExpr struct {
	ExprImpl
	Lhs      Expr
	Operator string
	Rhs      Expr
}

func (e *BinOpExpr) expr() {
	print("* BinOpExpr: ", e.Operator, "\n")
	e.Lhs.expr()
	e.Rhs.expr()
}

// FuncExpr provide function expression.
type FuncExpr struct {
	ExprImpl
	Name  string
	Stmts []Stmt
	Args  []string
}

func (e *FuncExpr) expr() {
	print("* FuncExpr: ", e.Name, "\n")
	print("** Args:", e.Args, "\n")
	rangeStmt(e.Stmts)
}

// CallExpr ...
type CallExpr struct {
	ExprImpl
	Func     interface{}
	Name     string
	SubExprs []Expr
}

func (e *CallExpr) expr() {
	print("* CallExpr: :", e.Name, "\n")
	print("** Args:")
	rangeExpr(e.SubExprs)
}

// ConstExpr provide expression for constant variable.
type ConstExpr struct {
	ExprImpl
	Value string
}

// utils
func rangeExpr(exprs []Expr) {
	for _, exp := range exprs {
		exp.expr()
	}
}
