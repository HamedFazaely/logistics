package providerclient

import (
	"testing"
	"time"
)


func TestUpdateSataus(t *testing.T) {
	url:="http://localhost:9090/v1/status/update/hamed"
	ord:=&Order{
		ID:10,
		Status: "0",
		CreatedAt: time.Now(),
	}
	err:=NotifiyProvider(url,ord)
	if err!=nil {
		t.Fatal(err)
	}

}