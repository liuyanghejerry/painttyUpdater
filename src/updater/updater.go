package main

import (
	"archive/zip"
	"bytes"
	"easycp"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"ipv6support"
	"language"
	"log"
	"net/http"
	"os"
	"os/exec"
	"osext"
	"path/filepath"
	"platform"
	"projectconst"
	"strconv"
	"strings"
	//"time"
)

const (
	SERVER_ADDR_IPV4           = projectconst.SERVER_ADDR_IPV4
	SERVER_ADDR_IPV6           = projectconst.SERVER_ADDR_IPV6
	DEFAULT_NEW_PACKAGE        = projectconst.DEFAULT_NEW_PACKAGE
	UPDATER_VER         uint64 = projectconst.UPDATER_VER
)

var to_be_killed uint64
var to_be_installed string
var last_version uint64

type VersionCheckReq struct {
	Request  string `json:"request"`
	Platform string `json:"platform"`
	Language string `json:"language"`
}

type VersionInfoRes struct {
	Version   uint64 `json:"version"`
	Level     uint64 `json:"level"`
	Changelog string `json:"changelog"`
	Url       string `json:"url,omitempty"`
}

type UpdaterRes struct {
	Version uint64 `json:"version"`
}

type VersionCheckRes struct {
	Response string         `json:"response"`
	Result   bool           `json:"result"`
	Info     VersionInfoRes `json:"info,omitempty"`
	Updater  UpdaterRes     `json:"updater,omitempty"`
}

type CopyFailedError struct {
	message string
}

func (self CopyFailedError) Error() string {
	return self.message
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyDir(src string, dest string) bool {
	return easycp.CopyDir(src, dest)
}

func oldVersion() uint64 {
	return last_version
}

func queryPlatform() string {
	p := platform.GetPlatformName()
	a := platform.GetPlatformArch()
	if p == "mac" {
		return p
	}
	return p + " " + a
}

func queryLanguage() string {
	return language.GetSystemLanguage()
}

// bool indicates whether should we go further
func complete() (bool, error) {
	if to_be_killed <= 0 || len(to_be_installed) <= 0 {
		return true, nil
	}

	old_proc, err := os.FindProcess(int(to_be_killed))

	if err != nil {
		log.Println(err)
		return false, err
	}

	err = old_proc.Kill()

	if err != nil {
		log.Println(err)
		return false, err
	}

	if len(to_be_installed) <= 0 {
		return false, nil
	}

	// NOTE: to_be_installed is surrounded by quotes
	// to_be_installed = strings.Replace(to_be_installed, "\"", "", -1)
	to_be_installed = to_be_installed[1:len(to_be_installed)-1]

	info, err := os.Stat(to_be_installed)
	if err != nil || !info.IsDir() {
		log.Println(err)
		return false, err
	}

	current_path, err := osext.ExecutableFolder()

	if err != nil {
		log.Println(err)
		return false, err
	}

	info, err = os.Stat(current_path)

	if err != nil || !info.IsDir() {
		log.Println(err)
		return false, err
	}

	log.Println("copying ", current_path, " to ", to_be_installed)

	result := CopyDir(current_path, to_be_installed)

	if !result {
		return false, CopyFailedError{
			"Cannot copy files from " + current_path + " to " + to_be_installed}
	}

	b := start(to_be_installed)

	if !b {
		log.Println("start new MrPaint failed")
		// return false, nil
	}

	return false, nil
}

func check() (bool, string, error) {
	m := VersionCheckReq{"version", queryPlatform(), queryLanguage()}
	jsonRequest, err := json.Marshal(m)
	if err != nil {
		return false, "", err
	}

	var jsonPkg bytes.Buffer
	json.Compact(&jsonPkg, jsonRequest)
	var addr string
	if ipv6support.IsIPv6Supported() {
		addr = SERVER_ADDR_IPV6
	} else {
		addr = SERVER_ADDR_IPV4
	}
	resp, err := http.Post(addr, "application/json", &jsonPkg)
	if err != nil {
		log.Println(err)
		log.Println("error when post")
		return false, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println("error while read response")
		return false, "", err
	}

	var version_info VersionCheckRes
	err = json.Unmarshal(respBody, &version_info)
	if err != nil {
		log.Println("error when unmarshal")
		return false, "", err
	}
	log.Println(version_info.Info)

	if !version_info.Result {
		log.Println("server rejected request")
		return false, "", nil
	}

	if version_info.Info.Version > oldVersion() {
		if len(strings.TrimSpace(version_info.Info.Url)) <= 0 {
			return true, DEFAULT_NEW_PACKAGE, nil
		}

		return true, version_info.Info.Url, nil
	}

	return false, "", nil
}

func download(address string) (bool, *os.File, error) {
	res, err := http.Get(address)
	if err != nil {
		log.Println("failed to download new package")
		return false, nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println("failed to download new package")
		return false, nil, err
	}
	f, err := ioutil.TempFile("", "paintty_update_")
	if err != nil {
		log.Println("failed to write new package to disk")
		return false, nil, err
	}

	f.Write(data)

	return true, f, nil
}

func uncompress(file *os.File) (bool, string, error) {
	dir_name := file.Name() + "_dir"
	err := Unzip(file.Name(), dir_name)
	if err != nil {
		return false, "", err
	}

	return true, dir_name, nil
}

func install(src string) error {
	current_path, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}

	if to_be_killed != 0 {
		old_proc, err := os.FindProcess(int(to_be_killed))

		if err != nil {
			return err
		}

		err = old_proc.Kill()

		if err != nil {
			return err
		}
	}

	s := string(os.PathSeparator)
	new_updater := filepath.FromSlash(src + s + "updater.exe")
	log.Println("new updater: ", new_updater)
	pid := os.Getpid()

	cmd := exec.Command(new_updater, "-i", "\""+current_path+"\"", "-p", strconv.Itoa(pid))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	// args := []string{"-i", "\""+current_path+"\"", "-p", strconv.Itoa(pid)}
	// log.Println("start cmd: ", args)
	// var procAttr os.ProcAttr 
 //    procAttr.Files = []*os.File{nil, nil, nil} 
	// prcs, err := os.StartProcess(new_updater, args, &procAttr)
	// if err != nil {
	// 	return err
	// }

	// // wait 30 sec at most and killed
	// // go func() {
	// // 	time.Sleep(time.Second * 30)
	// // 	if prcs != nil {
	// // 		_ = prcs.Kill()
	// // 	}
	// // }()

	// // if everything goes smooth, updater will be killed before Wait() returns
	// _, err = prcs.Wait()

	return err
}

