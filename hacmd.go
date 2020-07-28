package main

import (
	"fmt"
	"hacmd"
	"os"
)

func main() {

	// Default configuration JSON file
	configStr := "config.json"

	// Check length of args
	if len(os.Args[1:]) > 0 {
		// Is there a first argument
		if os.Args[1] != "" {
			fmt.Println("Arg[1]: " + os.Args[1])
			configStr = os.Args[1]
		}
	}

	fmt.Println(configStr)
	commandCenter := hacmd.New(configStr)

	for {
		select {
		case msgStr := <-commandCenter.CmdMessages:
			procIDCalled, action, actiontype, commands := commandCenter.ReadCommands(msgStr)
			if (procIDCalled == commandCenter.ProcID) && (actiontype == "config") {
				fmt.Println("Reading Configuration")
				commandCenter.Configured = true
			}
			if commandCenter.Configured && procIDCalled == commandCenter.ProcID {
				for _, command := range commands {
					// Send Commands to API
					if action == "api" {
						go commandCenter.APICommand(commandCenter.ProcID, command.Hubid, command.URL)
					}
					if action == "lutron" {
						go commandCenter.LutronCommand(command.URL)
					}
				}
			}
		}
	}
}
