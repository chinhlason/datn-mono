package helper

import (
	"fmt"
	"github.com/rainycape/unidecode"
	"strings"
)

func GenPatientCode(name string, phoneNumber string, ccid string) string {
	name = unidecode.Unidecode(name)

	nameParts := strings.Fields(name)
	lastName := nameParts[len(nameParts)-1]

	lastFourDigitsPhone := phoneNumber[len(phoneNumber)-4:]

	lastFourDigitsCCID := ccid[len(ccid)-4:]

	patientID := fmt.Sprintf("BN%s%s%s", lastName, lastFourDigitsPhone, lastFourDigitsCCID)

	return patientID
}
