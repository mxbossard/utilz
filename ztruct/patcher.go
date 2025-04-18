package ztruct

import (
	"encoding/json"
	"fmt"
	_ "log"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/errorz"
)

// Patcher API inspired by https://www.rfc-editor.org/rfc/rfc6902

type (
	operation interface {
		Transform(map[string]any) (map[string]any, error)
	}

	basicOp struct {
		op    string
		path  string
		from  string
		value any
	}

	testOp struct {
		basicOp
		swallowError bool
		thenOps      []operation
		elseOps      []operation
	}
)

func (o *basicOp) Transform(in map[string]any) (out map[string]any, err error) {
	switch o.op {
	case "add":
		out, err = treeAdd(in, o.path, o.value)
	case "default":
		out, err = treeDefault(in, o.path, o.value)
	case "remove":
		out, err = treeRemove(in, o.path)
	case "replace":
		out, err = treeReplace(in, o.path, o.value)
	case "move":
		out, err = treeMove(in, o.from, o.path)
	case "copy":
		out, err = treeCopy(in, o.from, o.path)
	default:
		err = fmt.Errorf("Not supported patch operation: %s !", o.op)
	}
	return
}

func (o *testOp) Transform(in map[string]any) (out map[string]any, err error) {
	switch o.op {
	case "test":
		err = treeTest(in, o.path, o.value)
		out = in
		if err != nil {
			// Swallow test error
			for _, op := range o.elseOps {
				out, err = op.Transform(out)
				if err != nil {
					return nil, err
				}
			}
		} else {
			for _, op := range o.thenOps {
				out, err = op.Transform(out)
				if err != nil {
					return nil, err
				}
			}
		}
		if o.swallowError {
			err = nil
		}
	default:
		err = fmt.Errorf("Not supported patch operation: %s !", o.op)
	}
	return
}

func OpAdd(path string, value any) *basicOp {
	return &basicOp{op: "add", path: path, value: value}
}

func OpDefault(path string, value any) *basicOp {
	// Not in standard
	return &basicOp{op: "default", path: path, value: value}
}

func OpRemove(path string) *basicOp {
	return &basicOp{op: "remove", path: path}
}

func OpReplace(path string, value any) *basicOp {
	return &basicOp{op: "replace", path: path, value: value}
}

func OpMove(from, path string) *basicOp {
	return &basicOp{op: "move", from: from, path: path}
}

func OpCopy(from, path string) *basicOp {
	return &basicOp{op: "copy", from: from, path: path}
}

func OpTest(path string, value any, thenOp, elseOp operation, swallowError bool) *testOp {
	basicOp := basicOp{op: "test", path: path, value: value}
	var thenOps, elseOps []operation
	if thenOp != nil {
		thenOps = append(thenOps, thenOp)
	}
	if elseOp != nil {
		elseOps = append(elseOps, elseOp)
	}
	return &testOp{
		basicOp:      basicOp,
		thenOps:      thenOps,
		elseOps:      elseOps,
		swallowError: swallowError,
	}
}

type (
	patcher struct {
		tree       []byte
		mappedTree map[string]any
		ops        []operation
		outFormat  string
		err        errorz.Aggregated
	}

	pThen struct {
		*patcher
	}

	pElse struct {
		*patcher
	}

	patcherOrThen struct {
		*patcher
		pThen
	}

	patcherOrElse struct {
		*patcher
		pElse
	}

	pThenOrElse struct {
		pThen
		pElse
	}
)

func (p *patcher) ResolveMap() (map[string]any, error) {
	var err error
	if p.mappedTree == nil {
		err = yaml.Unmarshal(p.tree, &p.mappedTree)
		if err != nil {
			return nil, err
		}
	}
	buffer := p.mappedTree
	for _, op := range p.ops {
		buffer, err = op.Transform(buffer)
		if err != nil {
			return nil, err
		}
	}
	return buffer, err
}

func (p *patcher) Resolve() (out []byte, err error) {
	res, err := p.ResolveMap()
	if err != nil {
		return nil, err
	}

	if p.outFormat == "yaml" {
		out, err = yaml.Marshal(res)
	} else if p.outFormat == "json" {
		out, err = json.Marshal(res)
	} else {
		err = fmt.Errorf("Not supported our format: %s !", p.outFormat)
		return nil, err
	}
	return
}

func (p *patcher) ResolveString() (string, error) {
	res, err := p.Resolve()
	if err != nil {
		return "", err
	}
	return string(res), err
}

func (p *patcher) OutFormat(format string) *patcher {
	p.outFormat = format
	return p
}

func (p *patcher) Add(path string, value any) *patcher {
	p.ops = append(p.ops, OpAdd(path, value))
	return p
}

func (p *patcher) Default(path string, value any) *patcher {
	p.ops = append(p.ops, OpDefault(path, value))
	return p
}

func (p *patcher) Remove(path string) *patcher {
	p.ops = append(p.ops, OpRemove(path))
	return p
}

func (p *patcher) Replace(path string, value any) *patcher {
	p.ops = append(p.ops, OpReplace(path, value))
	return p
}

