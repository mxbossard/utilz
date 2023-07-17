package serializ

import (
	"encoding/json"
	"fmt"
	_ "log"

	"gopkg.in/yaml.v3"
	"mby.fr/utils/errorz"
)

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
		thenOps []operation
		elseOps []operation
	}
)

func (o *basicOp) Transform(in map[string]any) (map[string]any, error) {
	switch o.op {
	case "add":

	case "remove":

	case "replace":

	case "move":

	case "copy":

	}
	return nil, nil
}

func (o *testOp) Transform(in map[string]any) (map[string]any, error) {
	return nil, nil
}

func OpAdd(path string, value any) *basicOp {
	return &basicOp{op: "add", path: path, value: value}
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

func OpTest(path string, value any, thenOp, elseOp operation) *testOp {
	basicOp := basicOp{op: "test", path: path, value: value}
	return &testOp{
		basicOp: basicOp,
		thenOps: []operation{thenOp},
		elseOps: []operation{elseOp},
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

func (p *patcher) Add(path string, value any) *patcher {
	p.ops = append(p.ops, OpAdd(path, value))
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
	p.ops = append(p.ops, OpTest(path, value, nil, nil))
	pt := pThen{p}
	pe := pElse{p}
	ptoe := pThenOrElse{pt, pe}
	return &ptoe
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
	return &patcher{mappedTree: yamlIn}
}

func Patcher(yamlIn []byte) *patcher {
	return &patcher{tree: yamlIn}
}

func PatcherString(yamlIn string) *patcher {
	return &patcher{tree: []byte(yamlIn)}
}

func treeAdd(tree map[string]any, path string, value any) error {

}

func treeRemove(tree map[string]any, path string) error {

}

func treeReplace(tree map[string]any, path string, value any) error {

}

func treeMove(tree map[string]any, from, path string) error {

}

func treeCopy(tree map[string]any, from, path string) error {

}

func treeTest(tree map[string]any, path string, value any) (bool, error) {

}
