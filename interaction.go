package interaction

import (
	"fmt"
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats"
    "io/ioutil"
    "object"
    "encoding/json"
    r "github.com/dancannon/gorethink"
    "time"
    _ "strconv"
    "strings"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 2 * time.Second

	// Time allowed to read the next pong message from the peer.
	//pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	//pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	//maxMessageSize = 512
)

var Upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type Initiator struct {
	Id  					string  `gorethink:"id,omitempty"`
	ID  					string  `gorethink:"IDKey"`
    Tag 					int   	`gorethink:"tagKey"`
    Admin					bool	`gorethink:"adminKey"`
    LocationKey 			string  `gorethink:"locationKey"`    
    WDUsing 				string  `gorethink:"wDUsingKey"`
    CollaborationPartner 	string  `gorethink:"collaborationPartnerKey"`
    RewardDBPtr				string	`gorethink:"rewardDBPtrKey"`
}

func (i *Initiator) updateParam(session *r.Session, param string, newVal string) {

	resultB, err := r.Table("initiators").Get(i.Id).Run(session)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("*** Prior to update: ***")
	printObj(resultB)
	fmt.Println("\n\n\n\n")

	switch param {

		case "LocationKey":
			i.LocationKey = newVal
			break
		case "WDUsing":
			i.WDUsing = newVal
			break
		case "CollaborationPartner":
			i.CollaborationPartner = newVal
			break
		case "RewardDBPtr":
			i.RewardDBPtr = newVal
			break
		default:
			fmt.Printf("Initiator.updateParam does not support updating of %s to value %s\n", param, newVal)

	}

	resultA, err := r.Table("initiators").Get(i.Id).Update(i).RunWrite(session)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("*** Update result: ***")
	printObj(resultA)
	fmt.Println("\n\n\n\n")

}

type Provider struct {
	Id  					string 	`gorethink:"id,omitempty"`
	IDKey  					string 	`gorethink:"IDKey"`
	TagKey 					int   	`gorethink:"tagKey"`	
	LocationKey				string 	`gorethink:"locationKey"`
	StateKey  				string 	`gorethink:"stateKey"`
	StateContextKey  		string 	`gorethink:"stateContextKey"`
	Initiator1Key	  		string 	`gorethink:"initiator1"`
	Initiator2Key	  		string 	`gorethink:"initiator2"`
	CivicFeatureKey	  		bool 	`gorethink:"civicFeatureKey"`
	CollaborationFeatureKey	bool 	`gorethink:"collaborationFeatureKey"`	
	ReclamationFeatureKey	bool 	`gorethink:"reclamationFeatureKey"`
}

func (p *Provider) queryUpdateParam(op string, i *Initiator, session *r.Session, param string, newStrVal string, newBoolVal bool) {

		canChg := false		// safe initial state which can be overridden
		if i.Admin == true || param == "StateKey" || param == "StateContextKey" || param == "Initiator1Key" || param == "Initiator2Key" {canChg = true}	

		resultB, err := r.Table("providers").Get(p.Id).Run(session)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("*** Prior to update: ***")
		printObj(resultB)
		fmt.Println("\n\n\n\n")

	// fmt.Printf("%s an Admin?: %t: affecting ability to change %s\n", i.ID, isAdmin, param)

	if op == "U" {

		// add type checking to this switch statement
		// manually manage until then
		switch param {

			case "LocationKey":
				if canChg == true {
					p.LocationKey = newStrVal
				} else {
					fmt.Printf("%s value must be changed by an Admin to a proper type, or the param must be a parameter %s (%s) is allowed to update\n", param, i.Id, i.ID)
				}
				break
			case "StateKey":
				p.StateKey = newStrVal
				break
			case "StateContextKey":
				p.StateContextKey = newStrVal
				break
			case "Initiator1Key":
				p.Initiator1Key = newStrVal
				break
			case "Initiator2Key":
				p.Initiator2Key = newStrVal
				break
			case "CivicFeatureKey":
				if canChg == true {
					p.CivicFeatureKey = newBoolVal
				} else {
					fmt.Printf("%s value must be changed by an Admin to a proper type\n", param)
				}			
				break
			case "CollaborationFeatureKey":
				if canChg == true {
					p.CollaborationFeatureKey = newBoolVal
				} else {
					fmt.Printf("%s value must be changed by an Admin to a proper type\n", param)
				}
				break
			case "ReclamationFeatureKey":
				if canChg == true {
					p.ReclamationFeatureKey = newBoolVal
				} else {
					fmt.Printf("%s value must be changed by an Admin to a proper type\n", param)
				}
				break
			default:
				fmt.Printf("Provider.updateParam does not support updating of %s\n", param)

		}	

		resultA, err := r.Table("providers").Get(p.Id).Update(p).RunWrite(session)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("*** Update result: ***")
		printObj(resultA)
		fmt.Println("\n\n\n\n")

	}

}

type Quota struct {
	Id        string    `gorethink:"id, omitempty"`
	Name      string    `gorethink:"name"`
	OwnerName string    `gorethink:"ownername"` // Name of the user/org who owns the quota
	WDID      string    `gorethink:"wdid"`      // optional parameter if quota is also specific to a certain provider
	Limit     float64   `gorethink:"limit"`
	StartDate time.Time `gorethink:"startdate"` // The first day of the month/week/day that this quota starts
	Duration  string    `gorethink:"duration"` // default: month
}

