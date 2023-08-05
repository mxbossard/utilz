package structz

import (
	"encoding/json"
	"errors"
	"fmt"
	_ "log"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrBadPathFormat  = errors.New("Bad path format")
	ErrPathDontExists = errors.New("Path don't exists")
	ErrBadElementType = errors.New("Bad element type")
)

type (
	unmarshaler interface {
		unmarshal(serialized []byte) (mapped map[string]any, err error)
	}

	explorer interface {
		Explore(path string) (result any, err error)
	}

	resolver[T any] interface {
		Resolve(path string) (result T, err error)
		ResolveArray(path string) (result []T, err error)
	}

	jsonUnmarshaler struct {
	}

	yamlUnmarshaler struct {
	}

	basicExplorer struct {
		unmarshaler
		serialized []byte
		mapped     map[string]any
		//path       []string
		//pointer any
		//err errorz.Aggregated
	}

	basicResolver[T any] struct {
		//resolver[T]
		explorer explorer
	}

	/*
		yamlResolver[T any] struct {
			explorer *explorer
		}
	*/
)

var (
	jsonUnmarshaler2 = func(serialized []byte) (mapped map[string]any, err error) {
		err = json.Unmarshal(serialized, &mapped)
		return
	}
)

func (u jsonUnmarshaler) unmarshal(serialized []byte) (mapped map[string]any, err error) {
	err = json.Unmarshal(serialized, &mapped)
	return
}

func (u yamlUnmarshaler) unmarshal(serialized []byte) (mapped map[string]any, err error) {
	err = yaml.Unmarshal(serialized, &mapped)
	return
}

/*
	func (e *basicExplorer) Get(key string) explorer {
		e.path = append(e.path, key)
		return e
	}

	func (e *basicExplorer) Path(path string) explorer {
		if path == "" {
			return e
		}
		if path[0:1] != "/" {
			err := fmt.Errorf("%w: path %s does not start with / !", ErrBadPathFormat, path)
			e.err.Add(err)
			return e
		}
		path = path[1:len(path)]
		e.path = append(e.path, strings.Split(path, "/")...)
		return e
	}
*/

func (e *basicExplorer) Explore(path string) (result any, err error) {
	/*
		if e.err.GotError() {
			return nil, e.err
		}
	*/

	if e.mapped == nil {
		e.mapped, err = e.unmarshal(e.serialized)
		if err != nil {
			return
		}
	}

	if path == "" {
		return e.mapped, nil
	}

	if path[0:1] != "/" {
		err := fmt.Errorf("%w: path %s does not start with /", ErrBadPathFormat, path)
		//e.err.Add(err)
		return nil, err
	}

	path = path[1:len(path)]
	pathSplit := strings.Split(path, "/")
	var p any
	p = e.mapped
	var browsingPath []string
	for _, key := range pathSplit {
		if p == nil {
			err = fmt.Errorf("%w. Path: [%s] is nil cannot resolve path: [%s} ! In json: [%s}", ErrPathDontExists, strings.Join(browsingPath, "."), path, e.serialized)
			return
		}
		if key == "" {
			return p, nil
		}
		browsingPath = append(browsingPath, key)
		if m, ok := p.(map[string]any); ok {
			if p, ok = m[key]; !ok {
				err = fmt.Errorf("%w. Path: [%s] does not exists ! In json: [%s]", ErrPathDontExists, strings.Join(browsingPath, "."), e.serialized)
				return
			}
		} else {
			err = fmt.Errorf("%w. path: [%s] exists but is not a map ! In json: [%s]", ErrPathDontExists, strings.Join(browsingPath, "."), e.serialized)
			return
		}
	}
	return p, err
}

