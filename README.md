# README #

Purpose:
This repository will contain all of the files necessary to construct
a webserver and websocket-based server. Both these servers are capable of
running on the ARTIK10 embedded board and are required to compete in the
'MAD Challenge'. Running on ARTIK10 requires setting command line flags to
emulationSvr.go as described below (note in emulationSvr.go there is an
issue with setting the emulationSvr.go flag '-artik' to 'false')

Instructions for use:
- install go 1.5+
- copy objectLogic.go and interaction.go to $GOPATH/src
- 'go get' all of the imports in the files being used
- set up rethinkdb given the information and table entry schema shown in
rethinkEntriesProviders and rethinkEntriesInitiators:

	NOTE: The 'ID' and 'id' conflict has been fixed by renaming fields per 
	rethinkEntriesProviders and rethinkEntriesInitiators.

- launch mobileAppSvr.go, which defaults to point to FEfiles. Use mobileAppSvr.go to
serve the files. mobileAppSvr takes -h, -p, and -dir flags to change the
default host, port and directory, respectively
- launch emulationSvr.go, which defaults to the following:
	host 		:= flag.String	("h", "192.168.1.119", "host server")
	port 		:= flag.Int		("p", 8080, "port to serve on")
	// Bug to be fixed: overriding artik10host to be false on the command line
	// does not work. It is possible to set '-artik true' to override the default 
	// of false below however
	var artik10host *bool
	artik10host	= flag.Bool	("artik", false, "hosting on ARTIK10")
	-h, -p and -artik can be overidden (-artik has a default of false, which can
	be reliably overridden to true if running on ARTIK10 is needed. If running on
	ARTIK10, the following command must be issued for emulationSvr.go to launch without
	errors (this cmd sets up the GPIO9 for use:
	echo 9 >> /sys/class/gpio/export))

* [Learn Markdown](https://bitbucket.org/tutorials/markdowndemo)

### How do I get set up? ###
To be described as features are completed
* Summary of set up
* Configuration
* Dependencies
* Database configuration
* How to run tests
* Deployment instructions

### Contribution guidelines ###

* Writing tests
* Code review
* Other guidelines

### To be added

- host IP and port for rethinkdb, which is currently hard coded


### Who do I talk to? ###

* Repo owner or admin --> jrotella@nycap.rr.com (Jason Rotella)
* Other community or team contact