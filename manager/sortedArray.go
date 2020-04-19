package manager

import "encoding/json"

type sortedArray [][2]interface{}

func (a sortedArray) Len() int {
	return len(a)
}

func (a sortedArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a *sortedArray) Add(key string, val interface{}) {
	*a = append(*a, [2]interface{}{key, val})
}

func (a sortedArray) Less(i, j int) bool {
	switch a[i][0].(type) {
	case string:
	default:
		panic("sortArray first element of each element must string")
	}
	switch a[j][0].(type) {
	case string:
	default:
		panic("sortArray first element of each element must string")
	}
	return a[i][0].(string) < a[j][0].(string)
}

func (a sortedArray) ToJsonObject() (ret []byte, err error) {
	ret = append(ret, byte('{'))
	for i, v := range a {
		key := v[0]
		val := v[1]
		switch key.(type) {
		case string:
		default:
			panic("sortArray first element of each element must string")
		}
		kbs, err := json.Marshal(key.(string))
		if err != nil {
			return nil, err
		}
		ret = append(ret, kbs...)
		ret = append(ret, []byte(":")...)

		vbs, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		ret = append(ret, vbs...)

		if i < len(a)-1 {
			ret = append(ret, []byte(",")...)
		}

	}
	ret = append(ret, byte('}'))

	return
}
