package main

import (
	"log"

	"github.com/0736b/registry-finder-gui/gui"
	"github.com/0736b/registry-finder-gui/usecases"
)

func main() {

	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatalln("failed to create cpu profile", err.Error())
	// }
	// defer f.Close()

	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	log.Fatalln("failed to start cpu profile", err.Error())
	// }
	// defer pprof.StopCPUProfile()

	usecase := usecases.NewRegistryUsecase()

	app, err := gui.NewAppWindow(usecase)
	if err != nil {
		log.Fatalln("failed to create app window", err.Error())
	}

	app.Run()

}
