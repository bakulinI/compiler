package parser

func GrammarRules() string {
	return `Program
├── KEYWORD package
├── IDENTIFIER
├── ImportDecl*
│   ├── KEYWORD import
│   └── CONSTANT_STR
└── FuncDecl+
    ├── KEYWORD func
    ├── IDENTIFIER
    ├── ParameterList
    │   ├── DELIMITER (
    │   ├── Param (DELIMITER , Param)*
    │   │   ├── IDENTIFIER
    │   │   └── Type
    │   └── DELIMITER )
    ├── ReturnType?
    └── Block
        ├── DELIMITER {
        ├── Statement*
        │   ├── VarDecl: KEYWORD var IDENTIFIER Type (OPERATOR = Expression)?
        │   ├── AssignStmt: IDENTIFIER (OPERATOR = | OPERATOR :=) Expression
        │   ├── ReturnStmt: KEYWORD return Expression?
        │   ├── IfStmt: KEYWORD if Expression Block (KEYWORD else Block)?
        │   └── ExprStmt: Expression
        └── DELIMITER }

Expression
├── LogicalOr: LogicalAnd (OPERATOR || LogicalAnd)*
├── LogicalAnd: Equality (OPERATOR && Equality)*
├── Equality: Comparison ((OPERATOR == | OPERATOR !=) Comparison)*
├── Comparison: Term ((OPERATOR < | <= | > | >=) Term)*
├── Term: Factor ((OPERATOR + | -) Factor)*
├── Factor: Unary ((OPERATOR * | / | %) Unary)*
└── Primary
    ├── IDENTIFIER
    ├── CONSTANT_INT | CONSTANT_REAL | CONSTANT_STR | CONSTANT_BOOL
    ├── DELIMITER ( Expression DELIMITER )
    ├── SelectorExpr: Primary DELIMITER . IDENTIFIER
    └── CallExpr: Primary DELIMITER ( ArgumentList? DELIMITER )`
}
