package main

import (
	"log"

	"github.com/0736b/registry-finder-gui/gui"
	"github.com/0736b/registry-finder-gui/usecases"
)

func main() {

	ru := usecases.NewRegistryUsecase()

	app, err := gui.NewAppWindow(ru)
	if err != nil {
		log.Fatalln("failed to create app window", err.Error())
	}

	app.Run()

}
