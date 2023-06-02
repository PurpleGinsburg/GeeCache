package geecache

/*
   ByteView——只读数据结构 用于表示缓存值(lru中的Vlaue实例化)
*/

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte
}

//ByteView是只读数据结构 故采用值传递

// 实现Value接口
// Len returns the view's length
func (bv ByteView) Len() int {
	return len(bv.b)
}

// ByteView是只读数据结构，返回拷贝值防止修改
// ByteSlice returns a copy of the data as abyte slice.
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// String returns the data as a string,making a copy if necessary.
func (bv ByteView) String() string {
	return string(bv.b)
}

// deep copy 实现深拷贝
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
