package models

import (
	"github.com/0736b/registry-finder-gui/entities"
	"github.com/lxn/walk"
)

type RegistryTableModel struct {
	walk.TableModelBase
	items []entities.Registry
}

func NewRegistryTableModel() *RegistryTableModel {

	m := new(RegistryTableModel)
	m.RowsReset()
	return m
}

func (m *RegistryTableModel) RowCount() int {
	return len(m.items)
}

func (m *RegistryTableModel) Value(row, col int) interface{} {

	item := m.items[row]

	switch col {
	case 0:
		return item.Path
	case 1:
		return item.ValueName
	case 2:
		return item.ValueType
	case 3:
		return item.Value
	}

	panic("unexpected col")
}
