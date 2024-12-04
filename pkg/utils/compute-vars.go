package utils

import (
	"os"
	"strconv"
	"strings"
)

import (
	"github.com/dop251/goja"
)

var formulaVM *goja.Runtime
var formulaEvaluate goja.Callable

var candiedJsData = ""

func GetDBCDecoder(dbcContent string) (func(int, string) string, error) {
	jsPath := "./js/candied.js"
	jsDataBytes, err := os.ReadFile(jsPath)
	if err != nil {
		return nil, err
	}
	candiedJsData = string(jsDataBytes)
	jsStr := string(candiedJsData) + "let bmsDbcFileContent = `" + dbcContent + "`;\n" + `
	    let dbc = new candied.Dbc();
        let can  = new candied.Can();
        can.database = dbc.load(bmsDbcFileContent);
		function decode(frameID, frameData) {
			let logstr = ""
			canFrame = can.createFrame(frameID, JSON.parse(frameData), true);
			boundMsg = can.decode(canFrame);
			 if (boundMsg) { //通道相关的dbc定义
				for ([key, value] of boundMsg.boundSignals.entries()) {
					logstr += ' ' + key + ':' + value.value
				}
			}
			return logstr
		}
	`
	vm := goja.New()

	_, err = vm.RunString(jsStr)
	if err != nil {
		return nil, err
	}
	var decoder func(int, string) string
	err = vm.ExportTo(vm.Get("decode"), &decoder)
	if err != nil {
		return nil, err
	}
	return decoder, nil
}

func ComputeDBCVars(frameID int, frameData string, decoder func(int, string) string) (map[string]float64, error) {
	//startTime := time.Now().UnixMicro()
	dbcSignals := make(map[string]float64)
	result := decoder(frameID, frameData)
	for _, sigVar := range strings.Split(result, " ") {
		if sigVar != "" {
			split := strings.Split(sigVar, ":")
			if len(split) == 2 {
				sigName := split[0]
				sigVal := split[1]
				if val, err := strconv.ParseFloat(sigVal, 64); err == nil {
					dbcSignals[sigName] = val
				}
			}
		}
	}
	//fmt.Printf("cost: %v\n", time.Now().UnixMicro()-startTime)
	return dbcSignals, nil
}
