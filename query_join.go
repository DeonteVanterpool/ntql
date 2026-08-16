package ntql

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type JoinQueryOptions struct {
	Distinct bool
}

func BuildSQLJoinQuery(expr QueryExpr, opts JoinQueryOptions) (string, error) {
	if expr == nil {
		return "", errors.New("query expression cannot be nil")
	}
	if len(schemaTables) == 0 {
		return "", errors.New("schema does not define any tables")
	}

	usedTables := map[string]struct{}{}
	conditionFieldMeta := map[*QueryCondition]subjectFieldMeta{}
	if err := collectJoinMetadata(expr, usedTables, conditionFieldMeta); err != nil {
		return "", err
	}

	baseTable := selectBaseTable(usedTables)
	aliasByTable := map[string]string{baseTable: "t0"}
	joinedTables := map[string]struct{}{baseTable: {}}
	joinClauses := make([]string, 0)
	nextAliasIndex := 1

	orderedTables := make([]string, 0, len(usedTables))
	for _, table := range schemaTables {
		if _, ok := usedTables[table.Name]; ok {
			orderedTables = append(orderedTables, table.Name)
		}
	}

	joinedEdges := map[string]struct{}{}
	for _, table := range orderedTables {
		if table == baseTable {
			continue
		}
		path, err := resolveJoinPath(baseTable, table)
		if err != nil {
			return "", err
		}
		for _, step := range path {
			if _, ok := aliasByTable[step.leftTable]; !ok {
				aliasByTable[step.leftTable] = fmt.Sprintf("t%d", nextAliasIndex)
				nextAliasIndex++
			}
			if _, ok := aliasByTable[step.rightTable]; !ok {
				aliasByTable[step.rightTable] = fmt.Sprintf("t%d", nextAliasIndex)
				nextAliasIndex++
			}

			edgeKey := step.edgeKey()
			if _, seen := joinedEdges[edgeKey]; seen {
				continue
			}

			if _, ok := joinedTables[step.rightTable]; !ok {
				joinClauses = append(joinClauses, fmt.Sprintf(
					"LEFT JOIN %s %s ON %s.%s = %s.%s",
					step.rightTable,
					aliasByTable[step.rightTable],
					aliasByTable[step.leftTable],
					step.leftKey,
					aliasByTable[step.rightTable],
					step.rightKey,
				))
				joinedTables[step.rightTable] = struct{}{}
				joinedEdges[edgeKey] = struct{}{}
				continue
			}

			if _, ok := joinedTables[step.leftTable]; !ok {
				joinClauses = append(joinClauses, fmt.Sprintf(
					"LEFT JOIN %s %s ON %s.%s = %s.%s",
					step.leftTable,
					aliasByTable[step.leftTable],
					aliasByTable[step.rightTable],
					step.rightKey,
					aliasByTable[step.leftTable],
					step.leftKey,
				))
				joinedTables[step.leftTable] = struct{}{}
			}
			joinedEdges[edgeKey] = struct{}{}
		}
	}

	whereSQL, err := buildJoinWhereSQL(expr, conditionFieldMeta, aliasByTable)
	if err != nil {
		return "", err
	}

	distinctClause := ""
	if opts.Distinct {
		distinctClause = "DISTINCT "
	}
	sql := fmt.Sprintf("SELECT %s%s.* FROM %s %s", distinctClause, aliasByTable[baseTable], baseTable, aliasByTable[baseTable])
	if len(joinClauses) > 0 {
		sql += " " + strings.Join(joinClauses, " ")
	}
	sql += " WHERE " + whereSQL

	return sql, nil
}

type subjectFieldMeta struct {
	table string
	field string
	dtype DType
}

func collectJoinMetadata(expr QueryExpr, usedTables map[string]struct{}, conditionFieldMeta map[*QueryCondition]subjectFieldMeta) error {
	switch node := expr.(type) {
	case *QueryCondition:
		meta, err := resolveSubjectFieldMeta(node.Field)
		if err != nil {
			return err
		}
		usedTables[meta.table] = struct{}{}
		conditionFieldMeta[node] = meta
		return nil
	case *QueryBinaryOp:
		if err := collectJoinMetadata(node.Left, usedTables, conditionFieldMeta); err != nil {
			return err
		}
		return collectJoinMetadata(node.Right, usedTables, conditionFieldMeta)
	case *QueryUnaryOp:
		return collectJoinMetadata(node.Operand, usedTables, conditionFieldMeta)
	default:
		return errors.New("unsupported query expression node")
	}
}

func selectBaseTable(usedTables map[string]struct{}) string {
	for _, table := range schemaTables {
		if table.Name == "tasks" {
			return table.Name
		}
	}
	for _, table := range schemaTables {
		if _, ok := usedTables[table.Name]; ok {
			return table.Name
		}
	}
	return schemaTables[0].Name
}

func resolveSubjectFieldMeta(field string) (subjectFieldMeta, error) {
	subject, err := getSubject(field)
	if err != nil {
		return subjectFieldMeta{}, fmt.Errorf("field %s is not defined in schema subjects", field)
	}
	if subject.Table == "" {
		return subjectFieldMeta{}, fmt.Errorf("subject %s is missing a table mapping", subject.Name)
	}
	column := subject.Column
	if column == "" {
		column = subject.Name
	}
	if len(subject.ValidTypes) == 0 {
		return subjectFieldMeta{}, fmt.Errorf("subject %s has no valid types", subject.Name)
	}
	return subjectFieldMeta{
		table: subject.Table,
		field: column,
		dtype: subject.ValidTypes[0],
	}, nil
}

type joinStep struct {
	leftTable  string
	rightTable string
	leftKey    string
	rightKey   string
}

