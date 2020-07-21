package aux

type nullval struct{}
type List map[string]nullval

func NewList(keys ...string) List {
	l := make(List)
	for _, key := range keys {
		l.Add(key)
	}
	return l
}

func (l List) Add(key string) {
	l[key] = nullval{}
}

func (l List) Has(key string) (in bool) {
	_, in = l[key]
	return
}

func IsIn_Fancy(key string, keys ...string) bool {
	return NewList(keys...).Has(key)
}

func IsIn(key string, keys ...string) bool {
	for i := range keys {
		if key == keys[i] {
			return true
		}
	}
	return false
}