type LogEntry struct {
	Id           string    `gorethink:"id,omitempty"`
	WDID         string    `gorethink:"wdid"`     // The water device being used
	UserName     string    `gorethink:"username"` // The name of the user/org using the water
	Timestamp    time.Time `gorethink:"timestamp"`
	FlowRate     float64   `gorethink:"flowrate"`
	DiffConsumed float64   `gorethink:"diffconsumed"` //The water consumed since the last log was taken (water consumed during this use)
}

// ProvFSM state function datatype and
// associated struct for its input
// conditions
type InitFSM_StateFn func(*InitFSMIn) InitFSM_StateFn

// input struct for the InitFSM
// InitFSM only populates the mobile application with
// information about the initiators
type InitFSMIn struct {
	InitStr			string			// the string of attributes of this initiator
	RDBsession		*r.Session 		// session for database interactions	
	Init 			Initiator 		// a nested Initiator struct
	ProvS 			[]Provider		// the slice of providers to enable the InitFSM to modify
	RxFEmsg 		chan string		// message received from the FE
	TxFEmsg 		chan string		// message sent to the FE
}

// The single state for the InitFSM
func InitFSM_StateInit(fsmIn *InitFSMIn) InitFSM_StateFn {
	
	RxStr := <- fsmIn.RxFEmsg
	fmt.Printf("In InitFSM_StateInit, The channel has value: %s\n", RxStr)
	RxStrS := strings.Split(RxStr,".")
	if RxStrS[0] != "iMsgRx" {
		fmt.Printf("Not processing %s in InitFSM\n", RxStr)
		return InitFSM_StateInit
	} else if RxStrS[0] == "iMsgRx" {
		fmt.Printf("Processing %s in InitFSM\n", RxStr)
		if fsmIn.Init.Admin {fmt.Printf("%s is an %s\n", fsmIn.Init.ID, "Admin of the water system")}
		if RxStrS[1] == "INIT" && RxStrS[2] == "INITS" {
			fmt.Printf("Sending InitStr of: %s to the mobile app\n", fsmIn.InitStr)
			TxStr := "iMsgTx.INIT.INITS."+fsmIn.InitStr
			for i:=0;i<4;i++ {
				goR0("Sending copy # (InitFSM):", 5)
				fmt.Printf("%d \n",i)
				fsmIn.TxFEmsg <- TxStr
			}

			/*
			for j, p := range fsmIn.ProvS {
				//fmt.Printf("Before changing, StateKey for %s (%s) was %s\n", fsmIn.ProvS[j].IDKey, fsmIn.ProvS[j].Id, fsmIn.ProvS[j].StateKey)
				p_m, _ := json.Marshal(p)
				p_s := string(p_m)
				fmt.Printf("P marshalled before is %s\n", p_s)
				// func (p *Provider) queryUpdateParam(op string, i *Initiator, session *r.Session, param string, newStrVal string, newBoolVal bool)

				//fsmIn.ProvS[j].queryUpdateParam("U", &fsmIn.Init, fsmIn.RDBsession, "StateKey", "permissive", false)
				//fsmIn.ProvS[j].queryUpdateParam("U", &fsmIn.Init, fsmIn.RDBsession, "StateContextKey", "permissive", false)
				//fsmIn.ProvS[j].queryUpdateParam("U", &fsmIn.Init, fsmIn.RDBsession, "Initiator1Key", "", false)
				//fsmIn.ProvS[j].queryUpdateParam("U", &fsmIn.Init, fsmIn.RDBsession, "Initiator2Key", "", false)
				//fmt.Printf("After changing, StateKey for %s (%s) is %s\n", fsmIn.ProvS[j].IDKey, fsmIn.ProvS[j].Id, fsmIn.ProvS[j].StateKey)
				p_m, _ = json.Marshal(p)
				p_s = string(p_m)
				fmt.Printf("P marshalled after is %s\n", p_s)
			}

			*/

			return InitFSM_StateInit			
		} else {
			fmt.Printf("NOT sending InitStr to the mobile app\n")
			fmt.Printf("%s is staying in InitFSM_StateInit\n", fsmIn.Init.ID)
			return InitFSM_StateInit
		}
	} else if RxStrS[0] == "qMsgRx" {
		fmt.Printf("Processing Quota Functionality %s in InitFSM\n", RxStr)
		return InitFSM_StateInit // move to another state in the final implementation
	} else if RxStrS[0] == "rMsgRx" {
		fmt.Printf("Processing Reward Functionality %s in InitFSM\n", RxStr)
		return InitFSM_StateInit // move to another state in the final implementation
	} else {
		fmt.Printf("An unexpected input was received; %s is staying in InitFSM_StateInit\n", fsmIn.Init.ID)
		return InitFSM_StateInit
	}

}

// the run() method for the InitFSM
func Run_InitFSM(fsmIn *InitFSMIn) {
	fmt.Printf("The instance of the FSM for Initiator %s has been created\n", fsmIn.Init.Id)
	for state := InitFSM_StateInit; state != nil; {
		state = state(fsmIn)
	}
}

// ProvFSM state function datatype and
// associated struct for its input
// conditions
type ProvFSM_StateFn func(*ProvFSMIn) ProvFSM_StateFn

