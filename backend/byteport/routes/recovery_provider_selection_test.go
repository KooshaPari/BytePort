package routes

import (
	"byteport/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTerminateInstanceRejectsZeroProviderResources(t *testing.T) {
	db := testDB(t)
	stub := newNVMSStub(t)
	alice := models.User{UUID: "alice-uuid"}
	if err := db.Create(&models.Project{UUID:"p-zero", ID:"p-zero", Owner:alice.UUID, Name:"zero"}).Error; err != nil { t.Fatal(err) }

	w:=httptest.NewRecorder()
	TerminateInstance(handlerContext(w,"/terminate",`{"uuid":"p-zero"}`,alice,true))
	if w.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s",w.Code,w.Body.String()) }
	if stub.stopCalls()!=0 { t.Fatalf("stop calls=%d want 0",stub.stopCalls()) }
}

func TestTerminateInstanceRejectsMultipleProviderResources(t *testing.T) {
	db := testDB(t)
	stub := newNVMSStub(t)
	alice := models.User{UUID: "alice-uuid"}
	p:=models.Project{UUID:"p-many",ID:"p-many",Owner:alice.UUID,Name:"many"}
	p.SetDeploy(map[string]models.Instance{
		"a":{UUID:"sandbox-a",Owner:alice.UUID,ResUUID:p.UUID},
		"b":{UUID:"sandbox-b",Owner:alice.UUID,ResUUID:p.UUID},
	})
	if err:=db.Create(&p).Error; err!=nil { t.Fatal(err) }

	w:=httptest.NewRecorder()
	TerminateInstance(handlerContext(w,"/terminate",`{"uuid":"p-many"}`,alice,true))
	if w.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s",w.Code,w.Body.String()) }
	if stub.stopCalls()!=0 { t.Fatalf("stop calls=%d want 0",stub.stopCalls()) }
}
