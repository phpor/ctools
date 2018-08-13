package main

import (
	ui "github.com/cjbassi/termui"
	"time"
	"github.com/phpor/ctools/cpu"
	"os"
	"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/phpor/ctools/mem"
)
func main() {
	spew.Dump(os.Args)

	if len(os.Args) > 1 && os.Args[1] == "play" {
		mainPlay()

	} else {
		mainUI()
	}
}

func mainPlay()  {
	fmt.Printf("CpuUsage: %f%%\n", cpu.GetCpuUsage()*100)
	memstat,err := mem.Usage()
	if err != nil {
		spew.Dump(err. Error())
	}
	spew.Dump(memstat)

	sysMemStat, _ := mem.GetSystemMemStat()
	spew.Dump(sysMemStat)
}

func mainUI() {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()


	g := ui.NewGauge()
	g.X = 40
	g.Y = 1
	g.Percent = int(cpu.GetCpuUsageNoDelay() * 100)
	g.Label = "Gauge CPU"

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			g.Percent = int(cpu.GetCpuUsage() * 100)
			ui.Render(g)
		}
	}()
	ui.Render(g)
	// quits
	ui.On("q", "<C-c>", func(e ui.Event) {
		ui.StopLoop()
	})

	ui.Loop()
}
