package ansi

import "regexp"

type Color string

const (
	None  = Color("")
	Reset = Color("\033[0m")

	Black  = Color("\033[0;30m")
	Red    = Color("\033[0;31m")
	Green  = Color("\033[0;32m")
	Yellow = Color("\033[0;33m")
	Blue   = Color("\033[0;34m")
	Purple = Color("\033[0;35m")
	Cyan   = Color("\033[0;36m")
	White  = Color("\033[0;37m")

	HiBlack  = Color("\033[0;90m")
	HiRed    = Color("\033[0;91m")
	HiGreen  = Color("\033[0;92m")
	HiYellow = Color("\033[0;93m")
	HiBlue   = Color("\033[0;94m")
	HiPurple = Color("\033[0;95m")
	HiCyan   = Color("\033[0;96m")
	HiWhite  = Color("\033[0;97m")

	BackBlack  = Color("\033[0;40m")
	BackRed    = Color("\033[0;41m")
	BackGreen  = Color("\033[0;42m")
	BackYellow = Color("\033[0;43m")
	BackBlue   = Color("\033[0;44m")
	BackPurple = Color("\033[0;45m")
	BackCyan   = Color("\033[0;46m")
	BackWhite  = Color("\033[0;47m")

	BackHiBlack  = Color("\033[0;100m")
	BackHiRed    = Color("\033[0;101m")
	BackHiGreen  = Color("\033[0;102m")
	BackHiYellow = Color("\033[0;103m")
	BackHiBlue   = Color("\033[0;104m")
	BackHiPurple = Color("\033[0;105m")
	BackHiCyan   = Color("\033[0;106m")
	BackHiWhite  = Color("\033[0;107m")

	BoldBlack  = Color("\033[1;30m")
	BoldRed    = Color("\033[1;31m")
	BoldGreen  = Color("\033[1;32m")
	BoldYellow = Color("\033[1;33m")
	BoldBlue   = Color("\033[1;34m")
	BoldPurple = Color("\033[1;35m")
	BoldCyan   = Color("\033[1;36m")
	BoldWhite  = Color("\033[1;37m")

	BoldHiBlack  = Color("\033[1;90m")
	BoldHiRed    = Color("\033[1;91m")
	BoldHiGreen  = Color("\033[1;92m")
	BoldHiYellow = Color("\033[1;93m")
	BoldHiBlue   = Color("\033[1;94m")
	BoldHiPurple = Color("\033[1;95m")
	BoldHiCyan   = Color("\033[1;96m")
	BoldHiWhite  = Color("\033[1;97m")

	HilightBlack  = Color("\033[0;97;40m")
	HilightRed    = Color("\033[0;97;41m")
	HilightGreen  = Color("\033[0;97;42m")
	HilightYellow = Color("\033[0;31;43m")
	HilightBlue   = Color("\033[0;97;44m")
	HilightPurple = Color("\033[0;97;45m")
	HilightCyan   = Color("\033[0;97;46m")
	HilightWhite  = Color("\033[0;90;47m")
)

var ansiRulePattern = regexp.MustCompile(`\033\[\d(;\d{2,3}){0,2}m`)

func Unformat(in string) string {
	return ansiRulePattern.ReplaceAllString(in, "")
}
