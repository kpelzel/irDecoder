package main

import (
	"fmt"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

var inputPin = 10
var timeout = 100000

// SignalPair is a pair of values that show the state of the ir receiver
type SignalPair struct {
	state rpio.State // State of the ir receiver (0 = pulse, 1 = gap)
	time  int64      // Time the ir receiver is in the state
}

func main() {
	rawSignal := decodeSignal(inputPin)
	gapValues, pulseValues := parseSignal(rawSignal)
	gapBinaryString := parseGapValues(gapValues, rawSignal)
	pulseBinaryString := parsePulseValues(pulseValues, rawSignal)
	fmt.Printf("Binary from gaps = %v\n", gapBinaryString)
	fmt.Printf("Binary from pulses = %v\n", pulseBinaryString)
}

func decodeSignal(inPin int) []SignalPair {
	err := rpio.Open()
	if err != nil {
		fmt.Printf("Error when opening rpio: %v", err)
	}
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

		// timout (global variable) is arbitrary, adjust as nessesary
		if highCounter > timeout {
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
		// If the state is a gap
		if pair.state == rpio.State(rpio.High) {
			fmt.Printf("Gap time: %v\n", pair.time)
			gapValues = addOrFindPulseGap(pair, gapValues)
		} else if pair.state == rpio.State(rpio.Low) {
			fmt.Printf("Pulse time: %v\n", pair.time)
			pulseValues = addOrFindPulseGap(pair, pulseValues)
		}
	}
	return gapValues, pulseValues
}

func addOrFindPulseGap(inputPair SignalPair, GapOrPulseValues [][]int64) [][]int64 {
	var foundAverageCategory bool

	// If GapOrPulseValues is totally empty
	if GapOrPulseValues == nil {
		s := []int64{inputPair.time}
		GapOrPulseValues = append(GapOrPulseValues, s)
	} else {
		// Get averages of the slices within GapOrPulseValues and see if the pair time fits somewhere
		foundAverageCategory = false
		var averages []int64
		for _, timeCategory := range GapOrPulseValues {
			averages = append(averages, averageOfSlice(timeCategory))
		}
		for index, average := range averages {
			if inputPair.time <= (average+250000) && inputPair.time >= (average-250000) && !foundAverageCategory {
				GapOrPulseValues[index] = append(GapOrPulseValues[index], inputPair.time)
				foundAverageCategory = true
				fmt.Printf("Found category. Inserting: %v into category: %v\n", inputPair.time, average)
			}
		}
		// If the pair time doesn't fit somewhere, create a new category for it
		if !foundAverageCategory {
			fmt.Printf("Category was not found. Adding %v to new category\n", inputPair.time)
			s := []int64{inputPair.time}
			GapOrPulseValues = append(GapOrPulseValues, s)
		}
	}
	return GapOrPulseValues

}

func parseGapValues(inputGapValues [][]int64, inputSignal []SignalPair) string {
	var binarySlice []byte
	var finalGapValues []int64
	var smallestGapIndex int

	for _, gapSlice := range inputGapValues {
		finalGapValues = append(finalGapValues, averageOfSlice(gapSlice))
	}
	fmt.Printf("Final gap values: \n")
	for _, value := range finalGapValues {
		fmt.Println(value)
	}

	fmt.Printf("Calculating binary string...\n")
	for _, pair := range inputSignal {
		if pair.state == rpio.State(rpio.High) {
			smallestGapIndex = indexOfSmallest(finalGapValues)

			if pair.time > finalGapValues[smallestGapIndex]+300000 {
				fmt.Printf("%v gapTime: %v = 1\n", finalGapValues[smallestGapIndex], pair.time)
				binarySlice = append(binarySlice, '1')
			} else {
				fmt.Printf("%v gapTime: %v = 0\n", finalGapValues[smallestGapIndex], pair.time)
				binarySlice = append(binarySlice, '0')
			}
		}
	}

	//remove first character because it's the leading bit
	binarySlice = binarySlice[1:]

	return string(binarySlice)
}

func parsePulseValues(inputPulseValues [][]int64, inputSignal []SignalPair) string {
	var binarySlice []byte
	var finalPulseValues []int64
	var smallestPulseIndex int

	for _, pulseSlice := range inputPulseValues {
		finalPulseValues = append(finalPulseValues, averageOfSlice(pulseSlice))
	}
	fmt.Printf("Final pulse values: \n")
	for _, value := range finalPulseValues {
		fmt.Println(value)
	}

	fmt.Printf("Calculating binary string...\n")
	for _, pair := range inputSignal {
		if pair.state == rpio.State(rpio.High) {
			smallestPulseIndex = indexOfSmallest(finalPulseValues)

			if pair.time > finalPulseValues[smallestPulseIndex]+300000 {
				fmt.Printf("%v gapTime: %v = 1\n", finalPulseValues[smallestPulseIndex], pair.time)
				binarySlice = append(binarySlice, '1')
			} else {
				fmt.Printf("%v gapTime: %v = 0\n", finalPulseValues[smallestPulseIndex], pair.time)
				binarySlice = append(binarySlice, '0')
			}
		}
	}

	//remove first character because it's the leading bit
	binarySlice = binarySlice[1:]

	return string(binarySlice)
}

func averageOfSlice(input []int64) int64 {
	var sum int64
	for _, value := range input {
		sum += value
	}
	return (sum / int64(len(input)))
}

//finds the index of the smallest value in a slice
func indexOfSmallest(inputSlice []int64) int {
	var smallestValue int64
	var smallestIndex int
	for index, value := range inputSlice {
		if smallestValue == 0 {
			smallestValue = value
			smallestIndex = index
		} else if value < smallestValue {
			smallestValue = value
			smallestIndex = index
		}
	}
	return smallestIndex
}
