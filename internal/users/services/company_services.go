package services

import (
	"context"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
)

type CompanyServicesInt interface {
	GetIndustries() (*[]models.GetIndustriesResponse, error)
}

type CompanyServices struct {
	companyRepo repositories.CompanyRepositoryInt
}

func NewCompanyServices(companyRepo repositories.CompanyRepositoryInt) CompanyServicesInt {
	return &CompanyServices{
		companyRepo: companyRepo,
	}
}

func (s *CompanyServices) GetIndustries() (*[]models.GetIndustriesResponse, error) {
	industries, err := s.companyRepo.GetIndustries(context.Background())
	if err != nil {
		return nil, err
	}
	response := make([]models.GetIndustriesResponse, len(*industries))
	for i, industry := range *industries {
		response[i] = industry.ToGetIndustriesResponse()
	}
	return &response, nil
}
