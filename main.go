package main

import (
	"os"
	"log"
	"fmt"
	"time"
	"flag"
	"encoding/json"
	"path/filepath"
	"github.com/avecost/netmon-c/ip"
	"github.com/kardianos/service"
	"github.com/matishsiao/goInfo"
	"golang.org/x/net/websocket"
)

type program struct {}

type Configuration struct {
	Outlet string
	Terminal_account string
	Interval int64
}

var logger service.Logger
//var terminal terminal
var config Configuration
// var IP address
var privIpAddr string
var pubIpAddr string

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()

	return nil
}

func (p *program) run() {

	gi := goInfo.GetInfo()
	// build pc os info
	pc := fmt.Sprintf("OS: %s Build: %s Platform: %s CPUs: %d", gi.OS, gi.Core, gi.Platform, int(gi.CPUs))

	LOOP:
	for {
		origin := "http://localhost:9000/"
		url := "ws://localhost:9000/ws"
		//origin := "http://whisky.24bet7.com:9000/"
		//url := "ws://whisky.24bet7.com:9000/ws"
		ws, err := websocket.Dial(url, "", origin)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		for {
			// get client Private IP
			privIpAddr, _ = ip.GetClientIP()
			// get client Public IP
			pubIpAddr = string(ip.GetPublicIP())

			mapD := map[string]string{"event": "KEEP-ALIVE",
				"outlet": config.Outlet,
				"acct": config.Terminal_account,
				"privateip": privIpAddr,
				"publicip": pubIpAddr,
				"os": pc}
			mapB, _ := json.Marshal(mapD)
			if _, err := ws.Write(mapB); err != nil {
				time.Sleep(5 * time.Second)
				continue LOOP
			}

			time.Sleep(time.Duration(config.Interval) * time.Second)
		}
	}

}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {

	flag.Parse()

	svcConfig := &service.Config{
		Name:        "netmon-c",
		DisplayName: "Netmon-Client",
		Description: "Network Monitoring Client service.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(flag.Args()) > 0 {
		err = service.Control(s, flag.Args()[0])
		if err != nil {
			logger.Error(err)
		}
		return
	}

	// get the directory where the exe was run
	curDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(curDir)
	confFile := fmt.Sprintf("%s/config.json", curDir)

	// read the config file
	file, err := os.Open(confFile)
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}