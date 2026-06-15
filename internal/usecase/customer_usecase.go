package usecase

import (
	"context"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type CustomerUsecase struct {
	repo *repository.CustomerRepository
}

func NewCustomerUsecase(repo *repository.CustomerRepository) *CustomerUsecase {
	return &CustomerUsecase{repo: repo}
}

func (u *CustomerUsecase) Create(ctx context.Context, req *model.CreateCustomerRequest) (*model.Customer, error) {
	if req.CompanyName == "" {
		return nil, ErrCustomerCompanyRequired
	}
	if req.PICName == "" {
		return nil, ErrCustomerPICRequired
	}
	if req.Phone == "" {
		return nil, ErrCustomerPhoneRequired
	}

	customer := &model.Customer{
		CompanyName: req.CompanyName,
		PICName:     req.PICName,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		NPWP:        req.NPWP,
		IsActive:    true,
	}

	if err := u.repo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (u *CustomerUsecase) List(ctx context.Context) ([]model.Customer, error) {
	return u.repo.FetchAll(ctx)
}

func (u *CustomerUsecase) GetByID(ctx context.Context, id int) (*model.Customer, error) {
	if id <= 0 {
		return nil, ErrCustomerInvalidID
	}
	return u.repo.GetByID(ctx, id)
}

func (u *CustomerUsecase) Update(ctx context.Context, id int, req *model.UpdateCustomerRequest) (*model.Customer, error) {
	if id <= 0 {
		return nil, ErrCustomerInvalidID
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.CompanyName != nil {
		existing.CompanyName = *req.CompanyName
	}
	if req.PICName != nil {
		existing.PICName = *req.PICName
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Address != nil {
		existing.Address = *req.Address
	}
	if req.NPWP != nil {
		existing.NPWP = *req.NPWP
	}

	if existing.CompanyName == "" {
		return nil, ErrCustomerCompanyRequired
	}
	if existing.PICName == "" {
		return nil, ErrCustomerPICRequired
	}
	if existing.Phone == "" {
		return nil, ErrCustomerPhoneRequired
	}

	if err := u.repo.Update(ctx, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (u *CustomerUsecase) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrCustomerInvalidID
	}
	return u.repo.Delete(ctx, id)
}
