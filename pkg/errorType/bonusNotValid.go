package errorType

import "fmt"

type (
	ErrBonusNotValid struct {
		BonusID  int64
		Message  string
		Location string
	}
)

func (err ErrBonusNotValid) Error() string {
	return fmt.Sprintf("%v: %v {BonusID=%v}", err.Location, err.Message, err.BonusID)
}