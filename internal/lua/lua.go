package lua

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	humanize "github.com/dustin/go-humanize"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

var L *lua.LState

// LoadFile loads a lua file
func LoadFile(module string, file string, data []byte) error {
	pluginDef := []byte("module(\"" + module + "\", package.seeall)")

	if fn, err := L.Load(bytes.NewReader(append(pluginDef, data...)), file); err != nil {
		return err
	} else {
		L.Push(fn)
		return L.PCall(0, lua.MultRet, nil)
	}
}

// Import allows a lua plugin to import a package
func Import(pkg string) *lua.LTable {
	switch pkg {
	case "fmt":
		return importFmt()
	case "io":
		return importIo()
	case "io/ioutil", "ioutil":
		return importIoUtil()
	case "net":
		return importNet()
	case "math":
		return importMath()
	case "math/rand":
		return importMathRand()
	case "os":
		return importOs()
	case "runtime":
		return importRuntime()
	case "path":
		return importPath()
	case "path/filepath", "filepath":
		return importFilePath()
	case "strings":
		return importStrings()
	case "regexp":
		return importRegexp()
	case "errors":
		return importErrors()
	case "time":
		return importTime()
	case "unicode/utf8", "utf8":
		return importUtf8()
	case "humanize":
		return importHumanize()
	case "net/http", "http":
		return importHTTP()
	case "archive/zip":
		return importArchiveZip()
	default:
		return nil
	}
}

func importFmt() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Errorf", luar.New(L, fmt.Errorf))
	L.SetField(pkg, "Fprint", luar.New(L, fmt.Fprint))
	L.SetField(pkg, "Fprintf", luar.New(L, fmt.Fprintf))
	L.SetField(pkg, "Fprintln", luar.New(L, fmt.Fprintln))
	L.SetField(pkg, "Fscan", luar.New(L, fmt.Fscan))
	L.SetField(pkg, "Fscanf", luar.New(L, fmt.Fscanf))
	L.SetField(pkg, "Fscanln", luar.New(L, fmt.Fscanln))
	L.SetField(pkg, "Print", luar.New(L, fmt.Print))
	L.SetField(pkg, "Printf", luar.New(L, fmt.Printf))
	L.SetField(pkg, "Println", luar.New(L, fmt.Println))
	L.SetField(pkg, "Scan", luar.New(L, fmt.Scan))
	L.SetField(pkg, "Scanf", luar.New(L, fmt.Scanf))
	L.SetField(pkg, "Scanln", luar.New(L, fmt.Scanln))
	L.SetField(pkg, "Sprint", luar.New(L, fmt.Sprint))
	L.SetField(pkg, "Sprintf", luar.New(L, fmt.Sprintf))
	L.SetField(pkg, "Sprintln", luar.New(L, fmt.Sprintln))
	L.SetField(pkg, "Sscan", luar.New(L, fmt.Sscan))
	L.SetField(pkg, "Sscanf", luar.New(L, fmt.Sscanf))
	L.SetField(pkg, "Sscanln", luar.New(L, fmt.Sscanln))

	return pkg
}

func importIo() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Copy", luar.New(L, io.Copy))
	L.SetField(pkg, "CopyN", luar.New(L, io.CopyN))
	L.SetField(pkg, "EOF", luar.New(L, io.EOF))
	L.SetField(pkg, "ErrClosedPipe", luar.New(L, io.ErrClosedPipe))
	L.SetField(pkg, "ErrNoProgress", luar.New(L, io.ErrNoProgress))
	L.SetField(pkg, "ErrShortBuffer", luar.New(L, io.ErrShortBuffer))
	L.SetField(pkg, "ErrShortWrite", luar.New(L, io.ErrShortWrite))
	L.SetField(pkg, "ErrUnexpectedEOF", luar.New(L, io.ErrUnexpectedEOF))
	L.SetField(pkg, "LimitReader", luar.New(L, io.LimitReader))
	L.SetField(pkg, "MultiReader", luar.New(L, io.MultiReader))
	L.SetField(pkg, "MultiWriter", luar.New(L, io.MultiWriter))
	L.SetField(pkg, "NewSectionReader", luar.New(L, io.NewSectionReader))
	L.SetField(pkg, "Pipe", luar.New(L, io.Pipe))
	L.SetField(pkg, "ReadAll", luar.New(L, io.ReadAll))
	L.SetField(pkg, "ReadAtLeast", luar.New(L, io.ReadAtLeast))
	L.SetField(pkg, "ReadFull", luar.New(L, io.ReadFull))
	L.SetField(pkg, "TeeReader", luar.New(L, io.TeeReader))
	L.SetField(pkg, "WriteString", luar.New(L, io.WriteString))

	return pkg
}

