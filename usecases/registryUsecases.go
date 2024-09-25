package usecases

import (
	"strings"

	"github.com/0736b/registry-finder-gui/entities"
	"github.com/0736b/registry-finder-gui/repositories"
	"github.com/0736b/registry-finder-gui/utils"
)

type RegistryUsecase interface {
	StreamRegistry() <-chan *entities.Registry
	FilterByKeyword(reg *entities.Registry, keyword string) bool
	FilterByKey(reg *entities.Registry, filterKey string) bool
	FilterByType(reg *entities.Registry, filterType string) bool
	OpenInRegedit(reg *entities.Registry)
}

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

func (u *RegistryUsecaseImpl) StreamRegistry() <-chan *entities.Registry {

	return u.registryRepository.StreamRegistry()
}

func (u *RegistryUsecaseImpl) FilterByKeyword(reg *entities.Registry, keyword string) bool {

	if len(keyword) == 0 {
		return true
	}

	keyword = strings.ToLower(keyword)
	keyword = strings.ReplaceAll(keyword, " ", "")
	keyword = strings.TrimSpace(keyword)

	info := reg.Path + " " + reg.Name + " " + reg.Value

	return strings.Contains(info, keyword)
}

func (u *RegistryUsecaseImpl) FilterByKey(reg *entities.Registry, filterKey string) bool {

	return true
}

func (u *RegistryUsecaseImpl) FilterByType(reg *entities.Registry, filterType string) bool {

	return true
}

func (u *RegistryUsecaseImpl) OpenInRegedit(reg *entities.Registry) {

	utils.OpenRegeditAtPath(reg.Path)
}
