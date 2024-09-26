package usecases

import (
	"strings"
	"sync"

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

var (
	singletonRegistryRepository repositories.RegistryRepository = nil

	keywordCache   = make(map[string]string)
	keywordCacheMu sync.RWMutex
)

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

	if keyword == "" {
		return true
	}

	keywordCacheMu.RLock()
	processedKeyword, exists := keywordCache[keyword]
	keywordCacheMu.RUnlock()

	if !exists {
		processedKeyword = utils.PreProcessStr(keyword)
		keywordCacheMu.Lock()
		keywordCache[keyword] = processedKeyword
		keywordCacheMu.Unlock()
	}

	return strings.Contains(utils.PreProcessStr(reg.Path), processedKeyword) ||
		strings.Contains(utils.PreProcessStr(reg.Name), processedKeyword) ||
		strings.Contains(utils.PreProcessStr(reg.Value), processedKeyword)
}

func (u *RegistryUsecaseImpl) FilterByKey(reg *entities.Registry, filterKey string) bool {

	return strings.HasPrefix(reg.Path, filterKey)
}

func (u *RegistryUsecaseImpl) FilterByType(reg *entities.Registry, filterType string) bool {

	return reg.Type == filterType
}

func (u *RegistryUsecaseImpl) OpenInRegedit(reg *entities.Registry) {

	utils.OpenRegeditAtPath(reg.Path)
}
