package vm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"errors"

	"../parse"
)

var (
	NilValue   = reflect.ValueOf((*interface{})(nil))
	NilType    = reflect.TypeOf((*interface{})(nil))
	TrueValue  = reflect.ValueOf(true)
	FalseValue = reflect.ValueOf(false)
)

var (
	BreakError    = errors.New("Unexpected break statement")
	ContinueError = errors.New("Unexpected continue statement")
	ReturnError   = errors.New("Unexpected return statement")
	//InterruptError = errors.New("Execution interrupted")
)

//////////////////////////////
// error
//////////////////////////////
type Error struct {
	Message string
	Pos     parse.Position
}

func NewStringError(pos parse.Pos, err string) error {
	if pos == nil {
		return &Error{Message: err, Pos: parse.Position{1, 1}}
	}
	return &Error{Message: err, Pos: pos.Position()}

}
func NewErrorf(pos parse.Pos, format string, args ...interface{}) error {
	return &Error{Message: fmt.Sprintf(format, args...), Pos: pos.Position()}
}
func NewError(pos parse.Pos, err error) error {

	if err == nil {
		return nil
	}
	if err == BreakError || err == ContinueError || err == ReturnError {
		return err
	}
	//if pe, ok := err.(*parser.Error); ok {
	//    return pe
	//}
	if ee, ok := err.(*Error); ok {
		return ee
	}
	return &Error{Message: err.Error(), Pos: pos.Position()}
}
func (e *Error) Error() string {

	return e.Message
}

//
type Func func(args ...reflect.Value) (reflect.Value, error)

func (f Func) String() string {

	return fmt.Sprintf("[Func: %p]", f)
}
func ToFunc(f Func) reflect.Value {
	return reflect.ValueOf(f)
}

//////////////////////////////
// stmt
//////////////////////////////
func Run(stmts []parse.Stmt, env *Env) (reflect.Value, error) {
	rv := NilValue
	var err error
	for _, stmt := range stmts {
		// TODO
		//if _, ok := stmt.(*parse.BreakStmt); ok {
		//    return NilValue, BreakError
		//}
		// TODO
		//if _, ok := stmt.(*parse.ContinueStmt); ok {
		//    return NilValue, ContinueError
		//}
		rv, err = RunSingleStmt(stmt, env)
		if err != nil {
			return rv, err
		}
		if _, ok := stmt.(*parse.ReturnStmt); ok {
			return reflect.ValueOf(rv), ReturnError
		}
	}
	print("=====")
	print(rv.String())
	return rv, nil
}

func RunSingleStmt(stmt parse.Stmt, env *Env) (reflect.Value, error) {
	switch stmt := stmt.(type) {
	case *parse.ExprStmt:
		rv, err := invokeExpr(stmt.Expr, env)
		if err != nil {
			return rv, NewError(stmt, err)
		}
		return rv, nil
	case *parse.IfStmt:
		rv, err := invokeExpr(stmt.Condition, env)
		if err != nil {
			return rv, NewError(stmt, err)
		}
		// if true
		if toBool(rv) {
			newEnv := env.NewEnv()
			defer newEnv.Destroy()
			rv, err = Run(stmt.Do, newEnv)
			if err != nil {
				return rv, NewError(stmt, err)
			}
			return rv, nil
		}
		// elif
		done := false
		if len(stmt.Elif) > 0 {
			for _, stmt := range stmt.Elif {
				stmtIf := stmt.(*parse.IfStmt)
				rv, err = invokeExpr(stmtIf.Condition, env)
				if err != nil {
					return rv, NewError(stmt, err)
				}
				if !toBool(rv) {
					continue
				}
				// 成功
				done = true
				rv, err = Run(stmtIf.Do, env)
				if err != nil {
					return rv, NewError(stmt, err)
				}
				break
			}
		}
		if !done && len(stmt.Else) > 0 {
			// Else
			newEnv := env.NewEnv()
			defer newEnv.Destroy()
			rv, err = Run(stmt.Else, newEnv)
			if err != nil {
				return rv, NewError(stmt, err)
			}
		}
		return rv, nil
	case *parse.ForStmt:
		newEnv := env.NewEnv()
		defer newEnv.Destroy()
		_, err := invokeExpr(stmt.Initial, newEnv)
		if err != nil {
			return NilValue, err
		}
		for {
			fb, err := invokeExpr(stmt.Condition, newEnv)
			if err != nil {
				return NilValue, err
			}
			if !toBool(fb) {
				break
			}

			rv, err := Run(stmt.Do, newEnv)
			if err != nil {
				// TODO break continue
				if err == ReturnError {
					return rv, err
				}
				return rv, NewError(stmt, err)
			}
			_, err = invokeExpr(stmt.After, newEnv)
			if err != nil {
				return NilValue, err
			}
		}
		return NilValue, nil
	case *parse.ReturnStmt:
		//rvs := []interface{}{}
		// TODO 单个返回值
		rv, err := invokeExpr(stmt.Expr, env)
		if err != nil {
			return rv, NewError(stmt, err)
		}
		return rv, nil
	default:
		return NilValue, NewStringError(stmt, "unknown statement")
	}
}

