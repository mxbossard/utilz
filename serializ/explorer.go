package serializ

import (
	"encoding/json"
	"errors"
	"fmt"
	_ "log"
	"strings"

	"mby.fr/utils/errorz"
)

var (
	ErrBatPathFormat = errors.New("Bad path format")
	ErrPathDontExists      = errors.New("Path don't exists")
	ErrBadElementType      = errors.New("Bad element type")
)

type (
	jsonExplorer struct {
		json []byte
		path []string
		tree map[string]any
		//pointer any
		err errorz.Aggregated
	}

	jsonResolver[T any] struct {
		explorer *jsonExplorer
	}
)

func (e *jsonExplorer) Get(key string) *jsonExplorer {
	e.path = append(e.path, key)
	return e
}

func (e *jsonExplorer) Path(path string) *jsonExplorer {
	if path == "" {
		return e
	}
	if path[0:1] != "/" {
		err := fmt.Errorf("%w: path %s does not start with / !", ErrBatPathFormat, path)
		e.err.Add(err)
		return e
	}
	path = path[1:len(path)]
	e.path = append(e.path, strings.Split(path, "/")...)
	return e
}

func (e jsonExplorer) Resolve() (result any, err error) {
	if e.err.GotError() {
		return nil, e.err
	}
	if e.tree == nil {
		err = json.Unmarshal(e.json, &e.tree)
		if err != nil {
			return
		}
	}
	var p any
	p = e.tree
	var browsingPath []string
	for _, key := range e.path {
		if p == nil {
			err = fmt.Errorf("%w: path %s is nil cannot resolve path %s ! In json: %s", ErrPathDontExists, strings.Join(browsingPath, "."), e.path, e.json)
			return
		}
		browsingPath = append(browsingPath, key)
		if m, ok := p.(map[string]any); ok {
			if p, ok = m[key]; !ok {
				err = fmt.Errorf("%w: path %s does not exists ! In json: %s", ErrPathDontExists, strings.Join(browsingPath, "."), e.json)
				return
			}
		} else {
			err = fmt.Errorf("%w: path %s exists but is not a map ! In json: %s", ErrPathDontExists, strings.Join(browsingPath, "."), e.json)
			return
		}
	}
	return p, err
}

func (e jsonResolver[T]) ResolveAny() (result any, err error) {
	// TODO
	return
}

func (e jsonResolver[T]) Resolve() (result T, err error) {
	res, err := e.explorer.Resolve()
	if err != nil {
		return result, err
	}
	//log.Printf("resolved: %v\n", res)

	var ok bool
	switch r := res.(type) {
	case map[string]any:
		result, err = map2Struct[T](r)
		if err != nil {
			err = fmt.Errorf("%s: cannot map2struct into type [%T] ! Caused by %w", ErrBadElementType, result, err)
			return
		}
	case []any:
		err = fmt.Errorf("%s: cannot resolve array into type [%T] ! Use ResolveArray() instead.", ErrBadElementType, result)
		return
	case any:
		if result, ok = r.(T); !ok {
			err = fmt.Errorf("%s: cannot cast [%T] into type [%T] !", ErrBadElementType, r, result)
			return
		}
	default:
		err = fmt.Errorf("Not support type: %T !", r)
		return
	}
	//log.Printf("result: %v\n", result)
	return
}
func (e jsonResolver[T]) ResolveArray() (result []T, err error) {
	res, err := e.explorer.Resolve()
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
					err = fmt.Errorf("%s: cannot cast [%T] into type [%T] !", ErrBadElementType, i, casted)
					return
				}
			}
		}
	default:
		err = fmt.Errorf("%w: can resolve only array !", ErrBadElementType)
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

func JsonExplorer(json []byte) *jsonExplorer {
	return &jsonExplorer{
		json: json,
	}
}

func JsonStringExplorer(json string) *jsonExplorer {
	return JsonExplorer([]byte(json))
}

func JsonMapExplorer(json map[string]any) *jsonExplorer {
	return &jsonExplorer{
		tree: json,
	}
}

func JsonResolver[T any](json []byte, path string) *jsonResolver[T] {
	return &jsonResolver[T]{JsonExplorer(json).Path(path)}
}

func JsonStringResolver[T any](json string, path string) *jsonResolver[T] {
	return JsonResolver[T]([]byte(json), path)
}

func JsonMapResolver[T any](json map[string]any, path string) *jsonResolver[T] {
	return &jsonResolver[T]{JsonMapExplorer(json).Path(path)}
}

func ResolveJson[T any](json []byte, path string) (T, error) {
	return JsonResolver[T](json, path).Resolve()
}

func ResolveJsonString[T any](json string, path string) (T, error) {
	return JsonStringResolver[T](json, path).Resolve()
}

func ResolveJsonMap[T any](json map[string]any, path string) (T, error) {
	return JsonMapResolver[T](json, path).Resolve()
}
