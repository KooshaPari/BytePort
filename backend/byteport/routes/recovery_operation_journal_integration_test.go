package routes

import (
	"byteport/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeployOperationIDReplaysRealizedResourceWithoutSecondProviderCall(t *testing.T) {
	testDB(t)
	stub:=newNVMSStub(t)
	alice:=models.User{UUID:"alice-uuid"}
	body:=`{"operation_id":"op-stable","name":"app","platform":"linux"}`

	w1:=httptest.NewRecorder()
	DeployProject(handlerContext(w1,"/deploy",body,alice,true))
	if w1.Code!=http.StatusOK { t.Fatalf("first status=%d body=%s",w1.Code,w1.Body.String()) }

	w2:=httptest.NewRecorder()
	DeployProject(handlerContext(w2,"/deploy",body,alice,true))
	if w2.Code!=http.StatusOK { t.Fatalf("retry status=%d body=%s",w2.Code,w2.Body.String()) }
	if stub.deployCalls()!=1 { t.Fatalf("provider deploy calls=%d want 1",stub.deployCalls()) }

	var count int64
	if err:=models.DB.Model(&models.RuntimeOperationRecord{}).Where("id = ?","op-stable").Count(&count).Error;err!=nil{t.Fatal(err)}
	if count!=1 { t.Fatalf("operation rows=%d want 1",count) }
}

func TestDeployOperationIDRejectsDifferentFingerprintBeforeProviderMutation(t *testing.T) {
	testDB(t)
	stub:=newNVMSStub(t)
	alice:=models.User{UUID:"alice-uuid"}

	w1:=httptest.NewRecorder()
	DeployProject(handlerContext(w1,"/deploy",`{"operation_id":"op-conflict","name":"app-a","platform":"linux"}`,alice,true))
	if w1.Code!=http.StatusOK { t.Fatalf("first status=%d body=%s",w1.Code,w1.Body.String()) }

	w2:=httptest.NewRecorder()
	DeployProject(handlerContext(w2,"/deploy",`{"operation_id":"op-conflict","name":"app-b","platform":"linux"}`,alice,true))
	if w2.Code!=http.StatusConflict { t.Fatalf("conflict status=%d body=%s",w2.Code,w2.Body.String()) }
	if stub.deployCalls()!=1 { t.Fatalf("provider deploy calls=%d want 1",stub.deployCalls()) }
}

func TestDeployPersistsApplyingOperationBeforeProviderTransportFailure(t *testing.T) {
	testDB(t)
	alice:=models.User{UUID:"alice-uuid"}
	old:=nvmsHTTPClient
	nvmsHTTPClient=&http.Client{Transport:roundTripperFunc(func(*http.Request)(*http.Response,error){return nil,fmt.Errorf("lost response")})}
	t.Cleanup(func(){nvmsHTTPClient=old})

	w:=httptest.NewRecorder()
	DeployProject(handlerContext(w,"/deploy",`{"operation_id":"op-lost","name":"app","platform":"linux"}`,alice,true))
	if w.Code!=http.StatusInternalServerError { t.Fatalf("status=%d body=%s",w.Code,w.Body.String()) }

	var op models.RuntimeOperationRecord
	if err:=models.DB.First(&op,"id = ?","op-lost").Error;err!=nil{t.Fatal(err)}
	if op.State!=models.RuntimeOperationUnknown { t.Fatalf("state=%s want UNKNOWN",op.State) }
}