func importIoUtil() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "ReadAll", luar.New(L, ioutil.ReadAll))
	L.SetField(pkg, "ReadDir", luar.New(L, ioutil.ReadDir))
	L.SetField(pkg, "ReadFile", luar.New(L, ioutil.ReadFile))
	L.SetField(pkg, "WriteFile", luar.New(L, ioutil.WriteFile))

	return pkg
}

func importNet() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "CIDRMask", luar.New(L, net.CIDRMask))
	L.SetField(pkg, "Dial", luar.New(L, net.Dial))
	L.SetField(pkg, "DialIP", luar.New(L, net.DialIP))
	L.SetField(pkg, "DialTCP", luar.New(L, net.DialTCP))
	L.SetField(pkg, "DialTimeout", luar.New(L, net.DialTimeout))
	L.SetField(pkg, "DialUDP", luar.New(L, net.DialUDP))
	L.SetField(pkg, "DialUnix", luar.New(L, net.DialUnix))
	L.SetField(pkg, "ErrWriteToConnected", luar.New(L, net.ErrWriteToConnected))
	L.SetField(pkg, "FileConn", luar.New(L, net.FileConn))
	L.SetField(pkg, "FileListener", luar.New(L, net.FileListener))
	L.SetField(pkg, "FilePacketConn", luar.New(L, net.FilePacketConn))
	L.SetField(pkg, "FlagBroadcast", luar.New(L, net.FlagBroadcast))
	L.SetField(pkg, "FlagLoopback", luar.New(L, net.FlagLoopback))
	L.SetField(pkg, "FlagMulticast", luar.New(L, net.FlagMulticast))
	L.SetField(pkg, "FlagPointToPoint", luar.New(L, net.FlagPointToPoint))
	L.SetField(pkg, "FlagUp", luar.New(L, net.FlagUp))
	L.SetField(pkg, "IPv4", luar.New(L, net.IPv4))
	L.SetField(pkg, "IPv4Mask", luar.New(L, net.IPv4Mask))
	L.SetField(pkg, "IPv4allrouter", luar.New(L, net.IPv4allrouter))
	L.SetField(pkg, "IPv4allsys", luar.New(L, net.IPv4allsys))
	L.SetField(pkg, "IPv4bcast", luar.New(L, net.IPv4bcast))
	L.SetField(pkg, "IPv4len", luar.New(L, net.IPv4len))
	L.SetField(pkg, "IPv4zero", luar.New(L, net.IPv4zero))
	L.SetField(pkg, "IPv6interfacelocalallnodes", luar.New(L, net.IPv6interfacelocalallnodes))
	L.SetField(pkg, "IPv6len", luar.New(L, net.IPv6len))
	L.SetField(pkg, "IPv6linklocalallnodes", luar.New(L, net.IPv6linklocalallnodes))
	L.SetField(pkg, "IPv6linklocalallrouters", luar.New(L, net.IPv6linklocalallrouters))
	L.SetField(pkg, "IPv6loopback", luar.New(L, net.IPv6loopback))
	L.SetField(pkg, "IPv6unspecified", luar.New(L, net.IPv6unspecified))
	L.SetField(pkg, "IPv6zero", luar.New(L, net.IPv6zero))
	L.SetField(pkg, "InterfaceAddrs", luar.New(L, net.InterfaceAddrs))
	L.SetField(pkg, "InterfaceByIndex", luar.New(L, net.InterfaceByIndex))
	L.SetField(pkg, "InterfaceByName", luar.New(L, net.InterfaceByName))
	L.SetField(pkg, "Interfaces", luar.New(L, net.Interfaces))
	L.SetField(pkg, "JoinHostPort", luar.New(L, net.JoinHostPort))
	L.SetField(pkg, "Listen", luar.New(L, net.Listen))
	L.SetField(pkg, "ListenIP", luar.New(L, net.ListenIP))
	L.SetField(pkg, "ListenMulticastUDP", luar.New(L, net.ListenMulticastUDP))
	L.SetField(pkg, "ListenPacket", luar.New(L, net.ListenPacket))
	L.SetField(pkg, "ListenTCP", luar.New(L, net.ListenTCP))
	L.SetField(pkg, "ListenUDP", luar.New(L, net.ListenUDP))
	L.SetField(pkg, "ListenUnix", luar.New(L, net.ListenUnix))
	L.SetField(pkg, "ListenUnixgram", luar.New(L, net.ListenUnixgram))
	L.SetField(pkg, "LookupAddr", luar.New(L, net.LookupAddr))
	L.SetField(pkg, "LookupCNAME", luar.New(L, net.LookupCNAME))
	L.SetField(pkg, "LookupHost", luar.New(L, net.LookupHost))
	L.SetField(pkg, "LookupIP", luar.New(L, net.LookupIP))
	L.SetField(pkg, "LookupMX", luar.New(L, net.LookupMX))
	L.SetField(pkg, "LookupNS", luar.New(L, net.LookupNS))
	L.SetField(pkg, "LookupPort", luar.New(L, net.LookupPort))
	L.SetField(pkg, "LookupSRV", luar.New(L, net.LookupSRV))
	L.SetField(pkg, "LookupTXT", luar.New(L, net.LookupTXT))
	L.SetField(pkg, "ParseCIDR", luar.New(L, net.ParseCIDR))
	L.SetField(pkg, "ParseIP", luar.New(L, net.ParseIP))
	L.SetField(pkg, "ParseMAC", luar.New(L, net.ParseMAC))
	L.SetField(pkg, "Pipe", luar.New(L, net.Pipe))
	L.SetField(pkg, "ResolveIPAddr", luar.New(L, net.ResolveIPAddr))
	L.SetField(pkg, "ResolveTCPAddr", luar.New(L, net.ResolveTCPAddr))
	L.SetField(pkg, "ResolveUDPAddr", luar.New(L, net.ResolveUDPAddr))
	L.SetField(pkg, "ResolveUnixAddr", luar.New(L, net.ResolveUnixAddr))
	L.SetField(pkg, "SplitHostPort", luar.New(L, net.SplitHostPort))

	return pkg
}

