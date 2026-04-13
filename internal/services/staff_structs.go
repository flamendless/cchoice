package services

import "database/sql"

type StaffRowBase struct {
	ID              int64
	FirstName       string
	MiddleName      sql.NullString
	LastName        string
	Birthdate       string
	Sex             string
	DateHired       string
	TimeInSchedule  sql.NullString
	TimeOutSchedule sql.NullString
	Position        string
	UserType        string
	Email           string
	MobileNo        string
	RequireInShop   bool
	CreatedAt       string
	UpdatedAt       string
}

type StaffRow struct {
	ID                          int64
	StaffID                     int64
	ForDate                     string
	TimeIn                      sql.NullString
	TimeOut                     sql.NullString
	InLocation                  sql.NullString
	OutLocation                 sql.NullString
	InUseragentID               sql.NullInt64
	OutUseragentID              sql.NullInt64
	LunchBreakIn                sql.NullString
	LunchBreakOut               sql.NullString
	LunchBreakInLocation        sql.NullString
	LunchBreakOutLocation       sql.NullString
	LunchBreakInUseragentID     sql.NullInt64
	LunchBreakOutUseragentID    sql.NullInt64
	CreatedAt                   string
	UpdatedAt                   string
	FirstName                   string
	MiddleName                  sql.NullString
	LastName                    string
	InBrowser                   sql.NullString
	InBrowserVersion            sql.NullString
	InOs                        sql.NullString
	InDevice                    sql.NullString
	OutBrowser                  sql.NullString
	OutBrowserVersion           sql.NullString
	OutOs                       sql.NullString
	OutDevice                   sql.NullString
	LunchBreakInBrowser         sql.NullString
	LunchBreakInBrowserVersion  sql.NullString
	LunchBreakInOs              sql.NullString
	LunchBreakInDevice          sql.NullString
	LunchBreakOutBrowser        sql.NullString
	LunchBreakOutBrowserVersion sql.NullString
	LunchBreakOutOs             sql.NullString
	LunchBreakOutDevice         sql.NullString
}

type UpdateProfileParams struct {
	ID         string
	FirstName  string
	MiddleName string
	LastName   string
	MobileNo   string
	Birthdate  string
	DateHired  string
}
