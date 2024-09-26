package repositories

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"
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

func queryValue(hkey *registry.Key, name string) (string, string, error) {

	n, valType, err := hkey.GetValue(name, nil)
	if err != nil {
		if err == registry.ErrNotExist {
			return "", "", fmt.Errorf("value does not exist: %w", err)
		}
		return "", "", fmt.Errorf("failed to get value info: %w", err)
	}

	buf := make([]byte, n)
	_, _, err = hkey.GetValue(name, buf)
	if err != nil {
		return "", "", fmt.Errorf("failed to get value data: %w", err)
	}

	var strValue string
	var typeStr string

	switch valType {
	case registry.NONE:
		typeStr = utils.STR_NONE
	case registry.SZ, registry.EXPAND_SZ, registry.LINK:
		strValue = utils.BytesToString(buf)
		typeStr = utils.GetTypeString(valType)
	case registry.BINARY, registry.FULL_RESOURCE_DESCRIPTOR:
		strValue = fmt.Sprintf("%x", buf)
		typeStr = utils.GetTypeString(valType)
	case registry.DWORD:
		if len(buf) >= 4 {
			strValue = fmt.Sprintf("0x%08x", binary.LittleEndian.Uint32(buf))
		}
		typeStr = utils.STR_REG_DWORD
	case registry.DWORD_BIG_ENDIAN:
		if len(buf) >= 4 {
			strValue = fmt.Sprintf("0x%08x", binary.BigEndian.Uint32(buf))
		}
		typeStr = utils.STR_REG_DWORD_BIG_ENDIAN
	case registry.QWORD:
		if len(buf) >= 8 {
			strValue = fmt.Sprintf("0x%016x", binary.LittleEndian.Uint64(buf))
		}
		typeStr = utils.STR_REG_QWORD
	case registry.MULTI_SZ:
		strValue = strings.Join(utils.MultiSZToStringSlice(buf), ", ")
		typeStr = utils.STR_REG_MULTI_SZ
	default:
		strValue = string(buf)
		typeStr = utils.GetTypeString(valType)
	}

	return strValue, typeStr, nil
}
