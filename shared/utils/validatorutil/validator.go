package validatorutil

import "regexp"

func ValidateTelefone(telefone string) bool {
	telefone = regexp.MustCompile(`[^\d]`).ReplaceAllString(telefone, "")
	match, _ := regexp.MatchString(`^\d{2}9\d{8}$`, telefone)
	return match
}