func importMath() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Abs", luar.New(L, math.Abs))
	L.SetField(pkg, "Acos", luar.New(L, math.Acos))
	L.SetField(pkg, "Acosh", luar.New(L, math.Acosh))
	L.SetField(pkg, "Asin", luar.New(L, math.Asin))
	L.SetField(pkg, "Asinh", luar.New(L, math.Asinh))
	L.SetField(pkg, "Atan", luar.New(L, math.Atan))
	L.SetField(pkg, "Atan2", luar.New(L, math.Atan2))
	L.SetField(pkg, "Atanh", luar.New(L, math.Atanh))
	L.SetField(pkg, "Cbrt", luar.New(L, math.Cbrt))
	L.SetField(pkg, "Ceil", luar.New(L, math.Ceil))
	L.SetField(pkg, "Copysign", luar.New(L, math.Copysign))
	L.SetField(pkg, "Cos", luar.New(L, math.Cos))
	L.SetField(pkg, "Cosh", luar.New(L, math.Cosh))
	L.SetField(pkg, "Dim", luar.New(L, math.Dim))
	L.SetField(pkg, "Erf", luar.New(L, math.Erf))
	L.SetField(pkg, "Erfc", luar.New(L, math.Erfc))
	L.SetField(pkg, "Exp", luar.New(L, math.Exp))
	L.SetField(pkg, "Exp2", luar.New(L, math.Exp2))
	L.SetField(pkg, "Expm1", luar.New(L, math.Expm1))
	L.SetField(pkg, "Float32bits", luar.New(L, math.Float32bits))
	L.SetField(pkg, "Float32frombits", luar.New(L, math.Float32frombits))
	L.SetField(pkg, "Float64bits", luar.New(L, math.Float64bits))
	L.SetField(pkg, "Float64frombits", luar.New(L, math.Float64frombits))
	L.SetField(pkg, "Floor", luar.New(L, math.Floor))
	L.SetField(pkg, "Frexp", luar.New(L, math.Frexp))
	L.SetField(pkg, "Gamma", luar.New(L, math.Gamma))
	L.SetField(pkg, "Hypot", luar.New(L, math.Hypot))
	L.SetField(pkg, "Ilogb", luar.New(L, math.Ilogb))
	L.SetField(pkg, "Inf", luar.New(L, math.Inf))
	L.SetField(pkg, "IsInf", luar.New(L, math.IsInf))
	L.SetField(pkg, "IsNaN", luar.New(L, math.IsNaN))
	L.SetField(pkg, "J0", luar.New(L, math.J0))
	L.SetField(pkg, "J1", luar.New(L, math.J1))
	L.SetField(pkg, "Jn", luar.New(L, math.Jn))
	L.SetField(pkg, "Ldexp", luar.New(L, math.Ldexp))
	L.SetField(pkg, "Lgamma", luar.New(L, math.Lgamma))
	L.SetField(pkg, "Log", luar.New(L, math.Log))
	L.SetField(pkg, "Log10", luar.New(L, math.Log10))
	L.SetField(pkg, "Log1p", luar.New(L, math.Log1p))
	L.SetField(pkg, "Log2", luar.New(L, math.Log2))
	L.SetField(pkg, "Logb", luar.New(L, math.Logb))
	L.SetField(pkg, "Max", luar.New(L, math.Max))
	L.SetField(pkg, "Min", luar.New(L, math.Min))
	L.SetField(pkg, "Mod", luar.New(L, math.Mod))
	L.SetField(pkg, "Modf", luar.New(L, math.Modf))
	L.SetField(pkg, "NaN", luar.New(L, math.NaN))
	L.SetField(pkg, "Nextafter", luar.New(L, math.Nextafter))
	L.SetField(pkg, "Pow", luar.New(L, math.Pow))
	L.SetField(pkg, "Pow10", luar.New(L, math.Pow10))
	L.SetField(pkg, "Remainder", luar.New(L, math.Remainder))
	L.SetField(pkg, "Signbit", luar.New(L, math.Signbit))
	L.SetField(pkg, "Sin", luar.New(L, math.Sin))
	L.SetField(pkg, "Sincos", luar.New(L, math.Sincos))
	L.SetField(pkg, "Sinh", luar.New(L, math.Sinh))
	L.SetField(pkg, "Sqrt", luar.New(L, math.Sqrt))
	L.SetField(pkg, "Tan", luar.New(L, math.Tan))
	L.SetField(pkg, "Tanh", luar.New(L, math.Tanh))
	L.SetField(pkg, "Trunc", luar.New(L, math.Trunc))
	L.SetField(pkg, "Y0", luar.New(L, math.Y0))
	L.SetField(pkg, "Y1", luar.New(L, math.Y1))
	L.SetField(pkg, "Yn", luar.New(L, math.Yn))

	return pkg
}

