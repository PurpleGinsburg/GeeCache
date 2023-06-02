package geecache

/*
   接口型函数
*/

// 函数类型实现某一个接口，称之为接口型函数
// 实现接口型函数 好处：该接口（Getter）作为某一函数形参时，既可以传GetterFunc的函数，亦可以传实现了该接口（Getter）的结构体作为参数
// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

/*
   定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己
   E函数（参数返回值定义与 F 一致）强转为 F ，即可以转换为接口 A 。
*/
// A GerrerFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}
