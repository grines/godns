package completion

var connected bool
var target string

var linecache []string
var connectedHosts []string
var currentDir string
var curcmd string
var sessionID string
var csessionID string

var records = map[string]string{}

type help struct {
	helpText     string
	infoText     string
	autocomplete string
}
