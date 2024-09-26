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
	UPDATE_INTERVAL   time.Duration = 250 * time.Millisecond

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
	updateShowed   chan bool

	debounce   *time.Timer
	debounceMu sync.Mutex

	keywordChan chan string

	keyChan          chan string
	keyEnabledChan   chan bool
	filterKeyEnabled bool

	typeChan          chan string
	typeEnabledChan   chan bool
	filterTypeEnabled bool

	*walk.MainWindow
	searchBox *walk.LineEdit

	keyCheckBox  *walk.CheckBox
	typeCheckBox *walk.CheckBox
	keyComboBox  *walk.ComboBox
	typeComboBox *walk.ComboBox

	regKeyModel  *[]string
	regTypeModel *[]string

	resultTable   *walk.TableView
	regTableModel *models.RegistryTableModel
}

func NewAppWindow(usecase usecases.RegistryUsecase) (*AppWindow, error) {

	app := &AppWindow{usecase: usecase,
		collectedResult: make([]*entities.Registry, 0), showedResult: make([]*entities.Registry, 0),
		regTableModel: models.NewRegistryTableModel(), updateShowed: make(chan bool),
		keywordChan:       make(chan string),
		keyEnabledChan:    make(chan bool),
		typeEnabledChan:   make(chan bool),
		keyChan:           make(chan string),
		typeChan:          make(chan string),
		filterKeyEnabled:  false,
		filterTypeEnabled: false,
		regKeyModel:       models.NewRegistryKeyModel(),
		regTypeModel:      models.NewRegistryTypeModel(),
	}

	var icon, _ = walk.NewIconFromResourceId(2)

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

			Composite{
				Layout: HBox{},
				Children: []Widget{
					CheckBox{
						AssignTo:       &app.keyCheckBox,
						Text:           "Filter Key",
						TextOnLeftSide: true,
						OnClicked: func() {
							app.onFilterKeyChecked()
						},
					},
					ComboBox{
						AssignTo:     &app.keyComboBox,
						Editable:     false,
						Model:        *app.regKeyModel,
						CurrentIndex: 0,
						OnCurrentIndexChanged: func() {
							app.onFilterKeyChanged()
						},
					},
					CheckBox{
						AssignTo:       &app.typeCheckBox,
						Text:           "Filter Type",
						TextOnLeftSide: true,
						OnClicked: func() {
							app.onFilterTypeChecked()
						},
					},
					ComboBox{
						AssignTo:     &app.typeComboBox,
						Editable:     false,
						Model:        *app.regTypeModel,
						CurrentIndex: 0,
						OnCurrentIndexChanged: func() {
							app.onFilterTypeChanged()
						},
					},
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

	_ = app.SetIcon(icon)

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

// TODO find better way
func (app *AppWindow) processingShowResult() {

	var currKeyword string
	var currKeyEnabled bool
	var currTypeEnabled bool
	var currKey string
	var currType string

	updateAndFilter := func(forceUpdate bool) {

		go func(keyword string, keyEnabled bool, typeEnabled bool, filterKey, filterType string) {

			app.collectedResultMu.Lock()
			collectedCopy := make([]*entities.Registry, len(app.collectedResult))
			copy(collectedCopy, app.collectedResult)
			app.collectedResultMu.Unlock()

			filtered := make([]*entities.Registry, 0, len(collectedCopy))

			for _, reg := range collectedCopy {
				if app.usecase.FilterByKeyword(reg, keyword) {
					if keyEnabled && !typeEnabled && app.usecase.FilterByKey(reg, filterKey) {
						filtered = append(filtered, reg)
					} else if typeEnabled && !keyEnabled && app.usecase.FilterByType(reg, filterType) {
						filtered = append(filtered, reg)
					} else if keyEnabled && typeEnabled && app.usecase.FilterByKey(reg, filterKey) && app.usecase.FilterByType(reg, filterType) {
						filtered = append(filtered, reg)
					} else if !keyEnabled && !typeEnabled {
						filtered = append(filtered, reg)
					}
				}
			}

			app.showedResultMu.Lock()
			app.showedResult = filtered
			app.showedResultMu.Unlock()

			if forceUpdate {
				select {
				case app.updateShowed <- true:
				default:
				}
			}

		}(currKeyword, currKeyEnabled, currTypeEnabled, currKey, currType)

	}

	for {

		select {

		case newKeyword := <-app.keywordChan:
			if newKeyword != currKeyword {
				currKeyword = newKeyword
				updateAndFilter(true)
			}

		case newKeyEnabled := <-app.keyEnabledChan:
			if newKeyEnabled != currKeyEnabled {
				currKeyEnabled = newKeyEnabled
				updateAndFilter(true)
			}

		case newTypeEnabled := <-app.typeEnabledChan:
			if newTypeEnabled != currTypeEnabled {
				currTypeEnabled = newTypeEnabled
				updateAndFilter(true)
			}

		case newKey := <-app.keyChan:
			if newKey != currKey {
				currKey = newKey
				if currKeyEnabled {
					updateAndFilter(true)
				}
			}

		case newType := <-app.typeChan:
			if newType != currType {
				currType = newType
				if currTypeEnabled {
					updateAndFilter(true)
				}
			}

		case <-time.After(UPDATE_INTERVAL):
			updateAndFilter(false)
		}

	}

}

func (app *AppWindow) updatingTable() {

	ticker := time.NewTicker(UPDATE_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			app.updateTable(false)
		case inv := <-app.updateShowed:
			app.updateTable(inv)
		}
	}
}

func (app *AppWindow) updateTable(invalidate bool) {

	app.Synchronize(func() {

		app.showedResultMu.Lock()
		showedCopy := make([]*entities.Registry, len(app.showedResult))
		copy(showedCopy, app.showedResult)
		app.showedResultMu.Unlock()

		if invalidate {
			app.regTableModel.Items = showedCopy
			app.regTableModel.PublishRowsReset()
			app.resultTable.Invalidate()
		} else {
			if len(showedCopy) > len(app.regTableModel.Items) {
				app.regTableModel.Items = showedCopy
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

func (app *AppWindow) onFilterKeyChecked() {

	app.debounceMu.Lock()
	defer app.debounceMu.Unlock()

	if app.debounce != nil {
		app.debounce.Stop()
	}

	app.debounce = time.AfterFunc(0, func() {
		app.filterKeyEnabled = !app.filterKeyEnabled
		app.keyCheckBox.SetChecked(app.filterKeyEnabled)
		select {
		case app.keyEnabledChan <- app.filterKeyEnabled:
			filterKey := app.keyComboBox.Text()
			app.keyChan <- filterKey
		default:
		}
	})

}

func (app *AppWindow) onFilterKeyChanged() {

	filterKey := app.keyComboBox.Text()

	app.debounceMu.Lock()
	defer app.debounceMu.Unlock()

	if app.debounce != nil {
		app.debounce.Stop()
	}

	app.debounce = time.AfterFunc(0, func() {
		select {
		case app.keyChan <- filterKey:
		default:
		}
	})

}

func (app *AppWindow) onFilterTypeChecked() {

	app.debounceMu.Lock()
	defer app.debounceMu.Unlock()

	if app.debounce != nil {
		app.debounce.Stop()
	}

	app.debounce = time.AfterFunc(0, func() {
		app.filterTypeEnabled = !app.filterTypeEnabled
		app.typeCheckBox.SetChecked(app.filterTypeEnabled)
		select {
		case app.typeEnabledChan <- app.filterTypeEnabled:
			filterType := app.typeComboBox.Text()
			app.typeChan <- filterType
		default:
		}
	})
}

func (app *AppWindow) onFilterTypeChanged() {

	filterType := app.typeComboBox.Text()

	app.debounceMu.Lock()
	defer app.debounceMu.Unlock()

	if app.debounce != nil {
		app.debounce.Stop()
	}

	app.debounce = time.AfterFunc(0, func() {
		select {
		case app.typeChan <- filterType:
		default:
		}
	})

}
