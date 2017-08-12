package parse

import (
	"fmt"
)

// Tree 语法树
type Tree struct {
	Root      []Stmt
	text      string
	lex       *lexer
	token     [2]token
	peekCount int
}

//////////////////////////////
// start
//////////////////////////////

// Parse 解析文本并返回语法树
func Parse(text string) (*Tree, error) {
	t := &Tree{text: text}

	// 词法分析
	print("# TOKEN\n")
	t.lex = lex(text)

	// 语法分析
	_, err := t.Parse()

	return t, err
}

// Parse 语法分析,生成语法树
func (t *Tree) Parse() (*Tree, error) {

	for t.peek().typ != EOF {
		n := t.parseStmt()
		if n != nil {
			t.Root = append(t.Root, n)
		}
	}
	// 打印语法树
	print("\n# PARSE\n")
	for _, stmt := range t.Root {
		stmt.stmt()
	}

	t.lex = nil
	return t, nil
}

//////////////////////////////
// token获取以及移动
//////////////////////////////

var numToken = 1

// 返回下一个token
func (t *Tree) next() token {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextToken()
	}

	print(fmt.Sprintf("\n[%v]：(%v, %v)\n", numToken, t.token[t.peekCount].typ, t.token[t.peekCount].val))
	numToken++
	return t.token[t.peekCount]
}

// 返回下一个token，但是不消耗token
func (t *Tree) peek() token {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextToken()
	return t.token[0]
}

func (t *Tree) peekNotNone() (tok token) {
	for {
		tok := t.peek()
		typ := tok.typ

		if typ != EOL && typ != SPACE {
			break
		}
		t.next()
	}
	return tok

}

// 返回下两个token，但是不消耗token
func (t *Tree) peek2() token {
	if t.peekCount == 1 {
		t.peekCount++
		t.token[1] = t.token[0]
		t.token[0] = t.lex.nextToken()
	} else if t.peekCount == 0 {
		t.peekCount = 2
		t.token[1] = t.lex.nextToken()
		t.token[0] = t.lex.nextToken()
	}
	return t.token[t.peekCount-2]
}

// 判断是否类型匹配
func (t *Tree) match(typ TokenType) token {
	token := t.next()
	if token.typ != typ {
		panic(fmt.Sprintf("match: 类型不对, give type is: (%v), match type is: (%v, %v)", typ, token.typ, token.val))
	}
	return token
}

//////////////////////////////
// new stmt
//////////////////////////////
func (t *Tree) newExprStmt() *ExprStmt {
	tok := t.peek()
	stmt := &ExprStmt{}
	stmt.SetPosition(tok.Position())
	return stmt
}
func (t *Tree) newIfStmt() *IfStmt {
	tok := t.peek()
	stmt := &IfStmt{}
	stmt.SetPosition(tok.Position())
	return stmt

}
func (t *Tree) newForStmt() *ForStmt {
	tok := t.peek()
	stmt := &ForStmt{}
	stmt.SetPosition(tok.Position())
	return stmt

}
func (t *Tree) newReturnStmt() *ReturnStmt {
	tok := t.peek()
	stmt := &ReturnStmt{}
	stmt.SetPosition(tok.Position())
	return stmt

}
func (t *Tree) newBreakStmt() *BreakStmt {
	tok := t.peek()
	stmt := &BreakStmt{}
	stmt.SetPosition(tok.Position())
	return stmt

}
func (t *Tree) newContinueStmt() *ContinueStmt {
	tok := t.peek()
	stmt := &ContinueStmt{}
	stmt.SetPosition(tok.Position())
	return stmt

}

