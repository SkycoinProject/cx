package actions

import (
	cxcore "github.com/skycoin/cx/cx"
)

func SelectProgram(prgrm *cxcore.CXProgram) {
	PRGRM = prgrm
}

func SetCorrectArithmeticOp(expr *cxcore.CXExpression) {
	if expr.Operator == nil || len(expr.Outputs) < 1 {
		return
	}

	code := expr.Operator.OpCode
	if code > cxcore.START_OF_OPERATORS && code < cxcore.END_OF_OPERATORS {
		// TODO: argument type are not fully resolved here, should be move elsewhere.
		//expr.Operator = cxcore.GetTypedOperator(cxcore.GetType(expr.ProgramInput[0]), code)
	}
}

// hasDeclSpec determines if an argument has certain declaration specifier
func hasDeclSpec(arg *cxcore.CXArgument, spec int) bool {
	found := false
	for _, s := range arg.DeclarationSpecifiers {
		if s == spec {
			found = true
		}
	}
	return found
}

// hasDerefOp determines if an argument has certain dereference operation
func hasDerefOp(arg *cxcore.CXArgument, spec int) bool {
	found := false
	for _, s := range arg.DereferenceOperations {
		if s == spec {
			found = true
		}
	}
	return found
}

// This function writes those bytes to PRGRM.Data
func WritePrimary(typ int, byts []byte, isGlobal bool) []*cxcore.CXExpression {
	if pkg := PRGRM.GetCurrentPackage(); pkg != nil {
		arg := cxcore.MakeArgument("", CurrentFile, LineNo)
		arg.AddType(cxcore.TypeNames[typ])
		arg.Package = pkg

		var size = len(byts)

		arg.Size = cxcore.GetArgSize(typ)
		arg.TotalSize = size
		arg.Offset = DataOffset

		if arg.Type == cxcore.TYPE_STR || arg.Type == cxcore.TYPE_AFF {
			arg.PassBy = cxcore.PASSBY_REFERENCE
			arg.Size = cxcore.TYPE_POINTER_SIZE
			arg.TotalSize = cxcore.TYPE_POINTER_SIZE
		}

		// A CX program allocates min(INIT_HEAP_SIZE, MAX_HEAP_SIZE) bytes
		// after the stack segment. These bytes are used to allocate the data segment
		// at compile time. If the data segment is bigger than min(INIT_HEAP_SIZE, MAX_HEAP_SIZE),
		// we'll start appending the bytes to PRGRM.Memory.
		// After compilation, we calculate how many bytes we need to add to have a heap segment
		// equal to `minHeapSize()` that is allocated after the data segment.
		if size+DataOffset > len(PRGRM.Memory) {
			var i int
			// First we need to fill the remaining free bytes in
			// the current `PRGRM.Memory` slice.
			for i = 0; i < len(PRGRM.Memory)-DataOffset; i++ {
				PRGRM.Memory[DataOffset+i] = byts[i]
			}
			// Then we append the bytes that didn't fit.
			PRGRM.Memory = append(PRGRM.Memory, byts[i:]...)
		} else {
			for i, byt := range byts {
				PRGRM.Memory[DataOffset+i] = byt
			}
		}
		DataOffset += size

		expr := cxcore.MakeExpression(nil, CurrentFile, LineNo)
		expr.Package = pkg
		expr.Outputs = append(expr.Outputs, arg)
		return []*cxcore.CXExpression{expr}
	} else {
		panic("WritePrimary(): error, PRGRM.GetCurrentPackage is nil")
	}
}

func TotalLength(lengths []int) int {
	var total int = 1
	for _, i := range lengths {
		total *= i
	}
	return total
}

func StructLiteralFields(ident string) *cxcore.CXExpression {
	if pkg := PRGRM.GetCurrentPackage(); pkg != nil {
		arg := cxcore.MakeArgument("", CurrentFile, LineNo)
		arg.AddType(cxcore.TypeNames[cxcore.TYPE_IDENTIFIER])
		arg.Name = ident
		arg.Package = pkg

		expr := cxcore.MakeExpression(nil, CurrentFile, LineNo)
		expr.Outputs = []*cxcore.CXArgument{arg}
		expr.Package = pkg

		return expr
	} else {
		panic("StructLiteralFields(): error, PRGRM.GetCurrentPackage is nil")
	}
}

