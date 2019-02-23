package engine

import (
	cmap "github.com/flywithbug-gopkg/concurrent-map"
)

var (
	syncMap = cmap.New()
)

func SetKeyValue(key string, value interface{}) {
	syncMap.Set(key, value)
}

func RemoveKey(key string) {
	syncMap.Remove(key)
}

func Value(key string) (interface{}, bool) {
	return syncMap.Get(key)
}
