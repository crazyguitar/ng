// Copyright 2015 The Numgrad Authors. All rights reserved.
// See the LICENSE file for rights to use this source code.

package parser

import (
	"fmt"
	"math/big"
	"testing"

	"numgrad.io/lang/expr"
	"numgrad.io/lang/stmt"
	"numgrad.io/lang/tipe"
	"numgrad.io/lang/token"
)

type parserTest struct {
	input string
	want  expr.Expr
}

var parserTests = []parserTest{
	{"foo", &expr.Ident{"foo"}},
	{"x + y", &expr.Binary{token.Add, &expr.Ident{"x"}, &expr.Ident{"y"}}},
	{
		"x + y + 9",
		&expr.Binary{
			token.Add,
			&expr.Binary{token.Add, &expr.Ident{"x"}, &expr.Ident{"y"}},
			&expr.BasicLiteral{big.NewInt(9)},
		},
	},
	{
		"x + (y + 7)",
		&expr.Binary{
			token.Add,
			&expr.Ident{"x"},
			&expr.Unary{
				Op: token.LeftParen,
				Expr: &expr.Binary{
					token.Add,
					&expr.Ident{"y"},
					&expr.BasicLiteral{big.NewInt(7)},
				},
			},
		},
	},
	{
		"x + y * z",
		&expr.Binary{
			token.Add,
			&expr.Ident{"x"},
			&expr.Binary{token.Mul, &expr.Ident{"y"}, &expr.Ident{"z"}},
		},
	},
	{
		"quit()",
		&expr.Call{Func: &expr.Ident{Name: "quit"}},
	},
	{
		"foo(4)",
		&expr.Call{
			Func: &expr.Ident{Name: "foo"},
			Args: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(4)}},
		},
	},
	{
		"min(1, 2)",
		&expr.Call{
			Func: &expr.Ident{Name: "min"},
			Args: []expr.Expr{
				&expr.BasicLiteral{Value: big.NewInt(1)},
				&expr.BasicLiteral{Value: big.NewInt(2)},
			},
		},
	},
	{
		"func() integer { return 7 }",
		&expr.FuncLiteral{
			Type: &tipe.Func{Results: &tipe.Tuple{Elems: []tipe.Type{tipe.Integer}}},
			Body: &stmt.Block{[]stmt.Stmt{
				&stmt.Return{Exprs: []expr.Expr{&expr.BasicLiteral{big.NewInt(7)}}},
			}},
		},
	},
	{
		"func(x, y val) (r0 val, r1 val) { return x, y }",
		&expr.FuncLiteral{
			Type: &tipe.Func{
				Params: &tipe.Tuple{Elems: []tipe.Type{
					&tipe.Unresolved{Name: "val"},
					&tipe.Unresolved{Name: "val"},
				}},
				Results: &tipe.Tuple{Elems: []tipe.Type{
					&tipe.Unresolved{Name: "val"},
					&tipe.Unresolved{Name: "val"},
				}},
			},
			ParamNames:  []string{"x", "y"},
			ResultNames: []string{"r0", "r1"},
			Body: &stmt.Block{[]stmt.Stmt{
				&stmt.Return{Exprs: []expr.Expr{
					&expr.Ident{Name: "x"},
					&expr.Ident{Name: "y"},
				}},
			}},
		},
	},
	{
		`func() int64 {
			x := 7
			return x
		}`,
		&expr.FuncLiteral{
			Type:        &tipe.Func{Results: &tipe.Tuple{Elems: []tipe.Type{tipe.Int64}}},
			ResultNames: []string{""},
			Body: &stmt.Block{[]stmt.Stmt{
				&stmt.Assign{
					Left:  []expr.Expr{&expr.Ident{"x"}},
					Right: []expr.Expr{&expr.BasicLiteral{big.NewInt(7)}},
				},
				&stmt.Return{Exprs: []expr.Expr{&expr.Ident{"x"}}},
			}},
		},
	},
	{
		`func() int64 {
			if x := 9; x > 3 {
				return x
			} else {
				return 1-x
			}
		}`,
		&expr.FuncLiteral{
			Type:        &tipe.Func{Results: &tipe.Tuple{Elems: []tipe.Type{tipe.Int64}}},
			ResultNames: []string{""},
			Body: &stmt.Block{[]stmt.Stmt{&stmt.If{
				Init: &stmt.Assign{
					Left:  []expr.Expr{&expr.Ident{"x"}},
					Right: []expr.Expr{&expr.BasicLiteral{big.NewInt(9)}},
				},
				Cond: &expr.Binary{
					Op:    token.Greater,
					Left:  &expr.Ident{"x"},
					Right: &expr.BasicLiteral{big.NewInt(3)},
				},
				Body: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{&expr.Ident{"x"}}},
				}},
				Else: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{
						&expr.Binary{
							Op:    token.Sub,
							Left:  &expr.BasicLiteral{big.NewInt(1)},
							Right: &expr.Ident{"x"},
						},
					}},
				}},
			}}},
		},
	},
	{
		"func(x val) val { return 3+x }(1)",
		&expr.Call{
			Func: &expr.FuncLiteral{
				Type: &tipe.Func{
					Params:  &tipe.Tuple{Elems: []tipe.Type{&tipe.Unresolved{Name: "val"}}},
					Results: &tipe.Tuple{Elems: []tipe.Type{&tipe.Unresolved{Name: "val"}}},
				},
				ParamNames:  []string{""},
				ResultNames: []string{""},
				Body: &stmt.Block{[]stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{
						&expr.Binary{
							Op:    token.Add,
							Left:  &expr.BasicLiteral{big.NewInt(3)},
							Right: &expr.Ident{"x"},
						},
					}},
				}},
			},
			Args: []expr.Expr{&expr.BasicLiteral{big.NewInt(1)}},
		},
	},
	{
		"func() { x = -x }",
		&expr.FuncLiteral{
			Type: &tipe.Func{},
			Body: &stmt.Block{[]stmt.Stmt{&stmt.Assign{
				Left:  []expr.Expr{&expr.Ident{"x"}},
				Right: []expr.Expr{&expr.Unary{Op: token.Sub, Expr: &expr.Ident{"x"}}},
			}}},
		},
	},
	{"x.y.z", &expr.Selector{&expr.Selector{&expr.Ident{"x"}, &expr.Ident{"y"}}, &expr.Ident{"z"}}},
	{"y * /* comment */ z", &expr.Binary{token.Mul, &expr.Ident{"y"}, &expr.Ident{"z"}}},
	//TODO{"y * z//comment", &expr.Binary{token.Mul, &expr.Ident{"y"}, &expr.Ident{"z"}}},
	{`"hello"`, &expr.BasicLiteral{"hello"}},
	{`"hello \"numgrad\""`, &expr.BasicLiteral{`hello \"numgrad\"`}},
	//TODO{`"\""`, &expr.BasicLiteral{`"\""`}}
	{"x[4]", &expr.TableIndex{Expr: &expr.Ident{"x"}, Cols: expr.Range{Exact: &expr.BasicLiteral{big.NewInt(4)}}}},
	{"x[1+2]", &expr.TableIndex{
		Expr: &expr.Ident{"x"},
		Cols: expr.Range{
			Exact: &expr.Binary{Op: token.Add,
				Left:  &expr.BasicLiteral{big.NewInt(1)},
				Right: &expr.BasicLiteral{big.NewInt(2)},
			},
		},
	}},
	{"x[1:3]", &expr.TableIndex{Expr: &expr.Ident{"x"}, Cols: expr.Range{Start: &expr.BasicLiteral{big.NewInt(1)}, End: &expr.BasicLiteral{big.NewInt(3)}}}},
	{"x[1:]", &expr.TableIndex{Expr: &expr.Ident{"x"}, Cols: expr.Range{Start: &expr.BasicLiteral{big.NewInt(1)}}}},
	{"x[:3]", &expr.TableIndex{Expr: &expr.Ident{"x"}, Cols: expr.Range{End: &expr.BasicLiteral{big.NewInt(3)}}}},
	{"x[:]", &expr.TableIndex{Expr: &expr.Ident{"x"}}},
	{"x[,:]", &expr.TableIndex{Expr: &expr.Ident{"x"}}},
	{"x[:,:]", &expr.TableIndex{Expr: &expr.Ident{"x"}}},
	{`x["C1"|"C2"]`, &expr.TableIndex{Expr: &expr.Ident{"x"}, ColNames: []string{"C1", "C2"}}},
	{`x["C1",1:]`, &expr.TableIndex{
		Expr:     &expr.Ident{"x"},
		ColNames: []string{"C1"},
		Rows:     expr.Range{Start: &expr.BasicLiteral{big.NewInt(1)}},
	}},
	{"x[1:3,5:7]", &expr.TableIndex{
		Expr: &expr.Ident{"x"},
		Cols: expr.Range{Start: &expr.BasicLiteral{big.NewInt(1)}, End: &expr.BasicLiteral{big.NewInt(3)}},
		Rows: expr.Range{Start: &expr.BasicLiteral{big.NewInt(5)}, End: &expr.BasicLiteral{big.NewInt(7)}},
	}},
	{"[|]num{}", &expr.TableLiteral{Type: &tipe.Table{&tipe.Unresolved{Name: "num"}}}},
	{"[|]num{{0, 1, 2}}", &expr.TableLiteral{
		Type: &tipe.Table{&tipe.Unresolved{Name: "num"}},
		Rows: [][]expr.Expr{{basic(0), basic(1), basic(2)}},
	}},
	{`[|]num{{|"Col1"|}, {1}, {2}}`, &expr.TableLiteral{
		Type:     &tipe.Table{&tipe.Unresolved{Name: "num"}},
		ColNames: []expr.Expr{basic("Col1")},
		Rows:     [][]expr.Expr{{basic(1)}, {basic(2)}},
	}},
}

