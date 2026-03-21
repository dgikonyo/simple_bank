package util

// constants for all supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// The function `isSupportedCurrency` checks if a given currency is supported (USD, EUR, or CAD) and
// returns true if it is.
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}

	return false
}
