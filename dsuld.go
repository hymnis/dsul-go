// DSUL - Disturb State USB Light : Daemon application.
package main

import (
    "log"

    "hymnis/dsul/ipc"
    "hymnis/dsul/serial"
    "hymnis/dsul/settings"
)

func main() {
    log.Print("DSUL Daemon initializing.")

    // get settings
    settings.Placeholder()

    // start runners
    go serial.Runner()
    go ipc.ServerRunner()

    select {} // run until user exits
}
