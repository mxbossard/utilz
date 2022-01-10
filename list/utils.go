package list

import (
	"reflect"
)

// Flat all list objects return a slice of not list objects.
func Flatten(objects ...interface{}) (allObjects []interface{}) {
        for _, obj := range objects {
                // Recursive call if obj is an array or a slice
                t := reflect.TypeOf(obj)
                if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
                        arrayValue := reflect.ValueOf(obj)
                        for i := 0; i < arrayValue.Len(); i++ {
                                value := arrayValue.Index(i).Interface()
                                expanded := Flatten(value)
                                allObjects = append(allObjects, expanded...)
                        }
                        continue
                } else {
                        allObjects = append(allObjects, obj)
                }
        }
        return
}