//////////////////////////////
// expr
//////////////////////////////
func invokeExpr(expr parse.Expr, env *Env) (reflect.Value, error) {
	switch e := expr.(type) {
	case *parse.NumberExpr:
		if strings.Contains(e.Lit, ".") || strings.Contains(e.Lit, "e") {
			v, err := strconv.ParseFloat(e.Lit, 64)
			if err != nil {
				return NilValue, NewError(expr, err)
			}
			return reflect.ValueOf(float64(v)), nil
		}
		var i int64
		var err error
		if strings.HasPrefix(e.Lit, "0x") {
			i, err = strconv.ParseInt(e.Lit[2:], 16, 64)
		} else {
			i, err = strconv.ParseInt(e.Lit, 10, 64)
		}
		if err != nil {
			return NilValue, NewError(expr, err)
		}
		return reflect.ValueOf(i), nil
	case *parse.IdentExpr:
		return env.Get(e.Lit)
	case *parse.StringExpr:
		return reflect.ValueOf(e.Lit), nil
	case *parse.ParenExpr:
		v, err := invokeExpr(e.SubExpr, env)
		if err != nil {
			return v, NewError(expr, err)
		}
		return v, nil
	case *parse.FuncExpr:
		f := reflect.ValueOf(func(expr *parse.FuncExpr, env *Env) Func {
			return func(args ...reflect.Value) (reflect.Value, error) {
				newenv := env.NewEnv()
				for i, arg := range expr.Args {
					newenv.Define(arg, args[i])
				}
				rr, err := Run(expr.Stmts, newenv)
				if err == ReturnError {
					err = nil
					rr = rr.Interface().(reflect.Value)
				}
				return rr, err
			}
		}(e, env))
		env.Define(e.Name, f)
		return f, nil
	//case *parse.LetExpr:
	//    rv, err := invokeExpr(e.Rhs, env)
	//    if err != nil {
	//        return rv, NewError(e, err)
	//    }
	//    if rv.Kind() == reflect.Interface {
	//        rv = rv.Elem()
	//    }
	//    return invokeLetExpr(e.Lhs, rv, env)
	case *parse.BinOpExpr:
		lhsV := NilValue
		rhsV := NilValue
		var err error

		lhsV, err = invokeExpr(e.Lhs, env)
		if err != nil {
			return lhsV, NewError(expr, err)
		}
		if lhsV.Kind() == reflect.Interface {
			lhsV = lhsV.Elem()
		}
		if e.Rhs != nil {
			rhsV, err = invokeExpr(e.Rhs, env)
			if err != nil {
				return rhsV, NewError(expr, err)
			}
			if rhsV.Kind() == reflect.Interface {
				rhsV = rhsV.Elem()
			}
		}
		switch e.Operator {
		case "+":
			if lhsV.Kind() == reflect.String || rhsV.Kind() == reflect.String {
				return reflect.ValueOf(toString(lhsV) + toString(rhsV)), nil
			}
			if (lhsV.Kind() == reflect.Array || lhsV.Kind() == reflect.Slice) && (rhsV.Kind() != reflect.Array && rhsV.Kind() != reflect.Slice) {
				return reflect.Append(lhsV, rhsV), nil
			}
			if (lhsV.Kind() == reflect.Array || lhsV.Kind() == reflect.Slice) && (rhsV.Kind() == reflect.Array || rhsV.Kind() == reflect.Slice) {
				return reflect.AppendSlice(lhsV, rhsV), nil
			}
			if lhsV.Kind() == reflect.Float64 || rhsV.Kind() == reflect.Float64 {
				return reflect.ValueOf(toFloat64(lhsV) + toFloat64(rhsV)), nil
			}
			return reflect.ValueOf(toInt64(lhsV) + toInt64(rhsV)), nil
		case "-":
			if lhsV.Kind() == reflect.Float64 || rhsV.Kind() == reflect.Float64 {
				return reflect.ValueOf(toFloat64(lhsV) - toFloat64(rhsV)), nil
			}
			return reflect.ValueOf(toInt64(lhsV) - toInt64(rhsV)), nil
		case "*":
			if lhsV.Kind() == reflect.String && (rhsV.Kind() == reflect.Int || rhsV.Kind() == reflect.Int32 || rhsV.Kind() == reflect.Int64) {
				return reflect.ValueOf(strings.Repeat(toString(lhsV), int(toInt64(rhsV)))), nil
			}
			if lhsV.Kind() == reflect.Float64 || rhsV.Kind() == reflect.Float64 {
				return reflect.ValueOf(toFloat64(lhsV) * toFloat64(rhsV)), nil
			}
			return reflect.ValueOf(toInt64(lhsV) * toInt64(rhsV)), nil
		case "/":
			return reflect.ValueOf(toFloat64(lhsV) / toFloat64(rhsV)), nil
		case "%":
			return reflect.ValueOf(toInt64(lhsV) % toInt64(rhsV)), nil
		case "==":
			return reflect.ValueOf(equal(lhsV, rhsV)), nil
		case "!=":
			return reflect.ValueOf(equal(lhsV, rhsV) == false), nil
		case ">":
			return reflect.ValueOf(toFloat64(lhsV) > toFloat64(rhsV)), nil
		case ">=":
			return reflect.ValueOf(toFloat64(lhsV) >= toFloat64(rhsV)), nil
		case "<":
			return reflect.ValueOf(toFloat64(lhsV) < toFloat64(rhsV)), nil
		case "<=":
			return reflect.ValueOf(toFloat64(lhsV) <= toFloat64(rhsV)), nil
		case "|":
			return reflect.ValueOf(toInt64(lhsV) | toInt64(rhsV)), nil
		case "||":
			if toBool(lhsV) {
				return lhsV, nil
			}
			return rhsV, nil
		case "&":
			return reflect.ValueOf(toInt64(lhsV) & toInt64(rhsV)), nil
		case "&&":
			if toBool(lhsV) {
				return rhsV, nil
			}
			return lhsV, nil
		default:
			return NilValue, NewStringError(expr, "Unknown operator")
		}
	case *parse.ConstExpr:
		switch e.Value {
		case "true":
			return reflect.ValueOf(true), nil
		case "false":
			return reflect.ValueOf(false), nil
		}
		return reflect.ValueOf(nil), nil
	//case *parse.CallExpr:
	//    f := NilValue

	//    if e.Func != nil {
	//        f = e.Func.(reflect.Value)
	//    } else {
	//        var err error
	//        ff, err := env.Get(e.Name)
	//        if err != nil {
	//            return f, err
	//        }
	//        f = ff
	//    }
	//    _, isReflect := f.Interface().(Func)

	//    args := []reflect.Value{}
	//    l := len(e.SubExprs)
	//    for i, expr := range e.SubExprs {
	//        arg, err := invokeExpr(expr, env)
	//        if err != nil {
	//            return arg, NewError(expr, err)
	//        }

	//        if i < f.Type().NumIn() {
	//            if !f.Type().IsVariadic() {
	//                it := f.Type().In(i)
	//                if arg.Kind().String() == "unsafe.Pointer" {
	//                    arg = reflect.New(it).Elem()
	//                }
	//                if arg.Kind() != it.Kind() && arg.IsValid() && arg.Type().ConvertibleTo(it) {
	//                    arg = arg.Convert(it)
	//                } else if arg.Kind() == reflect.Func {
	//                    if _, isFunc := arg.Interface().(Func); isFunc {
	//                        rfunc := arg
	//                        arg = reflect.MakeFunc(it, func(args []reflect.Value) []reflect.Value {
	//                            for i := range args {
	//                                args[i] = reflect.ValueOf(args[i])
	//                            }
	//                            if e.Go {
	//                                go func() {
	//                                    rfunc.Call(args)
	//                                }()
	//                                return []reflect.Value{}
	//                            }
	//                            var rets []reflect.Value
	//                            for _, v := range rfunc.Call(args)[:it.NumOut()] {
	//                                rets = append(rets, v.Interface().(reflect.Value))
	//                            }
	//                            return rets
	//                        })
	//                    }
	//                } else if !arg.IsValid() {
	//                    arg = reflect.Zero(it)
	//                }
	//            }
	//        }
	//        if !arg.IsValid() {
	//            arg = NilValue
	//        }

	//        if !isReflect {
	//            args = append(args, arg)
	//        } else {
	//            if arg.Kind() == reflect.Interface {
	//                arg = arg.Elem()
	//            }
	//            args = append(args, reflect.ValueOf(arg))
	//        }
	//    }
	//    ret := NilValue
	//    var err error
	//    fnc := func() {
	//        defer func() {
	//            if os.Getenv("ANKO_DEBUG") == "" {
	//                if ex := recover(); ex != nil {
	//                    if e, ok := ex.(error); ok {
	//                        err = e
	//                    } else {
	//                        err = errors.New(fmt.Sprint(ex))
	//                    }
	//                }
	//            }
	//        }()
	//        if f.Kind() == reflect.Interface {
	//            f = f.Elem()
	//        }
	//        rets := f.Call(args)
	//        if isReflect {
	//            ev := rets[1].Interface()
	//            if ev != nil {
	//                err = ev.(error)
	//            }
	//            ret = rets[0].Interface().(reflect.Value)
	//        } else {
	//            for i, expr := range e.SubExprs {
	//                if ae, ok := expr.(*parse.AddrExpr); ok {
	//                    if id, ok := ae.Expr.(*parse.IdentExpr); ok {
	//                        invokeLetExpr(id, args[i].Elem().Elem(), env)
	//                    }
	//                }
	//            }
	//            if f.Type().NumOut() == 1 {
	//                ret = rets[0]
	//            } else {
	//                var result []interface{}
	//                for _, r := range rets {
	//                    result = append(result, r.Interface())
	//                }
	//                ret = reflect.ValueOf(result)
	//            }
	//        }
	//    }
	//    if e.Go {
	//        go fnc()
	//        return NilValue, nil
	//    }
	//    fnc()
	//    if err != nil {
	//        return ret, NewError(expr, err)
	//    }
	//    return ret, nil
	default:
		return NilValue, NewStringError(expr, "Unknown expression")
	}
}