func TestParseExpr(t *testing.T) {
	for _, test := range parserTests {
		fmt.Printf("Parsing %q\n", test.input)
		s, err := ParseStmt([]byte(test.input))
		if err != nil {
			t.Errorf("ParseExpr(%q): error: %v", test.input, err)
			continue
		}
		if s == nil {
			t.Errorf("ParseExpr(%q): nil stmt", test.input)
			continue
		}
		got := s.(*stmt.Simple).Expr
		if !EqualExpr(got, test.want) {
			t.Errorf("ParseExpr(%q):\n%v", test.input, DiffExpr(test.want, got))
		}
	}
}

type stmtTest struct {
	input string
	want  stmt.Stmt
}

var stmtTests = []stmtTest{
	{"for {}", &stmt.For{Body: &stmt.Block{}}},
	{"for ;; {}", &stmt.For{Body: &stmt.Block{}}},
	{"for true {}", &stmt.For{Cond: &expr.Ident{"true"}, Body: &stmt.Block{}}},
	{"for ; true; {}", &stmt.For{Cond: &expr.Ident{"true"}, Body: &stmt.Block{}}},
	{"for range x {}", &stmt.Range{Expr: &expr.Ident{"x"}, Body: &stmt.Block{}}},
	{"for k, v := range x {}", &stmt.Range{
		Key:  &expr.Ident{"k"},
		Val:  &expr.Ident{"v"},
		Expr: &expr.Ident{"x"},
		Body: &stmt.Block{},
	}},
	{"for k := range x {}", &stmt.Range{
		Key:  &expr.Ident{"k"},
		Expr: &expr.Ident{"x"},
		Body: &stmt.Block{},
	}},
	{
		"for i := 0; i < 10; i++ { x = i }",
		&stmt.For{
			Init: &stmt.Assign{
				Decl:  true,
				Left:  []expr.Expr{&expr.Ident{"i"}},
				Right: []expr.Expr{&expr.BasicLiteral{big.NewInt(0)}},
			},
			Cond: &expr.Binary{
				Op:    token.Less,
				Left:  &expr.Ident{"i"},
				Right: &expr.BasicLiteral{big.NewInt(10)},
			},
			Post: &stmt.Assign{
				Left: []expr.Expr{&expr.Ident{"i"}},
				Right: []expr.Expr{
					&expr.Binary{
						Op:    token.Add,
						Left:  &expr.Ident{"i"},
						Right: &expr.BasicLiteral{big.NewInt(1)},
					},
				},
			},
			Body: &stmt.Block{Stmts: []stmt.Stmt{&stmt.Assign{
				Left:  []expr.Expr{&expr.Ident{"x"}},
				Right: []expr.Expr{&expr.Ident{"i"}},
			}}},
		},
	},
	{"const x = 4", &stmt.Const{Name: "x", Value: &expr.BasicLiteral{big.NewInt(4)}}},
	{
		"const x int64 = 4",
		&stmt.Const{
			Name:  "x",
			Type:  tipe.Int64,
			Value: &expr.BasicLiteral{big.NewInt(4)},
		},
	},
	{
		`type a class {
			x integer
			y [|]int64

			func (a) f(x integer) integer {
				return a.x
			}
		}
		`,
		&stmt.ClassDecl{
			Name: "a",
			Type: &tipe.Class{
				Tags: []string{"x", "y", "f"},
				Fields: []tipe.Type{
					tipe.Integer,
					&tipe.Table{tipe.Int64},
					&tipe.Func{
						Params:  &tipe.Tuple{Elems: []tipe.Type{tipe.Integer}},
						Results: &tipe.Tuple{Elems: []tipe.Type{tipe.Integer}},
					},
				},
			},
			Methods: []*expr.FuncLiteral{{
				Name:            "f",
				ReceiverName:    "a",
				PointerReceiver: true,
				Type: &tipe.Func{
					Params:  &tipe.Tuple{Elems: []tipe.Type{tipe.Integer}},
					Results: &tipe.Tuple{Elems: []tipe.Type{tipe.Integer}},
				},
				ParamNames: []string{"x"},
				Body: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{&expr.Selector{
						Left:  &expr.Ident{"a"},
						Right: &expr.Ident{"x"},
					}}},
				}},
			}},
		},
	},
	{"x.y", &stmt.Simple{&expr.Selector{&expr.Ident{"x"}, &expr.Ident{"y"}}}},
}

func TestParseStmt(t *testing.T) {
	for _, test := range stmtTests {
		fmt.Printf("Parsing stmt %q\n", test.input)
		got, err := ParseStmt([]byte(test.input))
		if err != nil {
			t.Errorf("ParseStmt(%q): error: %v", test.input, err)
			continue
		}
		if got == nil {
			t.Errorf("ParseStmt(%q): nil stmt", test.input)
			continue
		}
		if !EqualStmt(got, test.want) {
			t.Errorf("ParseStmt(%q):\n%v", test.input, DiffStmt(test.want, got))
		}
	}
}

func basic(x interface{}) *expr.BasicLiteral {
	switch x := x.(type) {
	case int:
		return &expr.BasicLiteral{big.NewInt(int64(x))}
	case int64:
		return &expr.BasicLiteral{big.NewInt(x)}
	case string:
		return &expr.BasicLiteral{x}
	default:
		panic(fmt.Sprintf("unknown basic %v (%T)", x, x))
	}
}
