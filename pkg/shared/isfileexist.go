package shared

import (
	"os"
)

// IsFileExist is from the https://stackoverflow.com/a/10510783/9469533
func IsFileExist(path string) (result bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