func (p *patcher) Move(from, path string) *patcher {
	p.ops = append(p.ops, OpMove(from, path))
	return p
}

func (p *patcher) Copy(from, path string) *patcher {
	p.ops = append(p.ops, OpCopy(from, path))
	return p
}

func (p *patcher) Test(path string, value any) *pThenOrElse {
	p.ops = append(p.ops, OpTest(path, value, nil, nil, false))
	pt := pThen{p}
	pe := pElse{p}
	ptoe := pThenOrElse{pt, pe}
	return &ptoe
}

func (p *pThenOrElse) SwallowError() *pThenOrElse {
	lastOp := p.pThen.ops[len(p.pThen.ops)-1]
	if testOp, ok := lastOp.(*testOp); ok {
		testOp.swallowError = true
	}
	return p
}

func (p *pThen) Then(ops ...operation) *patcherOrElse {
	// testOp is last op
	lastOp := p.ops[len(p.ops)-1]
	if testOp, ok := lastOp.(*testOp); ok {
		testOp.thenOps = ops
	}

	pe := pElse{p.patcher}
	poe := patcherOrElse{p.patcher, pe}
	return &poe
}

func (p *pElse) Else(ops ...operation) *patcherOrThen {
	// testOp is last op
	lastOp := p.ops[len(p.ops)-1]
	if testOp, ok := lastOp.(*testOp); ok {
		testOp.elseOps = ops
	}

	pt := pThen{p.patcher}
	pot := patcherOrThen{p.patcher, pt}
	return &pot
}

func PatcherMap(yamlIn map[string]any) *patcher {
	return &patcher{mappedTree: yamlIn, outFormat: "json"}
}

func Patcher(yamlIn []byte) *patcher {
	return &patcher{tree: yamlIn, outFormat: "json"}
}

func PatcherString(yamlIn string) *patcher {
	return &patcher{tree: []byte(yamlIn), outFormat: "json"}
}

func treeLeaf[T any](tree map[string]any, path string) (res T, err error) {
	if tree == nil {
		tree = map[string]any{}
	}
	if path == "" {
		path = "/"
	}
	if path[0:1] != "/" {
		err = fmt.Errorf("%w: path: [%s] must start with / !", ErrBadPathFormat, path)
		return
	}
	var p any
	p = tree
	path = strings.TrimLeft(path, "/")
	if path != "" {
		splitedPath := strings.Split(path, "/")
		var browsingPath []string
		for _, key := range splitedPath {
			if p == nil {
				err = fmt.Errorf("%w: path %s is nil cannot resolve path %s in tree: %s !", ErrPathDontExists, "/"+strings.Join(browsingPath, "/"), path, tree)
				return
			}
			browsingPath = append(browsingPath, key)
			if m, ok := p.(map[string]any); ok {
				if p, ok = m[key]; !ok {
					err = fmt.Errorf("%w: path %s not found in tree: %s !", ErrPathDontExists, "/"+strings.Join(browsingPath, "/"), tree)
					return
				}
			} else {
				err = fmt.Errorf("%w: path %s exists but is not a map in tree: %s !", ErrPathDontExists, "/"+strings.Join(browsingPath, "/"), tree)
				return
			}
		}
	}
	var ok bool
	if res, ok = p.(T); !ok {
		err = fmt.Errorf("Impossible to cast %T to %T !", p, res)
		return
	}
	return
}

func cloneLeaf(leaf any) any {
	v := reflect.ValueOf(leaf)
	switch v.Kind() {
	case reflect.Map:
		if m, ok := leaf.(map[string]any); ok {
			return collectionz.CloneMap(m)
		} else {
			// FIXME: what to do ?
			return map[string]any{}
		}
	case reflect.Slice:
		return collectionz.CloneSliceReflect(leaf)
	default:
		return leaf
	}
}

func parentPath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	path = strings.TrimLeft(path, "/")  // Remove heading /
	path = strings.TrimRight(path, "/") // Remove optional trailing /
	splitted := strings.Split(path, "/")
	return "/" + strings.Join(splitted[0:len(splitted)-1], "/")
}

func lastChild(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	path = strings.TrimLeft(path, "/")  // Remove heading /
	path = strings.TrimRight(path, "/") // Remove optional trailing /
	splitted := strings.Split(path, "/")
	return splitted[len(splitted)-1 : len(splitted)][0]
}

