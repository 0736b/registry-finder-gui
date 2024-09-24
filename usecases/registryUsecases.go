package usecases

import (
	"github.com/0736b/registry-finder-gui/models"
	"github.com/0736b/registry-finder-gui/repositories"
)

type RegistryUsecase interface{}

type RegistryUsecaseImpl struct {
	registryRepository repositories.RegistryRepository
}

var singletonRegistryRepository repositories.RegistryRepository = nil

func NewRegistryUsecase() *RegistryUsecaseImpl {

	if singletonRegistryRepository == nil {
		singletonRegistryRepository = repositories.NewRegistryRepository()
	}
	return &RegistryUsecaseImpl{registryRepository: singletonRegistryRepository}
}

func (rs *RegistryUsecaseImpl) StreamRegistry() <-chan models.Registry {
	return rs.registryRepository.StreamRegistry()
}

func (rs *RegistryUsecaseImpl) FilterByKeyword(reg models.Registry, keyword string) bool {
	return true
}

func (rs *RegistryUsecaseImpl) FilterByKey(reg models.Registry, filterKey string) bool {
	return true
}

func (rs *RegistryUsecaseImpl) FilterByType(reg models.Registry, filterType string) bool {
	return true
}