func importMathRand() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "ExpFloat64", luar.New(L, rand.ExpFloat64))
	L.SetField(pkg, "Float32", luar.New(L, rand.Float32))
	L.SetField(pkg, "Float64", luar.New(L, rand.Float64))
	L.SetField(pkg, "Int", luar.New(L, rand.Int))
	L.SetField(pkg, "Int31", luar.New(L, rand.Int31))
	L.SetField(pkg, "Int31n", luar.New(L, rand.Int31n))
	L.SetField(pkg, "Int63", luar.New(L, rand.Int63))
	L.SetField(pkg, "Int63n", luar.New(L, rand.Int63n))
	L.SetField(pkg, "Intn", luar.New(L, rand.Intn))
	L.SetField(pkg, "NormFloat64", luar.New(L, rand.NormFloat64))
	L.SetField(pkg, "Perm", luar.New(L, rand.Perm))
	L.SetField(pkg, "Seed", luar.New(L, rand.Seed))
	L.SetField(pkg, "Uint32", luar.New(L, rand.Uint32))

	return pkg
}

func importOs() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Args", luar.New(L, os.Args))
	L.SetField(pkg, "Chdir", luar.New(L, os.Chdir))
	L.SetField(pkg, "Chmod", luar.New(L, os.Chmod))
	L.SetField(pkg, "Chown", luar.New(L, os.Chown))
	L.SetField(pkg, "Chtimes", luar.New(L, os.Chtimes))
	L.SetField(pkg, "Clearenv", luar.New(L, os.Clearenv))
	L.SetField(pkg, "Create", luar.New(L, os.Create))
	L.SetField(pkg, "DevNull", luar.New(L, os.DevNull))
	L.SetField(pkg, "Environ", luar.New(L, os.Environ))
	L.SetField(pkg, "ErrExist", luar.New(L, os.ErrExist))
	L.SetField(pkg, "ErrInvalid", luar.New(L, os.ErrInvalid))
	L.SetField(pkg, "ErrNotExist", luar.New(L, os.ErrNotExist))
	L.SetField(pkg, "ErrPermission", luar.New(L, os.ErrPermission))
	L.SetField(pkg, "Exit", luar.New(L, os.Exit))
	L.SetField(pkg, "Expand", luar.New(L, os.Expand))
	L.SetField(pkg, "ExpandEnv", luar.New(L, os.ExpandEnv))
	L.SetField(pkg, "FindProcess", luar.New(L, os.FindProcess))
	L.SetField(pkg, "Getegid", luar.New(L, os.Getegid))
	L.SetField(pkg, "Getenv", luar.New(L, os.Getenv))
	L.SetField(pkg, "Geteuid", luar.New(L, os.Geteuid))
	L.SetField(pkg, "Getgid", luar.New(L, os.Getgid))
	L.SetField(pkg, "Getgroups", luar.New(L, os.Getgroups))
	L.SetField(pkg, "Getpagesize", luar.New(L, os.Getpagesize))
	L.SetField(pkg, "Getpid", luar.New(L, os.Getpid))
	L.SetField(pkg, "Getuid", luar.New(L, os.Getuid))
	L.SetField(pkg, "Getwd", luar.New(L, os.Getwd))
	L.SetField(pkg, "Hostname", luar.New(L, os.Hostname))
	L.SetField(pkg, "Interrupt", luar.New(L, os.Interrupt))
	L.SetField(pkg, "IsExist", luar.New(L, os.IsExist))
	L.SetField(pkg, "IsNotExist", luar.New(L, os.IsNotExist))
	L.SetField(pkg, "IsPathSeparator", luar.New(L, os.IsPathSeparator))
	L.SetField(pkg, "IsPermission", luar.New(L, os.IsPermission))
	L.SetField(pkg, "Kill", luar.New(L, os.Kill))
	L.SetField(pkg, "Lchown", luar.New(L, os.Lchown))
	L.SetField(pkg, "Link", luar.New(L, os.Link))
	L.SetField(pkg, "Lstat", luar.New(L, os.Lstat))
	L.SetField(pkg, "Mkdir", luar.New(L, os.Mkdir))
	L.SetField(pkg, "MkdirAll", luar.New(L, os.MkdirAll))
	L.SetField(pkg, "ModeAppend", luar.New(L, os.ModeAppend))
	L.SetField(pkg, "ModeCharDevice", luar.New(L, os.ModeCharDevice))
	L.SetField(pkg, "ModeDevice", luar.New(L, os.ModeDevice))
	L.SetField(pkg, "ModeDir", luar.New(L, os.ModeDir))
	L.SetField(pkg, "ModeExclusive", luar.New(L, os.ModeExclusive))
	L.SetField(pkg, "ModeNamedPipe", luar.New(L, os.ModeNamedPipe))
	L.SetField(pkg, "ModePerm", luar.New(L, os.ModePerm))
	L.SetField(pkg, "ModeSetgid", luar.New(L, os.ModeSetgid))
	L.SetField(pkg, "ModeSetuid", luar.New(L, os.ModeSetuid))
	L.SetField(pkg, "ModeSocket", luar.New(L, os.ModeSocket))
	L.SetField(pkg, "ModeSticky", luar.New(L, os.ModeSticky))
	L.SetField(pkg, "ModeSymlink", luar.New(L, os.ModeSymlink))
	L.SetField(pkg, "ModeTemporary", luar.New(L, os.ModeTemporary))
	L.SetField(pkg, "ModeType", luar.New(L, os.ModeType))
	L.SetField(pkg, "NewFile", luar.New(L, os.NewFile))
	L.SetField(pkg, "NewSyscallError", luar.New(L, os.NewSyscallError))
	L.SetField(pkg, "O_APPEND", luar.New(L, os.O_APPEND))
	L.SetField(pkg, "O_CREATE", luar.New(L, os.O_CREATE))
	L.SetField(pkg, "O_EXCL", luar.New(L, os.O_EXCL))
	L.SetField(pkg, "O_RDONLY", luar.New(L, os.O_RDONLY))
	L.SetField(pkg, "O_RDWR", luar.New(L, os.O_RDWR))
	L.SetField(pkg, "O_SYNC", luar.New(L, os.O_SYNC))
	L.SetField(pkg, "O_TRUNC", luar.New(L, os.O_TRUNC))
	L.SetField(pkg, "O_WRONLY", luar.New(L, os.O_WRONLY))
	L.SetField(pkg, "Open", luar.New(L, os.Open))
	L.SetField(pkg, "OpenFile", luar.New(L, os.OpenFile))
	L.SetField(pkg, "PathListSeparator", luar.New(L, os.PathListSeparator))
	L.SetField(pkg, "PathSeparator", luar.New(L, os.PathSeparator))
	L.SetField(pkg, "Pipe", luar.New(L, os.Pipe))
	L.SetField(pkg, "ReadDir", luar.New(L, os.ReadDir))
	L.SetField(pkg, "ReadFile", luar.New(L, os.ReadFile))
	L.SetField(pkg, "Readlink", luar.New(L, os.Readlink))
	L.SetField(pkg, "Remove", luar.New(L, os.Remove))
	L.SetField(pkg, "RemoveAll", luar.New(L, os.RemoveAll))
	L.SetField(pkg, "Rename", luar.New(L, os.Rename))
	L.SetField(pkg, "SEEK_CUR", luar.New(L, io.SeekCurrent))
	L.SetField(pkg, "SEEK_END", luar.New(L, io.SeekEnd))
	L.SetField(pkg, "SEEK_SET", luar.New(L, io.SeekStart))
	L.SetField(pkg, "SameFile", luar.New(L, os.SameFile))
	L.SetField(pkg, "Setenv", luar.New(L, os.Setenv))
	L.SetField(pkg, "StartProcess", luar.New(L, os.StartProcess))
	L.SetField(pkg, "Stat", luar.New(L, os.Stat))
	L.SetField(pkg, "Stderr", luar.New(L, os.Stderr))
	L.SetField(pkg, "Stdin", luar.New(L, os.Stdin))
	L.SetField(pkg, "Stdout", luar.New(L, os.Stdout))
	L.SetField(pkg, "Symlink", luar.New(L, os.Symlink))
	L.SetField(pkg, "TempDir", luar.New(L, os.TempDir))
	L.SetField(pkg, "Truncate", luar.New(L, os.Truncate))
	L.SetField(pkg, "UserHomeDir", luar.New(L, os.UserHomeDir))
	L.SetField(pkg, "WriteFile", luar.New(L, os.WriteFile))

	return pkg
}

