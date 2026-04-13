package services

import (
	"cchoice/internal/enums"
)

type RegisterCustomerParams struct {
	FirstName    string
	MiddleName   string
	LastName     string
	Birthdate    string
	Sex          string
	Email        string
	MobileNo     string
	Password     string
	CompanyName  string
	CustomerType enums.CustomerType
}

type UpdateCustomerProfileParams struct {
	ID         string
	FirstName  string
	MiddleName string
	LastName   string
	MobileNo   string
	Birthdate  string
	Sex        string
}

type CustomerListItem struct {
	ID           string
	Email        string
	FirstName    string
	MiddleName   string
	LastName     string
	Birthdate    string
	Sex          string
	CompanyName  string
	CreatedAt    string
	CustomerType enums.CustomerType
	IsVerified   enums.CustomerStatus
}
