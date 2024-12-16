package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type ModeSMessage struct {
	DF           int64  // Downlink Format (5 bits)
	CA           int64  // Capability (3 bits)
	ICAO         string // ICAO address (24 bits)
	Callsign     string // Up to 8 characters
	RawBits      string // Entire message in binary
	WakeCategory string
}

func main() {
	// Correct example
	hexExample := "8D4840D6202CC371C32CE0576098"

	// Decode the hex message into a ModeSMessage struct
	message, err := DecodeModeSMessage(hexExample)
	if err != nil {
		log.Fatalf("Failed to decode Mode S message: %v", err)
	}

	// Print the decoded DF (Downlink Format)
	fmt.Println("Decoded Mode S Message:")
	fmt.Printf("DF (Downlink Format): %d\n", message.DF)
	fmt.Printf("CA (Capability): %d\n", message.CA)
	fmt.Printf("ICAO address: %s\n", message.ICAO)
	fmt.Printf("Callsign: %s\n", message.Callsign)
	fmt.Printf("Wake vortex category: %s\n", message.WakeCategory)
	fmt.Printf("Raw Bits: %s\n", message.RawBits)
}

// DecodeModeSMessage decodes a Mode S hex string into a ModeSMessage struct
func DecodeModeSMessage(hexString string) (*ModeSMessage, error) {
	var message ModeSMessage

	// Convert the hex string to binary
	binaryString, err := ModeSHexToBin(hexString)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hex to binary: %w", err)
	}

	// Validate length
	if len(binaryString) != 112 {
		return nil, errors.New("Mode S message must be 112 bits")
	}

	message.RawBits = binaryString

	// Parse the DF (Downlink Format) (5)
	df, err := strconv.ParseInt(binaryString[:5], 2, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DF: %w", err)
	}

	message.DF = df
	binaryString = binaryString[5:]

	// Parse CA (Capability) (3)
	ca, err := strconv.ParseInt(binaryString[:3], 2, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA: %w", err)
	}

	message.CA = ca
	binaryString = binaryString[3:]

	// Parse ICAO address (24)
	icao_int, err := strconv.ParseInt(binaryString[:24], 2, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA: %w", err)
	}

	icao := fmt.Sprintf("%X", icao_int)
	message.ICAO = icao
	fmt.Println(binaryString[:24])
	binaryString = binaryString[24:]

	// *********************
	//        MESSAGE
	// *********************

	// Parse TC (Type code) (5)
	tc, err := strconv.ParseInt(binaryString[:5], 2, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TC: %w", err)
	}
	fmt.Println(tc)
	switch {
	case tc >= 1 && tc <= 4:
		callsign, wakeCategory, err := ProcessAircraftIdentification(binaryString[:56])
		if err != nil {
			return nil, fmt.Errorf("failed to parse Aircraft identification: %w", err)
		}
		message.Callsign = callsign
		message.WakeCategory = wakeCategory
	}

	binaryString = binaryString[56:]

	// Return the parsed message as a struct
	return &message, nil
}

func ProcessAircraftIdentification(binaryString string) (string, string, error) {
	// Parse TC (Type code) (5)
	tc, err := strconv.ParseInt(binaryString[:5], 2, 64)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse TC: %w", err)
	}
	binaryString = binaryString[5:]

	ca, err := strconv.ParseInt(binaryString[:3], 2, 64)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA: %w", err)
	}
	binaryString = binaryString[3:]

	wakeVortexCategory := getWakeTurbulenceCategory(tc, ca)

	var callsignBuilder strings.Builder

	for len(binaryString) != 0 {
		characterBinary := binaryString[:6]
		binaryString = binaryString[6:]

		characterInt, err := strconv.ParseInt(characterBinary, 2, 64)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse callsign character: %w", err)
		}

		var character rune

		switch {
		case 1 <= characterInt && characterInt <= 26:
			character = rune(characterInt + 64)
		case 48 <= characterInt && characterInt <= 57:
			character = rune(characterInt)
		case characterInt == 32:
			character = rune(characterInt)
		}

		callsignBuilder.WriteRune(character)
	}

	return callsignBuilder.String(), wakeVortexCategory, nil
}

func getWakeTurbulenceCategory(tc int64, ca int64) string {
	wakeVortexCategorylookupTable := map[int64]map[int64]string{
		1: {
			0: "No category information", // TC = 1, CA = 0
		},
		2: {
			1: "Surface emergency vehicle",
			3: "Surface service vehicle",
			4: "Ground obstruction",
			5: "Reserved",
		},
		3: {
			1: "Glider, sailplane",
			2: "Lighter-than-air",
			3: "Parachutist, skydiver",
			4: "Ultralight, hang-glider, paraglider",
			5: "Reserved",
			6: "Unmanned aerial vehicle",
			7: "Space or transatmospheric vehicle",
		},
		4: {
			1: "Light (less than 7000 kg)",
			2: "Medium 1 (between 7000 kg and 34000 kg)",
			3: "Medium 2 (between 34000 kg and 136000 kg)",
			4: "High vortex aircraft",
			5: "Heavy (larger than 136000 kg)",
			6: "High performance (>5 g acceleration and high speed (>400 kt))",
			7: "Rotorcraft",
		},
	}

	if tc == 1 {
		return "Reserved"
	} else if ca == 0 {
		return "No category information"
	} else {
		return wakeVortexCategorylookupTable[tc][ca]
	}
}

// ModeSHexToBin converts a hex string to a binary string
func ModeSHexToBin(hexString string) (string, error) {
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %w", err)
	}

	var bits string
	for _, b := range bytes {
		bits += fmt.Sprintf("%08b", b)
	}

	return bits, nil
}
