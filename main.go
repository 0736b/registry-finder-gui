package main

import (
	"bytes"
	"fmt"
	"log"
	"slices"

	"golang.org/x/sys/windows/registry"
)

func BytesToString(b []byte) string {

	n := bytes.Index(b, []byte{0, 0})
	if n == -1 {
		n = len(b)
	}

	return string(b[:n])
}

type RegistryModel struct {
	Path      string
	ValueName string
	ValueType string
	Value     string
}

func QueryEnumKeys(hkey registry.Key, path string) ([]RegistryModel, error) {

	// log.Println("key: ", key)
	// _hkey, _ := registry.OpenKey(hkey, key, registry.READ)
	// defer _hkey.Close()

	results := make([]RegistryModel, 0)

	hkeyStat, err := hkey.Stat()
	if err != nil {
		// log.Println("failed stat", err.Error())
		return make([]RegistryModel, 0), fmt.Errorf("QueryEnumKeys failed to get stat: %w", err)
	}

	if hkeyStat.SubKeyCount == 0 {
		return QueryEnumValues(hkey, path)
	}

	subKeys, err := hkey.ReadSubKeyNames(-1)
	if err != nil {
		// log.Println("can't get subkeys", err.Error())
		return make([]RegistryModel, 0), fmt.Errorf("QueryEnumKeys failed to get subkeys: %w", err)
	}

	values, _ := QueryEnumValues(hkey, path)

	for _, subkey := range subKeys {
		_hkey, _ := registry.OpenKey(hkey, subkey, registry.READ)
		defer _hkey.Close()
		enumsValues, _ := QueryEnumKeys(_hkey, path+"\\"+subkey)
		results = append(results, enumsValues...)
	}

	results = append(results, values...)

	go func() {
		for _, r := range results {
			log.Println(r)
		}
	}()

	return results, fmt.Errorf("QueryEnumKeys failed to enum: %w", err)
}

func QueryEnumValues(hkey registry.Key, path string) ([]RegistryModel, error) {

	hkeyStat, err := hkey.Stat()
	if err != nil {
		return make([]RegistryModel, 0), fmt.Errorf("QueryEnumValues failed to get stat: %w", err)
	}

	values := make([]RegistryModel, 0)

	if hkeyStat.ValueCount == 0 {
		// values = append(values, RegistryModel{path, "", "", ""})
		log.Println(RegistryModel{path, "", "", ""})
		return values, nil
	}

	valNames, err := hkey.ReadValueNames(-1)
	if err != nil {
		// log.Println(RegistryModel{path, "", "", ""})
		return values, fmt.Errorf("QueryEnumValues failed to get names: %w", err)
	}

	log.Println(RegistryModel{path, "", "", ""})
	for _, name := range valNames {

		val, valType, err := QueryValue(hkey, name)
		if err != nil {
			// log.Println(RegistryModel{path, "", "", ""})
			return values, fmt.Errorf("QueryEnumValues - %w", err)
		}

		// values = append(values, RegistryModel{Path: path,
		// 	ValueName: name, ValueType: valType, Value: val})
		log.Println(RegistryModel{Path: path,
			ValueName: name, ValueType: valType, Value: val})
	}

	return values, nil
}

func QueryValue(hkey registry.Key, name string) (string, string, error) {

	// log.Println("name", name)

	value := make([]byte, 1024)

	n, valtype, err := hkey.GetValue(name, value)
	if err != nil {
		// log.Println("QueryValue", err.Error())
		return "", "", fmt.Errorf("QueryValue error: %w", err)
	}

	value = value[:n]

	// log.Println("Type:", valtype, "Value:", value)

	switch valtype {

	case registry.NONE:
		return "", "NONE", nil

	case registry.SZ:
		return BytesToString(value), "REG_SZ", nil

	case registry.EXPAND_SZ:
		return BytesToString(value), "REG_EXPAND_SZ", nil

	case registry.BINARY:
		return fmt.Sprintf("%-1x", value), "REG_BINARY", nil

	case registry.DWORD:
		// slices.Reverse(value) // To Little Endian
		return fmt.Sprintf("0x%x", BytesToString(value)), "REG_DWORD", nil

	case registry.DWORD_BIG_ENDIAN:
		return fmt.Sprintf("0x%x", BytesToString(value)), "REG_DWORD_BIG_ENDIAN", nil

	case registry.LINK:
		return BytesToString(value), "REG_LINK", nil

	case registry.MULTI_SZ:
		val, _, _ := hkey.GetStringsValue(name)
		return fmt.Sprintf("%s", val), "REG_MULTI_SZ", nil

	case registry.RESOURCE_LIST:
		return BytesToString(value), "REG_RESOURCE_LIST", nil

	case registry.FULL_RESOURCE_DESCRIPTOR:
		return fmt.Sprintf("%-2x", value), "REG_FULL_RESOURCE_DESCRIPTOR", nil

	case registry.RESOURCE_REQUIREMENTS_LIST:
		return BytesToString(value), "REG_RESOURCE_REQUIREMENTS_LIST", nil

	case registry.QWORD:
		slices.Reverse(value)
		return fmt.Sprintf("%x", BytesToString(value)), "REG_QWORD", nil

	}

	return "", "", fmt.Errorf("QueryValue error: %w", err)
}

func QueryRegistry(keyword string) {

}

func main() {

	log.Println("[BETTER-REG-QUERY]")

	hkey, _ := registry.OpenKey(registry.LOCAL_MACHINE, "", registry.READ)
	defer hkey.Close()

	QueryEnumKeys(hkey, "HKEY_LOCAL_MACHINE")
	// log.Println(results)

	// if err != nil {
	// 	log.Fatalln("failed open key", err.Error())
	// }

	// val, valtype, err := QueryValue(hkey, "Configuration Data")
	// if err != nil {
	// 	log.Fatalln(err.Error())
	// }
	// log.Println(valtype, val)

	// for {
	// 	time.Sleep(1 * time.Second)
	// }

}
