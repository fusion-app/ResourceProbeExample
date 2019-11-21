package parser

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
	//res, err := PKUAPIParse([]byte(FailedString))
	//t.Logf("API Parse: %s, %+v", res, err)
	t.Logf("%v", 15.6)
	//val, err := JQParse(res, ".duration", Int)
	//t.Logf("JQ Parse: %v, %v", val, err)
}

const AppInstanceString string = `
[{"action_id":"sid-3602ABBD-309E-472B-84B8-A0E918BF78DC","action_name":"\u5236\u4f5c\u5496\u5561","resource_id":"sid-6073FE36-655E-462A-A821-7FD84733684D","resource_instance_id":{"kind":"Edge","name":"coffee-maker-fudan","namespace":"fusion-app-resources","uid":"db225aa7-3cbb-40ea-ae51-85310394effd"},"state":"2"},{"action_id":"sid-75CF55E1-180B-4AC7-9788-229935C9065C","action_name":"\u5496\u5561\u5236\u4f5c\u5b8c\u6210","resource_id":"sid-6073FE36-655E-462A-A821-7FD84733684D","resource_instance_id":{"kind":"Edge","name":"coffee-maker-fudan","namespace":"fusion-app-resources","uid":"db225aa7-3cbb-40ea-ae51-85310394effd"},"state":"1"},{"action_id":"sid-6448512B-5539-4D02-9AE6-C040E28BB63D","action_name":"\u9001\u5496\u5561","resource_id":"sid-A222DD91-4BA4-4055-B064-22DA36C46BB8","resource_instance_id":null,"state":""}]
`

func TestAppEngineStatus(t *testing.T) {
	val, err := JQParse([]byte(AppInstanceString), ".", Any)
	t.Logf("JQ Parse: %v, %v", val, err)
}