func AffordanceStructs(pkg *cxcore.CXPackage, currentFile string, lineNo int) {
	// Argument type
	argStrct := cxcore.MakeStruct("Argument")
	// argStrct.Size = cxcore.GetArgSize(cxcore.TYPE_STR) + cxcore.GetArgSize(cxcore.TYPE_STR)

	argFldName := cxcore.MakeField("Name", cxcore.TYPE_STR, "", 0)
	argFldName.TotalSize = cxcore.GetArgSize(cxcore.TYPE_STR)
	argFldIndex := cxcore.MakeField("Index", cxcore.TYPE_I32, "", 0)
	argFldIndex.TotalSize = cxcore.GetArgSize(cxcore.TYPE_I32)
	argFldType := cxcore.MakeField("Type", cxcore.TYPE_STR, "", 0)
	argFldType.TotalSize = cxcore.GetArgSize(cxcore.TYPE_STR)

	argStrct.AddField(argFldName)
	argStrct.AddField(argFldIndex)
	argStrct.AddField(argFldType)

	pkg.AddStruct(argStrct)

	// Expression type
	exprStrct := cxcore.MakeStruct("Expression")
	// exprStrct.Size = cxcore.GetArgSize(cxcore.TYPE_STR)

	exprFldOperator := cxcore.MakeField("Operator", cxcore.TYPE_STR, "", 0)

	exprStrct.AddField(exprFldOperator)

	pkg.AddStruct(exprStrct)

	// Function type
	fnStrct := cxcore.MakeStruct("Function")
	// fnStrct.Size = cxcore.GetArgSize(cxcore.TYPE_STR) + cxcore.GetArgSize(cxcore.TYPE_STR) + cxcore.GetArgSize(cxcore.TYPE_STR)

	fnFldName := cxcore.MakeField("Name", cxcore.TYPE_STR, "", 0)
	fnFldName.TotalSize = cxcore.GetArgSize(cxcore.TYPE_STR)

	fnFldInpSig := cxcore.MakeField("InputSignature", cxcore.TYPE_STR, "", 0)
	fnFldInpSig.Size = cxcore.GetArgSize(cxcore.TYPE_STR)
	fnFldInpSig = DeclarationSpecifiers(fnFldInpSig, []int{0}, cxcore.DECL_SLICE)

	fnFldOutSig := cxcore.MakeField("OutputSignature", cxcore.TYPE_STR, "", 0)
	fnFldOutSig.Size = cxcore.GetArgSize(cxcore.TYPE_STR)
	fnFldOutSig = DeclarationSpecifiers(fnFldOutSig, []int{0}, cxcore.DECL_SLICE)

	fnStrct.AddField(fnFldName)
	fnStrct.AddField(fnFldInpSig)

	fnStrct.AddField(fnFldOutSig)

	pkg.AddStruct(fnStrct)

	// Structure type
	strctStrct := cxcore.MakeStruct("Structure")
	// strctStrct.Size = cxcore.GetArgSize(cxcore.TYPE_STR)

	strctFldName := cxcore.MakeField("Name", cxcore.TYPE_STR, "", 0)
	strctFldName.TotalSize = cxcore.GetArgSize(cxcore.TYPE_STR)

	strctStrct.AddField(strctFldName)

	pkg.AddStruct(strctStrct)

	// Package type
	pkgStrct := cxcore.MakeStruct("Structure")
	// pkgStrct.Size = cxcore.GetArgSize(cxcore.TYPE_STR)

	pkgFldName := cxcore.MakeField("Name", cxcore.TYPE_STR, "", 0)

	pkgStrct.AddField(pkgFldName)

	pkg.AddStruct(pkgStrct)

	// Caller type
	callStrct := cxcore.MakeStruct("Caller")
	// callStrct.Size = cxcore.GetArgSize(cxcore.TYPE_STR) + cxcore.GetArgSize(cxcore.TYPE_I32)

	callFldFnName := cxcore.MakeField("FnName", cxcore.TYPE_STR, "", 0)
	callFldFnName.TotalSize = cxcore.GetArgSize(cxcore.TYPE_STR)
	callFldFnSize := cxcore.MakeField("FnSize", cxcore.TYPE_I32, "", 0)
	callFldFnSize.TotalSize = cxcore.GetArgSize(cxcore.TYPE_I32)

	callStrct.AddField(callFldFnName)
	callStrct.AddField(callFldFnSize)

	pkg.AddStruct(callStrct)

	// Program type
	prgrmStrct := cxcore.MakeStruct("Program")
	// prgrmStrct.Size = cxcore.GetArgSize(cxcore.TYPE_I32) + cxcore.GetArgSize(cxcore.TYPE_I64)

	prgrmFldCallCounter := cxcore.MakeField("CallCounter", cxcore.TYPE_I32, "", 0)
	prgrmFldCallCounter.TotalSize = cxcore.GetArgSize(cxcore.TYPE_I32)
	prgrmFldFreeHeap := cxcore.MakeField("HeapUsed", cxcore.TYPE_I64, "", 0)
	prgrmFldFreeHeap.TotalSize = cxcore.GetArgSize(cxcore.TYPE_I64)

	// prgrmFldCaller := cxcore.MakeField("Caller", cxcore.TYPE_CUSTOM, "", 0)
	prgrmFldCaller := DeclarationSpecifiersStruct(callStrct.Name, callStrct.Package.Name, false, currentFile, lineNo)
	prgrmFldCaller.Name = "Caller"

	prgrmStrct.AddField(prgrmFldCallCounter)
	prgrmStrct.AddField(prgrmFldFreeHeap)
	prgrmStrct.AddField(prgrmFldCaller)

	pkg.AddStruct(prgrmStrct)
}