func importRuntime() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "GC", luar.New(L, runtime.GC))
	L.SetField(pkg, "GOARCH", luar.New(L, runtime.GOARCH))
	L.SetField(pkg, "GOMAXPROCS", luar.New(L, runtime.GOMAXPROCS))
	L.SetField(pkg, "GOOS", luar.New(L, runtime.GOOS))
	L.SetField(pkg, "GOROOT", luar.New(L, runtime.GOROOT))

	return pkg
}

func importPath() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Base", luar.New(L, path.Base))
	L.SetField(pkg, "Clean", luar.New(L, path.Clean))
	L.SetField(pkg, "Dir", luar.New(L, path.Dir))
	L.SetField(pkg, "ErrBadPattern", luar.New(L, path.ErrBadPattern))
	L.SetField(pkg, "Ext", luar.New(L, path.Ext))
	L.SetField(pkg, "IsAbs", luar.New(L, path.IsAbs))
	L.SetField(pkg, "Join", luar.New(L, path.Join))
	L.SetField(pkg, "Match", luar.New(L, path.Match))
	L.SetField(pkg, "Split", luar.New(L, path.Split))

	return pkg
}

func importFilePath() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Abs", luar.New(L, filepath.Abs))
	L.SetField(pkg, "Base", luar.New(L, filepath.Base))
	L.SetField(pkg, "Clean", luar.New(L, filepath.Clean))
	L.SetField(pkg, "Dir", luar.New(L, filepath.Dir))
	L.SetField(pkg, "EvalSymlinks", luar.New(L, filepath.EvalSymlinks))
	L.SetField(pkg, "Ext", luar.New(L, filepath.Ext))
	L.SetField(pkg, "FromSlash", luar.New(L, filepath.FromSlash))
	L.SetField(pkg, "Glob", luar.New(L, filepath.Glob))
	L.SetField(pkg, "HasPrefix", luar.New(L, filepath.HasPrefix))
	L.SetField(pkg, "IsAbs", luar.New(L, filepath.IsAbs))
	L.SetField(pkg, "Join", luar.New(L, filepath.Join))
	L.SetField(pkg, "Match", luar.New(L, filepath.Match))
	L.SetField(pkg, "Rel", luar.New(L, filepath.Rel))
	L.SetField(pkg, "Split", luar.New(L, filepath.Split))
	L.SetField(pkg, "SplitList", luar.New(L, filepath.SplitList))
	L.SetField(pkg, "ToSlash", luar.New(L, filepath.ToSlash))
	L.SetField(pkg, "VolumeName", luar.New(L, filepath.VolumeName))

	return pkg
}

