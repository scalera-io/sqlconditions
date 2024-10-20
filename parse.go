package sqlconditions

import (
	"fmt"
	"log"
	"strings"
)

type Tokens []interface{}

type ExprSeparator string

func (s ExprSeparator) ToSQL(argsMap FilterArgs) (string, error) {
	return string(s), nil
}

type ExprOperator string

func (s ExprOperator) ToSQL(argsMap FilterArgs) (string, error) {
	return " " + string(s) + " ", nil
}

// ParseCondition parses a string and returns a Condition (an ExprElt)
func ParseCondition(s string) (Condition, error) {
	c := Condition{}

	if strings.HasPrefix(s, "AND ") {
		s = strings.TrimPrefix(s, "AND ")
		c.LinkOperator = "AND"
	} else if strings.HasPrefix(s, "OR ") {
		s = strings.TrimPrefix(s, "OR ")
		c.LinkOperator = "OR"
	}

	if strings.HasPrefix(s, "if_present ") {
		s = strings.TrimPrefix(s, "if_present ")
		c.Modality = "if_present"
	} else {
		c.Modality = "must"
	}

	condParts := strings.Split(s, " ")

	if len(condParts) != 3 {
		return c, fmt.Errorf("malformed, expected: columnName operator namedArg")
	}

	c.ColumnName = condParts[0]
	c.Operator = condParts[1]
	c.ArgName = condParts[2]

	return c, nil
}

// transform a YAML condition into an expr
func (expr Tokens) Parse() (CondExpr, error) {
	var exprElts CondExpr

	for idx, exprElt := range expr {
		switch v := exprElt.(type) {
		case string:
			if v == "(" || v == ")" {
				exprElts = append(exprElts, ExprSeparator(v))
				continue
			}

			if strings.TrimSpace(v) == "AND" {
				exprElts = append(exprElts, ExprOperator(v))
				continue
			}

			if strings.TrimSpace(v) == "OR" {
				exprElts = append(exprElts, ExprOperator(v))
				continue
			}

			c, err := ParseCondition(v)
			if err != nil {
				log.Fatalf("ParseCondition err: %v on: <%v>", err, v)
			}

			exprElts = append(exprElts, c)

		case []interface{}:
			var subExpr Tokens = Tokens(v)
			subExprElts, err := subExpr.Parse()
			if err != nil {
				log.Fatalf("ParseExpr #v err: %v on: <%v>", idx, err, v)
				return nil, err
			}
			exprElts = append(exprElts, subExprElts...)

		default:
			log.Fatalln("Tokens.Parse unhandled type: %T %v", exprElt, exprElt)
		}
	}

	return exprElts, nil
}
