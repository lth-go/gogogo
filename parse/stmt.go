package parse

// Stmt provides all of interfaces for statement.
type Stmt interface {
	Pos
	stmt()
}

// StmtImpl provide commonly implementations for Stmt..
type StmtImpl struct {
	PosImpl // StmtImpl provide Pos() function.
}

// stmt provide restraint interface.
func (s *StmtImpl) stmt() {}

// ExprStmt provide expression statement.
type ExprStmt struct {
	StmtImpl
	Expr Expr
}

func (s *ExprStmt) stmt() {
	print("## ExprStmt:\n")
	s.Expr.expr()
}

// IfStmt provide "if/else" statement.
type IfStmt struct {
	StmtImpl
	Condition     Expr
	Do   []Stmt
	Elif []Stmt // This is array of IfStmt
	Else   []Stmt
}

func (s *IfStmt) stmt() {
	print("## IfStmt: \n")
	print("### Condition: \n")
	s.Condition.expr()
	print("### Do: \n")
	rangeStmt(s.Do)
	print("### Elif: \n")
	rangeStmt(s.Elif)
	print("### Else: \n")
	rangeStmt(s.Else)
}

// ForStmt provide C-style "for (;;)" expression statement.
type ForStmt struct {
	StmtImpl
	Initial Expr
	Condition Expr
	After Expr
	Do []Stmt
}
func (s *ForStmt) stmt() {
	print("## ForStmt: \n")
	print("### Initial: \n")
	s.Initial.expr()
	print("### Condition: \n")
	s.Condition.expr()
	print("### After: \n")
	s.After.expr()
	print("### Do: \n")
	rangeStmt(s.Do)
}

// BreakStmt provide "break" expression statement.
type BreakStmt struct {
	StmtImpl
}
func (s *BreakStmt) stmt() {
	print("## BreakStmt: \n")
}

// ContinueStmt provide "continue" expression statement.
type ContinueStmt struct {
	StmtImpl
}
func (s *ContinueStmt) stmt() {
	print("## ContinueStmt: \n")
}

// ForStmt provide "return" expression statement.
type ReturnStmt struct {
	StmtImpl
	Expr Expr
}
func (s *ReturnStmt) stmt() {
	print("## ReturnStmt: \n")
	s.Expr.expr()
}

// utils

func rangeStmt(s []Stmt) {
	for _, stmt := range s {
		stmt.stmt()
	}	
}
