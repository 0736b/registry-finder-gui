package repositories

import (
	"fmt"
	"log"

	"github.com/0736b/registry-finder-gui/entities"
	"github.com/0736b/registry-finder-gui/utils"
	"golang.org/x/sys/windows/registry"
)

type RegistryRepository interface {
	StreamRegistry() <-chan *entities.Registry
}

type RegistryRepositoryImpl struct{}

func NewRegistryRepository() *RegistryRepositoryImpl {
	return &RegistryRepositoryImpl{}
}

func (r *RegistryRepositoryImpl) StreamRegistry() <-chan *entities.Registry {

	fromHKCR := generateRegistryByKey(registry.CLASSES_ROOT)
	fromHKCU := generateRegistryByKey(registry.CURRENT_USER)
	fromHKLM := generateRegistryByKey(registry.LOCAL_MACHINE)
	fromHKCC := generateRegistryByKey(registry.CURRENT_CONFIG)
	fromHKU := generateRegistryByKey(registry.USERS)

	results := fanInRegistry(fromHKCR, fromHKCU, fromHKLM, fromHKCC, fromHKU)

	return results
}

func fanInRegistry(hkcr, hkcu, hklm, hkcc, hku <-chan *entities.Registry) <-chan *entities.Registry {

	resultsChan := make(chan *entities.Registry)

	go func() {

		defer close(resultsChan)

		for hkcr != nil || hkcu != nil || hklm != nil || hkcc != nil || hku != nil {

			select {

			case reg, ok := <-hkcr:
				if !ok {
					hkcr = nil
				}
				if reg != nil {
					resultsChan <- reg
				}

			case reg, ok := <-hkcu:
				if !ok {
					hkcu = nil
				}
				if reg != nil {
					resultsChan <- reg
				}

			case reg, ok := <-hklm:
				if !ok {
					hklm = nil
				}
				if reg != nil {
					resultsChan <- reg
				}

			case reg, ok := <-hkcc:
				if !ok {
					hkcc = nil
				}
				if reg != nil {
					resultsChan <- reg
				}

			case reg, ok := <-hku:
				if !ok {
					hku = nil
				}
				if reg != nil {
					resultsChan <- reg
				}

			}
		}

	}()

	return resultsChan
}

func generateRegistryByKey(key registry.Key) <-chan *entities.Registry {

	regChan := make(chan *entities.Registry)
	// defer close(regChan)
	// log.Println("generateRegistryByKey regChan", &regChan)

	hkey, err := registry.OpenKey(key, "", registry.READ)
	if err != nil {
		log.Println("generate failed", err.Error())
		return nil
	}
	// defer hkey.Close()

	go queryEnumKeys(&hkey, utils.KeyToString(key), regChan)

	return regChan
}

func queryEnumKeys(hkey *registry.Key, path string, regChan chan *entities.Registry) {

	// log.Println("queryEnumKeys regChan", regChan)

	// log.Println("queryEnumKeys", hkey)

	_, err := hkey.Stat()
	if err != nil {
		regChan = nil
		return
	}

	// if hkeyStat.SubKeyCount == 0 {
	// 	queryEnumValues(hkey, path, regChan)
	// }

	subKeys, err := hkey.ReadSubKeyNames(-1)
	if err != nil {
		regChan = nil
		return
	}

	queryEnumValues(hkey, path, regChan)

	for _, subkey := range subKeys {
		_hkey, _ := registry.OpenKey(*hkey, subkey, registry.READ)
		// defer _hkey.Close()
		queryEnumKeys(&_hkey, path+"\\"+subkey, regChan)
	}

}

func queryEnumValues(hkey *registry.Key, path string, regChan chan *entities.Registry) {

	hkeyStat, err := hkey.Stat()
	if err != nil {
		log.Println("queryEnumValues failed to get stat", err.Error())
	}

	if hkeyStat.ValueCount == 0 {
		regChan <- &entities.Registry{Path: path, Name: "", Type: "", Value: ""}
		return
	}

	valNames, err := hkey.ReadValueNames(-1)
	if err != nil {
		log.Println("queryEnumValues failed to get names", err.Error())
		hkey.Close()
		return
	}

	regChan <- &entities.Registry{Path: path, Name: "", Type: "", Value: ""}

	for _, name := range valNames {

		val, valType, err := queryValue(hkey, name)
		if err != nil {
			log.Println("QueryEnumValues failed to query value", err.Error())
			return
		}

		regChan <- &entities.Registry{Path: path, Name: name, Type: valType, Value: val}
	}

	// hkey.Close()

}

func queryValue(hkey *registry.Key, name string) (string, string, error) {

	value := make([]byte, 1024)

	n, valType, err := hkey.GetValue(name, value)
	if err != nil && err != registry.ErrShortBuffer {
		return "", "", fmt.Errorf("QueryValue failed to get value: %w", err)
	} else if err != nil && err == registry.ErrShortBuffer {
		value = make([]byte, n)
		hkey.GetValue(name, value)
	}

	value = value[:n]

	switch valType {

	case registry.NONE:
		return "", utils.STR_NONE, nil

	case registry.SZ:
		strValue := utils.BytesToString(value)
		return strValue, utils.STR_REG_SZ, nil

	case registry.EXPAND_SZ:
		strValue := utils.BytesToString(value)
		return strValue, utils.STR_REG_EXPAND_SZ, nil

	case registry.BINARY:
		strValue := fmt.Sprintf("%x", value)
		return strValue, utils.STR_REG_BINARY, nil

	case registry.DWORD:
		// slices.Reverse(value)
		strValue := fmt.Sprintf("0x%x", utils.BytesToString(value))
		return strValue, utils.STR_REG_DWORD, nil

	case registry.DWORD_BIG_ENDIAN:
		strValue := fmt.Sprintf("0x%x", utils.BytesToString(value))
		return strValue, utils.STR_REG_DWORD_BIG_ENDIAN, nil

	case registry.LINK:
		strValue := utils.BytesToString(value)
		return strValue, utils.STR_REG_LINK, nil

	case registry.MULTI_SZ:
		val, _, _ := hkey.GetStringsValue(name)
		strValue := fmt.Sprintf("%s", val)
		return strValue, utils.STR_REG_MULTI_SZ, nil

	case registry.RESOURCE_LIST:
		strValue := utils.BytesToString(value)
		return strValue, utils.STR_REG_RESOURCE_LIST, nil

	case registry.FULL_RESOURCE_DESCRIPTOR:
		strValue := fmt.Sprintf("%x", value)
		return strValue, utils.STR_REG_FULL_RESOURCE_DESCRIPTOR, nil

	case registry.RESOURCE_REQUIREMENTS_LIST:
		strValue := utils.BytesToString(value)
		return strValue, utils.STR_REG_RESOURCE_REQUIREMENTS_LIST, nil

	case registry.QWORD:
		// slices.Reverse(value)
		strValue := fmt.Sprintf("%x", utils.BytesToString(value))
		return strValue, utils.STR_REG_QWORD, nil

	}

	return "", "", fmt.Errorf("queryValue error: %w", err)
}
