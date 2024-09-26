package models

import "github.com/0736b/registry-finder-gui/utils"

var registryKeyModel []string = []string{utils.STR_HKEY_CLASSES_ROOT, utils.STR_HKEY_CURRENT_USER, utils.STR_HKEY_LOCAL_MACHINE, utils.STR_HKEY_USERS, utils.STR_HKEY_CURRENT_CONFIG}

func NewRegistryKeyModel() *[]string {

	return &registryKeyModel
}