func (s joinStep) edgeKey() string {
	if s.leftTable < s.rightTable {
		return fmt.Sprintf("%s.%s:%s.%s", s.leftTable, s.leftKey, s.rightTable, s.rightKey)
	}
	return fmt.Sprintf("%s.%s:%s.%s", s.rightTable, s.rightKey, s.leftTable, s.leftKey)
}

func resolveJoinPath(baseTable, targetTable string) ([]joinStep, error) {
	if baseTable == targetTable {
		return nil, nil
	}

	type pathNode struct {
		table string
		path  []joinStep
	}
	queue := []pathNode{{table: baseTable, path: []joinStep{}}}
	visited := map[string]struct{}{baseTable: {}}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for _, join := range schemaJoins {
			var next string
			var step joinStep
			switch {
			case join.FromTable == node.table:
				next = join.ToTable
				step = joinStep{
					leftTable:  join.FromTable,
					rightTable: join.ToTable,
					leftKey:    join.FromKey,
					rightKey:   join.ToKey,
				}
			case join.ToTable == node.table:
				next = join.FromTable
				step = joinStep{
					leftTable:  join.ToTable,
					rightTable: join.FromTable,
					leftKey:    join.ToKey,
					rightKey:   join.FromKey,
				}
			default:
				continue
			}

			if _, seen := visited[next]; seen {
				continue
			}
			newPath := append(append([]joinStep{}, node.path...), step)
			if next == targetTable {
				return newPath, nil
			}
			visited[next] = struct{}{}
			queue = append(queue, pathNode{table: next, path: newPath})
		}
	}

	return nil, fmt.Errorf("no join path found from %s to %s", baseTable, targetTable)
}

func buildJoinWhereSQL(expr QueryExpr, conditionFieldMeta map[*QueryCondition]subjectFieldMeta, aliasByTable map[string]string) (string, error) {
	switch node := expr.(type) {
	case *QueryCondition:
		meta, ok := conditionFieldMeta[node]
		if !ok {
			return "", fmt.Errorf("missing metadata for field %s", node.Field)
		}
		alias, ok := aliasByTable[meta.table]
		if !ok {
			return "", fmt.Errorf("missing alias for table %s", meta.table)
		}
		return joinConditionToSQL(node, fmt.Sprintf("%s.%s", alias, meta.field), meta.dtype)
	case *QueryBinaryOp:
		left, err := buildJoinWhereSQL(node.Left, conditionFieldMeta, aliasByTable)
		if err != nil {
			return "", err
		}
		right, err := buildJoinWhereSQL(node.Right, conditionFieldMeta, aliasByTable)
		if err != nil {
			return "", err
		}
		switch node.Operator {
		case OperatorAnd:
			return "(" + left + " AND " + right + ")", nil
		case OperatorOr:
			return "(" + left + " OR " + right + ")", nil
		case OperatorXor:
			return "(" + left + " XOR " + right + ")", nil
		default:
			return "", errors.New("invalid operator: " + node.Operator.ToStr())
		}
	case *QueryUnaryOp:
		operand, err := buildJoinWhereSQL(node.Operand, conditionFieldMeta, aliasByTable)
		if err != nil {
			return "", err
		}
		switch node.Operator {
		case OperatorNot:
			return "(NOT (" + operand + "))", nil
		default:
			return "", errors.New("invalid operator: " + node.Operator.ToStr())
		}
	default:
		return "", errors.New("unsupported query expression node")
	}
}

func joinConditionToSQL(c *QueryCondition, fieldRef string, dtype DType) (string, error) {
	switch dtype {
	case DTypeDate, DTypeDateTime:
		if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}Z)?$`).MatchString(c.Value) {
			return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
		switch c.Operator {
		case OperatorEq:
			return fieldRef + " = '" + c.Value + "'", nil
		case OperatorNeq:
			return fieldRef + " != '" + c.Value + "'", nil
		case OperatorGt:
			return fieldRef + " > '" + c.Value + "'", nil
		case OperatorLT:
			return fieldRef + " < '" + c.Value + "'", nil
		case OperatorGte:
			return fieldRef + " >= '" + c.Value + "'", nil
		case OperatorLte:
			return fieldRef + " <= '" + c.Value + "'", nil
		default:
			return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	case DTypeInt:
		if _, err := strconv.Atoi(c.Value); err != nil {
			return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
		switch c.Operator {
		case OperatorEq:
			return fieldRef + " = " + c.Value, nil
		case OperatorNeq:
			return fieldRef + " != " + c.Value, nil
		case OperatorGt:
			return fieldRef + " > " + c.Value, nil
		case OperatorLT:
			return fieldRef + " < " + c.Value, nil
		case OperatorGte:
			return fieldRef + " >= " + c.Value, nil
		case OperatorLte:
			return fieldRef + " <= " + c.Value, nil
		default:
			return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	default:
		if !regexp.MustCompile(`^[a-zA-Z0-9\-/: ]+$`).MatchString(c.Value) {
			return "", errors.New("invalid value: " + c.Value + " for field: " + c.Field)
		}
		switch c.Operator {
		case OperatorEq:
			return fieldRef + " = '" + c.Value + "'", nil
		case OperatorNeq:
			return fieldRef + " != '" + c.Value + "'", nil
		case OperatorCnt:
			return fieldRef + " LIKE '%" + c.Value + "%'", nil
		case OperatorSW:
			return fieldRef + " LIKE '" + c.Value + "%'", nil
		case OperatorEw:
			return fieldRef + " LIKE '%" + c.Value + "'", nil
		default:
			return "", errors.New("invalid operator: " + c.Operator.ToStr() + " for field: " + c.Field)
		}
	}
}
