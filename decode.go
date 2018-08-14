package main

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio"
)

// SignalPair is a pair of values that show the state of the ir receiver
type SignalPair struct {
	state rpio.State // State of the ir receiver (0 = pulse, 1 = gap)
	time  int64      // Time the ir receiver is in the state
}

func main() {

}

func decodeSignal(inPin int) []SignalPair {

	pin := rpio.Pin(inPin)
	pin.Input()
	currentState := rpio.State(rpio.High)

	var command []SignalPair

	// Loop until we see a low state
	for currentState == rpio.State(rpio.High) {
		currentState = pin.Read()
	}

	startTime := time.Now()

	highCounter := 0
	previousState := rpio.State(rpio.Low)

	for {
		if currentState != previousState {
			// The state has changed, so calculate the length of this run
			now := time.Now()
			elapsed := now.Sub(startTime)
			startTime = now

			var mySignalPair SignalPair
			mySignalPair.state = previousState
			mySignalPair.time = elapsed.Nanoseconds()

			command = append(command, mySignalPair)
		}

		if currentState == rpio.State(rpio.High) {
			highCounter++
		} else {
			highCounter = 0
		}

		// 10000 is arbitrary, adjust as nessesary
		if highCounter > 100000 {
			break
		}

		previousState = currentState
		currentState = pin.Read()
	}
	return command
}

func parseSignal(inputSignal []SignalPair) ([][]int64, [][]int64) {
	var gapValues [][]int64
	var pulseValues [][]int64

	for _, pair := range inputSignal {
		if pair.state == rpio.State(rpio.High) {
			fmt.Printf("Gap time: %v\n", pair.time)
			if gapValues == nil {
				s := []int64{pair.time}
				gapValues = append(gapValues, s)
			} else {

			}
		}
	}
	return gapValues, pulseValues
}
