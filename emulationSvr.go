// This go program has been tested with go 1.5.1
// This program establishes a Websocket connection with client code that is contained
// directly in this file. The client shows direction on the use of the browser interface
// The intention is to test that Websockets work and that a handler can be created in which
// arbitrary emulation logic can be placed.

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
    "fmt"
	_ "github.com/gorilla/websocket"
    _ "strings"
    _ "bufio"
    "io/ioutil"
    _ "os"
    _ "os/exec"
    "object"
    "interaction"
    _ "encoding/json"
    r "github.com/dancannon/gorethink"

)

// set up the RethinkDB sessions
var sessionOrg0 *r.Session
var sessionOrg1 *r.Session

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := interaction.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// wDInteraction forms a closure around the anonymous function that will be returned to http.HandleFunc in main()
// The closure enables additional parameters to be passed to wDInteraction, thereby enabling the implementation of
// arbitrary emulation logic in the confines of wDInteraction and via functions called from wDInteraction



func check(e error) {
    if e != nil {
        panic(e)
    }
}

// setQuota() internals can be replaced with a database query 

func setQuota(user string) float64 {

	switch user { 
		case "user0":	return 0050.000
		case "user1": 	return 0100.000
		case "user2": 	return 0300.000
		case "user3": 	return 0150.000
		default: 		return 50000.000
    }

}

func setupWSListener (hdlr http.HandlerFunc, org string, session *r.Session, wD []*object.WaterDevice, cwD []*object.ComplexWaterDevice, u []*object.User, artik10host *bool) {
	http.HandleFunc("/"+org, interaction.InitiatorProviderInteraction(hdlr, org, session, wD[:], cwD[:], artik10host))
}


func init() {
	fmt.Println("Performing the init functions")

	// command line flags
	hostI 		:= flag.String	("hI", "192.168.1.119", "host server")
	portI 		:= flag.Int		("pI", 28016, "port to serve on")

	flag.Parse()

    var err1,err2 error
    sessionOrg0, err1 = r.Connect(r.ConnectOpts{
        Address:  fmt.Sprintf("%s:%d", *hostI, *portI),
        Database: "org0",
    })
    if err1 != nil {
        fmt.Println(err1)
        return
    }
	sessionOrg1, err2 = r.Connect(r.ConnectOpts{
        Address:  fmt.Sprintf("%s:%d", *hostI, *portI),
        Database: "org1",
    })
    if err2 != nil {
        fmt.Println(err2)
        return
    }

}
// Main begins here

func main() {

	// command line flags
	host 		:= flag.String	("h", "192.168.1.119", "host server")
	port 		:= flag.Int		("p", 8080, "port to serve on")

	// Bug to be fixed: overriding artik10host to be fasle on the command line
	// does not work. It is possible to set '-artik 1' to override the default 
	// of false below however
	var artik10host *bool
	artik10host	= flag.Bool	("artik", false, "hosting on ARTIK10")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)

//	interaction.FetchAllRecordsInitiators(sessionOrg0)
//	interaction.FetchAllRecordsProviders(sessionOrg0)

	interaction.FetchAllRecordsInitiators(sessionOrg1)
	interaction.FetchAllRecordsProviders(sessionOrg1)


	fmt.Println(*artik10host)

	if (*artik10host) {
		interaction.PrintStr("setting up for ARTIK10")
		gpioNumShowerString := "9"
		gpioNumShower := []byte("9")
		gpioDirShower := []byte("out")
		gpioValShower := []byte("0")
	//	err0 := ioutil.WriteFile("/sys/class/gpio/export", 							gpioNumShower, 0644)
	//	check (err0)
		defer   ioutil.WriteFile("/sys/class/gpio/unexport", 							gpioNumShower, 0644)
		err1 := ioutil.WriteFile("/sys/class/gpio/gpio"+gpioNumShowerString+"/direction", gpioDirShower, 0644)
		check(err1)
		err2 := ioutil.WriteFile("/sys/class/gpio/gpio"+gpioNumShowerString+"/value", gpioValShower, 0644)
	    check(err2)
	}