// input struct for the ProvFSM
// ProvFSM sets up the logic for each of the providers
// and populates the mobile application with information about all
// of the providers present
type ProvFSMIn struct {
	ProvStr 		string			// the string of attributes of this provider
	RDBsession		*r.Session 		// session for database interactions	
	Prov 			Provider 		// a nested Initiator struct
	InitS 			[]Initiator		// the slice of initiators to enable the InitFSM to modify
	RxFEmsg 		chan string		// message received from the FE
	TxFEmsg 		chan string		// message sent to the FE
	RxNATSmsg		chan string		// message received from NATS
	TxNATSmsg		chan string		// message sent to NATS
	ARTIK10H 		bool			// whether GPIO information is populated is based on
									// whether hosting is on am IoT module, or fully emulated
	GPIONumStr		string			// the GPIO allocated to this provider
	GPIODir			string			// the GPIO direction for this provider's ARTIK10 functionality
	GPIOInitVal		string			// whether the GPIO is initially ON or OFF (determines actuation
									// based on all connectivity and logical functionality of both
									// electrical and mechanical systems in combination
}

func (pFSM *ProvFSMIn) initGPIO() {
 
 	fmt.Printf("Calling initGPIO with %s %s %s\n", pFSM.GPIONumStr, pFSM.GPIODir, pFSM.GPIOInitVal)
	err := ioutil.WriteFile("/sys/class/gpio/export", []byte(pFSM.GPIONumStr), 0644)
	checkANDContinue(err)
	err  = ioutil.WriteFile("/sys/class/gpio/gpio"+pFSM.GPIONumStr+"/direction", []byte(pFSM.GPIODir), 0644)
	checkANDContinue(err)
	err  = ioutil.WriteFile("/sys/class/gpio/gpio"+pFSM.GPIONumStr+"/value", []byte(pFSM.GPIOInitVal), 0644)
	checkANDContinue(err)

}

func (pFSM *ProvFSMIn) chgGPIO(GPIOState string) {
 
	err := ioutil.WriteFile("/sys/class/gpio/gpio"+pFSM.GPIONumStr+"/value", []byte(GPIOState), 0644)
	checkANDContinue(err)

}

func (pFSM *ProvFSMIn) finishGPIO() {

	err := ioutil.WriteFile("/sys/class/gpio/unexport", []byte(pFSM.GPIONumStr), 0644)
	checkANDContinue(err)

}

// the states for ProvFSM

// the first state of the ProvFSM, StateInit_ProvFSM sets up the allowed functionality, 
// based on fields read into each ProvFSMIn as it is created
// this ensures that the proper device functionality is created at initialization
// if the functionality is changed, the composition of the FSM behavior
// would be different
// configuration changes require returning to the initial state of the ProvFSM

