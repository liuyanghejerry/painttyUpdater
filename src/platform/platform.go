package platform

import (
		"runtime"
		"path"
)

func GetPlatformName() string {
	if runtime.GOOS == "darwin" {
		return "mac"
	}
	return runtime.GOOS
}

func GetPlatformArch() (ret string) {
	switch runtime.GOARCH {
	case "amd64":
		ret = "x64"
	case "386":
		ret = "x86"
	default:
		ret = "x86"
	}
	return ret
}

func GetPlatformExcSuffix() (ret string) {
	switch runtime.GOOS {
	case "windows":
		ret = ".exe"
	default:
		ret = ""
	}
	return ret
}

func GetPlatformExec(dir string, exec string) (ret string) {
	switch runtime.GOOS {
	case "darwin":
		ret = path.Join(dir, "MrPaint.app", "Contents", "MacOS", exec)
	default:
		ret = path.Join(dir, exec + GetPlatformExcSuffix())
	}
	return ret
}

func GetPlatformExecRoot(dir string) (ret string) {
	switch runtime.GOOS {
	case "darwin":
		ret = path.Join(dir, "..", "..")
	default:
		ret = dir
	}
	return ret
}
