package main

import (
	ui "github.com/cjbassi/termui"
	"time"
	"github.com/phpor/ctools/cpu"
	"os"
	"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/phpor/ctools/mem"
	"github.com/phpor/ctools/utils"
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
	g.Label = "Gauge CPU"


	m := ui.NewGauge()
	m.YOffset = 4
	m.X = 40
	m.Y = 1
	m.Label = "Gauge Mem"

	var update = func() {
		g.Percent = int(cpu.GetCpuUsageNoDelay() * 100)
		memstat,_ := mem.Usage()
		percent := float64(memstat.Used) / float64(memstat.Total) * 100
		m.Percent = int(percent)
		memUsed, memUsedUnit := utils.FormatBytes(memstat.Used)
		memTotal, memTotalUnit := utils.FormatBytes(memstat.Total)
		m.Label = fmt.Sprintf("Mem: %3.1f%%  %.1f %s/%.1f %s", percent, memUsed, memUsedUnit, memTotal, memTotalUnit)

	}

	update()
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			update()
			ui.Render(g,m)
		}
	}()
	ui.Render(g,m)
	// quits
	ui.On("q", "<C-c>", func(e ui.Event) {
		ui.StopLoop()
	})

	ui.Loop()
}