func importStrings() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Contains", luar.New(L, strings.Contains))
	L.SetField(pkg, "ContainsAny", luar.New(L, strings.ContainsAny))
	L.SetField(pkg, "ContainsRune", luar.New(L, strings.ContainsRune))
	L.SetField(pkg, "Count", luar.New(L, strings.Count))
	L.SetField(pkg, "EqualFold", luar.New(L, strings.EqualFold))
	L.SetField(pkg, "Fields", luar.New(L, strings.Fields))
	L.SetField(pkg, "FieldsFunc", luar.New(L, strings.FieldsFunc))
	L.SetField(pkg, "HasPrefix", luar.New(L, strings.HasPrefix))
	L.SetField(pkg, "HasSuffix", luar.New(L, strings.HasSuffix))
	L.SetField(pkg, "Index", luar.New(L, strings.Index))
	L.SetField(pkg, "IndexAny", luar.New(L, strings.IndexAny))
	L.SetField(pkg, "IndexByte", luar.New(L, strings.IndexByte))
	L.SetField(pkg, "IndexFunc", luar.New(L, strings.IndexFunc))
	L.SetField(pkg, "IndexRune", luar.New(L, strings.IndexRune))
	L.SetField(pkg, "Join", luar.New(L, strings.Join))
	L.SetField(pkg, "LastIndex", luar.New(L, strings.LastIndex))
	L.SetField(pkg, "LastIndexAny", luar.New(L, strings.LastIndexAny))
	L.SetField(pkg, "LastIndexFunc", luar.New(L, strings.LastIndexFunc))
	L.SetField(pkg, "Map", luar.New(L, strings.Map))
	L.SetField(pkg, "NewReader", luar.New(L, strings.NewReader))
	L.SetField(pkg, "NewReplacer", luar.New(L, strings.NewReplacer))
	L.SetField(pkg, "Repeat", luar.New(L, strings.Repeat))
	L.SetField(pkg, "Replace", luar.New(L, strings.Replace))
	L.SetField(pkg, "Split", luar.New(L, strings.Split))
	L.SetField(pkg, "SplitAfter", luar.New(L, strings.SplitAfter))
	L.SetField(pkg, "SplitAfterN", luar.New(L, strings.SplitAfterN))
	L.SetField(pkg, "SplitN", luar.New(L, strings.SplitN))
	L.SetField(pkg, "Title", luar.New(L, strings.Title))
	L.SetField(pkg, "ToLower", luar.New(L, strings.ToLower))
	L.SetField(pkg, "ToLowerSpecial", luar.New(L, strings.ToLowerSpecial))
	L.SetField(pkg, "ToTitle", luar.New(L, strings.ToTitle))
	L.SetField(pkg, "ToTitleSpecial", luar.New(L, strings.ToTitleSpecial))
	L.SetField(pkg, "ToUpper", luar.New(L, strings.ToUpper))
	L.SetField(pkg, "ToUpperSpecial", luar.New(L, strings.ToUpperSpecial))
	L.SetField(pkg, "Trim", luar.New(L, strings.Trim))
	L.SetField(pkg, "TrimFunc", luar.New(L, strings.TrimFunc))
	L.SetField(pkg, "TrimLeft", luar.New(L, strings.TrimLeft))
	L.SetField(pkg, "TrimLeftFunc", luar.New(L, strings.TrimLeftFunc))
	L.SetField(pkg, "TrimPrefix", luar.New(L, strings.TrimPrefix))
	L.SetField(pkg, "TrimRight", luar.New(L, strings.TrimRight))
	L.SetField(pkg, "TrimRightFunc", luar.New(L, strings.TrimRightFunc))
	L.SetField(pkg, "TrimSpace", luar.New(L, strings.TrimSpace))
	L.SetField(pkg, "TrimSuffix", luar.New(L, strings.TrimSuffix))

	return pkg
}

