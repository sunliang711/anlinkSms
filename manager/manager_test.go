package manager

import "testing"

func TestSend(t *testing.T) {
	smsUrl := "http://tech-anlink-openapi-gateway.test.za-tech.net/x-man/api/v1/message/smssend"
	key := "24217b6b53254059a1967f95b2c1862c"    //test
	secret := "Etuq25NDwx12ZuDkhVa08FNokUDzqmJQ" //test
	taskCode := "XMAN202004204399"               //test
	channelType := "MESSAGE"

	receiver := "18019708955"

	man := NewAnlinkSmsManager(smsUrl, key, secret, taskCode, channelType)

	data := map[string]string{"vc": "9527"}
	err := man.Send(receiver, data)
	if err != nil {
		t.Fatalf("send error: %v", err)
	}
	t.Logf("send ok")
}