//////////////////////////////
// new expr
//////////////////////////////
func (t *Tree) newNumberExpr() *NumberExpr {
	tok := t.peek()
	expr := &NumberExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newStringExpr() *StringExpr {
	tok := t.peek()
	expr := &StringExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newIdentExpr() *IdentExpr {
	tok := t.peek()
	expr := &IdentExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newUnaryExpr() *UnaryExpr {
	tok := t.peek()
	expr := &UnaryExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newParenExpr() *ParenExpr {
	tok := t.peek()
	expr := &ParenExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newBinOpExpr() *BinOpExpr {
	tok := t.peek()
	expr := &BinOpExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newConstExpr() *ConstExpr {
	tok := t.peek()
	expr := &ConstExpr{}
	expr.SetPosition(tok.Position())
	return expr
}
func (t *Tree) newFuncExpr() *FuncExpr {
	tok := t.peek()
	expr := &FuncExpr{}
	expr.SetPosition(tok.Position())
	return expr
}

func (t *Tree) newCallExpr() *CallExpr {
	tok := t.peek()
	expr := &CallExpr{}
	expr.SetPosition(tok.Position())
	return expr
}

//////////////////////////////
// ## 语法分析
//////////////////////////////

// ## 语句

//
//IF
//FOR
//RETURN
//EXPRESSION
func (t *Tree) parseStmt() Stmt {

	token := t.peek()

	switch token.typ {
	case IF:
		n := t.parseIf()
		return n
	case FOR:
		n := t.parseFor()
		return n
	case RETURN:
		n := t.parseReturnStmt()
		return n
	case EOL:
		t.next()
		return nil
	case BREAK:
		n := t.parseBreakStmt()
		return n
	case CONTINUE:
		n := t.parseContinueStmt()
		return n
	default:
		n := t.newExprStmt()
		n.Expr = t.parseExpr()
		// 函数后面没分号
		switch n.Expr.(type) {
		case *FuncExpr:
			// nothing
		default:
			t.match(SEMICOLON)
		}
		return n
	}
}

// ### IF

// parseIf parse like
//if CONDITION {
//    DO
//} elif CONDITION {
//    DO
//} else {
//    DO
//}
func (t *Tree) parseIf() Stmt {

	n := t.newIfStmt()

	t.match(IF)

	n.Condition = t.parseExpr()

	n.Do = t.parseBlock()

	for t.peek().typ == ELIF {
		n.Elif = append(n.Elif, t.parseElif())
	}

	if t.peek().typ == ELSE {
		t.match(ELSE)
		n.Else = t.parseBlock()
	}

	return n

}

func (t *Tree) parseElif() Stmt {
	n := t.newIfStmt()
	t.match(ELIF)

	n.Condition = t.parseExpr()

	n.Do = t.parseBlock()

	return n

}

// ## FOR
func (t *Tree) parseFor() Stmt {
	n := t.newForStmt()

	t.match(FOR)

	if t.peek().typ != SEMICOLON {
		n.Initial = t.parseExpr()
	}
	t.match(SEMICOLON)

	if t.peek().typ != SEMICOLON {
		n.Condition = t.parseExpr()
	}
	t.match(SEMICOLON)

	if t.peek().typ != LC {
		n.After = t.parseExpr()
	}

	n.Do = t.parseBlock()

	return n

}

// ## break
func (t *Tree) parseBreakStmt() Stmt {
	n := t.newBreakStmt()
	t.match(BREAK)
	return n
}

// ## continue
func (t *Tree) parseContinueStmt() Stmt {
	n := t.newContinueStmt()
	t.match(CONTINUE)
	return n
}

// ## 函数

// parseFuncExpr
//func funcname(arg1,arg2, arg3) {
//    DO
//}
func (t *Tree) parseFuncExpr() Expr {
	n := t.newFuncExpr()

	t.match(FUNC)

	if t.peek().typ == IDENTI {
		n.Name = t.match(IDENTI).val
	}

	t.match(LP)

	if t.peek().typ != RP {
		n.Args = t.parseArgList()
	}

	t.match(RP)

	n.Stmts = t.parseBlock()

	return n
}

// 形参
func (t *Tree) parseArgList() []string {
	l := []string{}

	first := t.match(IDENTI)
	l = append(l, first.val)

	for t.peek().typ == COMMA {
		t.match(COMMA)
		item := t.match(IDENTI)
		l = append(l, item.val)
	}

	return l
}

// 实参
func (t *Tree) parseParameterList() []Expr {
	l := []Expr{}

	first := t.parseExpr()
	l = append(l, first)

	for t.peek().typ == COMMA {
		t.match(COMMA)
		p := t.parseExpr()
		l = append(l, p)
	}
	return l
}

// ## RETURN
func (t *Tree) parseReturnStmt() Stmt {
	n := t.newReturnStmt()
	t.match(RETURN)

	if t.peek().typ != SEMICOLON {
		n.Expr = t.parseExpr()
	}
	t.match(SEMICOLON)

	return n
}

// ## 块
func (t *Tree) parseBlock() []Stmt {
	n := []Stmt{}

	t.match(LC)

	for t.peekNotNone(); t.peek().typ != RC; t.peekNotNone() {
		statement := t.parseStmt()
		n = append(n, statement)
	}

	t.match(RC)

	return n
}

// ## 解析表达式

// ### Expr

// parseExpr ...
func (t *Tree) parseExpr() Expr {

	// 函数
	if t.peek().typ == FUNC {
		expr := t.parseFuncExpr()
		return expr
	}

	// 赋值
	if t.peek().typ == IDENTI && t.peek2().typ == EQ {
		expr := t.newBinOpExpr()

		token := t.match(IDENTI)
		LExpr := t.newIdentExpr()
		LExpr.Lit = token.val

		expr.Lhs = LExpr

		expr.Operator = t.peek().val

		t.match(EQ)

		expr.Rhs = t.parseExpr()

		return expr
	}

	expr := t.parseLogicalOrExp()

	return expr
}

// or表达式
func (t *Tree) parseLogicalOrExp() Expr {

	expr := t.newBinOpExpr()
	lExpr := t.parseLogicalAndExp()

	if t.peek().typ == OROR {
		expr.Lhs = lExpr
		expr.Operator = t.peek().val

		t.match(OROR)
		expr.Rhs = t.parseLogicalOrExp()
		return expr
	}
	return lExpr
}

// and 表达式
func (t *Tree) parseLogicalAndExp() Expr {

	expr := t.newBinOpExpr()
	lExpr := t.parseEqualityExp()

	if t.peek().typ == ANDAND {
		expr.Lhs = lExpr
		expr.Operator = t.peek().val

		t.match(ANDAND)
		expr.Rhs = t.parseLogicalAndExp()
		return expr
	}
	return lExpr
}

// 相等表达式
func (t *Tree) parseEqualityExp() Expr {
	expr := t.newBinOpExpr()
	lExpr := t.parseRelationalExp()

	switch typ := t.peek().typ; typ {
	case EQ, NEQ:
		expr.Lhs = lExpr
		expr.Operator = t.peek().val

		t.match(typ)
		expr.Rhs = t.parseEqualityExp()
		return expr
	}

	return lExpr
}

// 关系表达式
func (t *Tree) parseRelationalExp() Expr {
	expr := t.newBinOpExpr()

	lExpr := t.parseAdditiveExp()

	switch typ := t.peek().typ; typ {
	case GT, GE, LT, LE:
		expr.Lhs = lExpr
		expr.Operator = t.peek().val

		t.match(typ)
		expr.Rhs = t.parseRelationalExp()
		return expr
	}

	return lExpr
}

// 加法表达式
func (t *Tree) parseAdditiveExp() Expr {
	expr := t.newBinOpExpr()

	lExpr := t.parseMultiplicativeExp()

	switch typ := t.peek().typ; typ {
	case PLUS, MINUS:
		expr.Lhs = lExpr
		expr.Operator = t.peek().val

		t.match(typ)
		expr.Rhs = t.parseAdditiveExp()
		return expr
	}

	return lExpr
}

// 乘法表达式
func (t *Tree) parseMultiplicativeExp() Expr {
	expr := t.newBinOpExpr()

	lExpr := t.parseUnaryExp()

	switch typ := t.peek().typ; typ {
	case MULTIPLY, DIVIDE:
		expr.Lhs = lExpr
		expr.Operator = t.peek().val

		t.match(typ)
		expr.Rhs = t.parseMultiplicativeExp()
		return expr
	}

	return lExpr
}

// 一元表达式
func (t *Tree) parseUnaryExp() Expr {

	if t.peek().typ == PLUS {
		expr := t.newUnaryExpr()
		expr.Operator = t.peek().val

		t.match(PLUS)
		expr.Expr = t.parseUnaryExp()
		return expr
	}

	expr := t.parsePrimaryExp()

	return expr

}

// 表达式最小单元
func (t *Tree) parsePrimaryExp() Expr {

	switch t.peek().typ {
	case IDENTI:

		// identifier
		if t.peek2().typ != LP {
			expr := t.newStringExpr()
			expr.Lit = t.match(IDENTI).val
			return expr
		}

		// func call
		expr := t.newCallExpr()

		expr.Name = t.match(IDENTI).val
		t.match(LP)

		if t.peek().typ != RP {
			expr.SubExprs = t.parseParameterList()
		}
		t.match(RP)
		return expr

	case LP:
		// ()
		expr := t.newParenExpr()
		t.match(LP)
		expr.SubExpr = t.parseExpr()
		t.match(RP)
		return expr

	case NUMBER:
		// number
		expr := t.newNumberExpr()
		expr.Lit = t.match(NUMBER).val
		return expr
	case STRING:
		expr := t.newStringExpr()
		expr.Lit = t.match(STRING).val
		return expr
	case BOOL, NIL:
		expr := t.newConstExpr()
		expr.Value = t.match(BOOL).val
		return expr
	default:
		panic("fuck 不知道输入了什么鬼")
	}
}
