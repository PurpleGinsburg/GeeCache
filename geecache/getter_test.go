package geecache

import (
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	//匿名函数强制转换为GetterFunc
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed...")
	}
}
