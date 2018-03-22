package object

import "fmt"
import "time"

/*
//Objects-----

//Example Quotas
    //Quota1
        //David can use 600 gallons of showers this month
        //user or group, usageLimit, waterDevice

    //Quota2
        //The family can use 10000 gallons this month
        //user or group, usageLimit
    */
//
/*
NOTES:
 - Should the flow be controllable only by the electronic valve (if you do not have access to a device how to control the water)
 - The user needs to always be able to user their water (i.e. when they dont have their smart phone around, or set their quota too low (or even emergencies where they need to use water))
 - 


*/
//------------------------------------------Objects

/*
type Quota struct {     
    Name            string
    UsageLimit      int
    WaterDevice     string // optional 
}

*/
type Organization struct {
    Name            string
    Quota           float64
    Members         []User
    WaterDevices    []WaterDevice
}

type User struct {
    Name            string
    Userid          int
    Admin           bool
    TotalUsage      float64
    Quota           float64
}

type WaterDevice struct {
    Name            string
    Index           string // cast into an int if indexing is needed
    Room            int
    Permissive      bool 
    On              bool
    Usage           float64
    AccessLog       []LogEntry
}

type ComplexWaterDevice struct {
    BaseWaterDevice WaterDevice
    Civic           bool
    Collaborative   bool
}

type Quota struct {     
    UsageLimit      int
    Owner           string 
}

type LogEntry struct {
    Userid          int
    Time            time.Time
    FlowRate        float64
    DiffConsumed    float64 //The water consumed since the last log was taken (water consumed during this use)
}

//*************************************************
//------------------------------------------Methods
//*************************************************

// >> waterDevice Methods 
    
/*    
interface waterDevice {
    updateState // When flow reaches a setpoint, change state
    getState    // return true or false for on or off
    getWaterUsed  //amount consumed in this period (since the last state change)
    getFlowRate   //real time flow rate measurement
    getAccessLog  //Returns an array of maps, each map is a logEntry (i.e. {user: "David", time: "1:00, 1,5,2016", consumed: 7})
    //The waterDevice tracks short term usage history...
}
*/

func (wD *WaterDevice) UpdateUsage() float64 {
    
    wD.Usage += 100
    fmt.Println(wD.Name+".Usage is now: ", wD.Usage)
    return wD.Usage

}

func (wD *WaterDevice) newLogEntry(logUser int, fRate float64) float64 {
    
    newLog := LogEntry{
        Userid: logUser,
        Time: time.Now(),
        FlowRate: fRate,
    }
    
    wD.AccessLog = append(wD.AccessLog, newLog)
    
    flowConsumed := wD.measureFlow()
    
    return flowConsumed
}


func (wD *WaterDevice) measureFlow () float64 { //calculates diffConsumed and puts it into the mpLog

    logLen := len(wD.AccessLog)
    fmt.Printf("logLen:") 
    fmt.Println(logLen)
    if logLen >= 2 {
        lastFlowRate := wD.AccessLog[logLen-2].FlowRate //What was the flow rate of the last recorded log entry
        
        timeElapsedInt := wD.AccessLog[logLen-1].Time.Second() - wD.AccessLog[logLen-2].Time.Second()
        timeElapsed := float64(timeElapsedInt)
        flowConsumed := timeElapsed*lastFlowRate
        wD.AccessLog[logLen-1].DiffConsumed = flowConsumed
        
        return flowConsumed
        
    } else {
        
        return 0
        //wD.AccessLog[logLen].diffConsumed = 0
    }
}

/*  JER 01_15_2016

func (wD *waterDevice) updateState( "*FlowData" ) {
        setpoint = 0.4 int //gpm
        hysteresis = 0.2 int //gpm
        if  ((wD.getState() == "off") && (flow > setpoint)){
            wD.setState("on")
        } else if  ((wD.getState() == "on") && (flow < (setpoint-hysteresis))){
            wD.setState("off")
        }

    waterConsumed = calcWaterConsumedJustNow("*FlowData" )
    wD.newLogEntry(1, waterConsumed) //need to change this methods definition above

    //Need a separate "in use" function to consider turning on and off within small amount of time as 1 "use" to be recorded in the log
}

// >> user Methods

 /*
interface user {
    david.useDevice(sink) //change the device state
    david.updateTotalUsage //look in the access log for all devices and save as totalUsage (maybe unneccessary depending on DB implementation)
    david.getTotalUsage // return totalUsage
    david.getDeviceUsage(sink) // look in the sink's access log for all entrys by this user
*/


func (u *User) UseDevice() {

    fmt.Println("This person wants to use the device:", u.Name)

} 

/*  JER 01_15_2016
 
func (u *user) updateTotalUsage (deviceList [1]waterDevice) {
    //Check all devices to add up all the water used by this user
        //NOTE: may eventually implement to look in the database
    
    listLen := len(deviceList)
    
    for j := 0; j <listLen; j++ {
        wD := deviceList[j]     // For each wD in deviceList
        
        logLen := len(wD.AccessLog)
        for j := 0; j <logLen; j++ {
            logEntry := wD.AccessLog[j]     //For each logEntry in the AccessLog of that waterDevice
            
            if logEntry.userid == u.userid {    //If the logEntry was made by this user ("u")
                
                u.totalUsage += logEntry.diffConsumed    //add the diffConsumed to the users total "waterUsed" 
            
            }
        }
    }
}

JER 01_15_2016 */

// func (u *user) getTotalUsage (

// >> Quota Methods    
/*
interface quota{
    createQ() // initialize with name, limit and specific device(optional)
    update() // update the quota with the totalConsumed by the user or group
    quotaFilled() // return true if quota is full
}

func (*userQuota u) quotaFilled ( int totalConsumed) quotaFilled bool {
    
    if (quota reached) {
        return quotaFilled = true
    {
}
    
func (*userQuota u) updateQ ( int totalConsumed) quotaFilled bool {
 
    u.waterDevice = device
    u.usageLimit 
    
    if (quota reached) {
        return quotaFilled = true
    {
}
    
*/
//

//------------------------------------------Functions

/*  JER 01_15_2016
func initQuota(initName string, initUsageLimit int, initWaterDevice string) quota {
        return quota {
            name string
            usageLimit int
            waterDevice string  
        }
}

func initUser(initName string, initUserID int, initAdmin bool, initWaterUsed int) user {
        return user {
            name:   initName,
            userid: initUserID,
            admin:  initAdmin,
            waterUsed:  initWaterUsed,
        }
}
        
func initWaterDevice(initName string, initRoom int, initState bool) waterDevice {
        var log []logEntry

        return waterDevice {
            name: initName,
            room: initRoom,
            state: initState,
        }
}

JER 01_15_2016 */

    
  
  

//------------------------------------------Main

/*  JER 01_15_2016        
func main() {
    
        //Toggle the state of the controlPoint ON
        fmt.Println(valve1.changeState())
        //create a log in the measurePoint
        fmt.Println(flowMeter1.newLogEntry(1, 2.34))    //Kitchen Sink measuring 2.34 gpm
        
        Then := time.Now().Second()
        for time.Now().Second() < (Then+3){     //Homemade 3 second delay
            //do nothing
        }
        
        //Toggle the state of the controlPoint OFF
        fmt.Println(valve1.changeState())
        //create a log in the measurePoint
        fmt.Println(flowMeter1.newLogEntry(1, 0))       //Return how much water was consumed since last time
        
        
        mpList := [1]measurePoint{flowMeter1}
        //Diplay how much water the user consumed
        fmt.Println(david.calcWaterUsed(mpList))
}
JER 01_15_2016 */