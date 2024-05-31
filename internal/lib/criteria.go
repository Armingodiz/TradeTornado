package lib

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type FilterOperator string

var filterOperatorMap = map[string]FilterOperator{
	"Equal":   EqualOperator,
	"GT":      GTOperator,
	"GTE":     GTEOperator,
	"LT":      LTOperator,
	"LTE":     LTEOperator,
	"In":      InOperator,
	"Between": BetweenOperator,
	"Contain": ContainOperator,
}

func FilterOperatorFromStringName(name string) (FilterOperator, error) {
	value, ok := filterOperatorMap[name]
	if !ok {
		err := NewErrorNotification()
		err.Add("filter_operator", errors.New("invalid filter operator"))
		return EqualOperator, err
	}
	return value, nil
}

const (
	EqualOperator   FilterOperator = "="
	GTOperator      FilterOperator = ">"
	GTEOperator     FilterOperator = ">="
	LTOperator      FilterOperator = "<"
	LTEOperator     FilterOperator = "<="
	InOperator      FilterOperator = "IN"
	BetweenOperator FilterOperator = "between"
	ContainOperator FilterOperator = "Contain"
)

type Filter struct {
	Field    string
	Value    []string
	Operator FilterOperator
}

func NewFilter(field string, opr FilterOperator, values ...string) (Filter, error) {
	filter := Filter{
		Field:    field,
		Operator: opr,
		Value:    values,
	}
	return filter, filter.validate()
}

func (fil Filter) validate() error {
	switch fil.Operator {
	case BetweenOperator:
		if len(fil.Value) != 2 {
			return fmt.Errorf("value length should be 2 for operator %s", fil.Operator)
		}
	case InOperator:
		return nil
		// Other operators
	default:
		if len(fil.Value) > 1 {
			return fmt.Errorf("value length should be 1 for operator %s", fil.Operator)
		}
	}
	return nil
}

type LogicalOperator string

const (
	And LogicalOperator = "AND"
	Or  LogicalOperator = "OR"
)

type Pagination struct {
	Offset uint
	Limit  uint
}

func NewPagination(offset, limit uint) *Pagination {
	return &Pagination{
		Offset: offset,
		Limit:  limit,
	}
}

type SortOperator string

const (
	ASC  SortOperator = "ASC"
	DESC SortOperator = "DESC"
)

type Sort struct {
	Field    string
	Operator SortOperator
}

func NewSort(f string, opr SortOperator) Sort {
	return Sort{
		Field:    f,
		Operator: opr,
	}
}

type Criteria struct {
	Filters    []Filter
	Operator   LogicalOperator
	Pagination *Pagination
	Sorts      []Sort
}

func NewCriteria() *Criteria {
	return &Criteria{
		Operator: And,
		Filters:  make([]Filter, 0),
		Sorts:    make([]Sort, 0),
	}
}

func (cr *Criteria) AddFilter(f Filter) *Criteria {
	cr.Filters = append(cr.Filters, f)
	return cr
}

func (cr *Criteria) SetPagination(pg *Pagination) {
	cr.Pagination = pg
}

func (cr *Criteria) SetOperator(opr LogicalOperator) {
	cr.Operator = opr
}

func (cr *Criteria) AddSort(sr Sort) {
	cr.Sorts = append(cr.Sorts, sr)
}

func ParseCriteriaFromRequest(c *gin.Context) (*Criteria, error) {
	criteria := NewCriteria()

	filters := c.QueryArray("filters")
	for _, filterStr := range filters {
		parts := strings.Split(filterStr, ",")
		if len(parts) < 3 {
			return nil, errors.New("invalid filter format, expected field,operator,value")
		}

		field := parts[0]
		operator, err := FilterOperatorFromStringName(parts[1])
		if err != nil {
			return nil, err
		}

		values := parts[2:]
		filter, err := NewFilter(field, operator, values...)
		if err != nil {
			return nil, err
		}
		criteria.AddFilter(filter)
	}

	offset := c.Query("offset")
	limit := c.Query("limit")
	if offset != "" && limit != "" {
		pg := NewPagination(cast.ToUint(offset), cast.ToUint(limit))
		criteria.SetPagination(pg)
	}

	sorts := c.QueryArray("sorts")
	for _, sortStr := range sorts {
		parts := strings.Split(sortStr, ",")
		if len(parts) != 2 {
			return nil, errors.New("invalid sort format, expected field,operator")
		}

		field := parts[0]
		operator := SortOperator(parts[1])
		sort := NewSort(field, operator)
		criteria.AddSort(sort)
	}

	// Parse Logical Operator
	operator := c.Query("operator")
	if operator != "" {
		criteria.SetOperator(LogicalOperator(operator))
	}

	return criteria, nil
}

type ApplyGormCriteria func(qr *gorm.DB, structType any, criteria *Criteria) (*gorm.DB, error)