func importRegexp() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Match", luar.New(L, regexp.Match))
	L.SetField(pkg, "MatchReader", luar.New(L, regexp.MatchReader))
	L.SetField(pkg, "MatchString", luar.New(L, regexp.MatchString))
	L.SetField(pkg, "QuoteMeta", luar.New(L, regexp.QuoteMeta))
	L.SetField(pkg, "Compile", luar.New(L, regexp.Compile))
	L.SetField(pkg, "CompilePOSIX", luar.New(L, regexp.CompilePOSIX))
	L.SetField(pkg, "MustCompile", luar.New(L, regexp.MustCompile))
	L.SetField(pkg, "MustCompilePOSIX", luar.New(L, regexp.MustCompilePOSIX))

	return pkg
}

func importErrors() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "New", luar.New(L, errors.New))

	return pkg
}

func importTime() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Sleep", luar.New(L, time.Sleep))
	L.SetField(pkg, "Since", luar.New(L, time.Since))
	L.SetField(pkg, "FixedZone", luar.New(L, time.FixedZone))
	L.SetField(pkg, "LoadLocation", luar.New(L, time.LoadLocation))
	L.SetField(pkg, "Date", luar.New(L, time.Date))
	L.SetField(pkg, "Now", luar.New(L, time.Now))
	L.SetField(pkg, "Parse", luar.New(L, time.Parse))
	L.SetField(pkg, "ParseDuration", luar.New(L, time.ParseDuration))
	L.SetField(pkg, "ParseInLocation", luar.New(L, time.ParseInLocation))
	L.SetField(pkg, "Unix", luar.New(L, time.Unix))
	L.SetField(pkg, "Nanosecond", luar.New(L, time.Nanosecond))
	L.SetField(pkg, "Microsecond", luar.New(L, time.Microsecond))
	L.SetField(pkg, "Millisecond", luar.New(L, time.Millisecond))
	L.SetField(pkg, "Second", luar.New(L, time.Second))
	L.SetField(pkg, "Minute", luar.New(L, time.Minute))
	L.SetField(pkg, "Hour", luar.New(L, time.Hour))

	// TODO: these raw Go timer APIs don't provide any synchronization
	// with micro. Stop exposing them to lua once plugins switch to using
	// the safer micro.After() interface instead. See issue #3209
	L.SetField(pkg, "After", luar.New(L, time.After))
	L.SetField(pkg, "AfterFunc", luar.New(L, time.AfterFunc))
	L.SetField(pkg, "NewTicker", luar.New(L, time.NewTicker))
	L.SetField(pkg, "NewTimer", luar.New(L, time.NewTimer))
	L.SetField(pkg, "Tick", luar.New(L, time.Tick))

	return pkg
}

