package cxcore

import (
	"fmt"
	"github.com/skycoin/cx/cx/ast"
	"github.com/skycoin/cx/cx/constants"
	"strconv"
)

func buildString(inputs []ast.CXValue, outputs []ast.CXValue) []byte {
	fmtStr := inputs[0].Get_str()

	var res []byte
	var specifiersCounter int
	var lenStr = int(len(fmtStr))

	call := ast.PROGRAM.GetCurrentCall()
	expr := call.Operator.Expressions[call.Line]
    fp := inputs[0].FramePointer

    for c := 0; c < len(fmtStr); c++ {
		var nextCh byte
		ch := fmtStr[c]
		if c < lenStr-1 {
			nextCh = fmtStr[c+1]
		}
		if ch == '\\' {
			switch nextCh {
			case '%':
				c++
				res = append(res, nextCh)
				continue
			case 'n':
				c++
				res = append(res, '\n')
				continue
			default:
				res = append(res, ch)
				continue
			}
		}
		if ch == '%' {
			if specifiersCounter+1 == len(expr.Inputs) {
				res = append(res, []byte(fmt.Sprintf("%%!%c(MISSING)", nextCh))...)
				c++
				continue
			}

			inp := expr.Inputs[specifiersCounter+1]
			switch nextCh {
			case 's':
				res = append(res, []byte(CheckForEscapedChars(ast.ReadStr(fp, inp)))...)
			case 'd':
				switch inp.Type {
				case constants.TYPE_I8:
					res = append(res, []byte(strconv.FormatInt(int64(ast.ReadI8(fp, inp)), 10))...)
				case constants.TYPE_I16:
					res = append(res, []byte(strconv.FormatInt(int64(ast.ReadI16(fp, inp)), 10))...)
				case constants.TYPE_I32:
					res = append(res, []byte(strconv.FormatInt(int64(ast.ReadI32(fp, inp)), 10))...)
				case constants.TYPE_I64:
					res = append(res, []byte(strconv.FormatInt(ast.ReadI64(fp, inp), 10))...)
				case constants.TYPE_UI8:
					res = append(res, []byte(strconv.FormatUint(uint64(ast.ReadUI8(fp, inp)), 10))...)
				case constants.TYPE_UI16:
					res = append(res, []byte(strconv.FormatUint(uint64(ast.ReadUI16(fp, inp)), 10))...)
				case constants.TYPE_UI32:
					res = append(res, []byte(strconv.FormatUint(uint64(ast.ReadUI32(fp, inp)), 10))...)
				case constants.TYPE_UI64:
					res = append(res, []byte(strconv.FormatUint(ast.ReadUI64(fp, inp), 10))...)
				}
			case 'f':
				switch inp.Type {
				case constants.TYPE_F32:
					res = append(res, []byte(strconv.FormatFloat(float64(ast.ReadF32(fp, inp)), 'f', 7, 32))...)
				case constants.TYPE_F64:
					res = append(res, []byte(strconv.FormatFloat(ast.ReadF64(fp, inp), 'f', 16, 64))...)
				}
			case 'v':
				res = append(res, []byte(ast.GetPrintableValue(fp, inp))...)
			}
			c++
			specifiersCounter++
		} else {
			res = append(res, ch)
		}
	}

	if specifiersCounter != len(expr.Inputs)-1 {
		extra := "%!(EXTRA "
		// for _, inp := range expr.ProgramInput[:specifiersCounter] {
		lInps := len(expr.Inputs[specifiersCounter+1:])
		for c := 0; c < lInps; c++ {
			inp := expr.Inputs[specifiersCounter+1+c]
			elt := ast.GetAssignmentElement(inp)
			typ := ""
			_ = typ
			if elt.CustomType != nil {
				// then it's custom type
				typ = elt.CustomType.Name
			} else {
				// then it's native type
				typ = constants.TypeNames[elt.Type]
			}

			if c == lInps-1 {
				extra += fmt.Sprintf("%s=%s", typ, ast.GetPrintableValue(fp, elt))
			} else {
				extra += fmt.Sprintf("%s=%s, ", typ, ast.GetPrintableValue(fp, elt))
			}

		}

		extra += ")"

		res = append(res, []byte(extra)...)
	}

	return res
}

func opSprintf(inputs []ast.CXValue, outputs []ast.CXValue) {
    outputs[0].Set_str(string(buildString(inputs, outputs)))
}

func opPrintf(inputs []ast.CXValue, outputs []ast.CXValue) {
	fmt.Print(string(buildString(inputs, outputs)))
}

//Only used in op_fmt.go, once
func CheckForEscapedChars(str string) []byte {
	var res []byte
	var lenStr = int(len(str))
	for c := 0; c < len(str); c++ {
		var nextCh byte
		ch := str[c]
		if c < lenStr-1 {
			nextCh = str[c+1]
		}
		if ch == '\\' {
			switch nextCh {
			case '%':
				c++
				res = append(res, nextCh)
				continue
			case 'n':
				c++
				res = append(res, '\n')
				continue
			default:
				res = append(res, ch)
				continue
			}

		} else {
			res = append(res, ch)
		}
	}

	return res
}

