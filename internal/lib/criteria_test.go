package lib

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GenericGormCriteraTestSuit struct {
	// db       *gorm.DB
	// criteria *Criteria
	suite.Suite
}

func TestGenericGormCriteraTestSuit(t *testing.T) {
	suite.Run(t, new(GenericGormCriteraTestSuit))
}

func (suite *GenericGormCriteraTestSuit) SetupTest() {
	// suite.db = &gorm.DB{
	// 	Statement: &gorm.Statement{Clauses: make(map[string]clause.Clause)},
	// 	Dialector: gorm.Dialector,
	// }
	// suite.criteria = NewCriteria()
}

func (suite *GenericGormCriteraTestSuit) Cleanup() {}

// func (suite *GenericGormCriteraTestSuit) TestApplyPagination() {
// 	pagination := NewPagination(4, 5)
// 	suite.criteria.SetPagination(pagination)
// 	query, err := applyPagination(suite.db, suite.criteria)
// 	suite.Nil(err)
// 	fmt.Println(query)
// 	sql := query.ToSQL(func(query *gorm.DB) *gorm.DB {
// 		return query
// 	})
// 	fmt.Println(sql)

// }

func (suite *GenericGormCriteraTestSuit) TestIsFieldIndexed() {
	gormTag := "history:no;column:armin;test:yes"
	suite.Equal(false, isFieldIndexed(gormTag))
	gormTag = "column:user_id;uniqueIndex:type_user_id_user_feedback_unique_index"
	suite.Equal(true, isFieldIndexed(gormTag))
}

func (suite *GenericGormCriteraTestSuit) TestGetFieldName() {
	_, err := getFieldName("failed", mockStruct{})
	suite.NotNil(err)
	name, err := getFieldName("type", mockStruct{})
	suite.Nil(err)
	suite.Equal("type", name)
	_, err = getFieldName("name", mockStruct{})
	suite.NotNil(err)
}

func (suite *GenericGormCriteraTestSuit) TestGetFieldColumnName() {
	gormTag := "history:no;column:armin;test:yes"
	name, err := getFieldColumnName(gormTag)
	suite.Nil(err)
	suite.Equal("armin", name)
}

type mockStruct struct {
	Name   string `criteria:"name"`
	UserId string `criteria:"user_id" gorm:"column:user_id;uniqueIndex:type_user_id_user_feedback_unique_index"`
	Type   string `criteria:"type" gorm:"column:type;uniqueIndex:type_user_id_user_feedback_unique_index"`
}
