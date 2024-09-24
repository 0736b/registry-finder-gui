package repositories

import (
	"fmt"
	"log"
	"slices"

	"github.com/0736b/registry-finder-gui/entities"
	"github.com/0736b/registry-finder-gui/utils"
	"golang.org/x/sys/windows/registry"
)

type RegistryRepository interface {
	StreamRegistry() <-chan entities.Registry
}

type RegistryRepositoryImpl struct{}

func NewRegistryRepository() *RegistryRepositoryImpl {
	return &RegistryRepositoryImpl{}
}

func (r *RegistryRepositoryImpl) StreamRegistry() <-chan entities.Registry {

	fromHKCR := generateRegistryByKey(registry.CLASSES_ROOT)
	fromHKCU := generateRegistryByKey(registry.CURRENT_USER)
	fromHKLM := generateRegistryByKey(registry.LOCAL_MACHINE)
	fromHKCC := generateRegistryByKey(registry.CURRENT_CONFIG)
	fromHKU := generateRegistryByKey(registry.USERS)

	results := fanInRegistry(fromHKCR, fromHKCU, fromHKLM, fromHKCC, fromHKU)

	return results
}

func fanInRegistry(hkcr, hkcu, hklm, hkcc, hku <-chan entities.Registry) <-chan entities.Registry {

	resultsChan := make(chan entities.Registry)

	go func() {
		for {
			select {
			case reg := <-hkcr:
				resultsChan <- reg
			case reg := <-hkcu:
				resultsChan <- reg
			case reg := <-hklm:
				resultsChan <- reg
			case reg := <-hkcc:
				resultsChan <- reg
			case reg := <-hku:
				resultsChan <- reg
			}
		}
	}()

	return resultsChan
}

func generateRegistryByKey(key registry.Key) <-chan entities.Registry {

	regChan := make(chan entities.Registry)

	hkey, _ := registry.OpenKey(key, "", registry.READ)
	defer hkey.Close()

	go queryEnumKeys(hkey, utils.KeyToString(key), regChan)

	return regChan
}

func queryEnumKeys(hkey registry.Key, path string, regChan chan entities.Registry) {

	hkeyStat, err := hkey.Stat()
	if err != nil {
		return
	}

	if hkeyStat.SubKeyCount == 0 {
		queryEnumValues(hkey, path, regChan)
	}

	subKeys, err := hkey.ReadSubKeyNames(-1)
	if err != nil {
		return
	}

	queryEnumValues(hkey, path, regChan)

	for _, subkey := range subKeys {
		_hkey, _ := registry.OpenKey(hkey, subkey, registry.READ)
		defer _hkey.Close()
		queryEnumKeys(_hkey, path+"\\"+subkey, regChan)
	}

}

func queryEnumValues(hkey registry.Key, path string, regChan chan entities.Registry) {

	hkeyStat, err := hkey.Stat()
	if err != nil {
		log.Println("queryEnumValues failed to get stat", err.Error())
	}

	if hkeyStat.ValueCount == 0 {
		regChan <- entities.Registry{Path: path, ValueName: "", ValueType: "", Value: ""}
		return
	}

	valNames, err := hkey.ReadValueNames(-1)
	if err != nil {
		log.Println("queryEnumValues failed to get names", err.Error())
		return
	}

	regChan <- entities.Registry{Path: path, ValueName: "", ValueType: "", Value: ""}

	for _, name := range valNames {

		val, valType, err := queryValue(hkey, name)
		if err != nil {
			log.Println("QueryEnumValues failed to query value", err.Error())
			return
		}

		regChan <- entities.Registry{Path: path, ValueName: name, ValueType: valType, Value: val}
	}

}

func queryValue(hkey registry.Key, name string) (string, string, error) {

	value := make([]byte, 1024)

	n, valtype, err := hkey.GetValue(name, value)
	if err != nil {
		return "", "", fmt.Errorf("QueryValue failed to get value: %w", err)
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

	return "", "", fmt.Errorf("queryValue error: %w", err)
}