// 	flag.Parse()
//	log.SetFlags(0)
//	http.HandleFunc("/", home)

// create the users
//	var users [4]object.User
	user0 		:= object.User 					{"Dad", 	0, true, 	0050.000, setQuota("user0")}
	user1 		:= object.User 					{"Mom", 	1, true, 	0100.000, setQuota("user1")}
	user2 		:= object.User 					{"Sister", 	2, false, 	0300.000, setQuota("user2")}
	user3 		:= object.User 					{"Brother", 3, false, 	0200.000, setQuota("user3")}

// create the devices and arrays of them
// LogEntry is included with each device, but is not initialized as initialization occurs implicitly
// Notional 'Starter' Rooms (to be replaced with DB entries later):
// 1. Kitchen
// 2. Bathroom 1 (1/2 bathroom,  	first floor)
// 3. Bathroom 2 (full bathroom, 	second floor)
// 4. Bathroom 3 (Master bathroom, 	second floor)
// 5. Garage
// All devices are set to non=permissive and (as a result also) off when created, so these
// WaterDevice.Permissive and WaterDevice.Open are not initialized
// The declarations of WaterDevices below will come from database queries in future revisions	
	shower 		:= object.WaterDevice			{Name: "shower", Index: "0",	Room: 4, Usage: 0050.000}

	sink 		:= object.WaterDevice			{Name: "sink", 	 Index: "1",	Room: 2, Usage: 0050.000}

	dishwasher	:= object.ComplexWaterDevice	{Civic: true, Collaborative: true}	// Note: Would like to define each 'ComplexWaterDevice'
	dishwasher.BaseWaterDevice.Name 	= "dishwasher"								// as a one-liner as for the simpler 'WaterDevice'
	dishwasher.BaseWaterDevice.Index 	= "2"
	dishwasher.BaseWaterDevice.Room 	= 1											// This syntax works
	dishwasher.BaseWaterDevice.On 		= false										// in the meantime
	
	sprinkler	:= object.ComplexWaterDevice	{Civic: true, Collaborative: false}	// Note: Would like to define each 'ComplexWaterDevice'
	sprinkler.BaseWaterDevice.Name 		= "sprinkler"								// as a one-liner as for the simpler 'WaterDevice'
	sprinkler.BaseWaterDevice.Index 	= "3"
	sprinkler.BaseWaterDevice.Room 		= 1											// This syntax works
	sprinkler.BaseWaterDevice.On 		= false										// in the meantime	

	wDorg0 		:= [...]*object.WaterDevice{&shower,&sink}
	wDorg1		:= [...]*object.WaterDevice{&shower,&sink}
	cwDorg0		:= [...]*object.ComplexWaterDevice{&dishwasher,&sprinkler}
	cwDorg1		:= [...]*object.ComplexWaterDevice{&sprinkler}
	uorg0 		:= [...]*object.User{&user0, &user2}
	uorg1 		:= [...]*object.User{&user1, &user3}

// create the organizations
// 
	orgs 		:= [...]string{"org0", "org1"}

// create the websocket handlers

//    for _, org := range orgs {
//    	setupWSListener(echo, org, "wD"+org+"[:]", "cwD"+org[:], "u"+org[:])
//	}

	setupWSListener(echo, orgs[0], sessionOrg0, wDorg0[:], cwDorg0[:], uorg0[:], artik10host)
	setupWSListener(echo, orgs[1], sessionOrg1, wDorg1[:], cwDorg1[:], uorg1[:], artik10host)

//    http.HandleFunc("/org1", wDInteraction(echo, wD1[:], cwD0[:], u1[:]))    
	log.Fatal(http.ListenAndServe(addr, nil))
}

var homeTemplate = template.Must(template.ParseFiles("FEfiles/systemViewMobileApp.html"))