func GenericApplyGormCriteria(qr *gorm.DB, structType any, criteria *Criteria) (*gorm.DB, error) {
	qr, err := applyFilters(qr, structType, criteria)
	if err != nil {
		return nil, err
	}
	qr, err = applySorts(qr, structType, criteria)
	if err != nil {
		return nil, err
	}
	return applyPagination(qr, criteria)
}

func applyFilters(qr *gorm.DB, structType any, criteria *Criteria) (*gorm.DB, error) {
	for _, f := range criteria.Filters {
		fieldName, err := getFieldName(f.Field, structType)
		if err != nil {
			return nil, err
		}
		if f.Operator == ContainOperator {
			f.Value[0] = "%" + f.Value[0] + "%"
			f.Operator = "Like"
		}
		if criteria.Operator == And {
			if f.Operator == BetweenOperator && len(f.Value) == 2 {
				qr = qr.Where(fmt.Sprintf("%s %s ? and ?", fieldName, f.Operator), f.Value[0], f.Value[1])
			} else {
				qr = qr.Where(fmt.Sprintf("%s %s ?", fieldName, f.Operator), f.Value)
			}
		} else {
			if f.Operator == BetweenOperator && len(f.Value) == 2 {
				qr = qr.Or(fmt.Sprintf("%s %s ? and ?", fieldName, f.Operator), f.Value[0], f.Value[1])
			} else {
				qr = qr.Or(fmt.Sprintf("%s %s ?", fieldName, f.Operator), f.Value)
			}
		}
	}
	return qr, nil
}

func applySorts(qr *gorm.DB, structType any, criteria *Criteria) (*gorm.DB, error) {
	for _, sr := range criteria.Sorts {
		sortField, err := getFieldName(sr.Field, structType)
		if err != nil {
			return nil, err
		}
		qr = qr.Order(fmt.Sprintf("%s %s", sortField, sr.Operator))
	}
	return qr, nil
}

func applyPagination(qr *gorm.DB, criteria *Criteria) (*gorm.DB, error) {
	if criteria.Pagination != nil {
		if criteria.Pagination.Limit == 0 {
			criteria.Pagination.Limit = 100
		}
		qr = qr.Limit(int(criteria.Pagination.Limit)).Offset(int(criteria.Pagination.Offset))
	}
	return qr, nil
}

func isFieldIndexed(gormTag string) bool {
	tags := strings.Split(gormTag, ";")
	for _, tag := range tags { // TODO: check other kind of index tags
		parts := strings.Split(tag, ":")
		if parts[0] == "index" || parts[0] == "uniqueIndex" || parts[0] == "primarykey" {
			return true
		}
	}
	return false
}

func isFieldGormEmbedded(gormTag string) bool {
	tags := strings.Split(gormTag, ";")
	for _, tag := range tags {
		parts := strings.Split(tag, ":")
		if parts[0] == "embedded" {
			return true
		}
	}
	return false
}

func getFieldValue(structType interface{}, i int) interface{} {
	var embeddedField interface{}
	if reflect.ValueOf(structType).Field(i).Kind() == reflect.Ptr {
		embeddedField = reflect.ValueOf(structType).Field(i).Elem().Interface()
	} else {
		embeddedField = reflect.ValueOf(structType).Field(i).Interface()
	}
	return embeddedField
}

func getFieldName(field string, structType any) (string, error) {
	fieldName, ok, err := getCriteriaName(field, structType)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("field %s dosn't have the criteria tag, you cant filter it", field)
	}
	return fieldName, nil
}

func getCriteriaName(field string, structType any) (string, bool, error) {
	for i := 0; i < reflect.TypeOf(structType).NumField(); i++ {
		gormTag := reflect.TypeOf(structType).Field(i).Tag.Get("gorm")
		if isFieldGormEmbedded(gormTag) {
			columnName, ok, err := getCriteriaName(field, getFieldValue(structType, i))
			if err != nil {
				return "", false, err
			}
			if ok {
				return columnName, true, nil
			} else {
				continue
			}
		}

		if reflect.TypeOf(structType).Field(i).Tag.Get("criteria") == field {
			if gormTag == "" {
				return "", false, fmt.Errorf("field %s doesn't have gorm tag, you cant filter it", field)
			}
			if !isFieldIndexed(gormTag) {
				return "", false, fmt.Errorf("field %s is not indexed, you cant filter it", field)
			}
			columnName, err := getFieldColumnName(gormTag)
			if err != nil {
				return "", false, err
			}
			return columnName, true, nil
		}
	}
	return "", false, nil
}

func getFieldColumnName(gormTag string) (string, error) {
	tags := strings.Split(gormTag, ";")
	for _, tag := range tags {
		if strings.Contains(tag, "column") {
			vals := strings.Split(tag, ":")
			if len(vals) < 2 {
				return "", errors.New("invalid tag format for column name")
			}
			return vals[1], nil
		}
	}
	return "", errors.New("column name tag is not set for the strcut")
}
