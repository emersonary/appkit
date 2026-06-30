package social

import "strings"

const (
	sansSerifBoldUpperBase = 0x1D5D4 // 𝗔
	sansSerifBoldLowerBase = 0x1D5EE // 𝗮
	sansSerifBoldDigitBase = 0x1D7EC // 𝟬
)

// ToSansSerifBold maps ASCII letters and digits to Unicode mathematical sans-serif bold.
// Other runes (spaces, punctuation, accented letters) are left unchanged.
func ToSansSerifBold(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(sansSerifBoldUpperBase + (r - 'A'))
		case r >= 'a' && r <= 'z':
			b.WriteRune(sansSerifBoldLowerBase + (r - 'a'))
		case r >= '0' && r <= '9':
			b.WriteRune(sansSerifBoldDigitBase + (r - '0'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