func (e basicResolver[T]) Resolve(path string) (result T, err error) {
	res, err := e.explorer.Explore(path)
	if err != nil {
		return result, err
	}
	//log.Printf("resolved: %v\n", res)

	var ok bool
	switch r := res.(type) {
	case map[string]any:
		result, err = map2Struct[T](r)
		if err != nil {
			err = fmt.Errorf("%s. Cannot map2struct into type: [%T] ! Caused by %w", ErrBadElementType, result, err)
			return
		}
	case []any:
		err = fmt.Errorf("%s. Cannot resolve array into type: [%T] ! Use ResolveArray() instead", ErrBadElementType, result)
		return
	case any:
		if result, ok = r.(T); !ok {
			err = fmt.Errorf("%s. Cannot cast type: [%T] into type: [%T]", ErrBadElementType, r, result)
			return
		}
	default:
		err = fmt.Errorf("cannot resolve not supported type: [%T]", r)
		return
	}
	//log.Printf("result: %v\n", result)
	return
}
func (e basicResolver[T]) ResolveArray(path string) (result []T, err error) {
	res, err := e.explorer.Explore(path)
	if err != nil {
		return result, err
	}
	//log.Printf("resolved: %v\n", res)

	switch r := res.(type) {
	case []any:
		for _, i := range r {
			if m, ok := i.(map[string]any); ok {
				s, err := map2Struct[T](m)
				if err != nil {
					err = fmt.Errorf("%s: cannot map2struct into type [%T] ! Caused by %w", ErrBadElementType, result, err)
					return nil, err
				}
				result = append(result, s)
			} else {
				if casted, ok := i.(T); ok {
					result = append(result, casted)
				} else {
					err = fmt.Errorf("%s: cannot cast [%T] into type [%T]", ErrBadElementType, i, casted)
					return
				}
			}
		}
	default:
		err = fmt.Errorf("%w: can resolve only array", ErrBadElementType)
		return
	}
	//log.Printf("result: %v\n", result)
	return
}

func map2Struct[T any](in map[string]any) (res T, err error) {
	var buffer []byte
	buffer, err = json.Marshal(in)
	if err != nil {
		return
	}
	//log.Printf("buffer: %s\n", string(buffer))
	err = json.Unmarshal(buffer, &res)
	//log.Printf("struct: %v\n", res)
	return
}

func JsonExplorer(serialized []byte) explorer {
	return &basicExplorer{
		unmarshaler: jsonUnmarshaler{},
		serialized:  serialized,
	}
}

func JsonStringExplorer(json string) explorer {
	return JsonExplorer([]byte(json))
}

func JsonMapExplorer(mapped map[string]any) explorer {
	return &basicExplorer{
		unmarshaler: jsonUnmarshaler{},
		mapped:      mapped,
	}
}

func YamlExplorer(serialized []byte) explorer {
	return &basicExplorer{
		unmarshaler: yamlUnmarshaler{},
		serialized:  serialized,
	}
}

func YamlStringExplorer(json string) explorer {
	return YamlExplorer([]byte(json))
}

func YamlMapExplorer(mapped map[string]any) explorer {
	return &basicExplorer{
		unmarshaler: yamlUnmarshaler{},
		mapped:      mapped,
	}
}

func Resolve[T any](e explorer, path string) (T, error) {
	r := basicResolver[T]{explorer: e}
	return r.Resolve(path)
}

func ResolveArray[T any](e explorer, path string) ([]T, error) {
	r := basicResolver[T]{explorer: e}
	return r.ResolveArray(path)
}

func ResolveMap[T any](e explorer, path string) (map[string]T, error) {
	r := basicResolver[map[string]T]{explorer: e}
	return r.Resolve(path)
}

/*
func YamlResolver[T any](json []byte) resolver[T] {
	return &basicResolver[T]{explorer: YamlExplorer(json)}
}

func YamlStringResolver[T any](json string) resolver[T] {
	return YamlResolver[T]([]byte(json))
}

func YamlMapResolver[T any](json map[string]any, path string) resolver[T] {
	return &basicResolver[T]{explorer: YamlMapExplorer(json)}
}

func ResolveYaml[T any](json []byte, path string) (T, error) {
	return YamlResolver[T](json, path).Resolve()
}

func ResolveYamlString[T any](json string, path string) (T, error) {
	return YamlStringResolver[T](json, path).Resolve()
}

func ResolveYamlMap[T any](json map[string]any, path string) (T, error) {
	return YamlMapResolver[T](json, path).Resolve()
}
*/
