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

	DEBOUNCE_INTERVAL time.Duration = 250 * time.Millisecond
	UPDATE_INTERVAL   time.Duration = 500 * time.Millisecond

	COL_TITLE_PATH  string = "Path"
	COL_TITLE_NAME  string = "Name"
	COL_TITLE_TYPE  string = "Type"
	COL_TITLE_VALUE string = "Value"

	COL_WIDTH_PATH  float32 = 0.4
	COL_WIDTH_NAME  float32 = 0.1
	COL_WIDTH_TYPE  float32 = 0.1
	COL_WIDTH_VALUE float32 = 0.4
)

type AppWindow struct {
	usecase usecases.RegistryUsecase

	collectedResult   []*entities.Registry
	collectedResultMu sync.Mutex

	showedResult   []*entities.Registry
	showedResultMu sync.Mutex
	updateShowed   chan struct{}

	debounce    *time.Timer
	debounceMu  sync.Mutex
	keywordChan chan string

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

	app := &AppWindow{usecase: usecase, collectedResult: make([]*entities.Registry, 0), showedResult: make([]*entities.Registry, 0),
		regTableModel: models.NewRegistryTableModel(), keywordChan: make(chan string), updateShowed: make(chan struct{}, 1)}

	mw := MainWindow{

		AssignTo: &app.MainWindow,
		Title:    APP_TITLE,
		Size:     Size{Width: APP_WIDTH, Height: APP_HEIGHT},
		MinSize:  Size{Width: APP_WIDTH, Height: APP_HEIGHT},
		Layout:   VBox{},
		OnSizeChanged: func() {
			go app.handleOnSizeChanged()
		},

		Children: []Widget{
			LineEdit{
				AssignTo: &app.searchBox,
				OnTextChanged: func() {
					go app.handleOnKeywordChanged()
				},
			},
			TableView{
				AssignTo:         &app.resultTable,
				AlternatingRowBG: true,
				Columns: []TableViewColumn{
					{Name: COL_TITLE_PATH, Title: COL_TITLE_PATH, Width: int(COL_WIDTH_PATH * float32(APP_WIDTH))},
					{Name: COL_TITLE_NAME, Title: COL_TITLE_NAME, Width: int(COL_WIDTH_NAME * float32(APP_WIDTH))},
					{Name: COL_TITLE_TYPE, Title: COL_TITLE_TYPE, Width: int(COL_WIDTH_TYPE * float32(APP_WIDTH))},
					{Name: COL_TITLE_VALUE, Title: COL_TITLE_VALUE, Width: int(COL_WIDTH_VALUE * float32(APP_WIDTH))},
				},
				Model: app.regTableModel,
				OnItemActivated: func() {
					app.handleOnItemActivated()
				},
			},
		},
	}

	if err := mw.Create(); err != nil {
		return nil, err
	}

	go app.streamingRegistry()

	go app.processingShowResult()

	go app.updatingTable()

	return app, nil
}

func (app *AppWindow) streamingRegistry() {

	for reg := range app.usecase.StreamRegistry() {
		app.collectedResultMu.Lock()
		app.collectedResult = append(app.collectedResult, reg)
		app.collectedResultMu.Unlock()
	}
}

func (app *AppWindow) processingShowResult() {

	var prevKeyword string = ""
	var currKeyword string = ""

	for {

		select {
		case newKeyword := <-app.keywordChan:
			prevKeyword = currKeyword
			currKeyword = newKeyword
			app.updateShowed <- struct{}{}
		case <-time.After(UPDATE_INTERVAL):
		}

		go func() {

			app.showedResultMu.Lock()
			defer app.showedResultMu.Unlock()

			app.showedResult = make([]*entities.Registry, 0)
			for _, reg := range app.collectedResult {
				if app.usecase.FilterByKeyword(reg, currKeyword) {
					app.showedResult = append(app.showedResult, reg)
				}
			}

			if prevKeyword != currKeyword {
				select {
				case app.updateShowed <- struct{}{}:
					prevKeyword = currKeyword
				default:
				}
			}

		}()

	}

}

func (app *AppWindow) updatingTable() {

	ticker := time.NewTicker(UPDATE_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			app.updateTable(false)
		case <-app.updateShowed:
			app.updateTable(true)
		}
	}
}

func (app *AppWindow) updateTable(invalidate bool) {

	app.Synchronize(func() {

		app.showedResultMu.Lock()
		defer app.showedResultMu.Unlock()

		if invalidate {
			app.regTableModel.Items = app.showedResult
			app.regTableModel.PublishRowsReset()
			app.resultTable.Invalidate()
		} else {
			if len(app.showedResult) > len(app.regTableModel.Items) {
				app.regTableModel.Items = app.showedResult
				app.regTableModel.PublishRowsReset()
			}
		}

	})
}

func (app *AppWindow) handleOnKeywordChanged() {

	keyword := app.searchBox.Text()

	app.debounceMu.Lock()
	defer app.debounceMu.Unlock()

	if app.debounce != nil {
		app.debounce.Stop()
	}

	app.debounce = time.AfterFunc(DEBOUNCE_INTERVAL, func() {
		select {
		case app.keywordChan <- keyword:
		default:
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

func (app *AppWindow) handleOnSizeChanged() {

	app.Synchronize(func() {
		go app.resultTable.Columns().ByName(COL_TITLE_PATH).SetWidth(int(float32(app.Width()) * (COL_WIDTH_PATH)))
		go app.resultTable.Columns().ByName(COL_TITLE_NAME).SetWidth(int(float32(app.Width()) * (COL_WIDTH_NAME)))
		go app.resultTable.Columns().ByName(COL_TITLE_TYPE).SetWidth(int(float32(app.Width()) * (COL_WIDTH_TYPE)))
		go app.resultTable.Columns().ByName(COL_TITLE_VALUE).SetWidth(int(float32(app.Width()) * (COL_WIDTH_VALUE)))
	})

}
