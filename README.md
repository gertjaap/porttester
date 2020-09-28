# Super simple echo server

Listens with a TCP server on ports specified by environment variable `DUMMY_TCPPORTS` (separated by `|`) 
Listens with a UDP server on ports specified by environment variable `DUMMY_UDPPORTS` (separated by `|`) 

Listens on ALL interfaces

Will echo anything sent to it (for UDP echoes entire data, for TCP echoes once you enter a `\n` (newline))