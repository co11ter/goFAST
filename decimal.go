package fast

import (
	"github.com/shopspring/decimal"
	"math"
)

// TODO int will be able overflow if exponent < 0 ??
func newFloat(mantissa int64, exponent int32) (f float64) {
	return float64(mantissa)/math.Pow10(int(exponent) * -1)
}

func newMantExp(f float64) (int64, int32) {
	if f == 0 {
		return 0,0
	}
	d := decimal.NewFromFloat(f)
	return d.Coefficient().Int64(), d.Exponent()
}