func treeAdd(tree map[string]any, path string, value any) (map[string]any, error) {
	// If the target location specifies an array index, a new value is inserted into the array at the specified index.
	// If the target location specifies an object member that does not already exist, a new member is added to the object.
	// If the target location specifies an object member that does exist, that member's value is replaced.
	// When the operation is applied, the target location MUST reference one of:
	// - The root of the target document - whereupon the specified value becomes the entire content of the target document.
	// - A member to add to an existing object - whereupon the supplied value is added to that object at the indicated location. If the member already exists, it is replaced by the specified value.
	// - An element to add to an existing array - whereupon the supplied value is added to the array at the indicated location. Any elements at or above the specified index are shifted one position to the right.  The specified index MUST NOT be greater than the number of elements in the array.  If the "-" character is used to index the end of the array (see [RFC6901]), this has the effect of appending the value to the array.
	// TODO: implement last point with position in array

	if value == nil {
		err := fmt.Errorf("Added value must not be nil !")
		return nil, err
	}

	if path == "" || path == "/" {
		if m, ok := value.(map[string]any); ok {
			return m, nil
		}
		err := fmt.Errorf("Attempt to replace document root by a non map !")
		return nil, err
	}

	pPath := parentPath(path)
	lChild := lastChild(path)
	clone := collectionz.CloneMap(tree)
	parent, err := treeLeaf[map[string]any](clone, pPath)
	if err != nil {
		return nil, err
	}

	if parent[lChild] != nil {
		v := reflect.ValueOf(parent[lChild])
		if v.Kind() == reflect.Slice {
			// Check value not a slice nor a map (cannot add slice into slice !))!
			if reflect.ValueOf(value).Kind() == reflect.Slice {
				err := fmt.Errorf("Attempt to add a slice: [%s] into an array: [%s] !", value, parent[lChild])
				return nil, err
			}
			// Check slice type
			if reflect.TypeOf(parent[lChild]).Elem().Kind() != reflect.ValueOf(value).Kind() {
				err := fmt.Errorf("Attempt to add wrong object type: [%s] into array: [%s] !", value, parent[lChild])
				return nil, err
			}
			newSlice := reflect.Append(v, reflect.ValueOf(value))
			parent[lChild] = newSlice.Interface()
		} else {
			parent[lChild] = value
		}
	} else {
		parent[lChild] = value
	}

	return clone, nil
}

func treeDefault(tree map[string]any, path string, value any) (map[string]any, error) {
	// Not in standard
	// Should add only if do not exists

	pPath := parentPath(path)
	lChild := lastChild(path)
	clone := collectionz.CloneMap(tree)
	parent, err := treeLeaf[map[string]any](clone, pPath)
	if err != nil {
		return nil, err
	}

	if parent[lChild] == nil {
		parent[lChild] = value
	}

	return clone, nil
}

func treeRemove(tree map[string]any, path string) (map[string]any, error) {
	// The target location MUST exist for the operation to be successful.
	if path == "" || path == "/" {
		return map[string]any{}, nil
	}

	pPath := parentPath(path)
	lChild := lastChild(path)
	clone := collectionz.CloneMap(tree)
	parent, err := treeLeaf[map[string]any](clone, pPath)
	if err != nil {
		return nil, err
	}
	delete(parent, lChild)

	return clone, nil
}

func treeReplace(tree map[string]any, path string, value any) (map[string]any, error) {
	// The operation object MUST contain a "value" member whose content specifies the replacement value.
	// The target location MUST exist for the operation to be successful.

	// Assert path exists
	_, err := treeLeaf[any](tree, path)
	if err != nil {
		return nil, err
	}

	return treeAdd(tree, path, value)
}

func treeMove(tree map[string]any, from, path string) (map[string]any, error) {
	// The operation object MUST contain a "from" member, which is a string containing a JSON Pointer value that references the location in the target document to move the value from.
	// The "from" location MUST exist for the operation to be successful.
	// Equivalent to a remove then a add.

	// Assert path exists
	leaf, err := treeLeaf[any](tree, from)
	if err != nil {
		return nil, err
	}

	clone := collectionz.CloneMap(tree)

	clone, err = treeRemove(clone, from)
	if err != nil {
		return nil, err
	}

	// FIXME: should clone leaf !!!
	clone, err = treeAdd(clone, path, cloneLeaf(leaf))
	return clone, err
}

func treeCopy(tree map[string]any, from, path string) (map[string]any, error) {
	// The operation object MUST contain a "from" member, which is a string containing a JSON Pointer value that references the location in the target document to copy the value from.
	// The "from" location MUST exist for the operation to be successful.
	// Equivalent to an add.

	// Assert path exists
	leaf, err := treeLeaf[any](tree, from)
	if err != nil {
		return nil, err
	}

	clone := collectionz.CloneMap(tree)
	// FIXME: should clone leaf !!!
	clone, err = treeAdd(clone, path, cloneLeaf(leaf))
	return clone, err
}

func treeTest(tree map[string]any, path string, value any) error {
	// The operation object MUST contain a "value" member that conveys the value to be compared to the target location's value.
	// The target location MUST be equal to the "value" value for the operation to be considered successful.
	// Here, "equal" means that the value at the target location and the value conveyed by "value" are of the same JSON type, and that they are considered equal by the following rules for that type:
	// 		strings: are considered equal if they contain the same number of Unicode characters and their code points are byte-by-byte equal.
	//		numbers: are considered equal if their values are numerically equal.
	//		arrays: are considered equal if they contain the same number of values, and if each value can be considered equal to the value at the corresponding position in the other array, using this list of type-specific rules.

	// Assert path exists
	leaf, err := treeLeaf[any](tree, path)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(value, leaf) {
		return fmt.Errorf("Value: [%s] at path: [%s] not equals expected [%s] !", leaf, path, value)
	}

	return nil
}
