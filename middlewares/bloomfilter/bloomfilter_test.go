package bloomfilter

import "testing"

func TestQuery(t *testing.T) {
	names := []string{"jack", "lucy", "david"}
	BF := New(3)
	for _, name := range names {
		BF.Add(name)
	}
	for _, name := range names {
		if !BF.Query(name) {
			t.Fatal("bloomfilter error")
		}
	}
}
