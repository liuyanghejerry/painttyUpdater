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
	"osext"
	"path/filepath"
	"platform"
	"projectconst"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_ADDR_IPV4           = projectconst.SERVER_ADDR_IPV4
	SERVER_ADDR_IPV6           = projectconst.SERVER_ADDR_IPV6
	DEFAULT_NEW_PACKAGE        = projectconst.DEFAULT_NEW_PACKAGE
	UPDATER_VER         uint64 = projectconst.UPDATER_VER
)

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
	v := flag.Uint64("v", 40, "current version, unsigned integer")
	flag.Parse()
	return *v
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
	r := flag.String("i", "", "path to original updater")
	p := flag.Uint64("p", 0, "pid to original updater")
	flag.Parse()

	if *p <= 0 {
		return true, nil
	}

	old_proc, err := os.FindProcess(int(*p))

	if err != nil {
		return false, err
	}

	err = old_proc.Kill()

	if err != nil {
		return false, err
	}

	if len(*r) <= 0 {
		return false, nil
	}

	info, err := os.Stat(*r)
	if err != nil || !info.IsDir() {
		return false, err
	}

	current_path, err := osext.ExecutableFolder()

	if err != nil {
		return false, err
	}

	info, err = os.Stat(current_path)

	if err != nil || !info.IsDir() {
		return false, err
	}

	result := CopyDir(current_path, *r)

	if !result {
		return false, CopyFailedError{"Cannot copy files from " + current_path + " to " + *r}
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
		log.Panicln("error when post")
		return false, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Panicln("error while read response")
		return false, "", err
	}

	var version_info VersionCheckRes
	err = json.Unmarshal(respBody, &version_info)
	if err != nil {
		log.Panicln("error when unmarshal")
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
		log.Panicln("failed to download new package")
		return false, nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Panicln("failed to download new package")
		return false, nil, err
	}
	f, err := ioutil.TempFile("", "paintty_update_")
	if err != nil {
		log.Panicln("failed to write new package to disk")
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

func install(dest string) error {
	current_path, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	s := string(os.PathSeparator)
	new_updater := filepath.FromSlash(current_path + s + "updater")
	pid := os.Getpid()
	args := []string{"-i " + current_path + " ", "-p " + strconv.Itoa(pid) + " "}
	prcs, err := os.StartProcess(new_updater, args, nil)
	if err != nil {
		return err
	}

	// wait 30 sec at most and killed
	go func() {
		time.Sleep(time.Second * 30)
		if prcs != nil {
			_ = prcs.Kill()
		}
	}()

	// if everything goes smooth, updater will be killed before Wait() returns
	_, err = prcs.Wait()

	return err
}

func start(src string) bool {
	s := string(os.PathSeparator)
	_, err := os.StartProcess(src+s+"MrPaint", []string{}, nil)
	if err != nil {
		return false
	}

	return true
}

func main() {
	next, _ := complete()
	if !next {
		return
	}
	next, addr, err := check()
	if err != nil {
		log.Panicln("check failed")
		log.Panicln(err)
		return
	}

	if !next {
		return
	}

	next, f, err := download(addr)

	if f != nil {
		defer f.Close()
		defer os.Remove(f.Name())
	}

	if err != nil {
		log.Panicln("download failed")
		log.Panicln(err)
		return
	}

	if !next {
		return
	}

	log.Println(f.Name())

	next, file_path, err := uncompress(f)
	log.Println(file_path)

	if err != nil {
		log.Panicln("uncompress failed")
		log.Panicln(err)
		return
	}

	if !next {
		return
	}

	err = install(file_path)

	if err != nil {
		log.Panicln("install failed")
		log.Panicln(err)
		return
	}

	b := start(file_path)

	if !b {
		log.Panicln("start new MrPaint failed")
		return
	}
}