func PrimaryIdentifier(ident string) []*cxcore.CXExpression {
	if pkg := PRGRM.GetCurrentPackage(); pkg != nil {
		arg := cxcore.MakeArgument(ident, CurrentFile, LineNo) // fix: line numbers in errors sometimes report +1 or -1. Issue #195
		arg.AddType(cxcore.TypeNames[cxcore.TYPE_IDENTIFIER])
		// arg.Typ = "ident"
		arg.Name = ident
		arg.Package = pkg

		// expr := &cxcore.CXExpression{ProgramOutput: []*cxcore.CXArgument{arg}}
		expr := cxcore.MakeExpression(nil, CurrentFile, LineNo)
		expr.Outputs = []*cxcore.CXArgument{arg}
		expr.Package = pkg

		return []*cxcore.CXExpression{expr}
	} else {
		panic("PrimaryIdentifier(): error, PRGRM.GetCurrentPackage is nil")
	}
}

// IsArgBasicType returns true if `arg`'s type is a basic type, false otherwise.
func IsArgBasicType(arg *cxcore.CXArgument) bool {
	switch arg.Type {
	case cxcore.TYPE_BOOL,
		cxcore.TYPE_STR, //A STRING IS NOT AN ATOMIC TYPE
		cxcore.TYPE_F32,
		cxcore.TYPE_F64,
		cxcore.TYPE_I8,
		cxcore.TYPE_I16,
		cxcore.TYPE_I32,
		cxcore.TYPE_I64,
		cxcore.TYPE_UI8,
		cxcore.TYPE_UI16,
		cxcore.TYPE_UI32,
		cxcore.TYPE_UI64:
		return true
	}
	return false
}

// IsAllArgsBasicTypes checks if all the input arguments in an expressions are of basic type.
func IsAllArgsBasicTypes(expr *cxcore.CXExpression) bool {
	for _, inp := range expr.Inputs {
		if !IsArgBasicType(inp) {
			return false
		}
	}
	return true
}
