package config

import (
	"fmt"
	"github.com/spf13/cast"
	"reflect"
)

func convert(mp map[any]any) map[string]any {
	m := make(map[string]any, len(mp))
	for k, v := range mp {
		m[fmt.Sprintf("%v", k)] = v
	}
	return m
}

// 配置文件层级合并，key冲突时由读取顺序决定覆盖顺序(避免key重复)
func merge(dest, src map[string]any) {
	for sk, sv := range src {
		tv, ok := dest[sk]
		if !ok {
			dest[sk] = sv
			continue
		}

		if reflect.TypeOf(sv) != reflect.TypeOf(tv) {
			continue
		}

		switch ttv := tv.(type) {
		case map[any]any:
			tsv := sv.(map[any]any)
			ssv := convert(tsv)
			stv := convert(ttv)
			merge(stv, ssv)
			dest[sk] = stv
		case map[string]any:
			merge(ttv, sv.(map[string]any))
			dest[sk] = ttv
		default:
			dest[sk] = sv
		}
	}
}

// 搜索
func copier(m map[string]any, paths ...string) map[string]any {
	mp := make(map[string]any)
	for k, v := range m {
		mp[k] = v
	}
	for _, k := range paths {
		v, ok := mp[k]
		if !ok {
			vv := make(map[string]any)
			mp[k] = vv
			mp = vv
			continue
		}
		vv, err := cast.ToStringMapE(v)
		if err != nil {
			vv = make(map[string]any)
			mp[k] = vv
		}
		mp = vv
	}
	return mp
}

func search(m map[string]any, paths []string) map[string]any {
	for _, path := range paths {
		v, ok := m[path]
		if !ok {
			vv := make(map[string]any)
			m[path] = vv
			m = vv
			continue
		}
		vv, ok := v.(map[string]any)
		if !ok {
			vv = make(map[string]any)
			m[path] = vv
		}
		m = vv
	}
	return m
}

func find(src map[string]any, prefix, delimiter string) map[string]any {
	var data map[string]any
	for k, v := range src {
		p := fmt.Sprintf("%s%s%s", prefix, delimiter, k)
		if prefix == "" {
			p = k
		}
		m, err := cast.ToStringMapE(v)
		if err != nil {
			data[p] = v
		} else {
			for mk, mv := range find(m, p, delimiter) {
				data[mk] = mv
			}
		}
	}
	return data
}
