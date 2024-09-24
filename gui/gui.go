package gui

import (
	"time"

	"github.com/0736b/registry-finder-gui/gui/models"
	"github.com/0736b/registry-finder-gui/usecases"
	"github.com/lxn/walk"

	//lint:ignore ST1001 don't worry trust me
	. "github.com/lxn/walk/declarative"
)

const (
	APP_TITLE         string        = "Registry Finder"
	APP_WIDTH         int           = 1000
	APP_HEIGHT        int           = 800
	DEBOUNCE_INTERVAL time.Duration = 300 * time.Millisecond

	COL_TITLE_PATH  string = "Path"
	COL_TITLE_NAME  string = "Name"
	COL_TITLE_TYPE  string = "Type"
	COL_TITLE_VALUE string = "Value"
)

type AppWindow struct {
	*walk.MainWindow
	usecase   usecases.RegistryUsecase
	searchBox *walk.LineEdit
	// keyCheckBox  *walk.CheckBox
	// typeCheckBox *walk.CheckBox
	// keyComboBox  *walk.ComboBox
	// typeComboBox *walk.ComboBox
	resultTable *walk.TableView

	regTableModel *models.RegistryTableModel
}

func NewAppWindow(usecase usecases.RegistryUsecase) (*AppWindow, error) {

	app := &AppWindow{usecase: usecase, regTableModel: models.NewRegistryTableModel()}

	mw := MainWindow{

		AssignTo: &app.MainWindow,
		Title:    APP_TITLE,
		Size:     Size{Width: APP_WIDTH, Height: APP_HEIGHT},
		MinSize:  Size{Width: APP_WIDTH, Height: APP_HEIGHT},
		Layout:   VBox{},

		Children: []Widget{
			LineEdit{
				AssignTo: &app.searchBox,
				OnTextChanged: func() {
					// TODO keyword change
				},
			},
			TableView{
				AssignTo:         &app.resultTable,
				AlternatingRowBG: true,
				Columns: []TableViewColumn{
					{Title: COL_TITLE_PATH, Width: int(40 * APP_WIDTH / 100)},
					{Title: COL_TITLE_NAME, Width: int(10 * APP_WIDTH / 100)},
					{Title: COL_TITLE_TYPE, Width: int(10 * APP_WIDTH / 100)},
					{Title: COL_TITLE_VALUE, Width: int(40 * APP_WIDTH / 100)},
				},
				Model: app.regTableModel,
				OnItemActivated: func() {
					// TODO double-clicked to open regedit at item's key path
				},
			},
		},
	}

	if err := mw.Create(); err != nil {
		return nil, err
	}

	go func() {
		// collect streaming registry
	}()

	return app, nil
}
