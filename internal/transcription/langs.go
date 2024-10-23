package transcription

import (
	"fmt"
)

// Basically a dump of
// https://github.com/openai/whisper/blob/25639fc17ddc013d56c594bfbf7644f2185fad84/whisper/tokenizer.py#L226

// STTLang represents supported language codes.
type STTLang string

const (
	STTLangEN  STTLang = "en"
	STTLangZH  STTLang = "zh"
	STTLangDE  STTLang = "de"
	STTLangES  STTLang = "es"
	STTLangRU  STTLang = "ru"
	STTLangKO  STTLang = "ko"
	STTLangFR  STTLang = "fr"
	STTLangJA  STTLang = "ja"
	STTLangPT  STTLang = "pt"
	STTLangTR  STTLang = "tr"
	STTLangPL  STTLang = "pl"
	STTLangCA  STTLang = "ca"
	STTLangNL  STTLang = "nl"
	STTLangAR  STTLang = "ar"
	STTLangSV  STTLang = "sv"
	STTLangIT  STTLang = "it"
	STTLangID  STTLang = "id"
	STTLangHI  STTLang = "hi"
	STTLangFI  STTLang = "fi"
	STTLangVI  STTLang = "vi"
	STTLangHE  STTLang = "he"
	STTLangUK  STTLang = "uk"
	STTLangEL  STTLang = "el"
	STTLangMS  STTLang = "ms"
	STTLangCS  STTLang = "cs"
	STTLangRO  STTLang = "ro"
	STTLangDA  STTLang = "da"
	STTLangHU  STTLang = "hu"
	STTLangTA  STTLang = "ta"
	STTLangNO  STTLang = "no"
	STTLangTH  STTLang = "th"
	STTLangUR  STTLang = "ur"
	STTLangHR  STTLang = "hr"
	STTLangBG  STTLang = "bg"
	STTLangLT  STTLang = "lt"
	STTLangLA  STTLang = "la"
	STTLangMI  STTLang = "mi"
	STTLangML  STTLang = "ml"
	STTLangCY  STTLang = "cy"
	STTLangSK  STTLang = "sk"
	STTLangTE  STTLang = "te"
	STTLangFA  STTLang = "fa"
	STTLangLV  STTLang = "lv"
	STTLangBN  STTLang = "bn"
	STTLangSR  STTLang = "sr"
	STTLangAZ  STTLang = "az"
	STTLangSL  STTLang = "sl"
	STTLangKN  STTLang = "kn"
	STTLangET  STTLang = "et"
	STTLangMK  STTLang = "mk"
	STTLangBR  STTLang = "br"
	STTLangEU  STTLang = "eu"
	STTLangIS  STTLang = "is"
	STTLangHY  STTLang = "hy"
	STTLangNE  STTLang = "ne"
	STTLangMN  STTLang = "mn"
	STTLangBS  STTLang = "bs"
	STTLangKK  STTLang = "kk"
	STTLangSQ  STTLang = "sq"
	STTLangSW  STTLang = "sw"
	STTLangGL  STTLang = "gl"
	STTLangMR  STTLang = "mr"
	STTLangPA  STTLang = "pa"
	STTLangSI  STTLang = "si"
	STTLangKM  STTLang = "km"
	STTLangSN  STTLang = "sn"
	STTLangYO  STTLang = "yo"
	STTLangSO  STTLang = "so"
	STTLangAF  STTLang = "af"
	STTLangOC  STTLang = "oc"
	STTLangKA  STTLang = "ka"
	STTLangBE  STTLang = "be"
	STTLangTG  STTLang = "tg"
	STTLangSD  STTLang = "sd"
	STTLangGU  STTLang = "gu"
	STTLangAM  STTLang = "am"
	STTLangYI  STTLang = "yi"
	STTLangLO  STTLang = "lo"
	STTLangUZ  STTLang = "uz"
	STTLangFO  STTLang = "fo"
	STTLangHT  STTLang = "ht"
	STTLangPS  STTLang = "ps"
	STTLangTK  STTLang = "tk"
	STTLangNN  STTLang = "nn"
	STTLangMT  STTLang = "mt"
	STTLangSA  STTLang = "sa"
	STTLangLB  STTLang = "lb"
	STTLangMY  STTLang = "my"
	STTLangBO  STTLang = "bo"
	STTLangTL  STTLang = "tl"
	STTLangMG  STTLang = "mg"
	STTLangAS  STTLang = "as"
	STTLangTT  STTLang = "tt"
	STTLangHAW STTLang = "haw"
	STTLangLN  STTLang = "ln"
	STTLangHA  STTLang = "ha"
	STTLangBA  STTLang = "ba"
	STTLangJW  STTLang = "jw"
	STTLangSU  STTLang = "su"
	STTLangYUE STTLang = "yue"
)

