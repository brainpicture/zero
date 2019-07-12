package zero

import "strings"

// LangObj allow to work with localisation entry
type LangObj struct {
	str   string
	form2 string
	form3 string
}

// LangUndefined is an undefined lang entity
var LangUndefined = LangObj{str: "undefined"}

// LangPack is basic object to work with langpack
type LangPack map[string]LangObj

var langPacks = map[string]LangPack{}

// Lang return langpack
func Lang(language string) LangPack {
	langPack, ok := langPacks[language]
	if !ok {
		langPack, ok = langPacks["en"]
		if !ok {
			panic("zero: Lang: en langpack not set")
		}
	}
	return langPack
}

// Get fetch lang entry from langpack
func (lp LangPack) Get(name string) LangObj {
	langObj, ok := lp[name]
	if !ok {
		langObj.str = "undefined"
	}
	return langObj
}

// Plural select form of words passed based on number
func (lo *LangObj) Plural(num int64) {
	if num > 1 || num < -1 {
		lo.str = lo.form2
	}
}

// Format will format data of Lang Obj
func (lo *LangObj) Format(data S) string {
	str := lo.str
	for k, v := range data {
		str = strings.Replace(str, "$"+k, v, -1)
	}
	return str
}

// AddLangPack will set langpack
func AddLangPack(langKey string, langPack H) {
	pack, ok := langPacks[langKey]
	if !ok {
		pack = map[string]LangObj{}
		langPacks[langKey] = pack
		//panic("lang " + langKey + " not supported")
	}

	for key, lang := range langPack {
		langObj := LangObj{}
		switch langType := lang.(type) {
		case string:
			langObj.str = langType
		case []string: // vector
			if len(langType) > 0 {
				langObj.str = langType[0]
			}
			if len(langType) > 1 {
				langObj.form2 = langType[1]
			}
			if len(langType) > 2 {
				langObj.form3 = langType[2]
			}
		default:
			panic("lang " + langKey + " key " + key + " has unsupported type, try string")
		}
		pack[key] = langObj
	}
}
