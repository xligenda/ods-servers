package repo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (r *GenericRepository[I, T]) buildSelectQuery(filters []Filter, opts *QueryOptions) (string, []any) {
	query := fmt.Sprintf("SELECT * FROM %s", pq.QuoteIdentifier(r.tableName))

	whereClause, args := r.buildWhereClause(filters)
	query += whereClause

	if opts != nil {
		if opts.OrderBy != "" {
			query += fmt.Sprintf(" ORDER BY %s", pq.QuoteIdentifier(opts.OrderBy))
		}
		if opts.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		}
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	return query, args
}

func (r *GenericRepository[I, T]) buildWhereClause(filters []Filter) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []any
	argIndex := 1

	for _, filter := range filters {
		condition, arg := r.buildCondition(filter, argIndex)
		conditions = append(conditions, condition)

		if arg != nil {
			switch v := arg.(type) {
			case []any:
				args = append(args, v...)
				argIndex += len(v)
			default:
				args = append(args, v)
				argIndex++
			}
		}
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")
	return whereClause, args
}

func (r *GenericRepository[I, T]) buildCondition(filter Filter, argIndex int) (string, any) {
	field := pq.QuoteIdentifier(filter.Field)

	switch strings.ToUpper(filter.Operator) {
	case "OR":
		if orData, ok := filter.Value.(map[string]any); ok {
			if conditions, exists := orData["conditions"].([]map[string]any); exists {
				var orParts []string
				var orArgs []any
				currentArgIndex := argIndex

				for _, condition := range conditions {
					fieldName, _ := condition["field"].(string)
					operator, _ := condition["operator"].(string)
					value := condition["value"]

					quotedField := pq.QuoteIdentifier(fieldName)
					orParts = append(orParts, fmt.Sprintf("%s %s $%d", quotedField, operator, currentArgIndex))
					orArgs = append(orArgs, value)
					currentArgIndex++
				}

				condition := fmt.Sprintf("(%s)", strings.Join(orParts, " OR "))
				return condition, orArgs
			}
		}

	case "RAW":
		if rawSQL, ok := filter.Value.(string); ok {
			return rawSQL, nil
		}

	case "IN":
		if values, ok := filter.Value.([]any); ok {
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			}
			condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
			return condition, values
		}

	case "NOT IN":
		if values, ok := filter.Value.([]any); ok {
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			}
			condition := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
			return condition, values
		}

	case "LIKE", "ILIKE":
		condition := fmt.Sprintf("%s %s $%d", field, filter.Operator, argIndex)
		return condition, filter.Value

	default:
		condition := fmt.Sprintf("%s %s $%d", field, filter.Operator, argIndex)
		return condition, filter.Value
	}

	condition := fmt.Sprintf("%s = $%d", field, argIndex)
	return condition, filter.Value
}

func (r *GenericRepository[I, T]) buildInsertData(entity T) ([]string, []any, []string) {
	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var fields []string
	var values []any
	var placeholders []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldName := field.Tag.Get("db")
		if fieldName == "" || fieldName == "-" {
			continue
		}

		fieldValue := r.getFieldValue(value)

		if value.Kind() == reflect.Ptr && value.IsNil() {
			continue
		}

		fields = append(fields, pq.QuoteIdentifier(fieldName))
		values = append(values, fieldValue)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)))
	}

	return fields, values, placeholders
}

func (r *GenericRepository[I, T]) buildUpdateData(entity T) ([]string, []any) {
	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var fields []string
	var values []any

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldName := field.Tag.Get("db")
		if fieldName == "" || fieldName == "-" || fieldName == "id" {
			continue
		}

		fieldValue := r.getFieldValue(value)

		if value.Kind() == reflect.Ptr && value.IsNil() {
			continue
		}

		fields = append(fields, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(fieldName), len(values)+1))
		values = append(values, fieldValue)
	}

	return fields, values
}

func (r *GenericRepository[I, T]) getFieldValue(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		return r.getFieldValue(value.Elem())
	}

	switch value.Type() {
	case reflect.TypeOf(time.Time{}):
		return value.Interface().(time.Time)
	case reflect.TypeOf((*time.Time)(nil)).Elem():
		if value.IsZero() {
			return nil
		}
		return value.Interface().(time.Time)
	}

	switch value.Kind() {
	case reflect.Slice, reflect.Map, reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return value.Interface()
		}

		jsonData, err := json.Marshal(value.Interface())
		if err != nil {
			return nil
		}
		return jsonData
	}

	return value.Interface()
}

func NewFilter(field, operator string, value any) Filter {
	return Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

func NewORFilter(conditions ...Filter) Filter {
	conditionMaps := make([]map[string]any, len(conditions))
	for i, condition := range conditions {
		conditionMaps[i] = map[string]any{
			"field":    condition.Field,
			"operator": condition.Operator,
			"value":    condition.Value,
		}
	}

	return Filter{
		Field:    "custom_or",
		Operator: "OR",
		Value: map[string]any{
			"conditions": conditionMaps,
		},
	}
}

func NewRawFilter(sql string) Filter {
	return Filter{
		Field:    "raw_condition",
		Operator: "RAW",
		Value:    sql,
	}
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

func (opts *QueryOptions) WithOrderBy(orderBy string) *QueryOptions {
	opts.OrderBy = orderBy
	return opts
}

func (opts *QueryOptions) WithLimit(limit int) *QueryOptions {
	opts.Limit = limit
	return opts
}

func (opts *QueryOptions) WithOffset(offset int) *QueryOptions {
	opts.Offset = offset
	return opts
}
