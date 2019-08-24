package main

import (
	"flag"
	"fmt"
	"github.com/edison-moreland/go6502/c64Example/vic2"
	"github.com/edison-moreland/go6502/cpu"
	"log"
	"net/http"
	"path"
	"runtime"
	"time"
)

import _ "net/http/pprof"

var cpuprofile = flag.Bool("cpuprofile", false, "start profile server on localhost:6060")

type StopExecutionAddon struct {
	cpu.BaseAddon
	cycles int
}

func (se *StopExecutionAddon) AfterExecution() {
	if se.cycles >= 1000 {
		se.G6.StopEmulation()
		return
	}

	se.cycles++
	return
}

// Relative path to C64 ROM
const BASICRomPath = "./basic.901226-01.bin"
const KernalRomPath = "./kernal.901227-03.bin"

func main() {
	flag.Parse()
	if *cpuprofile {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	g6502 := cpu.Go6502{}

	// Register Addons
	g6502.RegisterAddons(
		&cpu.DebugAddon{SlowDown: 25 * time.Millisecond, Step: false, ShowZP: false},
		&vic2.Addon{},
		//&StopExecutionAddon{},
	)

	// Find location of this go file
	if _, filename, _, ok := runtime.Caller(0); ok {
		// rom paths are relative this file
		basicRomPath := path.Join(path.Dir(filename), BASICRomPath)
		kernalRomPath := path.Join(path.Dir(filename), KernalRomPath)

		// Load basic into memory
		if err := g6502.Mem.LoadMem(basicRomPath, 0xA000, 0xBFFF); err != nil {
			log.Panic(err)
		}

		// Load kernal into memory
		if err := g6502.Mem.LoadMem(kernalRomPath, 0xE000, 0xFFFF); err != nil {
			log.Panic(err)
		}
	} else {
		log.Panic("Could not find ROM path")
	}

	err := g6502.StartEmulation()
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
}
