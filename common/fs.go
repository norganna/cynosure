package common

import (
	"os"
)

// FileExists returns whether the path exists and is a plain file.
func FileExists(path string) bool {
	ok, dir := checkExists(path)
	return ok && !dir
}

// DirExists returns whether the path exists and is a directory.
func DirExists(path string) bool {
	ok, dir := checkExists(path)
	return ok && dir
}

// Exists returns whether the path exists.
func Exists(path string) bool {
	ok, _ := checkExists(path)
	return ok
}

func checkExists(path string) (ok, dir bool) {
	s, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	return true, s.IsDir()
}
