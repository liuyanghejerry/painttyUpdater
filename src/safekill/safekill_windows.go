// +build windows

package safekill

import "os"

func SafeKill(pid int) error {
	old_proc, err := os.FindProcess(pid)

	if err != nil {
		return err
	}

	return old_proc.Kill()
}