func ProvFSM_StateInit(fsmIn *ProvFSMIn) ProvFSM_StateFn {

	/*
	RxNATSStr := <- fsmIn.RxNATSmsg
	RxNATSStrS := strings.Split(RxNATSStr,".")
	fmt.Printf("%s received NATS message of %s\n", fsmIn.Prov.Id, RxNATSStr)

	if RxNATSStrS[0] == fsmIn.Prov.Id {

		f, _ := strconv.ParseFloat(RxNATSStr, 32)
		fmt.Printf("Incoming measurement in %s (%s) is: %f\n", fsmIn.Prov.IDKey, fsmIn.Prov.Id, RxNATSStr)
		if f > 10.0 {
			//fmt.Printf("%s is staying in ProvFSM_StateInit\n", fsmIn.Prov.Id)
			fmt.Printf("%s NATS message crossed the threshold!\n", fsmIn.Prov.Id)
			//return StateInitInits
		}
	} else if RxNATSStrS[0] != fsmIn.Prov.Id {
		fmt.Printf("Not processing NATS message %s in ProvFSM %s and continuing\n", RxNATSStr)
	}
	*/
	
	RxStr := <- fsmIn.RxFEmsg
	fmt.Printf("In ProvFSM_StateInit, The channel has value: %s\n", RxStr)
	RxStrS := strings.Split(RxStr,".")
	if RxStrS[0] != "pMsgRx" {
		fmt.Printf("Not processing %s in ProvFSM\n", RxStr)
		return ProvFSM_StateInit
	} else if RxStrS[0] == "pMsgRx" && RxStrS[1] == "INIT" {
		fmt.Printf("Processing %s in ProvFSM\n", RxStr)

		if fsmIn.Prov.CivicFeatureKey 			{fmt.Printf("%s has the %s\n", fsmIn.Prov.Id, "Civic Feature")}
		if fsmIn.Prov.CollaborationFeatureKey 	{fmt.Printf("%s has the %s\n", fsmIn.Prov.Id, "Collaboration Feature")}
		if fsmIn.Prov.ReclamationFeatureKey 	{fmt.Printf("%s has the %s\n", fsmIn.Prov.Id, "Reclamation Feature")}

		if RxStrS[1] == "INIT" && RxStrS[2] == "PROVS" {
			fmt.Printf("Sending ProvStr of: %s to the mobile app\n", fsmIn.ProvStr)
			TxStr := "pMsgTx.INIT.PROVS."+fsmIn.ProvStr
			for i:=0;i<9;i++ {
				goR0("Sending copy # (ProvFSM):", 10)
				fmt.Printf("%d \n",i)
				fsmIn.TxFEmsg <- TxStr
			}

			fmt.Printf("ARTIK10H and GPIONumStr are %t and %s", fsmIn.ARTIK10H, fsmIn.GPIONumStr)
			if fsmIn.ARTIK10H && fsmIn.GPIONumStr != "" {
				fmt.Printf("\n\n\nProvider %s of type %s is physically connected, so setup operations will be performed\n\n\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				fsmIn.initGPIO()
			} else {
				fmt.Printf("\n\n\nProvider %s of type %s will be emulated\n\n\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
			}

			//return ProvFSM_StateRequest
			return ProvFSM_StateRequest

		} else {
			fmt.Printf("NOT sending ProvStr to the mobile app\n")
			fmt.Printf("%s is staying in ProvFSM_StateInit\n", fsmIn.Prov.IDKey)
			return ProvFSM_StateRequest
		}

	} else if RxStrS[0] == "pMsgRx" && RxStrS[1] == fsmIn.Prov.Id {
		return ProvFSM_StateRequest		
	} else {
		fmt.Printf("An unexpected input was received; %s is returning to ProvFSM_StateRequest\n", fsmIn.Prov.IDKey)
		return ProvFSM_StateInit
	}

}

func ProvFSM_StateRequest(fsmIn *ProvFSMIn) ProvFSM_StateFn {
	
	//mStr := fsmIn.Prov.Id+"."+"REQUEST"

	RxStr := <- fsmIn.RxFEmsg
	fmt.Printf("In ProvFSM_StateRequest, The channel has value: %s\n", RxStr)
	RxStrS := strings.Split(RxStr,".")

	// mobile app commands from the nonCollab WDs
    // <button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"GETUSAGE") >Get Usage</button>                    
    // <button class="button button-energized button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"REQUEST") >Request</button>
    // <button class="button button-positive button-clear" ng-show=dev["Selected"]  ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"ON")      >Use</button>
    // <button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"OFF")     >Finished</button>

	// mobile app commands from the Collab WDs
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"GETUSAGE") >Get Usage</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"REQUEST") >Request</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"ON")      >Use</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"OFF")     >Finished</button>

	if RxStrS[0] != "pMsgRx" {
		fmt.Printf("Not processing %s in ProvFSM\n", RxStr)
		return ProvFSM_StateRequest
	} else if RxStrS[0] == "pMsgRx" {
		fmt.Printf("Processing %s in ProvFSM\n", RxStr)

		switch RxStrS[1] {

			case fsmIn.Prov.Id:
				fmt.Printf("Checking whether %s is the intended recipient\n", fsmIn.Prov.Id)
				fmt.Printf("In ProvFSM_StateRequest: This message IS intended for provider %s, of type %s\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				fmt.Printf("Taking appropriate actions for the users %s and %s in ProvFSM for %s (%s)\n", RxStrS[2], RxStrS[3], fsmIn.Prov.Id, fsmIn.Prov.IDKey)

				switch RxStrS[4] {

					case "GETUSAGE":

						fmt.Printf("GETUSAGE command received. Getting the usage of %s by user %s\n", RxStrS[1], RxStrS[2])
						return ProvFSM_StateRequest
						break

					case "REQUEST":
						fmt.Printf("REQUEST command received by %s, checking policy and state\n", fsmIn.Prov.Id)
						if (fsmIn.Prov.StateKey == "permissive") { // add the check of the quotas based on the device WDID and usage against that
							fmt.Printf("%s is transitioning from ProvFSM_StateRequest to ProvFSM_StateOn after marking the provider as being in use by initiator1: %s, as %s is not over quota and the device is permissive\n", fsmIn.Prov.IDKey, RxStrS[2], RxStrS[2])
							p_m, _ := json.Marshal(fsmIn.Prov)
							p_s := string(p_m)
							fmt.Printf("P marshalled before is %s\n", p_s)
							fmt.Printf("length of InitS is %d:\n", len(fsmIn.InitS))
							for _, i := range fsmIn.InitS {
								fmt.Printf("Id for this initiator is %s and REQUESTOR is %s\n", i.Id, RxStrS[2])
								if i.Id == RxStrS[2] {
									fmt.Printf("Setting the particular initiator to pass to queryUpdateParam. Command has a match with request from %s matching %s", RxStrS[2], i.Id)
									// func (p *Provider) queryUpdateParam(op string, i *Initiator, session *r.Session, param string, newStrVal string, newBoolVal bool)
									//fsmIn.Prov.queryUpdateParam("U", &i, fsmIn.RDBsession, "Initiator1Key", RxStrS[2], false)
									p_m, _ = json.Marshal(fsmIn.Prov)
									p_s = string(p_m)
									fmt.Printf("P marshalled after is %s\n", p_s)
								}
							}

							return ProvFSM_StateOn
						} else {
							fmt.Printf("%s is staying in ProvFSM_StateRequest, based on StateKey of: %s. Context is: %s\n", fsmIn.Prov.IDKey, fsmIn.Prov.StateKey, fsmIn.Prov.StateContextKey)
							return ProvFSM_StateRequest
						}
						break

					default:
						fmt.Printf("No valid command received by %s, returning to ProvFSM_StateRequest\n", fsmIn.Prov.Id)
						return ProvFSM_StateRequest
				}

			default:
		
				fmt.Printf("This message is not intended for provider %s, of type %s\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				return ProvFSM_StateInit

		}

	} else {
		fmt.Printf("An unexpected input was received; %s is staying in ProvFSM_StateRequest\n", fsmIn.Prov.IDKey)
		return ProvFSM_StateInit
	}

	return ProvFSM_StateRequest

}

func ProvFSM_StateOn(fsmIn *ProvFSMIn) ProvFSM_StateFn {
	
	RxStr := <- fsmIn.RxFEmsg
	fmt.Printf("In ProvFSM_StateOn, The channel has value: %s\n", RxStr)
	RxStrS := strings.Split(RxStr,".")
	
	// mobile app commands from the Collab WDs
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"GETUSAGE") >Get Usage</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"REQUEST") >Request</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"ON")      >Use</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"OFF")     >Finished</button>

	if RxStrS[0] != "pMsgRx" {
		fmt.Printf("Not processing %s in ProvFSM\n", RxStr)
		return ProvFSM_StateRequest
	} else if RxStrS[0] == "pMsgRx" {
		fmt.Printf("Processing %s in ProvFSM\n", RxStr)

		switch RxStrS[1] {

			case fsmIn.Prov.Id:
				fmt.Printf("Checking whether %s is the intended recipient\n", fsmIn.Prov.Id)
				fmt.Printf("In ProvFSM_StateOn: This message IS intended for provider %s, of type %s\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				fmt.Printf("Taking appropriate actions for the users %s and %s\n", RxStrS[2], RxStrS[3])

				switch RxStrS[4] {

					case "ON":

						fmt.Printf("ON command received, meaning the device %s was permissive for user %s\n", RxStrS[1], RxStrS[2])
						fsmIn.chgGPIO("0")
						return ProvFSM_StateOff
						break

					default:
						fmt.Printf("No valid command received by %s, returning to ProvFSM_StateRequest\n", fsmIn.Prov.Id)
						return ProvFSM_StateRequest
				}

			default:
		
				fmt.Printf("This message is not intended for provider %s, of type %s\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				return ProvFSM_StateInit

		}

	} else {
		fmt.Printf("An unexpected input was received; %s is staying in ProvFSM_StateRequest\n", fsmIn.Prov.IDKey)
		return ProvFSM_StateInit
	}

	return ProvFSM_StateRequest	
	/*
	if RxStr == mStr {
		fmt.Printf("ON command received by %s, checking policy and state\n", fsmIn.Prov.Id)
		fmt.Printf("%s is transitioning from ProvFSM_StateOn to ProvFSM_StateOff\n", fsmIn.Prov.IDKey)
		return ProvFSM_StateOff
	} else if RxStr != mStr {
		fmt.Printf("%s is staying in ProvFSM_StateOn\n", fsmIn.Prov.IDKey)
		return ProvFSM_StateOn
	} else {
		fmt.Printf("In ProvFSM_StateOn, taking the default case for %s", fsmIn.Prov.IDKey)
		return ProvFSM_StateOn
	}
	*/
}

func ProvFSM_StateOff(fsmIn *ProvFSMIn) ProvFSM_StateFn {
	
	RxStr := <- fsmIn.RxFEmsg
	fmt.Printf("In ProvFSM_StateOn, The channel has value: %s\n", RxStr)
	RxStrS := strings.Split(RxStr,".")
	
	// mobile app commands from the Collab WDs
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+"NULL"+"."+"GETUSAGE") >Get Usage</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"REQUEST") >Request</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"ON")      >Use</button>
    //<button class="button button-positive button-clear" ng-show=dev["Selected"] ng-click=devIntCtrl_0.sendws("pMsgRx"+"."+dev["Id"]+"."+devIntCtrl_0.OWNERID+"."+devIntCtrl_0.User2+"."+"OFF")     >Finished</button>

	if RxStrS[0] != "pMsgRx" {
		fmt.Printf("Not processing %s in ProvFSM\n", RxStr)
		return ProvFSM_StateRequest
	} else if RxStrS[0] == "pMsgRx" {
		fmt.Printf("Processing %s in ProvFSM_StateOff of ProvFSM\n", RxStr)

		switch RxStrS[1] {

			case fsmIn.Prov.Id:
				fmt.Printf("Checking whether %s is the intended recipient\n", fsmIn.Prov.Id)
				fmt.Printf("In ProvFSM_StateOff: This message IS intended for provider %s, of type %s\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				fmt.Printf("Taking appropriate actions for the users %s and %s\n", RxStrS[2], RxStrS[3])

				switch RxStrS[4] {

					case "OFF":

						fmt.Printf("OFF command received, meaning the device %s will now shut off %s\n", RxStrS[1], RxStrS[2])
						fsmIn.chgGPIO("1")
						return ProvFSM_StateRequest
						break

					default:
						fmt.Printf("No valid command received by %s, returning to ProvFSM_StateRequest\n", fsmIn.Prov.Id)
						return ProvFSM_StateRequest
				}

			default:
		
				fmt.Printf("This message is not intended for provider %s, of type %s\n", fsmIn.Prov.Id, fsmIn.Prov.IDKey)
				return ProvFSM_StateInit

		}

	} else {
		fmt.Printf("An unexpected input was received; %s is staying in ProvFSM_StateRequest\n", fsmIn.Prov.IDKey)
		return ProvFSM_StateInit
	}

	return ProvFSM_StateRequest

}
//func refState(fsmIn *ProvFSMIn) wDStateFn {
	//f, _ := strconv.ParseFloat(<-fsmIn.NATSmsg, 32)
	//fsmInt.Printf("Incoming measurement is: %f\n", f)
