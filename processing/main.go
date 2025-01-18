package main

import (
	"fmt"
	"log"

	decoder "processing/adsbdecoder"
)

func main() {
	// Correct example
	hexExample := "8D4840D6202CC371C32CE0576098"

	message, err := decoder.DecodeModeSMessage(hexExample)
	if err != nil {
		log.Fatalf("Failed to decode Mode S message: %v", err)
	}

	callsign, wakeVortexCategory, err := decoder.ProcessAircraftIdentification(message.RawBits)

	if err != nil {
		log.Fatalf("Failed to decode aircraft identification: %v", err)
	}

	message.Callsign = callsign
	message.WakeCategory = wakeVortexCategory

	// Print the decoded DF (Downlink Format)
	fmt.Println("Decoded Mode S Message:")
	fmt.Printf("DF (Downlink Format): %d\n", message.DF)
	fmt.Printf("CA (Capability): %d\n", message.CA)
	fmt.Printf("ICAO address: %s\n", message.ICAO)
	fmt.Printf("Callsign: %s\n", message.Callsign)
	fmt.Printf("Wake vortex category: %s\n", message.WakeCategory)
	fmt.Printf("Raw Bits: %s\n", message.RawBits)
}