func start(src string) bool {
	// _, err := os.StartProcess(src+"MrPaint.exe", []string{}, nil)
	cmd := exec.Command(src+"MrPaint.exe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}

func init() {
	i := flag.String("i", "", "path to original updater")
	p := flag.Uint64("p", 0, "pid to original updater")
	v := flag.Uint64("v", 40, "current version, unsigned integer")
	flag.Parse()

	to_be_installed = *i
	to_be_killed = *p
	last_version = *v

	lf, err := os.OpenFile("./updater.log", os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
	}
	log.SetOutput(lf)

	log.Println("my cmd: ", to_be_installed, to_be_killed, last_version)
}

func main() {
	next, _ := complete()
	if !next {
		return
	}
	next, addr, err := check()
	if err != nil {
		log.Println("check failed")
		log.Println(err)
		return
	}

	if !next {
		return
	}

	next, f, err := download(addr)

	if f != nil {
		defer f.Close()
		// defer os.Remove(f.Name())
	}

	if err != nil {
		log.Println("download failed")
		log.Println(err)
		return
	}

	if !next {
		return
	}

	log.Println(f.Name())

	next, file_path, err := uncompress(f)
	log.Println(file_path)

	if err != nil {
		log.Println("uncompress failed")
		log.Println(err)
		return
	}

	if !next {
		return
	}

	err = install(file_path)

	if err != nil {
		log.Println("install failed")
		log.Println(err)
		return
	}

	// b := start(file_path)

	// if !b {
	// 	log.Println("start new MrPaint failed")
	// 	return
	// }
}
