package bank

type OperationError struct {
	Op string
}

func (e *OperationError) Error() string {
	if e.Op != "" {
		return "bad operation: " + e.Op
	}
	return "bad operation"
}

type NotEnoughMoneyError struct{}

func (e *NotEnoughMoneyError) Error() string {
	return "balance has not enough money"
}

type NoAccountError struct {
	Info string
}

func (e *NoAccountError) Error() string {
	if e.Info != "" {
		return "account doesn't exist: " + e.Info
	}
	return "account doesn't exist"
}

type InvalidSumError struct {
	Err error
}

func (e *InvalidSumError) Error() string {
	if e.Err != nil {
		return e.Error()
	}
	return ""
}

func (e *InvalidSumError) Unwrap() error {
	return e.Err
}

type NegativeSumError struct {
	Method string
}

func (e *NegativeSumError) Error() string {
	if e.Method != "" {
		return "negative sum: " + e.Method
	}
	return "negative sum"
}

type ZeroSumError struct {
	Method string
}

func (e *ZeroSumError) Error() string {
	if e.Method != "" {
		return "zero sum: " + e.Method
	}
	return "zero sum"
}
