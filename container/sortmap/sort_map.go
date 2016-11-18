package sortmap

import (
	"sync"
)

type SortMap struct {

	// 保存所有key
	Keys []interface{}

	// k v存储
	Values map[interface{}]interface{}
	sync.RWMutex
}

func NewSortMap() *SortMap {

	sortMap := &SortMap{
		Keys:   make([]interface{}, 0),
		Values: make(map[interface{}]interface{}),
	}
	return sortMap
}

func (sortMap *SortMap) Set(key interface{}, value interface{}) {

	sortMap.Lock()
	defer sortMap.Unlock()

	temp := sortMap.Values[key]
	if nil == temp {
		sortMap.Keys = append(sortMap.Keys, key)
	}
	sortMap.Values[key] = value
}

func (sortMap *SortMap) Delete(key interface{}) {

	sortMap.Lock()
	defer sortMap.Unlock()
	delete(sortMap.Values, key)
	temp := make([]interface{}, 0)
	for _, value := range sortMap.Keys {

		if value != key {
			temp = append(temp, value)
		}
	}
	sortMap.Keys = temp
}

func (sortMap *SortMap) Get(key interface{}) interface{} {

	return sortMap.Values[key]

}

func (sortMap *SortMap) KeySet() []interface{} {
	return sortMap.Keys
}

func (sortMap *SortMap) ValueSet() []interface{} {

	var values = make([]interface{}, 0)

	for _, v := range sortMap.Keys {
		value := sortMap.Values[v]
		values = append(values, value)
	}
	return values
}
