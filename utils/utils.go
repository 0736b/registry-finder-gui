package utils

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	STR_HKEY_CLASSES_ROOT   string = "HKEY_CLASSES_ROOT"
	STR_HKEY_CURRENT_USER   string = "HKEY_CURRENT_USER"
	STR_HKEY_LOCAL_MACHINE  string = "HKEY_LOCAL_MACHINE"
	STR_HKEY_USERS          string = "HKEY_USERS"
	STR_HKEY_CURRENT_CONFIG string = "HKEY_CURRENT_CONFIG"
)

const (
	STR_EMPTY                          string = ""
	STR_NONE                           string = "NONE"
	STR_REG_SZ                         string = "REG_SZ"
	STR_REG_EXPAND_SZ                  string = "REG_EXPAND_SZ"
	STR_REG_BINARY                     string = "REG_BINARY"
	STR_REG_DWORD                      string = "REG_DWORD"
	STR_REG_DWORD_BIG_ENDIAN           string = "REG_DWORD_BIG_ENDIAN"
	STR_REG_LINK                       string = "REG_LINK"
	STR_REG_MULTI_SZ                   string = "REG_MULTI_SZ"
	STR_REG_RESOURCE_LIST              string = "REG_RESOURCE_LIST"
	STR_REG_FULL_RESOURCE_DESCRIPTOR   string = "REG_FULL_RESOURCE_DESCRIPTOR"
	STR_REG_RESOURCE_REQUIREMENTS_LIST string = "REG_RESOURCE_REQUIREMENTS_LIST"
	STR_REG_QWORD                      string = "REG_QWORD"
)

const (
	FLAG_CREATE_NO_WINDOW uint32 = 0x08000000
)

func BytesToString(b []byte) string {

	n := bytes.Index(b, []byte{0, 0})
	if n == -1 {
		n = len(b)
	}
	return strings.ReplaceAll(string(b[:n]), "\x00", "") // clean unexpected null characters
}

func GetTypeString(valType uint32) string {

	switch valType {
	case registry.SZ:
		return STR_REG_SZ
	case registry.EXPAND_SZ:
		return STR_REG_EXPAND_SZ
	case registry.BINARY:
		return STR_REG_BINARY
	case registry.DWORD:
		return STR_REG_DWORD
	case registry.DWORD_BIG_ENDIAN:
		return STR_REG_DWORD_BIG_ENDIAN
	case registry.LINK:
		return STR_REG_LINK
	case registry.MULTI_SZ:
		return STR_REG_MULTI_SZ
	case registry.RESOURCE_LIST:
		return STR_REG_RESOURCE_LIST
	case registry.FULL_RESOURCE_DESCRIPTOR:
		return STR_REG_FULL_RESOURCE_DESCRIPTOR
	case registry.RESOURCE_REQUIREMENTS_LIST:
		return STR_REG_RESOURCE_REQUIREMENTS_LIST
	case registry.QWORD:
		return STR_REG_QWORD
	case registry.NONE:
		return STR_NONE
	default:
		return STR_EMPTY
	}
}

func MultiSZToStringSlice(value []byte) []string {

	var result []string
	for _, s := range bytes.Split(value, []byte{0, 0}) {
		if len(s) > 0 {
			result = append(result, BytesToString(s))
		}
	}
	return result
}

func KeyToString(key registry.Key) string {

	switch key {
	case registry.CLASSES_ROOT:
		return STR_HKEY_CLASSES_ROOT
	case registry.CURRENT_USER:
		return STR_HKEY_CURRENT_USER
	case registry.LOCAL_MACHINE:
		return STR_HKEY_LOCAL_MACHINE
	case registry.USERS:
		return STR_HKEY_USERS
	case registry.CURRENT_CONFIG:
		return STR_HKEY_CURRENT_CONFIG
	default:
		return STR_EMPTY
	}
}

func PreProcessStr(s string) string {

	return strings.ToLower(strings.ReplaceAll(s, " ", ""))
}

func OpenRegeditAtPath(path string) {

	addLastKeyCmd := exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Applets\\Regedit", "/v", "LastKey", "/t", "REG_SZ", "/d", path, "/f")
	addLastKeyCmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: FLAG_CREATE_NO_WINDOW,
	}

	err := addLastKeyCmd.Run()
	if err != nil {
		log.Println("OpenRegeditAtPath failed to add last key", err.Error())
	}

	openRegeditCmd := exec.Command("cmd", "/c", "regedit", "/m")
	openRegeditCmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: FLAG_CREATE_NO_WINDOW,
	}

	err = openRegeditCmd.Run()
	if err != nil {
		log.Println("OpenRegeditAtPath failed to open regedit", err.Error())
	}
}