//////////////////////////////
// utils
//////////////////////////////

func toString(v reflect.Value) string {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	if !v.IsValid() {
		return "nil"
	}
	return fmt.Sprint(v.Interface())

}
func toBool(v reflect.Value) bool {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0.0
	case reflect.Int, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		if v.String() == "true" {
			return true
		}
		if toInt64(v) != 0 {
			return true
		}
	}
	return false
}

func toFloat64(v reflect.Value) float64 {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Int, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	}
	return 0.0
}

func isNil(v reflect.Value) bool {
	if !v.IsValid() || v.Kind().String() == "unsafe.Pointer" {
		return true
	}
	if (v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr) && v.IsNil() {
		return true
	}
	return false
}

func isNum(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		return true
	}
	return false

}

func equal(lhsV, rhsV reflect.Value) bool {
	lhsIsNil, rhsIsNil := isNil(lhsV), isNil(rhsV)
	if lhsIsNil && rhsIsNil {
		return true
	}
	if (!lhsIsNil && rhsIsNil) || (lhsIsNil && !rhsIsNil) {
		return false
	}
	return reflect.DeepEqual(lhsV, rhsV)
}
func toInt64(v reflect.Value) int64 {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return int64(v.Float())
	case reflect.Int, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.String:
		s := v.String()
		var i int64
		var err error
		if strings.HasPrefix(s, "0x") {
			i, err = strconv.ParseInt(s, 16, 64)
		} else {
			i, err = strconv.ParseInt(s, 10, 64)
		}
		if err == nil {
			return int64(i)
		}
	}
	return 0
}
