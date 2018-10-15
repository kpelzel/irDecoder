irDecoder
=======

NOTICE: This package is untested

irDecoder uses a raspberry pi and IR receiver to decode any 38khz IR signal into it's underlying binary (or at least tries it's best)

Dependencies
------------
- Go
- github.com/stianeikeland/go-rpio

Software Installation & Setup
-----------------------------
irDecoder is written entirely in go and uses the go-rpio package to access the rpio pins. It can be built using:
        
        go build decode.go

Then execute the program as usual. Be sure to specify the signal pin from the ir sensor:
        
        ./decode 16

Hardware Installation & Setup
-----------------------------
This package was designed to be used with a raspberry pi (any version should work) and an ir receiver (something similar to the TSOP38238). The input pin signal pin from the ir receiver is hard coded in as rpio pin 10. This can easily be changed by changing the value at the top of decode.go. 
