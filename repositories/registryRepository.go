package repositories

import (
	"fmt"
	"log"
	"sync"

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

// TODO logic/performance improving on all related to this
func (r *RegistryRepositoryImpl) StreamRegistry() <-chan *entities.Registry {

	genByHKCR := generateRegByKey(registry.CLASSES_ROOT)
	genByHKCU := generateRegByKey(registry.CURRENT_USER)
	genByHKLM := generateRegByKey(registry.LOCAL_MACHINE)
	genByHKCC := generateRegByKey(registry.CURRENT_CONFIG)
	genByHKU := generateRegByKey(registry.USERS)

	stream := fanInGenerate(genByHKCR, genByHKCU, genByHKLM, genByHKCC, genByHKU)

	return stream
}

func fanInGenerate(genChans ...<-chan *entities.Registry) <-chan *entities.Registry {

	var wg sync.WaitGroup
	streamChan := make(chan *entities.Registry)

	fanIn := func(gen <-chan *entities.Registry) {
		defer wg.Done()
		for gen != nil {
			reg, ok := <-gen
			if !ok {
				gen = nil
			}
			streamChan <- reg
		}
	}

	wg.Add(len(genChans))
	for _, gen := range genChans {
		go fanIn(gen)
	}

	go func() {
		wg.Wait()
		close(streamChan)
	}()

	return streamChan
}

func generateRegByKey(key registry.Key) <-chan *entities.Registry {

	regChan := make(chan *entities.Registry)

	hkey, err := registry.OpenKey(key, "", registry.READ)
	if err != nil {
		log.Println("generateRegByKey OpenKey failed", err.Error())
		return nil
	}

	go queryEnumKeys(&hkey, utils.KeyToString(key), regChan)

	return regChan
}

func queryEnumKeys(hkey *registry.Key, path string, regChan chan *entities.Registry) {

	_, err := hkey.Stat()
	if err != nil {
		return
	}

	subKeys, err := hkey.ReadSubKeyNames(-1)
	if err != nil {
		return
	}

	queryEnumValues(hkey, path, regChan)

	for _, subkey := range subKeys {
		_hkey, _ := registry.OpenKey(*hkey, subkey, registry.READ)
		queryEnumKeys(&_hkey, path+"\\"+subkey, regChan)
	}

}

func queryEnumValues(hkey *registry.Key, path string, regChan chan *entities.Registry) {

	hkeyStat, err := hkey.Stat()
	if err != nil {
		log.Println("queryEnumValues Stat() failed", err.Error())
	}

	if hkeyStat.ValueCount == 0 {
		regChan <- &entities.Registry{Path: path, Name: "", Type: "", Value: ""}
		return
	}

	valNames, err := hkey.ReadValueNames(-1)
	if err != nil {
		log.Println("queryEnumValues ReadValueNames() failed", err.Error())
		hkey.Close()
		return
	}

	regChan <- &entities.Registry{Path: path, Name: "", Type: "", Value: ""}

	for _, name := range valNames {

		val, valType, err := queryValue(hkey, name)
		if err != nil {
			log.Println("QueryEnumValues queryValue failed", err.Error())
			return
		}

		regChan <- &entities.Registry{Path: path, Name: name, Type: valType, Value: val}
	}

}

// TODO make value string look like in Regedit
func queryValue(hkey *registry.Key, name string) (string, string, error) {

	value := make([]byte, 1024)

	n, valType, err := hkey.GetValue(name, value)
	if err != nil && err != registry.ErrShortBuffer {
		return "", "", fmt.Errorf("%w", err)
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
		strValue := fmt.Sprintf("%x", utils.BytesToString(value))
		return strValue, utils.STR_REG_QWORD, nil

	}

	return "", "", nil
}