var languageMap = map[string]STTLang{
	"en":  STTLangEN,
	"zh":  STTLangZH,
	"de":  STTLangDE,
	"es":  STTLangES,
	"ru":  STTLangRU,
	"ko":  STTLangKO,
	"fr":  STTLangFR,
	"ja":  STTLangJA,
	"pt":  STTLangPT,
	"tr":  STTLangTR,
	"pl":  STTLangPL,
	"ca":  STTLangCA,
	"nl":  STTLangNL,
	"ar":  STTLangAR,
	"sv":  STTLangSV,
	"it":  STTLangIT,
	"id":  STTLangID,
	"hi":  STTLangHI,
	"fi":  STTLangFI,
	"vi":  STTLangVI,
	"he":  STTLangHE,
	"uk":  STTLangUK,
	"el":  STTLangEL,
	"ms":  STTLangMS,
	"cs":  STTLangCS,
	"ro":  STTLangRO,
	"da":  STTLangDA,
	"hu":  STTLangHU,
	"ta":  STTLangTA,
	"no":  STTLangNO,
	"th":  STTLangTH,
	"ur":  STTLangUR,
	"hr":  STTLangHR,
	"bg":  STTLangBG,
	"lt":  STTLangLT,
	"la":  STTLangLA,
	"mi":  STTLangMI,
	"ml":  STTLangML,
	"cy":  STTLangCY,
	"sk":  STTLangSK,
	"te":  STTLangTE,
	"fa":  STTLangFA,
	"lv":  STTLangLV,
	"bn":  STTLangBN,
	"sr":  STTLangSR,
	"az":  STTLangAZ,
	"sl":  STTLangSL,
	"kn":  STTLangKN,
	"et":  STTLangET,
	"mk":  STTLangMK,
	"br":  STTLangBR,
	"eu":  STTLangEU,
	"is":  STTLangIS,
	"hy":  STTLangHY,
	"ne":  STTLangNE,
	"mn":  STTLangMN,
	"bs":  STTLangBS,
	"kk":  STTLangKK,
	"sq":  STTLangSQ,
	"sw":  STTLangSW,
	"gl":  STTLangGL,
	"mr":  STTLangMR,
	"pa":  STTLangPA,
	"si":  STTLangSI,
	"km":  STTLangKM,
	"sn":  STTLangSN,
	"yo":  STTLangYO,
	"so":  STTLangSO,
	"af":  STTLangAF,
	"oc":  STTLangOC,
	"ka":  STTLangKA,
	"be":  STTLangBE,
	"tg":  STTLangTG,
	"sd":  STTLangSD,
	"gu":  STTLangGU,
	"am":  STTLangAM,
	"yi":  STTLangYI,
	"lo":  STTLangLO,
	"uz":  STTLangUZ,
	"fo":  STTLangFO,
	"ht":  STTLangHT,
	"ps":  STTLangPS,
	"tk":  STTLangTK,
	"nn":  STTLangNN,
	"mt":  STTLangMT,
	"sa":  STTLangSA,
	"lb":  STTLangLB,
	"my":  STTLangMY,
	"bo":  STTLangBO,
	"tl":  STTLangTL,
	"mg":  STTLangMG,
	"as":  STTLangAS,
	"tt":  STTLangTT,
	"haw": STTLangHAW,
	"ln":  STTLangLN,
	"ha":  STTLangHA,
	"ba":  STTLangBA,
	"jw":  STTLangJW,
	"su":  STTLangSU,
	"yue": STTLangYUE,
}

// ToString returns the string representation of STTLang.
func (s STTLang) ToString() string {
	return string(s)
}

// LoadLangFromValue creates an STTLang from a string value.
// It returns an error if the value is not a valid STTLang.
func LoadLangFromValue(value string) (STTLang, error) {
	if lang, exists := languageMap[value]; exists {
		return lang, nil
	}
	return "", fmt.Errorf("invalid STTLang value: %s", value)
}
