package repositories

import (
	"fmt"
	"slices"

	"github.com/0736b/registry-finder-gui/models"
	"github.com/0736b/registry-finder-gui/utils"
	"golang.org/x/sys/windows/registry"
)

type RegistryRepository interface {
	StreamRegistry() <-chan models.Registry
}

type RegistryRepositoryImpl struct{}

func NewRegistryRepository() *RegistryRepositoryImpl {
	return &RegistryRepositoryImpl{}
}

func (rr *RegistryRepositoryImpl) StreamRegistry() <-chan models.Registry {
	return nil
}

func (rr *RegistryRepositoryImpl) generateRegistryByKey(key *registry.Key) <-chan models.Registry {
	return nil
}

func (rr *RegistryRepositoryImpl) fanInRegistry(hkcr, hkcu, hklm, hkcc, hku <-chan models.Registry) <-chan models.Registry {
	return nil
}

func (rr *RegistryRepositoryImpl) queryEnumKeys(hkey registry.Key, path string) <-chan models.Registry {
	return nil
}

func (rr *RegistryRepositoryImpl) queryEnumValues(hkey registry.Key, path string) <-chan models.Registry {
	return nil
}

func (rr *RegistryRepositoryImpl) queryValue(hkey registry.Key, name string) (string, string, error) {

	value := make([]byte, 1024)

	n, valtype, err := hkey.GetValue(name, value)
	if err != nil {
		return "", "", fmt.Errorf("QueryValue error: %w", err)
	}

	value = value[:n]

	switch valtype {

	case registry.NONE:
		return "", utils.STR_NONE, nil

	case registry.SZ:
		return utils.BytesToString(value), utils.STR_REG_SZ, nil

	case registry.EXPAND_SZ:
		return utils.BytesToString(value), utils.STR_REG_EXPAND_SZ, nil

	case registry.BINARY:
		return fmt.Sprintf("%-1x", value), utils.STR_REG_BINARY, nil

	case registry.DWORD:
		slices.Reverse(value)
		return fmt.Sprintf("0x%x", utils.BytesToString(value)), utils.STR_REG_DWORD, nil

	case registry.DWORD_BIG_ENDIAN:
		return fmt.Sprintf("0x%x", utils.BytesToString(value)), utils.STR_REG_DWORD_BIG_ENDIAN, nil

	case registry.LINK:
		return utils.BytesToString(value), utils.STR_REG_LINK, nil

	case registry.MULTI_SZ:
		val, _, _ := hkey.GetStringsValue(name)
		return fmt.Sprintf("%s", val), utils.STR_REG_MULTI_SZ, nil

	case registry.RESOURCE_LIST:
		return utils.BytesToString(value), utils.STR_REG_RESOURCE_LIST, nil

	case registry.FULL_RESOURCE_DESCRIPTOR:
		return fmt.Sprintf("%-2x", value), utils.STR_REG_FULL_RESOURCE_DESCRIPTOR, nil

	case registry.RESOURCE_REQUIREMENTS_LIST:
		return utils.BytesToString(value), utils.STR_REG_RESOURCE_REQUIREMENTS_LIST, nil

	case registry.QWORD:
		slices.Reverse(value)
		return fmt.Sprintf("%x", utils.BytesToString(value)), utils.STR_REG_QWORD, nil

	}

	return "", "", fmt.Errorf("QueryValue error: %w", err)
}
