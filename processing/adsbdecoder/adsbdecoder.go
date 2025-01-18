package adsbdecoder

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ModeSMessage struct {
	DF           int64  // Downlink Format (5 bits)
	CA           int64  // Capability (3 bits)
	ICAO         string // ICAO address (24 bits)
	TC           int64  // Type Code (5 bits)
	Callsign     string // Up to 8 characters
	RawBits      string // Entire message in binary
	WakeCategory string
}

func HexToBinaryString(hexString string) (string, error) {
	const ModeSMessageLengthBits = 112

	bytes, err := hex.DecodeString(hexString)

	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %w", err)
	}

	var binaryString string
	for _, b := range bytes {
		binaryString += fmt.Sprintf("%08b", b)
	}

	if len(binaryString) != ModeSMessageLengthBits {
		return "", fmt.Errorf("message contains %d bits, expected %d", len(binaryString), ModeSMessageLengthBits)
	}

	return binaryString, nil
}

// Parse the DF (Downlink Format) (5)
func ParseDownlinkFormat(binaryString string) (int64, error) {
	df, err := strconv.ParseInt(binaryString[:5], 2, 64)

	if err != nil {
		return 0, fmt.Errorf("failed to parse DF: %w", err)
	}

	return df, nil
}

// Parse the CA (Capability) (3)
func ParseCapability(binaryString string) (int64, error) {
	ca, err := strconv.ParseInt(binaryString[5:8], 2, 64)

	if err != nil {
		return 0, fmt.Errorf("failed to parse CA: %w", err)
	}

	return ca, nil
}

// Parse the ICAO address (24)
func ParseICAOAddress(binaryString string) (string, error) {
	icaoAddressInt, err := strconv.ParseInt(binaryString[8:32], 2, 64)

	if err != nil {
		return "", fmt.Errorf("failed to parse ICAO address: %w", err)
	}

	icaoAddressHexadecimal := fmt.Sprintf("%X", icaoAddressInt)

	return icaoAddressHexadecimal, nil
}

// Parse the TC (Type code) (5)
func ParseTypeCode(binaryString string) (int64, error) {
	tc, err := strconv.ParseInt(binaryString[32:37], 2, 64)

	if err != nil {
		return 0, fmt.Errorf("failed to parse TC: %w", err)
	}

	return tc, nil
}

func DecodeModeSMessage(hexString string) (*ModeSMessage, error) {
	binaryString, err := HexToBinaryString(hexString)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hex to binary: %w", err)
	}

	df, err := ParseDownlinkFormat(binaryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse downlink format: %w", err)
	}

	ca, err := ParseCapability(binaryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse capability: %w", err)
	}

	icaoAddress, err := ParseICAOAddress(binaryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICAO address: %w", err)
	}

	tc, err := ParseTypeCode(binaryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse type code: %w", err)
	}

	message := &ModeSMessage{
		DF:      df,
		CA:      ca,
		ICAO:    icaoAddress,
		TC:      tc,
		RawBits: binaryString,
	}

	return message, nil
}

/*func ProcessAircraftLocationGlobally(binaryStringOdd string, binaryStringEven string) (string, string, error) {
	var nZ float64 = 15 // predefined n# of latitude zones for mode S

	// Parse TC (Type code) (5)
	tc, err := strconv.ParseInt(binaryString[:5], 2, 64)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse TC: %w", err)
	}

	cprFormat := binaryString[21] // 0 - even, 1  - odd

	nCprLat, err1 := strconv.ParseInt(binaryString[22:38], 2, 64)
	nCprLon, err2 := strconv.ParseInt(binaryString[39:55], 2, 64)

	if err1 != nil || err2 != nil {
		//return nil, fmt.Errorf("failed to parse TC: %w", err1)
	}

	latCpr := float64(nCprLat) / math.Pow(2, 17)
	lonCpr := float64(nCprLon) / math.Pow(2, 17)

	var dLat float64

	if cprFormat == 0 {
		dLat = 360 / (4 * nZ)
	} else {
		dLat = 360 / (4*nZ - 1)
	}

	latZoneIndex := math.Floor(59 * lat)
}*/

func ProcessAircraftIdentification(binaryString string) (string, string, error) {
	binaryString = binaryString[32:88]

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
		2: {
			1: "Surface emergency vehicle",
			3: "Surface service vehicle",
			4: "Ground obstruction",
			5: "Ground obstruction",
			6: "Ground obstruction",
			7: "Ground obstruction",
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

// Yield the longtitude zone number given latitude
func calculateLongtitudeZoneNumber(latitude float64) int64 {
	switch {
	case latitude == 0:
		return 59
	case math.Abs(latitude) == 87:
		return 2
	case math.Abs(latitude) > 87:
		return 1
	}

	pi := math.Pi
	var nZ float64 = 15 // predefined n# of latitude zones for mode S

	nominator := 2 * pi

	acosArgument := 1 - (1-math.Cos(pi/2/nZ))/(math.Pow(math.Cos(pi*latitude/180), 2))

	if acosArgument > 1 {
		acosArgument = 1
	} else if acosArgument < -1 {
		acosArgument = -1
	}

	denominator := math.Acos(acosArgument)

	return int64(math.Floor(nominator / denominator))
}
