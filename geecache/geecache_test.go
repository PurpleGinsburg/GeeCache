package geecache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"Ginsburg": "999",
	"Jack":     "589",
	"Purple":   "888",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	//创建一个geecache Group 内存大小为2<<10
	geecache := NewGroup("PrupleGinsburg", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SrcDB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		if view, err := geecache.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} //load from callback function
		if _, err := geecache.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := geecache.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
