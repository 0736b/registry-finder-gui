package models

import (
	"github.com/0736b/registry-finder-gui/entities"
	"github.com/lxn/walk"
)

type RegistryTableModel struct {
	walk.TableModelBase
	Items []*entities.Registry
}

func NewRegistryTableModel() *RegistryTableModel {

	m := new(RegistryTableModel)
	m.RowsReset()
	return m
}

func (m *RegistryTableModel) RowCount() int {

	return len(m.Items)
}

func (m *RegistryTableModel) Value(row, col int) interface{} {

	item := m.Items[row]

	switch col {
	case 0:
		return *item.Path
	case 1:
		return *item.Name
	case 2:
		return *item.Type
	case 3:
		return *item.Value
	}

	panic("unexpected col")
}
