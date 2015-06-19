package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	
	"github.com/cloudfoundry/cli/plugin"

	"path/filepath"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//Wildcard is this plugin
type Wildcard struct {
	matchedApps 	[]Apps
}



//GetMetadata returns metatada
func (cmd *Wildcard) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Wildcard",
		Version: plugin.VersionType{ //leavealone
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{  //****** array of command structures
			{
				Name:     "wildcard-apps",
				Alias:	  "wc-a"
				HelpText: "List all apps in the target space that matches the wildcard",
				UsageDetails: plugin.Usage{
					Usage: "cf wildcard-apps APP_NAME_WITH_WILDCARD",
				},
			}, 
			{
				Name:     "wildcard-delete-a",
				Alias:	  "wc-da"
				HelpText: "Delete all apps in the target space that matches the wildcard",
				UsageDetails: plugin.Usage{
					Usage: "cf wildcard-delete-a APP_NAME_WITH_WILDCARD",
				},
			},
			{
				Name:     "wildcard-delete-i",
				Alias:	  "wc-di"
				HelpText: "Interactively delete apps in the target space that matches the wildcard",
				UsageDetails: plugin.Usage{
					Usage: "cf wildcard-delete-i APP_NAME_WITH_WILDCARD",
				},
			},//one argument
			// {
			// 	Name:     "scaleover",
			// 	HelpText: "Roll http traffic from one application to another",
			// 	UsageDetails: plugin.Usage{
			// 		Usage: "cf wildcard delete APP1 APP2 ROLLOVER_DURATION [--no-route-check]",
			//		Usage: "cf wildcard apps APP1 APP2 ROLLOVER_DURATION [--no-route-check]",
			// 	},
			// }, //second command
		},

	} //no comma because not a sequence
}

func main() { //like Java's main
	plugin.Start(new(Wildcard)) //Wildcard is the implementing class name => <unique>
}
//cmd is a pointer to an object of type DeleteWildCmb. 
func (cmd *Wildcard) usage(args []string) error {
	//counts number of args to confirm.
	badArgs := 3 != len(args) //either 4 args

	if badArgs {
		return errors.New("Usage: cf wildcard-apps\n\tcf wildcard-apps APP_NAME_WITH_WILDCARD")
	}
	return nil
}


//Run runs the plugin
//called everytime user executes the command
func (cmd *Wildcard) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "wildcard-apps" { //checking is very imp.
		cmd.WildcardCommandApps(cliConnection, args)
	} else if args[0] == "wildcard-delete-a" {
		cmd.WildcardCommandDeleteAll(cliConnection, args)
	} else if args[0] == "wildcard-delete-i" {
		cmd.WildcardCommandDeleteInteractive(cliConnection, args)
	}
}
//Q: what is the error for?
func (cmd *Wildcard) getAppInformation(cliConnection plugin.CliConnection, name string) (AppStatus, error) {
	status := AppStatus {
		name:           	name,
		requested-state:   	0,
		instances: 			string,
		memory:          	string,
		disk:        		string,
		urls:				[]string{},
	}
}
//WildcardCommand creates a new instance of this plugin
//this is the actual implementation
//one method per command
func (cmd *Wildcard) WildcardCommandApps(cliConnection plugin.CliConnection, args []string) {

	if err := cmd.usage(args); nil != err { //usage is just confirmation for correct number of args
		fmt.Println(err) //printing
		os.Exit(1) //failure
	}
	//returns 2 vals
	rolloverTime, err := cmd.parseTime(args[3]) //go functions can ret multiple values
	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	// The getAppStatus calls will exit with an error if the named apps don't exist
	//lines 109 and 110 may be clearer than the ff
	if cmd.app1, err = cmd.getAppStatus(cliConnection, args[1]); nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	if cmd.app2, err = cmd.getAppStatus(cliConnection, args[2]); nil != err {
		fmt.Println(err)
		os.Exit(1)
	}
	cmd.showStatus()

	count := cmd.app1.countRequested
	if count == 0 {
		fmt.Println("There are no instances of the source app to scale over")
		os.Exit(0)
	}
	sleepInterval := time.Duration(rolloverTime.Nanoseconds() / int64(count))

	for count > 0 {
		count--
		cmd.app2.scaleUp(cliConnection) //
		cmd.app1.scaleDown(cliConnection)
		cmd.showStatus()
		if count > 0 {
			time.Sleep(sleepInterval)
		}
	}
	fmt.Println()
}

