package dtree

import (
	"fmt"
	"math"
	"strings"

	pbc "github.com/davedotdev/proto/eddie/proto_common"
)

var (
	// UnFoldCount is exported because of an intra-package error
	UnFoldCount int
	// UnFoldCountTN is exported because of an intra-package error
	UnFoldCountTN int
)

const (
	debugPrint = false
)

// Yes, I'm using global vars. Shoot me. I'll come back later and address this,
// providing someone hasn't actually shot me.
func init() {
	UnFoldCount = 0
	UnFoldCountTN = 0
	sliceOrderMap = make(map[string]map[int]interface{})
}

func printWrapper(stuffToPrint ...interface{}) {
	if debugPrint {
		fmt.Println(stuffToPrint...)
	}
}

// UnFoldTN unfolds the tree
func UnFoldTN(tn *TreeNode, father string, returnStr *string) {
	*returnStr += fmt.Sprintf("Node:Type -> %s\n", tn.Node.Type)
	if UnFoldCountTN != 0 {
		*returnStr += fmt.Sprintf("Node: Father -> %s\n", father)
	}
	if tn.IsSlice {
		*returnStr += fmt.Sprintf("Node: I am a slice member of slice: %s\n", tn.Node.Type)
	}

	*returnStr += fmt.Sprintf("Node:Kv -> %s\n", tn.Node.Kv)
	*returnStr += fmt.Sprintf("Node:Branches -> %d\n", len(tn.Branches))

	for _, v := range tn.Branches {
		// *returnStr += fmt.Sprintf("Node:Branches -> %d\n", len(v.Branches))
		UnFoldCountTN++
		*returnStr += "\n"
		UnFoldTN(v, tn.Node.Type, returnStr)
	}
}

// CreateTreeNode factory function
func CreateTreeNode() *TreeNode {
	tmp := &TreeNode{}
	// tmp.Branches = make(map[string]*TreeNode)
	tmp.Branches = make([]*TreeNode, 0, 1)
	return tmp
}

// TraverseTree is a recursive Tree lookup
func TraverseTree(v interface{}, TN *TreeNode, keydata string, isSlice bool) {
	kn := strings.ReplaceAll(keydata, "-", "_")
	kn = strings.ReplaceAll(kn, ":", "__")
	iterMap := func(x map[string]interface{}, TN *TreeNode, root string, isSlice bool) {
		for k, v := range x {
			TraverseTree(v, TN, k, isSlice)
		}
	}

	iterSlice := func(x []interface{}, TN *TreeNode, root string, isSlice bool) {

		for _, v := range x {
			// Let's look ahead and figure out if our slice has a map. If map, then create node!
			TraverseTree(v, TN, root, isSlice)
		}
	}

	switch vv := v.(type) {
	// Should we handle NIL?
	// case nil:
	// fmt.Printf("%s => (nil) null\n", kn)
	case string:
		if debugPrint {
			fmt.Printf("%s => (string) %q\n", kn, vv)
		}

		kvInsert := pbc.KV{}
		kvInsert.Key = kn
		kvInsert.Vtype = pbc.T_STRING
		kvInsert.Value = vv
		TN.Node.Kv = append(TN.Node.Kv, &kvInsert)
	case bool:
		if debugPrint {
			fmt.Printf("%s => (bool) %v\n", kn, vv)
		}

		kvInsert := pbc.KV{}
		kvInsert.Key = kn
		kvInsert.Vtype = pbc.T_BOOL
		if vv == true {
			kvInsert.Value = "true"
		} else {
			kvInsert.Value = "false"
		}
		TN.Node.Kv = append(TN.Node.Kv, &kvInsert)
	case float64:
		if debugPrint {
			fmt.Printf("%s => (float64) %f\n", kn, vv)
		}
		var tmpFloat float64
		tmpFloat = vv

		kvInsert := pbc.KV{}
		kvInsert.Key = kn
		kvInsert.Vtype = pbc.T_FLOAT64

		if tmpFloat == math.Trunc(tmpFloat) {
			kvInsert.Value = fmt.Sprintf("%d", (int64(vv)))
		} else {
			kvInsert.Value = fmt.Sprintf("%f", (vv))
		}
		TN.Node.Kv = append(TN.Node.Kv, &kvInsert)

	case map[string]interface{}:
		if debugPrint {
			fmt.Printf("%s => %s (map[string]interface{}) ...\n", kn, v)
		}
		if kn == "root" {
			TN.Node.Type = kn
			TN.IsSlice = isSlice
			iterMap(vv, TN, kn, isSlice)
		} else if TN.Node.Type == "root" {
			TN.Node.Type = kn
			TN.IsSlice = isSlice
			TN.Node.Operation = pbc.Operation_CREATE
			TN.Node.Kv = append(TN.Node.Kv, DKVIDs.Get()...)
			iterMap(vv, TN, kn, isSlice)
		} else {
			NTR := CreateTreeNode()
			if TN.Node.Type != "root" {
				TN.Branch(NTR)
			}
			NTR.Node.Type = kn
			NTR.IsSlice = isSlice
			NTR.Node.Operation = pbc.Operation_CREATE
			NTR.Node.Kv = append(NTR.Node.Kv, DKVIDs.Get()...)
			iterMap(vv, NTR, kn, isSlice)
		}
	case []interface{}:
		if debugPrint {
			fmt.Printf("%s => %s ([]interface{}) ...\n", kn, v)
		}
		if kn == "root" {
			iterSlice(vv, TN, kn, true)
		} else {
			for _, v1 := range vv {
				// Let's look ahead and figure out if our slice has a map. If map, then create node!
				if _, ok := v1.(map[string]interface{}); ok == true {
					printWrapper("FOUND SLICE OF MAP")
					TraverseTree(v1, TN, kn, true)
				} else if _, ok := v1.([]interface{}); ok == true {
					// G'damn it. A slice of slices.
					printWrapper("FOUND SLICE OF SLICE")
					NTR := CreateTreeNode()
					if TN.Node.Type != "root" {
						TN.Branch(NTR)
					}
					NTR.Node.Type = kn
					NTR.Node.Operation = pbc.Operation_CREATE
					NTR.IsSlice = true
					NTR.Node.Kv = append(NTR.Node.Kv, DKVIDs.Get()...)
					iterSlice(vv, NTR, kn, true)
				} else {
					printWrapper("FOUND SLICE OF SCALARs")
					NTR := CreateTreeNode()
					if TN.Node.Type != "root" {
						TN.Branch(NTR)
					}
					NTR.Node.Type = kn
					NTR.Node.Operation = pbc.Operation_CREATE
					NTR.IsSlice = true
					NTR.Node.Kv = append(NTR.Node.Kv, DKVIDs.Get()...)
					TraverseTree(v1, NTR, kn, true)
				}
			}
		}

	default:
		if debugPrint {
			fmt.Printf("%s => (unknown?) ...\n", kn)
		}
	}
}
