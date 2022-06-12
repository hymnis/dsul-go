# DSUL - Design Document

**dsuld** DSUL Daemon
  - settings: reading settings from file, environment or command line arguments
  - serial: reading and writing to the serial bus (the device)
  - ipc: reading and writing to the IPC bus (the client)

`dsulc/g, user data -> ipc -> main -> serial`

**dsulc** DSUL CLI
  - settings: reading settings from file, environment or command line arguments
  - ipc: reading and writing to the IPC bus (the daemon)

`user data -> main -> ipc -> dsuld`

**dsulg** DSUL GUI
  - settings: reading settings from file, environment or GUI
  - ipc: reading and writing to the IPC bus (the daemon)
  - gui: handle graphical interface and user interaction

`user interaction -> gui -> main -> ipc -> dsuld`
