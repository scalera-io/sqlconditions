package sqlconditions

import (
	"fmt"
	"strings"
)

// Config is used as the central container holding all SQL operations with their conditions.
// It can be loaded from a YAML file.
type Config struct {
	Version    uint
	Operations map[string]OperationConfig `yaml:"operations"`
}

func StrSliceContains(ss []string, searched string) bool {
	for _, s := range ss {
		if s == searched {
			return true
		}
	}

	return false
}

func (c Config) GetOperation(opName string, searchedTagNames []string) (OperationParams, error) {
	var opParam OperationParams

	opConfig, ok := c.Operations[opName]
	if !ok {
		return opParam, ErrNotFound
	}

	if searchedTagNames == nil {
		searchedTagNames = []string{"default"}
	}

	for opTagsSpaceSeparated, op := range opConfig.VariantsByTag {
		// TODO design question: space separated or comma separated with space allowed before after comma ?
		opTags := strings.Split(opTagsSpaceSeparated, " ")

		for _, searchedTag := range searchedTagNames {
			// match, _ := HasRole(ctx, RoleName(role)); match {
			if StrSliceContains(opTags, searchedTag) {
				return *op, nil
			}
		}
	}

	return opParam, ErrNotFound
}

// OperationConfig holds one or many OperationParams
type OperationConfig struct {
	VariantsByTag map[string]*OperationParams `yaml:"variants"`
}

// OperationParams holds the parameters of a condition
type OperationParams struct {
	Joins []string

	Tokens `yaml:"condition"`

	CondExpr `yaml:"ignore"`
}

// FilterArgs is used to pass the set of available named argument to ToSQL methods
// so that a Condition can be rendered with or without certains subconditions
type FilterArgs map[string]any

// An ExprElt is a constituent of a CondExpr
type ExprElt interface {
	ToSQL(h ParseHint, argsMap FilterArgs) (string, error)
}

// CondExpr is a list of ExprElt : either a Condition or another sub CondExpr.
type CondExpr []ExprElt

func (se CondExpr) String() string {
	s := ""
	for _, exprElt := range se {
		s += fmt.Sprintf("%v\n", exprElt)
	}
	return s
}

func ToSQL(se CondExpr, args FilterArgs) (string, error) {
	h := ParseHint{}
	return se.ToSQL(h, args)
}

type ParseHint struct {
	ExprNotEmpty bool
}

// ToSQL renders a CondExpr to an SQL string
// If a Condition is set to if_present, it will be rendered only if its expected named argument is present in argsMap
func (se CondExpr) ToSQL(h ParseHint, argsMap FilterArgs) (string, error) {

	sql := ""

	for idx, exprElt := range se {
		if idx == 0 {
			h.ExprNotEmpty = false
		}

		s, err := exprElt.ToSQL(h, argsMap)
		if err != nil {
			return "", err
		}

		if s != "" {
			sql += s

			if _, ok := exprElt.(Condition); ok {
				h.ExprNotEmpty = true
			}
		}
	}
	return sql, nil
}

// Condition holds all the parameters needed to render it in a SQL expression
// Current implementation expects ArgName to be a named argument prefixed with a @ character.
type Condition struct {
	// Modality describes whether the condition is optional or mandatory
	// If set to "if_present" the condition is optional : it will be rendered only if a named argument
	// having the same name as ArgName is found at expression evaluation time
	// Else the condition is always rendered.
	Modality string

	// LinkOperator describes how the condition should be chained
	// Currently only AND or OR operators are supported
	LinkOperator string

	// ColumnName is the name of column tested in the condition
	ColumnName string

	// Operator used to render the condition (AND, OR, LIKE expr).
	// Current parser implementation expects the operator to not contain any space characters.
	Operator string

	// ArgName is the expected named argument
	ArgName string
}

func (c Condition) ToSQL(h ParseHint, argsMap FilterArgs) (string, error) {
	s := ""

	if argsMap == nil {
		return "", fmt.Errorf("invalid (nil) param")
	}

	if len(c.ArgName) < 2 {
		return "", fmt.Errorf("Condition ArgName invalid (too short): %v", c.ArgName)
	}

	if c.Modality == "if_present" {
		// remove @ prefix from argName
		expectedArgName := c.ArgName[1:]

		if _, found := argsMap[expectedArgName]; !found {
			return "", nil
		}
	}

	if c.LinkOperator != "" && h.ExprNotEmpty {
		s += " " + c.LinkOperator + " "
	}
	return s + fmt.Sprintf("%v %v %v", c.ColumnName, c.Operator, c.ArgName), nil
}

func (c Condition) String() string {
	return fmt.Sprintf("%v %v %v %v", c.Modality, c.ColumnName, c.Operator, c.ArgName)
}
