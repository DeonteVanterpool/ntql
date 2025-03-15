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
}

// QueryBinaryOp represents a binary operation in a query.
type QueryBinaryOp struct {
	Left  QueryExpr `json:"left"`
	Right QueryExpr `json:"right"`
	Op    string    `json:"op"`
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
	switch q.Op {
	case "&":
		return "(" + left + " AND " + right + ")", nil
	case "|":
		return "(" + left + " OR " + right + ")", nil
	case "^":
		return "(" + left + " XOR " + right + ")", nil
	default:
		return "", errors.New("invalid operator")
	}
}

// QueryUnaryOp represents a unary operation in a query.
type QueryUnaryOp struct {
	Operator string    `json:"operator"`
	Operand  QueryExpr `json:"operand"`
}

func (q *QueryUnaryOp) ToSQL() (string, error) {
	operand, err := q.Operand.ToSQL()
	if err != nil {
		return "", err
	}
	switch q.Operator {
	case "!":
		return "(NOT " + "(" + operand + "))", nil
	default:
		return "", errors.New("invalid operator")
	}
}

// QueryCondition represents a condition in a query.
type QueryCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func (c *QueryCondition) ToSQL() (string, error) {
	if c.Field == "tag" {
		_, err := strconv.Atoi(c.Value)
		if err != nil {
			switch c.Operator {
			case "=":
				return "tag_id = (SELECT id FROM atomic_tags WHERE title = '" + c.Value + "')", nil
			case "!=":
				return "tag_id = (SELECT id FROM atomic_tags WHERE title != '" + c.Value + "')", nil
			case "contains":
				return "tag_id = (SELECT id FROM atomic_tags WHERE title LIKE '%" + c.Value + "%')", nil
			case "startswith":
				return "tag_id = (SELECT id FROM atomic_tags WHERE title LIKE '" + c.Value + "%')", nil
			case "endswith":
				return "tag_id = (SELECT id FROM atomic_tags WHERE title LIKE '%" + c.Value + "')", nil
			default:
				return "", errors.New("invalid operator")
			}
		} else {
			switch c.Operator {
			case "=":
				return "tag_id = (SELECT id FROM atomic_tags WHERE id = " + c.Value + ")", nil
			case "!=":
				return "tag_id = (SELECT id FROM atomic_tags WHERE id != " + c.Value + ")", nil
			default:
				return "", errors.New("invalid operator")
			}
		}
	}

	if c.Field == "completed" {
		if (c.Value == "true" && c.Operator == "=") || (c.Value == "false" && c.Operator == "!=") {
            // return if completed_at before now
            return "completed_at < NOW()", nil
		} else if (c.Value == "true" && c.Operator == "!=") || (c.Value == "false" && c.Operator == "=") {
            // return if completed_at after now or NULL
            return "completed_at > NOW() OR completed_at IS NULL", nil
		} else {
			return "", errors.New("invalid value")
		}
	}
	if slices.Contains(date_types, c.Field) {
        // check if datetime is in the ISO 8601 format
        if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`).MatchString(c.Value) {
            return "", errors.New("invalid value")
        }
		switch c.Operator {
		case "=":
			return c.Field + " = '" + c.Value + "'", nil
		case "!=":
			return c.Field + " != '" + c.Value + "'", nil
		case ">":
			return c.Field + " > '" + c.Value + "'", nil
		case "<":
			return c.Field + " < '" + c.Value + "'", nil
		case ">=":
			return c.Field + " >= '" + c.Value + "'", nil
		case "<=":
			return c.Field + " <= '" + c.Value + "'", nil
		default:
			return "", errors.New("invalid operator")
		}
	} else if slices.Contains(bool_types, c.Field) {
		switch c.Operator {
		case "=":
			return c.Field + " = " + c.Value, nil
		case "!=":
			return c.Field + " != " + c.Value, nil
		default:
			return "", errors.New("invalid operator")
		}
	} else if slices.Contains(string_types, c.Field) {
        if !regexp.MustCompile(`^[a-zA-Z0-9\-/: ]+$`).MatchString(c.Value) {
            return "", errors.New("invalid value")
        }
		switch c.Operator {
		case "=":
			return c.Field + " = '" + c.Value + "'", nil
		case "!=":
			return c.Field + " != '" + c.Value + "'", nil
		case "contains":
			return c.Field + " LIKE '%" + c.Value + "%'", nil
		case "startswith":
			return c.Field + " LIKE '" + c.Value + "%'", nil
		case "endswith":
			return c.Field + " LIKE '%" + c.Value + "'", nil
		default:
			return "", errors.New("invalid operator")
		}
	} else if slices.Contains(numeric_types, c.Field) {
		_, err := strconv.Atoi(c.Value)
		if err != nil {
			return "", errors.New("invalid value")
		}
		switch c.Operator {
		case "=":
			return c.Field + " = " + c.Value, nil
		case "!=":
			return c.Field + " != " + c.Value, nil
		case ">":
			return c.Field + " > " + c.Value, nil
		case "<":
			return c.Field + " < " + c.Value, nil
		case ">=":
			return c.Field + " >= " + c.Value, nil
		case "<=":
			return c.Field + " <= " + c.Value, nil
		default:
			return "", errors.New("invalid operator")
		}
	} else {
		return "", errors.New("invalid field")
	}
}

func buildQueryConditionFromMap(m map[string]interface{}) (*QueryCondition, error) {
	field, ok := m["field"].(string)
	if !ok {
		return nil, errors.New("invalid field")
	}
	operator, ok := m["operator"].(string)
	if !ok {
		return nil, errors.New("invalid operator")
	}
	value, ok := m["value"].(string)
	if !ok {
		return nil, errors.New("invalid value")
	}
	return &QueryCondition{Field: field, Operator: operator, Value: value}, nil
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
	return &QueryBinaryOp{Left: left_expr, Right: right_expr, Op: op}, nil
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
	return &QueryUnaryOp{Operator: operator, Operand: operand_expr}, nil
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

