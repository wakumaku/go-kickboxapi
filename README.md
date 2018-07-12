# kickbox API Client [![Build Status](https://travis-ci.org/wakumaku/go-kickboxapi.svg?branch=master)](https://travis-ci.org/wakumaku/go-kickboxapi) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/9b66f7d42dcb413bbf96f8f4d1471020)](https://www.codacy.com/app/wakumaku/go-kickboxapi?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=wakumaku/go-kickboxapi&amp;utm_campaign=Badge_Grade) [![Code Coverage](https://scrutinizer-ci.com/g/wakumaku/go-kickboxapi/badges/coverage.png?b=master)](https://scrutinizer-ci.com/g/wakumaku/go-kickboxapi/?branch=master) [![GoDoc](https://godoc.org/github.com/wakumaku/go-kickboxapi?status.svg)](https://godoc.org/github.com/wakumaku/go-kickboxapi)
### Source: https://docs.kickbox.com/v2.0

```
go get github.com/wakumaku/go-kickboxapi
```

Email validation:
```
client = kickboxapi.New(apiKey, nil)
response, err := client.Verify("email@domain.tld")
if err != nil {
    panic(err)
}
valid := response.IsValid()
```

Bulk Email validation:
```
csv := `"test1@test.com","Foo Bar1"
"test2@test.com","Foo Bar2"
"test3@test.com","Foo Bar3"
"test4@test.com","Foo Bar4"`

client = kickboxapi.New(apiKey, nil)
r, err := client.VerifyMultiple("http://callback.com", "filename.txt", []byte(csv))
if err != nil {
    t.Fatal(err)
}

jobID := r.ID

r, err := client.CheckJobStatus(jobID)
if err != nil {
    t.Fatal(err)
}

if r.IsCompleted() {
    fmt.Println("done!")
}
```

Get credits:
```
client = kickboxapi.New(apiKey, nil)
response, err := client.CreditBalance()
if err != nil {
    panic(err)
}

fmt.Println(response.Balance)
```

Makefile:
* `make test` Runs tests