//	if f > 0.015000 {
//		fmt.Printf("%s is transitioning from StateNext to StateInit\n", fsmIn.Prov.Id)
//		return StateInitInits
//	} else {
		//fsmInt.Printf("%s is staying in StateNext\n", fsmIn.Prov.Id)
//		return StateNext
//	}
//}

// the run() method for the ProvFSM

func Run_ProvFSM(fsmIn *ProvFSMIn) {
	fmt.Printf("The instance of the FSM for Provider %s has been created\n", fsmIn.Prov.Id)
	for state := ProvFSM_StateInit; state != nil; {
		state = state(fsmIn)
	}
}

func wsReader(ws *websocket.Conn, i int, RxchI []chan string, RxchP []chan string) {
	//defer ws.Close()
	//ws.SetReadLimit(512)
	//ws.SetReadDeadline(time.Now().Add(pongWait))
	//ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := ws.ReadMessage()
		fmt.Printf("Read in 'wsReader' func %s,%d\n", string(msg), i)	
		if err != nil {
			break
		}

		//goR0("Delaying initially sending to channels (before Prov and Init)", 5)

		for j_p_1 := 0; j_p_1 <= len(RxchP)-1; j_p_1++ {
			//fmt.Printf("Setting up readers for the ProvFSMs\n")
			//goR0("Delaying sending to Prov channel\n", 5)
			RxchP[j_p_1] <- string(msg)
		}
		//goR0("Delaying sending to channels (between Prov and Init)", 5)
		for j_i_1 := 0; j_i_1 <= len(RxchI)-1; j_i_1++ {
			//fmt.Printf("Setting up readers for the InitFSMs\n")
			//goR0("Delaying sending to Init channel\n", 5)
			RxchI[j_i_1] <- string(msg)
		}
	}
}

