package services

import (
	"context"
	"database/sql"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type CustomerService struct {
	encoder encode.IEncode
	dbRO    database.IService
	dbRW    database.IService
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
	if !strings.HasPrefix(params.MobileNo, constants.PHMobilePrefix) {
		return 0, errs.ErrValidationInvalidMobileNumber
	}

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
		Status:       enums.ParseCustomerStatusToEnum(customer.Status),
	}

	if profile.CustomerType == enums.CUSTOMER_TYPE_COMPANY {
		company, err := s.dbRO.GetQueries().GetCustomerCompanyByCustomerID(ctx, decodedID)
		if err == nil {
			profile.CompanyName = company.Name
		}
	}

	return profile, nil
}

func (s *CustomerService) GetAllCustomers(ctx context.Context) ([]CustomerListItem, error) {
	rows, err := s.dbRO.GetQueries().GetAllCustomersWithCompany(ctx)
	if err != nil {
		return nil, err
	}

	customers := make([]CustomerListItem, 0, len(rows))
	for _, r := range rows {
		customers = append(customers, CustomerListItem{
			ID:           s.encoder.Encode(r.ID),
			Email:        r.Email,
			FirstName:    r.FirstName,
			MiddleName:   r.MiddleName.String,
			LastName:     r.LastName,
			Birthdate:    r.Birthdate,
			Sex:          r.Sex,
			CustomerType: enums.ParseCustomerTypeToEnum(r.CustomerType),
			CompanyName:  r.CompanyName.String,
			IsVerified:   enums.ParseCustomerStatusToEnum(r.Status),
			CreatedAt:    r.CreatedAt,
		})
	}

	return customers, nil
}

func (s *CustomerService) FilterCustomers(
	ctx context.Context,
	email string,
	customerType string,
	status string,
) ([]CustomerListItem, error) {
	customers, err := s.GetAllCustomers(ctx)
	if err != nil {
		return nil, err
	}

	if email == "" && customerType == "" && status == "" {
		return customers, nil
	}

	filtered := make([]CustomerListItem, 0)
	for _, c := range customers {
		if email != "" && !containsIgnoreCase(c.Email, email) {
			continue
		}
		if customerType != "" && c.CustomerType.String() != customerType {
			continue
		}
		if status != "" && c.IsVerified.String() != status {
			continue
		}
		filtered = append(filtered, c)
	}

	return filtered, nil
}

func containsIgnoreCase(s, substr string) bool {
	return len(substr) == 0 || containsLower(toLower(s), toLower(substr))
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsLower(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (s *CustomerService) ID() string {
	return "Customer"
}

func (s *CustomerService) Log() {
	logs.Log().Info("[CustomerService] Loaded")
}

var _ IService = (*CustomerService)(nil)
