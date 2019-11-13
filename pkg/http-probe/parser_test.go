package httpprobe

import (
	"testing"
)

const SuccessString string = `
{
    "action": "onExecuteResult",
    "data": "{\"status\":\"Success\",\"result\":\"{\\\"resposeCode\\\":200,\\\"response\\\":\\\"{\\\\\\\"status\\\\\\\":\\\\\\\"done\\\\\\\",\\\\\\\"returnJSONStr\\\\\\\":\\\\\\\"{\\\\\\\\\\\\\\\"id\\\\\\\\\\\\\\\":\\\\\\\\\\\\\\\"54d6cb1100a4437cb67ffde1db767e5d\\\\\\\\\\\\\\\",\\\\\\\\\\\\\\\"userId\\\\\\\\\\\\\\\":0,\\\\\\\\\\\\\\\"weight\\\\\\\\\\\\\\\":57.1,\\\\\\\\\\\\\\\"pbf\\\\\\\\\\\\\\\":0,\\\\\\\\\\\\\\\"measurementDate\\\\\\\\\\\\\\\":\\\\\\\\\\\\\\\"2019-11-12 15:02:20\\\\\\\\\\\\\\\",\\\\\\\\\\\\\\\"deviceId\\\\\\\\\\\\\\\":\\\\\\\\\\\\\\\"92015500056c\\\\\\\\\\\\\\\",\\\\\\\\\\\\\\\"deviceUserNo\\\\\\\\\\\\\\\":0,\\\\\\\\\\\\\\\"ip\\\\\\\\\\\\\\\":\\\\\\\\\\\\\\\"10.10.109.74\\\\\\\\\\\\\\\",\\\\\\\\\\\\\\\"resistance50k\\\\\\\\\\\\\\\":0,\\\\\\\\\\\\\\\"created\\\\\\\\\\\\\\\":\\\\\\\\\\\\\\\"2019-11-12 15:02:22\\\\\\\\\\\\\\\",\\\\\\\\\\\\\\\"updated\\\\\\\\\\\\\\\":1573542141842,\\\\\\\\\\\\\\\"deleted\\\\\\\\\\\\\\\":0,\\\\\\\\\\\\\\\"duration\\\\\\\\\\\\\\\":252}\\\\\\\",\\\\\\\"errMsg\\\\\\\":\\\\\\\"\\\\\\\",\\\\\\\"androidId\\\\\\\":\\\\\\\"66f34f62c0651ddb\\\\\\\",\\\\\\\"spec\\\\\\\":{\\\\\\\"pkgName\\\\\\\":\\\\\\\"gz.lifesense.weidong\\\\\\\",\\\\\\\"versionName\\\\\\\":\\\\\\\"3.2\\\\\\\",\\\\\\\"methodName\\\\\\\":\\\\\\\"getUnknownWeight\\\\\\\",\\\\\\\"argsJSONStr\\\\\\\":\\\\\\\"{}\\\\\\\",\\\\\\\"timeoutSecs\\\\\\\":30}}\\\\n\\\"}\"}",
    "executeTime": 346
}
`

const FailedString string = `
{
    "action": "onExecuteResult",
    "data": "{\"status\":\"Success\",\"result\":\"{\\\"resposeCode\\\":200,\\\"response\\\":\\\"{\\\\\\\"status\\\\\\\":\\\\\\\"done\\\\\\\",\\\\\\\"returnJSONStr\\\\\\\":\\\\\\\"failed\\\\\\\",\\\\\\\"errMsg\\\\\\\":\\\\\\\"\\\\\\\",\\\\\\\"androidId\\\\\\\":\\\\\\\"66f34f62c0651ddb\\\\\\\",\\\\\\\"spec\\\\\\\":{\\\\\\\"pkgName\\\\\\\":\\\\\\\"gz.lifesense.weidong\\\\\\\",\\\\\\\"versionName\\\\\\\":\\\\\\\"3.2\\\\\\\",\\\\\\\"methodName\\\\\\\":\\\\\\\"getUnknownWeight\\\\\\\",\\\\\\\"argsJSONStr\\\\\\\":\\\\\\\"{}\\\\\\\",\\\\\\\"timeoutSecs\\\\\\\":30}}\\\\n\\\"}\"}",
    "executeTime": 145
}
`

func TestPKUAPIParse(t *testing.T) {
	res, err := PKUAPIParse([]byte(SuccessString))
	t.Logf("API Parse: %s, %+v", res, err)
	if err != nil {
		t.Errorf("API Parse error: %+v", err)
	}
}

func TestPKUAPIParse2(t *testing.T) {
	res, err := PKUAPIParse([]byte(FailedString))
	t.Logf("API Parse: %s, %+v", res, err)

	//val, err := JQParse(res, ".duration", Int)
	//t.Logf("JQ Parse: %v, %v", val, err)
}
