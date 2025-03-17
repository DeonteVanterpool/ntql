package tbql

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
)

var date_types = []string{"due_date", "do_date", "created_at", "updated_at", "completed_at"}
var bool_types = []string{"hide_from_calendar", "all_day", "completed"}
var numeric_types = []string{"priority"}
var string_types = []string{"title", "description"}

type QueryExpr interface {
	// ToSQL converts the query expression to a SQL string.
	ToSQL() (string, error)
    String() string
}

type Operator string // I chose string over iota to increase resilience to changes, especially for communication between the frontend and backend

const (
    // Binary operators
    OperatorAnd Operator = "AND"
    OperatorOr  Operator = "OR"
    OperatorXor Operator = "XOR"

    // Unary operators
    OperatorNot Operator = "NOT"

    // Comparison operators
    OperatorEq  Operator = "equals"
    OperatorNeq Operator = "notEquals"
    OperatorGt  Operator = "greaterThan"
    OperatorLt  Operator = "lessThan"
    OperatorGte Operator = "greaterThanOrEquals"
    OperatorLte Operator = "lessThanOrEquals"

    // String operators
    OperatorCnt Operator = "contains"
    OperatorSw  Operator = "startsWith"
    OperatorEw  Operator = "endsWith"
)

func NewOperator(s string) (Operator, error) {
    if s == "AND" {
        return OperatorAnd, nil
    } else if s == "OR" {
        return OperatorOr, nil
    } else if s == "XOR" {
        return OperatorXor, nil
    } else if s == "NOT" {
        return OperatorNot, nil
    } else if s == "equals" {
        return OperatorEq, nil
    } else if s == "notEquals" {
        return OperatorNeq, nil
    } else if s == "greaterThan" || s == "after" {
        return OperatorGt, nil
    } else if s == "lessThan" || s == "before" {
        return OperatorLt, nil
    } else if s == "greaterthanorequals" {
        return OperatorGte, nil
    } else if s == "lessthanorequals" {
        return OperatorLte, nil
    } else if s == "contains" {
        return OperatorCnt, nil
    } else if s == "startswith" {
        return OperatorSw, nil
    } else if s == "endsWith" {
        return OperatorEw, nil
    } else {
        return "", errors.New("invalid operator: " + s)
    }
}

func (o Operator) ToStr() string {
    return string(o)
}

// QueryBinaryOp represents a binary operation in a query.
type QueryBinaryOp struct {
	Left  QueryExpr `json:"left"`
	Right QueryExpr `json:"right"`
    Operator    Operator `json:"op"`
}

func (q *QueryBinaryOp) String() string {
    return q.Left.String() + " " + q.Operator.ToStr() + " " + q.Right.String()
}

func (q *QueryBinaryOp) ToSQL() (string, error) {
	left, err := q.Left.ToSQL()
	if err != nil {
		return "", err
	}
	right, err := q.Right.ToSQL()
	if err != nil {
		return "", err
	}
	switch q.Operator {
    case OperatorAnd:
		return "(" + left + " AND " + right + ")", nil
    case OperatorOr:
		return "(" + left + " OR " + right + ")", nil
    case OperatorXor:
		return "(" + left + " XOR " + right + ")", nil
	default:
        return "", errors.New("invalid operator: " + q.Operator.ToStr())
	}
}

// QueryUnaryOp represents a unary operation in a query.
type QueryUnaryOp struct {
	Operand  QueryExpr `json:"operand"`
    Operator Operator `json:"operator"`
}

func (q *QueryUnaryOp) ToSQL() (string, error) {
	operand, err := q.Operand.ToSQL()
	if err != nil {
		return "", err
	}
	switch q.Operator {
    case OperatorNot:
		return "(NOT " + "(" + operand + "))", nil
	default:
        return "", errors.New("invalid operator: " + q.Operator.ToStr())
	}
}

func (q *QueryUnaryOp) String() string {
    return q.Operator.ToStr() + " " + q.Operand.String()
}

// QueryCondition represents a condition in a query.
type QueryCondition struct {
	Field    string `json:"field"`
    Operator Operator `json:"operator"`
	Value    string `json:"value"`
}

