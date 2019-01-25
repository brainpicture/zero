package zero

import (
	"net"
	"strings"
)

type langString struct {
	str string
}

type langPack map[string]langString

var langPacks = map[string]langPack{}

const (
	// EnvPlatformNone platform not set yet
	EnvPlatformNone int = iota
	// EnvPlatformIOS is iphone device or ipad
	EnvPlatformIOS
	// EnvPlatformAndroid means its android
	EnvPlatformAndroid
	// EnvPlatformBot is for bots or index bot
	EnvPlatformBot
	// EnvPlatformTest means test sute
	EnvPlatformTest
)

// AddLangPack will set langpack
func AddLangPack(langKey string, langPack H) {
	pack, ok := langPacks[langKey]
	if !ok {
		pack = map[string]langString{}
		langPacks[langKey] = pack
		//panic("lang " + langKey + " not supported")
	}

	for key, lang := range langPack {
		langStr, ok := lang.(string)
		if !ok {
			panic("lang " + langKey + " key " + key + " has unsupported type, try string")
		}
		langObj := langString{
			str: langStr,
		}
		pack[key] = langObj
	}
}

// Environment defines user params like language and platform
type Environment struct {
	srv        *Server
	Language   string
	Version    int
	Platform   int
	AppVersion int
	Build      int
}

// Environment format in useragent
// {Api version} {platform}/{version}/{build} {language} {device}
// Example: 1 Android/3/23 ru nexus_6_345

// Env returns user environment
func Env(srv *Server) *Environment {
	// take care, this function should never fire any panic
	env := Environment{
		srv: srv,
	}
	ua := srv.GetUserAgent()
	parts := strings.Split(ua, " ")
	if len(parts) < 3 {
		return &env
	}
	env.Version = I(parts[0])
	platfromParts := strings.Split(parts[1], "/")
	lenParts := len(platfromParts)
	if lenParts > 1 {
		env.SetPlatform(platfromParts[0])
		if lenParts > 2 {
			env.AppVersion = I(platfromParts[1])
			env.Build = I(platfromParts[2])
		} else {
			env.Build = I(platfromParts[1])
		}
	}
	langStr := parts[2]
	if langStr == "" {
		langStr = "en"
	}

	env.Language = langStr
	return &env
}

// SetPlatform set platform from string
func (e *Environment) SetPlatform(platform string) {
	switch strings.ToLower(platform) {
	case "ios":
		e.Platform = EnvPlatformIOS
	case "android":
		e.Platform = EnvPlatformAndroid
	case "bot":
		e.Platform = EnvPlatformBot
	}
}

// PlatformString return string for platform
func (e *Environment) PlatformString() string {
	switch e.Platform {
	case EnvPlatformIOS:
		return "ios"
	case EnvPlatformAndroid:
		return "android"
	case EnvPlatformBot:
		return "bot"
	}
	return "unknown"
}

// Ver checks enviroment version
func (e *Environment) Ver(a int) bool {
	return e.Version >= a
}

// IsIOS checks if platform is iOS
func (e *Environment) IsIOS() bool {
	return e.Platform == 1
}

// IsAndroid checks if platform is Android
func (e *Environment) IsAndroid() bool {
	return e.Platform == 2
}

// Plural will return num with proper word
func (e *Environment) Plural(num int64, single, multiple string) string {
	if num > 1 || num < -1 {
		return J(num, " ", multiple)
	}
	return J(num, " ", single)
}

// Lang return langpacked string
func (e *Environment) Lang(name string) string {
	langPack, ok := langPacks[e.Language]
	if !ok {
		langPack, ok = langPacks["en"]
		if !ok {
			return "undefined"
		}
	}
	langObj, ok := langPack[name]
	if !ok {
		return name
	}
	return langObj.str
}

// LangToInt will convert language to int
func (e *Environment) LangToInt() int {
	switch e.Language {
	case "ru":
		return 1
	default: // en
		return 0
	}
}

// IP will return server IP
func (e *Environment) IP() net.IP {
	if e.srv == nil {
		return net.ParseIP("127.0.0.1")
	}
	ipStr := e.srv.GetHeader("X-Real-IP")
	return net.ParseIP(ipStr)
}
