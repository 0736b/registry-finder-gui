package models

import "github.com/0736b/registry-finder-gui/utils"

var registryTypeModel []string = []string{utils.STR_NONE, utils.STR_REG_SZ, utils.STR_REG_EXPAND_SZ, utils.STR_REG_BINARY, utils.STR_REG_DWORD,
	utils.STR_REG_DWORD_BIG_ENDIAN, utils.STR_REG_LINK, utils.STR_REG_MULTI_SZ, utils.STR_REG_RESOURCE_LIST, utils.STR_REG_FULL_RESOURCE_DESCRIPTOR,
	utils.STR_REG_RESOURCE_REQUIREMENTS_LIST, utils.STR_REG_QWORD}

func NewRegistryTypeModel() *[]string {

	return &registryTypeModel
}
