package gui

import (
	"sync"
	"time"

	"github.com/0736b/registry-finder-gui/entities"
	"github.com/0736b/registry-finder-gui/gui/models"
	"github.com/0736b/registry-finder-gui/usecases"
	"github.com/lxn/walk"

	//lint:ignore ST1001 don't worry trust me
	. "github.com/lxn/walk/declarative"
)

const (
	APP_TITLE  string = "Registry Finder"
	APP_WIDTH  int    = 1000
	APP_HEIGHT int    = 800

	DEBOUNCE_INTERVAL time.Duration = 300 * time.Millisecond
	UPDATE_INTERVAL   time.Duration = 300 * time.Millisecond
	MAX_SHOW_RESULT   int           = 1000

	COL_TITLE_PATH  string = "Path"
	COL_TITLE_NAME  string = "Name"
	COL_TITLE_TYPE  string = "Type"
	COL_TITLE_VALUE string = "Value"
)

type AppWindow struct {
	usecase           usecases.RegistryUsecase
	collectedResult   []*entities.Registry
	collectedResultMu sync.Mutex
	showedResult      []*entities.Registry

	*walk.MainWindow
	searchBox *walk.LineEdit
	// keyCheckBox  *walk.CheckBox
	// typeCheckBox *walk.CheckBox
	// keyComboBox  *walk.ComboBox
	// typeComboBox *walk.ComboBox
	resultTable   *walk.TableView
	regTableModel *models.RegistryTableModel
}

func NewAppWindow(usecase usecases.RegistryUsecase) (*AppWindow, error) {

	app := &AppWindow{usecase: usecase, collectedResult: make([]*entities.Registry, 0), showedResult: make([]*entities.Registry, 0), regTableModel: models.NewRegistryTableModel()}

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
					// TODO keyword changed, show result must change
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
					app.handleOnItemActivated()
				},
				OnSizeChanged: func() {
					// TODO adjust value column's size by app's size
				},
			},
		},
	}

	if err := mw.Create(); err != nil {
		return nil, err
	}

	go app.streamingRegistry()

	go app.updatingTable()

	return app, nil
}

func (app *AppWindow) streamingRegistry() {

	for reg := range app.usecase.StreamRegistry() {
		// log.Println(len(app.collectedResult))
		// app.collectedResultMu.Lock()
		app.collectedResult = append(app.collectedResult, reg)
		// app.collectedResultMu.Unlock()
		// log.Println(reg)
		// log.Println(len(app.collectedResult))
	}
}

func (app *AppWindow) updatingTable() {

	ticker := time.NewTicker(UPDATE_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		app.updateTable(false)
	}

}

func (app *AppWindow) updateTable(invalidate bool) {

	app.Synchronize(func() {

		app.showedResult = make([]*entities.Registry, 0)

		app.collectedResultMu.Lock()
		app.showedResult = append(app.showedResult, app.collectedResult...)
		app.collectedResultMu.Unlock()

		app.regTableModel.Items = app.showedResult
		app.regTableModel.PublishRowsReset()

		if invalidate {
			app.resultTable.Invalidate()
		}

	})
}

func (app *AppWindow) handleOnItemActivated() {

	index := app.resultTable.CurrentIndex()
	if index < 0 {
		return
	}

	item := app.regTableModel.Items[index]
	go app.usecase.OpenInRegedit(item)

}
