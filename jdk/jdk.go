package jdk

import (
	"io"
	"path/filepath"
	"os"
	"regexp"
	"strconv"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"io/ioutil"

	"github.com/heroku/java-buildpack/util"
	"github.com/buildpack/libbuildpack"
)

type Installer struct {
	In       []byte
	Out, Err io.Writer
	Version  Version
}

type Version struct {
	Major  int
	Tag    string
	Vendor string
}

const (
	DefaultJdkMajorVersion = 8
	DefaultVendor          = "openjdk"
	DefaultJdkBaseUrl      = "https://lang-jvm.s3.amazonaws.com/jdk"
)

var (
	DefaultVersionStrings = map[int]string{
		8:  "1.8.0_181",
		9:  "9.0.1",
		10: "10.0.2",
		11: "11.0.1",
	}
)

func (i *Installer) Init(appDir string) (error) {
	v, err := i.detectVersion(appDir)
	if err != nil {
		return err
	}

	i.Version = v

	return nil
}

func (i *Installer) Install(appDir string, cache libbuildpack.Cache, launchDir libbuildpack.Launch) (error) {
	i.Init(appDir)
	// check the build plan to see if another JDK has already been installed?

	jdkUrl, err := GetVersionUrl(i.Version)
	if err != nil {
		return err
	}

	if !IsValidJdkUrl(jdkUrl) {
		return errors.New("Invalid JDK version")
	}

	jdkLayer := launchDir.Layer("jdk")

	cmd := exec.Command("jdk-fetcher", jdkUrl, jdkLayer.Root)
	cmd.Env = os.Environ()
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// install cacerts
	// create profile.d
	// install pgconfig
	// install metrics agent

	// apply the overlay

	return nil
}

func (i *Installer) detectVersion(appDir string) (Version, error) {
	systemPropertiesFile := filepath.Join(appDir, "system.properties")
	if _, err := os.Stat(systemPropertiesFile); !os.IsNotExist(err) {
		sysProps, err := util.ReadPropertiesFile(systemPropertiesFile)
		if err != nil {
			return defaultVersion(), nil
		}

		if version, ok := sysProps["java.runtime.version"]; ok {
			return ParseVersionString(version)
		}
	}
	return defaultVersion(), nil
}

func defaultVersion() Version {
	version, _ := ParseVersionString(DefaultVersionStrings[DefaultJdkMajorVersion])
	return version
}

func ParseVersionString(v string) (Version, error) {
	if v == "10" || v == "11" {
		major, _ := strconv.Atoi(v)
		return ParseVersionString(DefaultVersionStrings[major])
	} else if m := regexp.MustCompile("^(1[0-1])\\.").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^1\\.([7-9])$").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    DefaultVersionStrings[major],
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^([7-9])$").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    DefaultVersionStrings[major],
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^1\\.([7-9])").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  major,
		}, nil
	}

	return Version{}, errors.New("unparseable version string")
}

func GetVersionUrl(v Version) (string, error) {
	baseUrl := DefaultJdkBaseUrl
	if customBaseUrl, ok := os.LookupEnv("DEFAULT_JDK_BASE_URL"); ok {
		baseUrl = customBaseUrl
	}

	stack, ok := os.LookupEnv("STACK")
	if !ok {
		return "", errors.New("missing stack")
	}

	return fmt.Sprintf("%s/%s/%s%s.tar.gz", baseUrl, stack, v.Vendor, v.Tag), nil
}

func IsValidJdkUrl(url string) bool {
	res, err := http.Head(url)
	if err != nil {
		return false
	}
	return res.StatusCode < 300
}