func (q *QueryCondition) String() string {
    return q.Field + " " + q.Operator.ToStr() + " " + q.Value
}

// TODO: Input sanitization
// Convert the query condition to a SQL string.
func (c *QueryCondition) ToSQL() (string, error) {
	if c.Field == "tag" {
		_, err := strconv.Atoi(c.Value)
		if err != nil {
			switch c.Operator {
            case OperatorEq:
				return "tag_id = (SELECT id FROM atomic_tags WHERE title = '" + c.Value + "')", nil
            case OperatorNeq:
				return "tag_id = (SELECT id FROM atomic_tags WHERE title != '" + c.Value + "')", nil
            case OperatorCnt:
				return "tag_id = (SELECT id FROM atomic_tags WHERE title LIKE '%" + c.Value + "%')", nil
            case OperatorSw:
				return "tag_id = (SELECT id FROM atomic_tags WHERE title LIKE '" + c.Value + "%')", nil
            case OperatorEw:
				return "tag_id = (SELECT id FROM atomic_tags WHERE title LIKE '%" + c.Value + "')", nil
			default:
                return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
			}
		} else {
			switch c.Operator {
            case OperatorEq:
				return "tag_id = (SELECT id FROM atomic_tags WHERE id = " + c.Value + ")", nil
            case OperatorNeq:
				return "tag_id = (SELECT id FROM atomic_tags WHERE id != " + c.Value + ")", nil
			default:
                return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
			}
		}
	}

	if c.Field == "completed" {
		if (c.Value == "true" && c.Operator == OperatorEq) || (c.Value == "false" && c.Operator == OperatorNeq) {
			// return if completed_at before now
			return "completed_at < NOW()", nil
		} else if (c.Value == "true" && c.Operator == OperatorNeq) || (c.Value == "false" && c.Operator == OperatorEq) {
			// return if completed_at after now or NULL
			return "completed_at > NOW() OR completed_at IS NULL", nil
		} else {
            return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
	}
	if slices.Contains(date_types, c.Field) {
		// check if datetime is in the ISO 8601 format
		if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}Z)?$`).MatchString(c.Value) {
            return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
		switch c.Operator {
		case OperatorEq:
			return c.Field + " = '" + c.Value + "'", nil
        case OperatorNeq:
			return c.Field + " != '" + c.Value + "'", nil
        case OperatorGt:
			return c.Field + " > '" + c.Value + "'", nil
        case OperatorLt:
			return c.Field + " < '" + c.Value + "'", nil
        case OperatorGte:
			return c.Field + " >= '" + c.Value + "'", nil
        case OperatorLte:
			return c.Field + " <= '" + c.Value + "'", nil
		default:
            return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	} else if slices.Contains(bool_types, c.Field) {
		switch c.Operator {
        case OperatorEq:
			return c.Field + " = " + c.Value, nil
        case OperatorNeq:
			return c.Field + " != " + c.Value, nil
		default:
            return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	} else if slices.Contains(string_types, c.Field) {
		if !regexp.MustCompile(`^[a-zA-Z0-9\-/: ]+$`).MatchString(c.Value) {
            return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
		switch c.Operator {
        case OperatorEq:
			return c.Field + " = '" + c.Value + "'", nil
        case OperatorNeq:
			return c.Field + " != '" + c.Value + "'", nil
        case OperatorCnt:
			return c.Field + " LIKE '%" + c.Value + "%'", nil
        case OperatorSw:
			return c.Field + " LIKE '" + c.Value + "%'", nil
        case OperatorEw:
			return c.Field + " LIKE '%" + c.Value + "'", nil
		default:
            return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	} else if slices.Contains(numeric_types, c.Field) {
		_, err := strconv.Atoi(c.Value)
		if err != nil {
            return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
		switch c.Operator {
        case OperatorEq:
			return c.Field + " = " + c.Value, nil
        case OperatorNeq:
			return c.Field + " != " + c.Value, nil
        case OperatorGt:
			return c.Field + " > " + c.Value, nil
        case OperatorLt:
			return c.Field + " < " + c.Value, nil
        case OperatorGte:
			return c.Field + " >= " + c.Value, nil
        case OperatorLte:
			return c.Field + " <= " + c.Value, nil
		default:
            return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	} else {
		return "", errors.New("invalid field")
	}
}

func NewQueryAnd(left QueryExpr, right QueryExpr) *QueryBinaryOp {
	return &QueryBinaryOp{Left: left, Right: right, Operator: OperatorAnd}
}

func NewQueryOr(left QueryExpr, right QueryExpr) *QueryBinaryOp {
	return &QueryBinaryOp{Left: left, Right: right, Operator: OperatorOr}
}

func NewQueryNot(expr QueryExpr) *QueryUnaryOp {
	return &QueryUnaryOp{Operand: expr, Operator: OperatorNot}
}

func NewQueryCondition(subject string, verb string, value string) (*QueryCondition, error) {
	op, err := NewOperator(verb)
	if err != nil {
		return nil, err
	}
	return &QueryCondition{Field: subject, Operator: op, Value: value}, nil
}

func buildQueryConditionFromMap(m map[string]interface{}) (*QueryCondition, error) {
	field, ok := m["field"].(string)
	if !ok {
		return nil, errors.New("invalid field")
	}
	operator, ok := m["operator"].(string)
	if !ok {
        return nil, errors.New("invalid operator: " + operator + " for field: " + field)
	}
	value, ok := m["value"].(string)
	if !ok {
        return nil, errors.New("invalid value: " + value + " for field: " + field)
	}
    op, err := NewOperator(operator)
    if err != nil {
        return nil, err
    }
	return &QueryCondition{Field: field, Operator: op, Value: value}, nil
}

func buildQueryBinaryOpFromMap(m map[string]interface{}) (*QueryBinaryOp, error) {
	left, ok := m["left"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid left operand")
	}
	left_expr, err := BuildQueryExprFromMap(left)
	if err != nil {
		return nil, err
	}
	right, ok := m["right"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid right operand")
	}
	right_expr, err := BuildQueryExprFromMap(right)
	if err != nil {
		return nil, err
	}
	op, ok := m["op"].(string)
	if !ok {
        return nil, errors.New("invalid operator")
	}

    operator, err := NewOperator(op)
    if err != nil {
        return nil, err
    }
	return &QueryBinaryOp{Left: left_expr, Right: right_expr, Operator: operator}, nil
}

func buildQueryUnaryOpFromMap(m map[string]interface{}) (*QueryUnaryOp, error) {
	operator, ok := m["operator"].(string)
	if !ok {
		return nil, errors.New("invalid operator")
	}
	operand, ok := m["operand"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid operand")
	}
	operand_expr, err := BuildQueryExprFromMap(operand)
	if err != nil {
		return nil, err
	}
    op, err := NewOperator(operator)
    if err != nil {
        return nil, err
    }
	return &QueryUnaryOp{Operator: op, Operand: operand_expr}, nil
}

func BuildQueryExprFromMap(m map[string]interface{}) (QueryExpr, error) {
	if len(m) == 0 {
		return nil, errors.New("empty query")
	}

	// check if it's a condition
	condition, err := buildQueryConditionFromMap(m)
	if err == nil {
		return condition, nil
	}

	// check if it's a binary operation
	binary_op, err := buildQueryBinaryOpFromMap(m)
	if err == nil {
		left, ok := m["left"].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid left operand")
		}
		right, ok := m["right"].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid right operand")
		}
		left_expr, err := BuildQueryExprFromMap(left)
		if err != nil {
			return nil, err
		}
		right_expr, err := BuildQueryExprFromMap(right)
		if err != nil {
			return nil, err
		}
		binary_op.Left = left_expr
		binary_op.Right = right_expr

		return binary_op, nil
	}

	// check if it's a unary operation
	unary_op, err := buildQueryUnaryOpFromMap(m)
	if err == nil {
		operand, ok := m["operand"].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid operand")
		}
		operand_expr, err := BuildQueryExprFromMap(operand)
		if err != nil {
			return nil, err
		}
		unary_op.Operand = operand_expr

		return unary_op, nil
	}

	return nil, errors.New("invalid query")
}

