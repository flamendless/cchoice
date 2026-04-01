package services

import (
	"context"
	"database/sql"

	"cchoice/cmd/web/models"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type CustomerService struct {
	encoder encode.IEncode
	dbRO    database.IService
	dbRW    database.IService
}

type RegisterCustomerParams struct {
	FirstName    string
	MiddleName   string
	LastName     string
	Birthdate    string
	Sex          string
	Email        string
	MobileNo     string
	Password     string
	CustomerType enums.CustomerType
	CompanyName  string
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

func NewCustomerService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
) *CustomerService {
	return &CustomerService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func (s *CustomerService) Register(ctx context.Context, params RegisterCustomerParams) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	middleNameNull := sql.NullString{String: params.MiddleName, Valid: params.MiddleName != ""}

	customerID, err := s.dbRW.GetQueries().CreateCustomer(ctx, queries.CreateCustomerParams{
		FirstName:    params.FirstName,
		MiddleName:   middleNameNull,
		LastName:     params.LastName,
		Birthdate:    params.Birthdate,
		Sex:          params.Sex,
		Email:        params.Email,
		MobileNo:     params.MobileNo,
		Password:     string(hash),
		CustomerType: params.CustomerType.String(),
	})
	if err != nil {
		return 0, err
	}

	if params.CustomerType == enums.CUSTOMER_TYPE_COMPANY && params.CompanyName != "" {
		_, err = s.dbRW.GetQueries().CreateCustomerCompany(ctx, queries.CreateCustomerCompanyParams{
			CustomerID: customerID,
			Name:       params.CompanyName,
		})
		if err != nil {
			return 0, err
		}
	}

	return customerID, nil
}

func (s *CustomerService) Login(ctx context.Context, email string, password string) (queries.GetCustomerByEmailRow, error) {
	customer, err := s.dbRO.GetQueries().GetCustomerByEmail(ctx, email)
	if err != nil {
		return queries.GetCustomerByEmailRow{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(password)); err != nil {
		return queries.GetCustomerByEmailRow{}, err
	}

	return customer, nil
}

func (s *CustomerService) GetByID(ctx context.Context, customerID string) (queries.GetCustomerByIDRow, error) {
	decodedID := s.encoder.Decode(customerID)
	customer, err := s.dbRO.GetQueries().GetCustomerByID(ctx, decodedID)
	if err != nil {
		return queries.GetCustomerByIDRow{}, err
	}
	return customer, nil
}

func (s *CustomerService) GetCompanyByCustomerID(ctx context.Context, customerID int64) (queries.GetCustomerCompanyByCustomerIDRow, error) {
	return s.dbRO.GetQueries().GetCustomerCompanyByCustomerID(ctx, customerID)
}

func (s *CustomerService) UpdatePassword(ctx context.Context, customerID string, password string) error {
	decodedID := s.encoder.Decode(customerID)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.dbRW.GetQueries().UpdateCustomerPassword(ctx, queries.UpdateCustomerPasswordParams{
		Password: string(hash),
		ID:       decodedID,
	})
	return err
}

func (s *CustomerService) UpdateProfile(ctx context.Context, params UpdateCustomerProfileParams) error {
	decodedID := s.encoder.Decode(params.ID)
	middleNameNull := sql.NullString{String: params.MiddleName, Valid: params.MiddleName != ""}

	_, err := s.dbRW.GetQueries().UpdateCustomerProfile(ctx, queries.UpdateCustomerProfileParams{
		FirstName:  params.FirstName,
		MiddleName: middleNameNull,
		LastName:   params.LastName,
		MobileNo:   params.MobileNo,
		Birthdate:  params.Birthdate,
		Sex:        params.Sex,
		ID:         decodedID,
	})
	return err
}

func (s *CustomerService) BuildProfile(ctx context.Context, customerID string) (models.CustomerProfile, error) {
	decodedID := s.encoder.Decode(customerID)
	customer, err := s.dbRO.GetQueries().GetCustomerByID(ctx, decodedID)
	if err != nil {
		return models.CustomerProfile{}, err
	}

	profile := models.CustomerProfile{
		FullName:     utils.BuildFullName(customer.FirstName, customer.MiddleName.String, customer.LastName),
		FirstName:    customer.FirstName,
		MiddleName:   customer.MiddleName.String,
		LastName:     customer.LastName,
		Birthdate:    customer.Birthdate,
		Sex:          customer.Sex,
		Email:        customer.Email,
		MobileNo:     customer.MobileNo,
		CustomerType: enums.ParseCustomerTypeToEnum(customer.CustomerType),
	}

	if profile.CustomerType == enums.CUSTOMER_TYPE_COMPANY {
		company, err := s.dbRO.GetQueries().GetCustomerCompanyByCustomerID(ctx, decodedID)
		if err == nil {
			profile.CompanyName = company.Name
		}
	}

	return profile, nil
}
