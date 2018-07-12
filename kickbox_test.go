package kickboxapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kickboxapi "github.com/wakumaku/go-kickboxapi"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	client *kickboxapi.Client
)

var (
	responseVerify200Ok = `{
		"result":"deliverable",
		"reason":"",
		"role":false,
		"free":false,
		"disposable":false,
		"accept_all":false,
		"did_you_mean":"bill.lumbergh@gmail.com",
		"sendex":0.23,
		"email":"bill.lumbergh@gamil.com",
		"user":"bill.lumbergh",
		"domain":"gamil.com",
		"success":true,
		"message":null
	  }`

	responseVerify200NOk = `{
		"result":"undeliverable",
		"reason":"rejected_email",
		"role":false,
		"free":false,
		"disposable":false,
		"accept_all":false,
		"did_you_mean":"bill.lumbergh@gmail.com",
		"sendex":0.23,
		"email":"bill.lumbergh@gamil.com",
		"user":"bill.lumbergh",
		"domain":"gamil.com",
		"success":true,
		"message":null
	  }`

	responseVerify400BAD = `{
		"message": "An error message describing the problem",
		"success": false
	  }`

	responseVerifyMultiple200OK = `{
		"id":123,
		"success":true,
		"message":null
	  }`

	responseVerifyMultiple400BAD = `{
		"message": "An error message describing the problem",
		"success": false
	  }`

	responseCheckJobStatus200OKStarting = `{
		"id": 465,
		"status": "starting",
		"success": true,
		"message": null
	  }`

	responseCheckJobStatus200OKProcessing = `{
		"id": 465,
		"status":"processing",
		"progress":{
		  "deliverable": 0,
		  "undeliverable": 0,
		  "risky": 0,
		  "unknown": 0,
		  "total": 0,
		  "unprocessed": 2
		},
		"success": true,
		"message": null
	  }`

	responseCheckJobStatus200OKCompleted = `{
		"id": 465,
		"name": "Batch API Process - 05-12-2015-01-58-08",
		"download_url": "https://{{DOWNLOAD_ADDRESS_HERE}}",
		"created_at": "2015-05-12T18:58:08.000Z",
		"status": "completed",
		"stats": {
			"deliverable": 1,
			"undeliverable": 1,
			"risky": 0,
			"unknown": 0,
			"sendex": 0.35,
			"addresses": 2
		},
		"error": null,
		"duration": 0,
		"success": true,
		"message": null
	}`

	responseCreditBalance200OK = `{
	  "balance": 1337,
	  "success": true,
	  "message": null
	}`

	responseCreditBalance200NOK = `{
		"balance": 0,
		"success": false,
		"message": "Error retrieving your balance"
	  }`

	responseDisposableCheck200OK = `{"disposable": true}`
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = kickboxapi.NewWith(server.URL, "a_valid_api_key", nil)

	return func() {
		server.Close()
	}
}

func TestVerify(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.Verify], segments, http.StatusOK, responseVerify200Ok)

	r, err := client.Verify("a_valid@email.com")
	if err != nil {
		t.Fatal(err)
	}

	if valid := r.IsValid(); !valid {
		t.Error("Should be valid")
	}
}

func TestVerifyNoOk(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.Verify], segments, http.StatusOK, responseVerify200NOk)

	r, err := client.Verify("a_valid@email.com")
	if err != nil {
		t.Fatal(err)
	}

	if valid := r.IsValid(); valid {
		t.Error("Should not be valid")
	}
}

func TestVerifyest(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.Verify], segments, http.StatusBadRequest, responseVerify400BAD)

	r, err := client.Verify("a_valid@email.com")
	if err != nil {
		t.Fatal(err)
	}

	if valid := r.IsValid(); valid {
		t.Error("Should not be valid")
	}

	if r.Error() == nil {
		t.Error("Should trhow an error")
	}
}

func TestVerifyMultiple(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.VerifyMultiple], segments, http.StatusOK, responseVerifyMultiple200OK)

	r, err := client.VerifyMultiple("http://callback.com", "filename.txt", []byte(`"test@test.com","Foo Bar"`))
	if err != nil {
		t.Fatal(err)
	}

	if r.ID != 123 {
		t.Error("Should be 123")
	}
}

func TestVerifyMultipleBadRequest(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.VerifyMultiple], segments, http.StatusBadRequest, responseVerifyMultiple400BAD)

	r, err := client.VerifyMultiple("http://callback.com", "filename.txt", []byte(`"test@test.com","Foo Bar"`))
	if err != nil {
		t.Fatal(err)
	}

	if r.Error() == nil {
		t.Error("Should trhow an error")
	}
}

func TestCheckJobStatusStarting(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2", "JOB_ID": "415"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.CheckJobStatus], segments, http.StatusOK, responseCheckJobStatus200OKStarting)

	r, err := client.CheckJobStatus(415)
	if err != nil {
		t.Fatal(err)
	}

	if !r.IsStarting() {
		t.Error("Should be starting")
	}
}

func TestCheckJobStatusProcessing(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2", "JOB_ID": "415"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.CheckJobStatus], segments, http.StatusOK, responseCheckJobStatus200OKProcessing)

	r, err := client.CheckJobStatus(415)
	if err != nil {
		t.Fatal(err)
	}

	if !r.IsProcessing() {
		t.Error("Should be processing")
	}
}

func TestCheckJobStatusCompleted(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2", "JOB_ID": "415"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.CheckJobStatus], segments, http.StatusOK, responseCheckJobStatus200OKCompleted)

	r, err := client.CheckJobStatus(415)
	if err != nil {
		t.Fatal(err)
	}

	if !r.IsCompleted() {
		t.Error("Should be completed")
	}
}

func TestCreditBalance(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.CreditBalance], segments, http.StatusOK, responseCreditBalance200OK)

	r, err := client.CreditBalance()
	if err != nil {
		t.Fatal(err)
	}

	if r.Balance != 1337 {
		t.Error("Amount does not match")
	}

	if r.Error() != nil {
		t.Error("Should no be errors")
	}
}

func TestCreditBalanceNoOK(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v2"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.CreditBalance], segments, http.StatusOK, responseCreditBalance200NOK)

	r, err := client.CreditBalance()
	if err != nil {
		t.Fatal(err)
	}

	if r.Error() == nil {
		t.Error("Should be an error")
	}
}

func TestDisposable(t *testing.T) {
	teardown := setup()
	defer teardown()

	segments := map[string]string{"API_VERSION": "v1", "EMAIL_ADDRESS": "email@address.com"}

	setMuxEndpoint(kickboxapi.Endpoints[kickboxapi.DisposableEmailCheck], segments, http.StatusOK, responseDisposableCheck200OK)

	r, err := client.Disposable("email@address.com")
	if err != nil {
		t.Fatal(err)
	}

	if r.Disposable == false {
		t.Error("Should be true")
	}
}

func setMuxEndpoint(endpoint kickboxapi.Endpoint, segments map[string]string, statusCode int, response string) {
	path := endpoint.Path
	for search, replace := range segments {
		path = strings.Replace(path, "{"+search+"}", replace, -1)
	}
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	})
}
