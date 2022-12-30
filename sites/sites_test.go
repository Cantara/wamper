package sites

/*
import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

var largestStore = 0
var stream = "store_test"
var key = "WBlW78fKy9Po2OgFVmPFMsaEcyCKuOhvq5gjhTI8rvI="
var id = crypto.SimpleHash("some id :P")
var username = "some_username_from_twitch"

func BenchmarkStore(b *testing.B) {
	os.RemoveAll("data")
	s := storeService{
		disc: diskv.New(diskv.Options{
			BasePath:     "data",
			Transform:    flatTransform,
			CacheSizeMax: 1024 * 1024,
		}),
		accountCach: make(map[string]cachData),
	}
	for i := 0; i < b.N; i++ {
		id := crypto.SimpleHash(strconv.Itoa(i))
		a := make(map[string]string)
		a["id"] = id
		data, err := json.Marshal(&a)
		if err != nil {
			b.Fatal(err)
		}
		err = s.store(id, data)
		if err != nil {
			b.Fatal(err)
		}
	}
	if b.N > largestStore {
		largestStore = b.N
	}
}

func TestPreInit(t *testing.T) {
	os.RemoveAll("data")
	err := event.PreInit()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegister(t *testing.T) {
	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()
	s, err := Init(stream, ctx)
	if err != nil {
		t.Fatal(err)
	}
	a := make(map[string]string)
	a["id"] = id
	data, err := json.Marshal(&a)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Register(data, id, key)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLink(t *testing.T) {
	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()
	s, err := Init(stream, ctx)
	if err != nil {
		t.Fatal(err)
	}
	a := make(map[string]string)
	a["id"] = id
	data, err := json.Marshal(&a)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Link(data, id, "TWITCH", username, key)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetById(t *testing.T) {
	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()
	s, err := Init(stream, ctx)
	if err != nil {
		t.Fatal(err)
	}
	a := make(map[string]string)
	a["id"] = id
	data, err := json.Marshal(&a)
	if err != nil {
		t.Fatal(err)
	}
	dataOut, err := s.GetById(id, key)
	if err != nil {
		t.Fatal(err)
	}
	if !testEq(dataOut, data) {
		t.Fatal("stored and got data is not equal")
	}
}

func TestGetByUsername(t *testing.T) {
	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()
	s, err := Init(stream, ctx)
	if err != nil {
		t.Fatal(err)
	}
	a := make(map[string]string)
	a["id"] = id
	data, err := json.Marshal(&a)
	if err != nil {
		t.Fatal(err)
	}
	dataOut, err := s.GetByUsername(username, key)
	if err != nil {
		t.Fatal(err)
	}
	if !testEq(dataOut, data) {
		t.Fatal("stored and got data is not equal")
	}
}

func testEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
*/