func wsWriter(ws *websocket.Conn, Txch chan string) {

	TxStr := <- Txch

	fmt.Printf("wsWriter called\n")
	defer func() {
		fmt.Printf("Error: Closing the WS\n")
		ws.Close()
	}()

	for {
		//fmt.Printf("Length of the Tx channel is %d\n", len(Txch))
		err := ws.WriteMessage(websocket.TextMessage, []byte(TxStr))
		//fmt.Printf("Length of the Tx channel is %i\n", len(Txch))
		if err != nil {
			fmt.Println("write:", err)
			break
		}		
	}	
}

func write(ws *websocket.Conn, mt int, payload []byte) error {
	ws.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.WriteMessage(mt, payload)
}

func wsWriter1(ws *websocket.Conn, TxFEmsgI []chan string, TxFEmsgP []chan string) {

	fmt.Printf("wsWriter called\n")
	defer func() {
		fmt.Printf("Error: Closing the WS\n")
		ws.Close()
	}()

	for {
		select {
			case message, ok := <- TxFEmsgI[0]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgI[1]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgI[2]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgI[3]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgP[0]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgP[1]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgP[2]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgP[3]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgP[4]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}
			case message, ok := <- TxFEmsgP[5]:
				if !ok {
					write(ws, websocket.CloseMessage, []byte{})
					return
				}
				if err := write(ws, websocket.TextMessage, []byte(message)); err != nil {
					return
				}

		}	
	}	
}

