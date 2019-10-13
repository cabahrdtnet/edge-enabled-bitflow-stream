package objects

import "fmt"

type ExportClientError struct {
	EngineName string
	AlreadyRegistered bool
	OtherError bool
	Err error
}

func (re ExportClientError) Error() string {
	if re.AlreadyRegistered {
		return fmt.Sprintf("%s is already registered as export client: %v", re.EngineName, re.Err)
	}

	if re.OtherError {
		return fmt.Sprintf("couldn't register %s as export client: %v", re.EngineName, re.Err)
	}

	return fmt.Sprintf("error flags are not set, but error occurred for %s: %v", re.EngineName, re.Err)
}
