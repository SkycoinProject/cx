// +build cxfx

package cxfx

import cxcore "github.com/skycoin/cx/cx"

type Func_i32_i32 func(a int32, b int32)

var Functions_i32_i32 []Func_i32_i32
var freeFns map[string]*func() = make(map[string]*func(), 0)
var cSources map[string]**uint8 = make(map[string]**uint8, 0)

func opGlfwFuncI32I32(inputs []cxcore.CXValue, outputs []cxcore.CXValue) {
	packageName := inputs[0].Get_str()
	functionName := inputs[1].Get_str()
	callback := func(a int32, b int32) {
		var inps [][]byte = make([][]byte, 2)
		inps[0] = cxcore.FromI32(a)
		inps[1] = cxcore.FromI32(b)
		if fn := cxcore.PROGRAM.GetFunction(functionName, packageName); fn != nil {
			cxcore.PROGRAM.Callback(fn, inps)
		}
	}

	Functions_i32_i32 = append(Functions_i32_i32, callback)
	outputs[0].Set_i32(int32(len(Functions_i32_i32) - 1))
}

func opGlfwCallI32I32(inputs []cxcore.CXValue, outputs []cxcore.CXValue) {
	index := inputs[0].Get_i32()
	count := int32(len(Functions_i32_i32))
	if index >= 0 && index < count {
		Functions_i32_i32[index](inputs[1].Get_i32(), inputs[2].Get_i32())
	}
}