func InitiatorProviderInteraction(fn http.HandlerFunc, org string, session *r.Session, artik10host *bool) http.HandlerFunc {

	// create channels for information from the mobile application and the NATS server
	//nm := make(chan string)

	// create a slice of InitFSMIn structs that will be assigned when initiator setup operations are executed
	InitFSMInS := make([]InitFSMIn,0)

	// create a slice of ProvFSMIn structs that will be assigned when provider setup operations are executed
	ProvFSMInS := make([]ProvFSMIn,0)


	// some feedback to the console
	//go goR0("Here we go 0: ", 1)
	//go goR1("Here we go 1: ")

	// auxilliary variable to control execution in the nested 
	// returned handler function, as needed
	var u uint = 0
	fmt.Printf("u 'outside' is: %d\n", u)

	// initialize the initiators and providers
	initiatorS := initInitiators(session)
	providersS := initProviders(session)

	// create communication channels


	// communication channels for the InitFSM instances
	RxFEmsgI 		:= make([]chan string, len(initiatorS))
	TxFEmsgI 		:= make([]chan string, len(initiatorS))
	
	for i_kk := range initiatorS {
		RxFEmsgI[i_kk]   = make(chan string)
		TxFEmsgI[i_kk]   = make(chan string, 256)
	}


	// communication channels for the ProvFSM instances
	RxFEmsgP 		:= make([]chan string, len(providersS))
	TxFEmsgP 		:= make([]chan string, len(providersS))
	RxNATSmsgP 		:= make([]chan string, len(providersS))
	TxNATSmsgP 		:= make([]chan string, len(providersS))
	
	for p_kk := range providersS {
		RxFEmsgP[p_kk]   = make(chan string)
		TxFEmsgP[p_kk]   = make(chan string, 256)
		RxNATSmsgP[p_kk] = make(chan string)		
		TxNATSmsgP[p_kk] = make(chan string)		
	}

	// NATS setup
	var server = "nats://192.168.1.100:4222"
	nc1, _ := nats.Connect(server)
	//nc1.Publish("WD1out", []byte("Hello Particle Photon!"))
	//8fc73061-3d40-4679-8450-1dee9ae64852 // sprinkler
	//3ca6094d-4d0c-422f-9bcf-75424294eda8 // shower
	//a26a4357-6631-4686-8926-8fa108c8f752 // sink

	//50c57d17-e0d5-4c07-b324-d682c8dd46b2 // washer
	//5386b83a-0545-4c07-b993-2a79abf86a58 // dishwasher
	//98ac10aa-b47f-4582-aa60-5153e09bbb10 // dryer
	
	nc1.Subscribe("8fc73061-3d40-4679-8450-1dee9ae64852.sprinkler.flowRate", func(m *nats.Msg) {
    	fmt.Printf("Received a message from the sprinkler: %s\n", string(m.Data))
    	// push the NATS data into the common channel
    	//nm <- string(m.Data)
		for j := range providersS {
			RxNATSmsgP[j] <- string(m.Data)
		}
	})

	nc1.Subscribe("3ca6094d-4d0c-422f-9bcf-75424294eda8.shower.flowRate", func(m *nats.Msg) {
    	fmt.Printf("Received a message from the shower: %s\n", string(m.Data))
	})

	nc1.Subscribe("a26a4357-6631-4686-8926-8fa108c8f752.sink.flowRate", func(m *nats.Msg) {
    	fmt.Printf("Received a message from the sink: %s\n", string(m.Data))
	})
	nc1.Subscribe("50c57d17-e0d5-4c07-b324-d682c8dd46b2.washer.flowRate", func(m *nats.Msg) {
    	fmt.Printf("Received a message from the washer: %s\n", string(m.Data))
	})
	nc1.Subscribe("5386b83a-0545-4c07-b993-2a79abf86a58.dishwasher.flowRates", func(m *nats.Msg) {
    	fmt.Printf("Received a message from the dishwasher: %s\n", string(m.Data))
 	})
	

	var p_i = 0
	var i_i = 0
	//var s_u = ""
	
	// perform setup operations for the providers

// snippet of the fields in the ProvFSMIn object
//	ARTIK10H 		bool			// whether GPIO information is populated is based on
									// whether hosting is on am IoT module, or fully emulated
//	GPIONumStr		string			// the GPIO allocated to this provider
//	GPIODir			string="out"	// the GPIO direction for this provider's ARTIK10 functionality
//	GPIOInitVal		string="0"		// whether the GPIO is initially ON or OFF (determines actuation

	if *artik10host {
		// do nothing ...
	}

	for _, p := range providersS {

		if p.Id != "" {

			p_m, _ := json.Marshal(p)
			p_s := string(p_m)
			fmt.Println(p_s)
			ProvFSMInS = append(ProvFSMInS, ProvFSMIn{ProvStr: p_s, RDBsession: session, Prov: providersS[p_i], InitS: initiatorS, RxFEmsg: RxFEmsgP[p_i], TxFEmsg: TxFEmsgP[p_i], RxNATSmsg: RxNATSmsgP[p_i], TxNATSmsg: TxNATSmsgP[p_i], ARTIK10H: *artik10host, GPIONumStr: "", GPIODir: "out", GPIOInitVal: "1"})
			fmt.Println("\n")

			// Mapping of GPIO to providers
			// GPIO 	Pin
			//    8		  2
			//    9		  3
			//   10		  4
			//  203		  5
			//  204		  6
			//   11		  7
			//   12		  8
			//   12+1	  9
			//   14		 10
			//   16		 11 
			//   21		 12 
			//   22		 12+1 

			
			if ProvFSMInS[p_i].ARTIK10H && ProvFSMInS[p_i].Prov.Id == "5386b83a-0545-4c07-b993-2a79abf86a58" {
				ProvFSMInS[p_i].GPIONumStr = "8"
				//nc1.Subscribe(ProvFSMInS[p_i].Prov.Id, func(m *nats.Msg) {
    			//	fmt.Printf("Received a message from the %s: %s\n", ProvFSMInS[p_i].Prov.IDKey, string(m.Data))
				//})
			} else if ProvFSMInS[p_i].ARTIK10H && ProvFSMInS[p_i].Prov.Id == "8fc73061-3d40-4679-8450-1dee9ae64852" {
				ProvFSMInS[p_i].GPIONumStr = "9"
				//nc1.Subscribe(ProvFSMInS[p_i].Prov.Id, func(m *nats.Msg) {
    			//	fmt.Printf("Received a message from the %s: %s\n", ProvFSMInS[p_i].Prov.IDKey, string(m.Data))
				//})
			} else if ProvFSMInS[p_i].ARTIK10H && ProvFSMInS[p_i].Prov.Id == "50c57d17-e0d5-4c07-b324-d682c8dd46b2" {
				ProvFSMInS[p_i].GPIONumStr = "11"
				//nc1.Subscribe(ProvFSMInS[p_i].Prov.Id, func(m *nats.Msg) {
    			//	fmt.Printf("Received a message from the %s: %s\n", ProvFSMInS[p_i].Prov.IDKey, string(m.Data))
				//})
			} else if ProvFSMInS[p_i].ARTIK10H && ProvFSMInS[p_i].Prov.Id == "3ca6094d-4d0c-422f-9bcf-75424294eda8" {
				ProvFSMInS[p_i].GPIONumStr = "203"
				//nc1.Subscribe(ProvFSMInS[p_i].Prov.Id, func(m *nats.Msg) {
    			//	fmt.Printf("Received a message from the %s: %s\n", ProvFSMInS[p_i].Prov.IDKey, string(m.Data))
				//})
			} else if ProvFSMInS[p_i].ARTIK10H && ProvFSMInS[p_i].Prov.Id == "a26a4357-6631-4686-8926-8fa108c8f752" {
				ProvFSMInS[p_i].GPIONumStr = "204"
				//nc1.Subscribe(ProvFSMInS[p_i].Prov.Id, func(m *nats.Msg) {
    			//	fmt.Printf("Received a message from the %s: %s\n", ProvFSMInS[p_i].Prov.IDKey, string(m.Data))
				//})
			}
			

			p_i++
		}
	}

	// initialize the GPIO for the reclamation pump also
	errRP := ioutil.WriteFile("/sys/class/gpio/export", []byte("14"), 0644)
	checkANDContinue(errRP)
	errRP  = ioutil.WriteFile("/sys/class/gpio/gpio14/direction", []byte("out"), 0644)
	checkANDContinue(errRP)
	errRP  = ioutil.WriteFile("/sys/class/gpio/gpio14/value", []byte("1"), 0644)
	checkANDContinue(errRP)

	// perform setup operations for the initiators		
	for _, i := range initiatorS {

		if i.Id != "" {

			i_m, _ := json.Marshal(i)
			i_s := string(i_m)
			fmt.Println(i_s)
			InitFSMInS = append(InitFSMInS, InitFSMIn{InitStr: i_s, RDBsession: session, Init: initiatorS[i_i], ProvS: providersS, RxFEmsg: RxFEmsgI[i_i], TxFEmsg: TxFEmsgI[i_i]})
			fmt.Println("\n")

			i_i++
		}
	}



	fmt.Printf("i_i is %d:\n", i_i)
	fmt.Printf("length of RxFEmsgI is %d:\n", len(RxFEmsgI))
	fmt.Printf("length of TxFEmsgI is %d:\n", len(TxFEmsgI))
	fmt.Printf("length of InitFSMInS is %d:\n", len(InitFSMInS))


	fmt.Printf("p_i is %d:\n", p_i)
	fmt.Printf("length of RxFEmsgP is %d:\n", len(RxFEmsgP))
	fmt.Printf("length of TxFEmsgP is %d:\n", len(TxFEmsgP))
	fmt.Printf("length of ProvFSMInS is %d:\n", len(ProvFSMInS))

	// run the initFSMs
	for j_i := 0; j_i < i_i; j_i++ {
		go Run_InitFSM(&InitFSMInS[j_i])
	}

	// run the ProvFSMs
	for j_p := 0; j_p < p_i; j_p++ {
		go Run_ProvFSM(&ProvFSMInS[j_p])
	}

	for j := range ProvFSMInS {defer ProvFSMInS[j].finishGPIO()}


	return func(w http.ResponseWriter, rs *http.Request) {

		// WS connection
		// **NOTE**:consider whether multiple connections would be advantageous
		c, err := Upgrader.Upgrade(w, rs, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		defer c.Close()

		/*
		// set up the WS write instances for the ProvFSMs		
		for j_p_1 := 0; j_p_1 <= len(ProvFSMInS)-1; j_p_1++ {
			fmt.Printf("Setting up a WS writer in loop1 for ProvFSMs\n")
			//goR0("Here we go 0: ", 5)
			go wsWriter(c, TxFEmsgP[j_p_1])
		}
		// set up the WS write instances for the InitFSMs
		for j_i_1 := 0; j_i_1 <= len(InitFSMInS)-1; j_i_1++ {
			fmt.Printf("Setting up a WS writer in loop1 for InitFSMs\n")
			//goR0("Here we go 0: ", 5)
			go wsWriter(c, TxFEmsgI[j_i_1])
		}
		*/

		// set up the WS reader
		go wsReader(c, 0, RxFEmsgI, RxFEmsgP)

		// set up the WS writer
		go wsWriter1(c, TxFEmsgI, TxFEmsgP)

		// empty for loop to preclude WS from closing
		for {
			// do nothing
			//func () { 	
			//	nc1.Publish("foo", []byte("Hello Particle Photon!"))
			//	goR0("test string", 50)
			//}()
		}

		fmt.Printf("u 'inside' is: %d\n", u)
		u++

	}
}

func initInitiators(session *r.Session) []Initiator {

    fmt.Println("entered initInitiators()")

    rows, err := r.Table("initiators").Run(session)
    if err != nil {
        fmt.Println(err)
        //return
    }

    var initiatorSlice []Initiator
    err2 := rows.All(&initiatorSlice)
    if err2 != nil {
        fmt.Println(err2)
        //return
    }

    fmt.Println("*** Fetch all rows: ***")
    for _, i := range initiatorSlice {
        printObj(i)
    }

    fmt.Println("\n")

    return initiatorSlice
}

func initProviders(session *r.Session) []Provider {

    fmt.Println("entered initProviders()")

    rows, err := r.Table("providers").Run(session)
    if err != nil {
        fmt.Println(err)
        //return
    }

    var providerSlice []Provider
    err2 := rows.All(&providerSlice)
    if err2 != nil {
        fmt.Println(err2)
        //return
    }

    fmt.Println("*** Fetch all rows: ***")
    for _, p := range providerSlice {
        printObj(p)
    }

    fmt.Println("\n")

    return providerSlice
}

func writeIDKeyInitiator(session *r.Session, initiator *object.User) {
    rows, err := r.Table("initiators").Run(session)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Read records into persons slice
    var initiators []Initiator
    err2 := rows.All(&initiators)
    if err2 != nil {
        fmt.Println(err2)
        return
    }

    PrintStr("*** Fetch all rows: ***")
    for _, i := range initiators {
        printObj(i)
        initiator.Name = i.ID
//     	fmt.Println("i is:", i)
    }
    PrintStr("\n")
}

func printObj(v interface{}) {
    vBytes, _ := json.Marshal(v)
    fmt.Println(string(vBytes))
}

func PrintStr(v string) {
    fmt.Println(v)
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func checkANDContinue(e error) {
    if e != nil {
    	fmt.Printf("\n\nioutil err, continuing\n\n")    
    }
}

func goR0(msg string, iter int) {

	const delay = 50 * time.Millisecond

	for i:=0; i<iter; i++ {
		fmt.Printf(msg+"%d\n", i)
		time.Sleep(delay)
	}
	return
}

func goR1(msg string) {

	const delay = 4000 * time.Millisecond

	for i:=0; i<25; i++ {
		fmt.Printf(msg+"%d\n", i)
		time.Sleep(delay)
	}
	return
}