func importUtf8() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "DecodeLastRune", luar.New(L, utf8.DecodeLastRune))
	L.SetField(pkg, "DecodeLastRuneInString", luar.New(L, utf8.DecodeLastRuneInString))
	L.SetField(pkg, "DecodeRune", luar.New(L, utf8.DecodeRune))
	L.SetField(pkg, "DecodeRuneInString", luar.New(L, utf8.DecodeRuneInString))
	L.SetField(pkg, "EncodeRune", luar.New(L, utf8.EncodeRune))
	L.SetField(pkg, "FullRune", luar.New(L, utf8.FullRune))
	L.SetField(pkg, "FullRuneInString", luar.New(L, utf8.FullRuneInString))
	L.SetField(pkg, "RuneCount", luar.New(L, utf8.RuneCount))
	L.SetField(pkg, "RuneCountInString", luar.New(L, utf8.RuneCountInString))
	L.SetField(pkg, "RuneLen", luar.New(L, utf8.RuneLen))
	L.SetField(pkg, "RuneStart", luar.New(L, utf8.RuneStart))
	L.SetField(pkg, "Valid", luar.New(L, utf8.Valid))
	L.SetField(pkg, "ValidRune", luar.New(L, utf8.ValidRune))
	L.SetField(pkg, "ValidString", luar.New(L, utf8.ValidString))

	return pkg
}

func importHumanize() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Bytes", luar.New(L, humanize.Bytes))
	L.SetField(pkg, "Ordinal", luar.New(L, humanize.Ordinal))

	return pkg
}

func importHTTP() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "Get", luar.New(L, http.Get))
	L.SetField(pkg, "Post", luar.New(L, http.Post))

	return pkg
}

func importArchiveZip() *lua.LTable {
	pkg := L.NewTable()

	L.SetField(pkg, "OpenReader", luar.New(L, zip.OpenReader))
	L.SetField(pkg, "NewReader", luar.New(L, zip.NewReader))
	L.SetField(pkg, "NewWriter", luar.New(L, zip.NewWriter))

	return pkg
}
