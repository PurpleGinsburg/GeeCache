package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	//测试时要知道传入key的哈希值，所以自定义哈希算法，传入数字返回对应的数字
	hash := NewMap(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	//Given the above hash function,this will give replicas with "hashes"
	//将节点"6","4","2"添加到Map.keys中
	//虚拟节点2，4，6，12，14，16，22，24，26
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2", //=2-2
		"11": "2", //6-12-2
		"23": "4", //22-24-4
		"27": "2", //>26-2
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s,should have yield %s", k, v)
		}
	}

	//Adds 8,18,28
	hash.Add("8")

	//27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s,should have yield %s", k, v)
		}
	}
}
