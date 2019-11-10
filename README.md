# KSysGuard AMDGPU Sensor

This is a small tool to graph AMD GPU values using KSysGuard. It Implements the KSysGuard Protocol
separately if anyone wants to reuse it.

## How To Use

Clone this repo and then use the Go toolchain to build

    go build .
    
You can then either run it as a daemon (listens on localhost:2635) or run it as a command.

In KSysGuard go to 

    File -> Monitor Remote Machine
   
Enter a value for host (e.g. 127.0.0.1), Choose "`Custom Command`" and 
enter `/the/full/path/ksysguard-amdgpu`. Hit `OK` and you should now see the some extra fields to
graph in your Sensor Browser!

Note: You might want to `File -> New Tab` to see the sensor browser

 