// func (cmd *Wildcard) getAppStatus(cliConnection plugin.CliConnection, name string) (AppStatus, error) {
// 	status := AppStatus{
// 		name:           name,
// 		countRunning:   0,
// 		countRequested: 0,
// 		state:          "unknown",
// 		routes:         []string{},
// 	}

// 	output, _ := cliConnection.CliCommandWithoutTerminalOutput("app", name)

// 	for idx, v := range output {
// 		v = strings.TrimSpace(v)
// 		if strings.HasPrefix(v, "FAILED") {
// 			e := output[idx+1]
// 			return status, errors.New(e)
// 		}
// 		if strings.HasPrefix(v, "requested state: ") {
// 			status.state = strings.TrimPrefix(v, "requested state: ")
// 		}
// 		if strings.HasPrefix(v, "instances: ") {
// 			instances := strings.TrimPrefix(v, "instances: ")
// 			split := strings.Split(instances, "/")
// 			status.countRunning, _ = strconv.Atoi(split[0])
// 			status.countRequested, _ = strconv.Atoi(split[1])
// 		}
// 		if strings.HasPrefix(v, "urls: ") {
// 			urls := strings.TrimPrefix(v, "urls: ")
// 			status.routes = strings.Split(urls, ", ")
// 		}
// 	}
// 	// Compensate for some CF weirdness that leaves the requested instances non-zero
// 	// even though the app is stopped
// 	if "stopped" == status.state {
// 		status.countRequested = 0
// 	}
// 	return status, nil
// }
//HELPFUL
// func (app *AppStatus) scaleUp(cliConnection plugin.CliConnection) { tot. specific
// 	// If not already started, start it
// 	if app.state != "started" {
// 		//invokes another 'cf start ---app name---'
// 		//output is simple not shows to user.  the output can be retrived and parsed through.
// 		//cf apps => outputs all the apps (list of apps deployed) => parse and check if match wildcard
// 		//Q&A for user to choose 
// 		cliConnection.CliCommandWithoutTerminalOutput("start", app.name)
// 		app.state = "started"
// 	}
// 	app.countRequested++
// 	cliConnection.CliCommandWithoutTerminalOutput("scale", "-i", strconv.Itoa(app.countRequested), app.name)
// }

// func (app *AppStatus) scaleDown(cliConnection plugin.CliConnection) {
// 	app.countRequested--
// 	// If going to zero, stop the app
// 	if app.countRequested == 0 {
// 		cliConnection.CliCommandWithoutTerminalOutput("stop", app.name)
// 		app.state = "stopped"
// 	} else {
// 		cliConnection.CliCommandWithoutTerminalOutput("scale", "-i", strconv.Itoa(app.countRequested), app.name)
// 	}
// }

// func (cmd *Wildcard) showStatus() {
// 	if termutil.Isatty(os.Stdout.Fd()) { \\Test if this is the terminal
// 		fmt.Printf("%s (%s) %s %s %s (%s) \r", // first param is the FORMAT. %s is insert string here
//\r is 'return'. usually use \n. ||\r goes to beginning of current line.|| \n 
// 			cmd.app1.name, cmd.app1.state,
// 			strings.Repeat("<", cmd.app1.countRequested),
// 			strings.Repeat(">", cmd.app2.countRequested),
// 			cmd.app2.name,
// 			cmd.app2.state,
// 		)
// 	} else { \\Then being called from a script and therefore logging
// 		fmt.Printf("%s (%s) %d instances, %s (%s) %d instances\n",
// 			cmd.app1.name,
// 			cmd.app1.state,
// 			cmd.app1.countRequested,
// 			cmd.app2.name,
// 			cmd.app2.state,
// 			cmd.app2.countRequested,
// 		)
// 	}
// } //no need to stress about this. 

// func (cmd *Wildcard) appsShareARoute() bool { //plugin specific
// 	for _, r1 := range cmd.app1.routes {
// 		for _, r2 := range cmd.app2.routes {
// 			if r1 == r2 {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func (cmd *Wildcard) errorIfNoSharedRoute() error {
// 	if cmd.appsShareARoute() {
// 		return nil
// 	}
// 	return errors.New("Apps do not share a route!")
// }


