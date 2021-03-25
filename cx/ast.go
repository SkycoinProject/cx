package cxcore

import (
	"errors"
	"fmt"
	"strings"
)

/*
 * The CXProgram struct contains a full program.
 *
 * It is the root data structures for all code, variable and data structures
 * declarations.
 */

// CXProgram is used to represent a full CX program.
//
// It is the root data structure for the declarations of all functions,
// variables and data structures.
//
type CXProgram struct {
	// Metadata
	Path string // Path to the CX project in the filesystem

	// Contents
	Packages []*CXPackage // Packages in a CX program

	// Runtime information
	Inputs       []*CXArgument // OS input arguments
	Outputs      []*CXArgument // outputs to the OS
	Memory       []byte        // Used when running the program
	
	StackSize    int           // This field stores the size of a CX program's stack
	StackPointer int           // At what byte the current stack frame is

	HeapSize     int           // This field stores the size of a CX program's heap
	HeapStartsAt int           // Offset at which the heap starts in a CX program's memory
	HeapPointer  int           // At what offset a CX program can insert a new object to the heap

	CallStack    []CXCall      // Collection of function calls
	CallCounter  int           // What function call is the currently being executed in the CallStack
	Terminated   bool          // Utility field for the runtime. Indicates if a CX program has already finished or not.
	Version      string        // CX version used to build this CX program.

	// Used by the REPL and parser
	CurrentPackage *CXPackage // Represents the currently active package in the REPL or when parsing a CX file.
	Error          error
}

// CXPackage is used to represent a CX package.
//
type CXPackage struct {
	// Metadata
	Name string // Name of the package

	// Contents
	Imports   []*CXPackage  // imported packages
	Functions []*CXFunction // declared functions in this package
	Structs   []*CXStruct   // declared structs in this package
	Globals   []*CXArgument // declared global variables in this package

	// Used by the REPL and parser
	CurrentFunction *CXFunction
	CurrentStruct   *CXStruct
}

// CXStruct is used to represent a CX struct.
//
type CXStruct struct {
	// Metadata
	Name    string     // Name of the struct
	Package *CXPackage // The package this struct belongs to
	Size    int        // The size in memory that this struct takes.

	// Contents
	Fields []*CXArgument // The fields of the struct
}

// CXFunction is used to represent a CX function.
//
type CXFunction struct {
	// Metadata
	Name     string     // Name of the function
	Package  *CXPackage // The package it's a member of
	IsNative bool       // True if the function is native to CX, e.g. int32.add()
	OpCode   int        // opcode if IsNative = true
    IntCode int // TODO: remove
	// Contents
	Inputs      []*CXArgument   // Input parameters to the function
	Outputs     []*CXArgument   // Output parameters from the function
	Expressions []*CXExpression // Expressions, including control flow statements, in the function
	Length      int             // number of expressions, pre-computed for performance
	Size        int             // automatic memory size

	// Debugging
	FileName string
	FileLine int

	// Used by the GC
	ListOfPointers []*CXArgument // Root pointers for the GC algorithm

	// Used by the REPL and parser
	CurrentExpression *CXExpression

    Version int // TODO: remove when all opcodes use new signature.
}

// CXExpression is used represent a CX expression.
//
// All statements in CX are expressions, including for loops and other control
// flow.
//
type CXExpression struct {
	// Contents
	Inputs   []*CXArgument
	Outputs  []*CXArgument
	Label    string
	Operator *CXFunction
	Function *CXFunction
	Package  *CXPackage

	// debugging
	FileName string
	FileLine int

	// used for jmp statements
	ThenLines int
	ElseLines int

	// 1 = start new scope; -1 = end scope; 0 = just regular expression
	ScopeOperation int

	IsMethodCall    bool
	IsStructLiteral bool
	IsArrayLiteral  bool
	IsUndType       bool
	IsBreak         bool
	IsContinue      bool
}


/*
grep -rn "IsShortDeclaration" .
IsShortDeclaration - is this CXArgument the result of a `CASSIGN` operation (`:=`)?
./cxgo/parser/cxgo.y:1158:							from.Outputs[0].IsShortDeclaration = true
./cxgo/parser/cxgo.y:1169:							from.Outputs[0].IsShortDeclaration = true
./cxgo/parser/cxgo.go:2366:							from.Outputs[0].IsShortDeclaration = true
./cxgo/parser/cxgo.go:2377:							from.Outputs[0].IsShortDeclaration = true
./cxgo/actions/functions.go:147:		if len(expr.Outputs) > 0 && len(expr.Inputs) > 0 && expr.Outputs[0].IsShortDeclaration && !expr.IsStructLiteral && !isParseOp(expr) {
./cxgo/actions/assignment.go:161:		sym.IsShortDeclaration = true
./cxgo/actions/assignment.go:167:			toExpr.Outputs[0].IsShortDeclaration = true
Binary file ./bin/cx matches
./docs/CompilerDevelopment.md:81:* IsShortDeclaration - is this CXArgument the result of a `CASSIGN` operation (`:=`)?
./cx/serialize.go:168:	IsShortDeclaration int32
./cx/serialize.go:337:	s.Arguments[argOff].IsShortDeclaration = serializeBoolean(arg.IsShortDeclaration)
./cx/serialize.go:1051:	arg.IsShortDeclaration = dsBool(sArg.IsShortDeclaration)
./cx/ast.go:234:	IsShortDeclaration    bool
./cx/ast.go:1499:	IsShortDeclaration    bool
*/

/*
grep -rn "IsRest" .
./cxgo/actions/postfix.go:226:		out.IsRest = true
./cxgo/actions/postfix.go:238:		inp.IsRest = true
./cxgo/actions/postfix.go:254:	if left.IsRest {
./cxgo/actions/postfix.go:255:		// right.IsRest = true
./cxgo/actions/postfix.go:264:	left.IsRest = true
Binary file ./bin/cx matches
./docs/CompilerDevelopment.md:79:* IsRest - if this is a package global, is this CXArgument representing the actual global variable from that package?
./cx/serialize.go:166:	IsRest             int32
./cx/serialize.go:335:	s.Arguments[argOff].IsRest = serializeBoolean(arg.IsRest)
./cx/serialize.go:1049:	arg.IsRest = dsBool(sArg.IsRest)
./cx/ast.go:252:	IsRest                bool // pkg.var <- var is rest
./cx/ast.go:1517:	IsRest                bool // pkg.var <- var is rest
./vendor/golang.org/x/sys/windows/security_windows.go:841:// IsRestricted reports whether the access token t is a restricted token.
./vendor/golang.org/x/sys/windows/security_windows.go:842:func (t Token) IsRestricted() (isRestricted bool, err error) {
	*/
	
// CXArgument is used to define local variables, global variables,
// literals (strings, numbers), inputs and outputs to function
// calls. All of the fields in this structure are determined at
// compile time.
type CXArgument struct {
	// Lengths is used if the `CXArgument` defines an array or a
	// slice. The number of dimensions for the array/slice is
	// equal to `len(Lengths)`, while the contents of `Lengths`
	// define the sizes of each dimension. In the case of a slice,
	// `Lengths` only determines the number of dimensions and the
	// sizes are all equal to 0 (these 0s are not used for any
	// computation).
	Lengths               []int
	// DereferenceOperations is a slice of integers where each
	// integer corresponds a `DEREF_*` constant (for example
	// `DEREF_ARRAY`, `DEREF_POINTER`.). A dereference is a
	// process where we consider the bytes at `Offset : Offset +
	// TotalSize` as an address in memory, and we use that address
	// to find the desired value (the referenced
	// value).
	DereferenceOperations []int
	// DeclarationSpecifiers is a slice of integers where each
	// integer corresponds a `DECL_*` constant (for example
	// `DECL_ARRAY`, `DECL_POINTER`.). Declarations are used to
	// create complex types such as `[5][]*Point` (an array of 5
	// slices of pointers to struct instances of type
	// `Point`).
	DeclarationSpecifiers []int
	// Indexes stores what indexes we want to access from the
	// `CXArgument`. A non-nil `Indexes` means that the
	// `CXArgument` is an index or a slice. The elements of
	// `Indexes` can be any `CXArgument` (for example, literals
	// and variables).
	Indexes               []*CXArgument
	// Fields stores what fields are being accessed from the
	// `CXArgument` and in what order. Whenever a `DEREF_FIELD` in
	// `DereferenceOperations` is found, we consume a field from
	// `Field` to determine the new offset to the desired
	// value.
	Fields                []*CXArgument
	// Inputs defines the input parameters of a first-class
	// function. The `CXArgument` is of type `TYPE_FUNC` if
	// `Inputs` is non-nil.
	Inputs                []*CXArgument
	// Outputs defines the output parameters of a first-class
	// function. The `CXArgument` is of type `TYPE_FUNC` if
	// `Outputs` is non-nil.
	Outputs               []*CXArgument
	// Name defines the name of the `CXArgument`. Most of the
	// time, this field will be non-nil as this defines the name
	// of a variable or parameter in source code, but some
	// exceptions exist, such as in the case of literals
	// (e.g. `4`, `"Hello world!"`, `[3]i32{1, 2, 3}`.)
	Name                  string
	// Type defines what's the basic or primitev type of the
	// `CXArgument`. `Type` can be equal to any of the `TYPE_*`
	// constants (e.g. `TYPE_STR`, `TYPE_I32`).
	Type                  int
	// Size determines the size of the basic type. For example, if
	// the `CXArgument` is of type `TYPE_CUSTOM` (i.e. a
	// user-defined type or struct) and the size of the struct
	// representing the custom type is 10 bytes, then `Size == 10`.
	Size                  int
	// TotalSize represents how many bytes are referenced by the
	// `CXArgument` in total. For example, if the `CXArgument`
	// defines an array of 5 struct instances of size 10 bytes,
	// then `TotalSize == 50`.
	TotalSize             int
	// Offset defines a relative memory offset (used in
	// conjunction with the frame pointer), in the case of local
	// variables, or it could define an absolute memory offset, in
	// the case of global variables and literals. It is used by
	// the CX virtual machine to find the bytes that represent the
	// value of the `CXArgument`.
	Offset                int
	// IndirectionLevels
	IndirectionLevels     int
	DereferenceLevels     int
	PassBy                int // pass by value or reference

	FileName              string
	FileLine              int

	CustomType            *CXStruct
	Package               *CXPackage
	IsSlice               bool
	IsArray               bool
	IsArrayFirst          bool // and then dereference
	IsPointer             bool
	IsReference           bool
	IsDereferenceFirst    bool // and then array
	IsStruct              bool
	IsRest                bool // pkg.var <- var is rest
	IsLocalDeclaration    bool
	IsShortDeclaration    bool
	IsInnerReference      bool // for example: &slice[0] or &struct.field
	PreviouslyDeclared    bool
	DoesEscape            bool
}

// CXCall ...
type CXCall struct {
	Operator     *CXFunction // What CX function will be called when running this CXCall in the runtime
	Line         int         // What line in the CX function is currently being executed
	FramePointer int         // Where in the stack is this function call's local variables stored
}

// MakeProgram ...
func MakeProgram() *CXProgram {
	minHeapSize := minHeapSize()
	newPrgrm := &CXProgram{
		Packages:    make([]*CXPackage, 0),
		CallStack:   make([]CXCall, CALLSTACK_SIZE),
		Memory:      make([]byte, STACK_SIZE+minHeapSize),
		StackSize:   STACK_SIZE,
		HeapSize:    minHeapSize,
		HeapPointer: NULL_HEAP_ADDRESS_OFFSET, // We can start adding objects to the heap after the NULL (nil) bytes.
	}
	return newPrgrm
}

// ----------------------------------------------------------------
//                             `CXProgram` Getters

// GetCurrentPackage ...
func (prgrm *CXProgram) GetCurrentPackage() (*CXPackage, error) {
	if prgrm.CurrentPackage != nil {
		return prgrm.CurrentPackage, nil
	}
	return nil, errors.New("current package is nil")

}

// GetCurrentStruct ...
func (prgrm *CXProgram) GetCurrentStruct() (*CXStruct, error) {
	if prgrm.CurrentPackage != nil {
		if prgrm.CurrentPackage.CurrentStruct != nil {
			return prgrm.CurrentPackage.CurrentStruct, nil
		}
		return nil, errors.New("current struct is nil")

	}
	return nil, errors.New("current package is nil")

}

// GetCurrentFunction ...
func (prgrm *CXProgram) GetCurrentFunction() (*CXFunction, error) {
	if prgrm.CurrentPackage != nil {
		if prgrm.CurrentPackage.CurrentFunction != nil {
			return prgrm.CurrentPackage.CurrentFunction, nil
		}
		return nil, errors.New("current function is nil")

	}
	return nil, errors.New("current package is nil")

}

// GetCurrentExpression ...
func (prgrm *CXProgram) GetCurrentExpression() (*CXExpression, error) {
	if prgrm.CurrentPackage != nil &&
		prgrm.CurrentPackage.CurrentFunction != nil &&
		prgrm.CurrentPackage.CurrentFunction.CurrentExpression != nil {
		return prgrm.CurrentPackage.CurrentFunction.CurrentExpression, nil
	}
	return nil, errors.New("current package, function or expression is nil")

}

// GetGlobal ...
func (prgrm *CXProgram) GetGlobal(name string) (*CXArgument, error) {
	mod, err := prgrm.GetCurrentPackage()
	if err != nil {
		return nil, err
	}

	var foundArgument *CXArgument
	for _, def := range mod.Globals {
		if def.Name == name {
			foundArgument = def
			break
		}
	}

	for _, imp := range mod.Imports {
		for _, def := range imp.Globals {
			if def.Name == name {
				foundArgument = def
				break
			}
		}
	}

	if foundArgument == nil {
		return nil, fmt.Errorf("global '%s' not found", name)
	}
	return foundArgument, nil
}

// GetPackage ...
func (prgrm *CXProgram) GetPackage(modName string) (*CXPackage, error) {
	if prgrm.Packages != nil {
		var found *CXPackage
		for _, mod := range prgrm.Packages {
			if modName == mod.Name {
				found = mod
				break
			}
		}
		if found != nil {
			return found, nil
		}
		return nil, fmt.Errorf("package '%s' not found", modName)

	}
	return nil, fmt.Errorf("package '%s' not found", modName)

}

// GetStruct ...
func (prgrm *CXProgram) GetStruct(strctName string, modName string) (*CXStruct, error) {
	var foundPkg *CXPackage
	for _, mod := range prgrm.Packages {
		if modName == mod.Name {
			foundPkg = mod
			break
		}
	}

	var foundStrct *CXStruct

	if foundPkg != nil {
		for _, strct := range foundPkg.Structs {
			if strct.Name == strctName {
				foundStrct = strct
				break
			}
		}
	}

	if foundStrct == nil {
		//looking in imports
		typParts := strings.Split(strctName, ".")

		if mod, err := prgrm.GetPackage(modName); err == nil {
			for _, imp := range mod.Imports {
				for _, strct := range imp.Structs {
					if strct.Name == typParts[0] {
						foundStrct = strct
						break
					}
				}
			}
		}
	}

	if foundPkg != nil && foundStrct != nil {
		return foundStrct, nil
	}
	return nil, fmt.Errorf("struct '%s' not found in package '%s'", strctName, modName)

}

// GetFunction ...
func (prgrm *CXProgram) GetFunction(fnName string, pkgName string) (*CXFunction, error) {
	// I need to first look for the function in the current package
	if pkg, err := prgrm.GetCurrentPackage(); err == nil {
		for _, fn := range pkg.Functions {
			if fn.Name == fnName {
				return fn, nil
			}
		}
	}

	var foundPkg *CXPackage
	for _, pkg := range prgrm.Packages {
		if pkgName == pkg.Name {
			foundPkg = pkg
			break
		}
	}

	var foundFn *CXFunction
	if foundPkg != nil {
		if foundPkg != nil {
			for _, fn := range foundPkg.Functions {
				if fn.Name == fnName {
					foundFn = fn
					break
				}
			}
		}
	} else {
		return nil, fmt.Errorf("package '%s' not found", pkgName)
	}

	if foundPkg != nil && foundFn != nil {
		return foundFn, nil
	}
	return nil, fmt.Errorf("function '%s' not found in package '%s'", fnName, pkgName)

}

// GetExpr returns the current CXExpression
func (prgrm *CXProgram) GetExpr() *CXExpression {
	//call := prgrm.GetCall()
	//return call.Operator.Expressions[call.Line]
	call := &prgrm.CallStack[prgrm.CallCounter]
	return call.Operator.Expressions[call.Line]
}

// GetCall returns the current CXCall
//TODO: What does this do?
func (prgrm *CXProgram) GetCall() *CXCall {
	return &prgrm.CallStack[prgrm.CallCounter]
}

// GetOpCode returns the current OpCode
func (prgrm *CXProgram) GetOpCode() int {
	return prgrm.GetExpr().Operator.OpCode
}

// GetFramePointer returns the current frame pointer
func (prgrm *CXProgram) GetFramePointer() int {
	return prgrm.GetCall().FramePointer
}

// ----------------------------------------------------------------
//                         `CXProgram` Package handling

// AddPackage ...
func (prgrm *CXProgram) AddPackage(mod *CXPackage) *CXProgram {
	found := false
	for _, md := range prgrm.Packages {
		if md.Name == mod.Name {
			prgrm.CurrentPackage = md
			found = true
			break
		}
	}
	if !found {
		prgrm.Packages = append(prgrm.Packages, mod)
		prgrm.CurrentPackage = mod
	}
	return prgrm
}

// RemovePackage ...
func (prgrm *CXProgram) RemovePackage(modName string) {
	lenMods := len(prgrm.Packages)
	for i, mod := range prgrm.Packages {
		if mod.Name == modName {
			if i == lenMods-1 {
				prgrm.Packages = prgrm.Packages[:len(prgrm.Packages)-1]
			} else {
				prgrm.Packages = append(prgrm.Packages[:i], prgrm.Packages[i+1:]...)
			}
			// This means that we're removing the package set to be the CurrentPackage.
			// If it is removed from the program's list of packages, prgrm.CurrentPackage
			// would be pointing to a package meant to be collected by the GC.
			// We fix this by pointing to the last package in the program's list of packages.
			if mod == prgrm.CurrentPackage {
				prgrm.CurrentPackage = prgrm.Packages[len(prgrm.Packages)-1]
			}
			break
		}
	}
}

// ----------------------------------------------------------------
//                             `CXProgram` Selectors

// SelectProgram sets `PROGRAM` to the the receiver `prgrm`. This is a utility function used mainly
// by CX chains. `PROGRAM` is used in multiple parts of the CX runtime as a convenience; instead of having
// to pass around a parameter of type CXProgram, the CX program currently being run is accessible through
// `PROGRAM`.
func (prgrm *CXProgram) SelectProgram() (*CXProgram, error) {
	PROGRAM = prgrm

	return PROGRAM, nil
}

// GetProgram returns the CX program assigned to global variable `PROGRAM`.
// This function is mainly used for CX chains.
func GetProgram() (*CXProgram, error) {
	if PROGRAM == nil {
		return nil, fmt.Errorf("a CX program has not been loaded")
	}
	return PROGRAM, nil
}

// SelectPackage ...
func (prgrm *CXProgram) SelectPackage(name string) (*CXPackage, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(prgrm *CXProgram) {
	// 		prgrm.SelectPackage(name)
	// 	},
	// }
	// saveProgramStep(prgrmStep, prgrm)

	var found *CXPackage
	for _, mod := range prgrm.Packages {
		if mod.Name == name {
			prgrm.CurrentPackage = mod
			found = mod
		}
	}

	if found == nil {
		return nil, fmt.Errorf("Package '%s' does not exist", name)
	}

	return found, nil
}

// SelectFunction ...
func (prgrm *CXProgram) SelectFunction(name string) (*CXFunction, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(prgrm *CXProgram) {
	// 		prgrm.SelectFunction(name)
	// 	},
	// }
	// saveProgramStep(prgrmStep, prgrm)

	mod, err := prgrm.GetCurrentPackage()
	if err == nil {
		return mod.SelectFunction(name)
	}
	return nil, err

}

// SelectStruct ...
func (prgrm *CXProgram) SelectStruct(name string) (*CXStruct, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(prgrm *CXProgram) {
	// 		prgrm.SelectStruct(name)
	// 	},
	// }
	// saveProgramStep(prgrmStep, prgrm)

	mod, err := prgrm.GetCurrentPackage()
	if err == nil {
		return mod.SelectStruct(name)
	}
	return nil, err

}

// SelectExpression ...
func (prgrm *CXProgram) SelectExpression(line int) (*CXExpression, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(prgrm *CXProgram) {
	// 		prgrm.SelectExpression(line)
	// 	},
	// }
	// saveProgramStep(prgrmStep, prgrm)

	mod, err := prgrm.GetCurrentPackage()
	if err == nil {
		return mod.SelectExpression(line)
	}
	return nil, err

}

// ----------------------------------------------------------------
//                             `CXProgram` Debugging

// PrintAllObjects prints all objects in a program
//
func (prgrm *CXProgram) PrintAllObjects() {
	fp := 0

	for c := 0; c <= prgrm.CallCounter; c++ {
		op := prgrm.CallStack[c].Operator

		for _, ptr := range op.ListOfPointers {
			heapOffset := Deserialize_i32(prgrm.Memory[fp+ptr.Offset : fp+ptr.Offset+TYPE_POINTER_SIZE])

			var byts []byte

			if ptr.CustomType != nil {
				// then it's a pointer to a struct
				// use CustomStruct to match the fields against the bytes
				// for _, fld := range ptr.Fields {

				// }

				byts = prgrm.Memory[int(heapOffset)+OBJECT_HEADER_SIZE : int(heapOffset)+OBJECT_HEADER_SIZE+ptr.CustomType.Size]
			}

			// var currLengths []int
			// var currCustom *CXStruct

			// for c := len(ptr.DeclarationSpecifiers) - 1; c >= 0; c-- {
			// 	// we need to go backwards in here

			// 	switch ptr.DeclarationSpecifiers[c] {
			// 	case DECL_POINTER:
			// 		// we might not need to do anything
			// 	case DECL_ARRAY:
			// 		currLengths = ptr.Lengths
			// 	case DECL_SLICE:
			// 	case DECL_STRUCT:
			// 		currCustom = ptr.CustomType
			// 	case DECL_BASIC:
			// 	}
			// }

			// if len(ptr.Lengths) > 0 {
			// 	fmt.Println("ARRAY")
			// }

			// if ptr.CustomType != nil {
			// 	fmt.Println("STRUCT")
			// }

			fmt.Println("declarat", ptr.DeclarationSpecifiers)

			fmt.Println("obj", ptr.Name, ptr.CustomType, prgrm.Memory[heapOffset:int(heapOffset)+op.Size], byts)
		}

		fp += op.Size
	}
}

// MakePackage creates a new empty CXPackage.
//
// It can be filled in later with imports, structs, globals and functions.
//
func MakePackage(name string) *CXPackage {
	return &CXPackage{
		Name:      name,
		Globals:   make([]*CXArgument, 0, 10),
		Imports:   make([]*CXPackage, 0),
		Structs:   make([]*CXStruct, 0),
		Functions: make([]*CXFunction, 0, 10),
	}
}

// ----------------------------------------------------------------
//                             `CXPackage` Getters

// GetImport ...
func (pkg *CXPackage) GetImport(impName string) (*CXPackage, error) {
	for _, imp := range pkg.Imports {
		if imp.Name == impName {
			return imp, nil
		}
	}
	return nil, fmt.Errorf("package '%s' not imported", impName)
}

// GetFunctions ...
func (pkg *CXPackage) GetFunctions() ([]*CXFunction, error) {
	// going from map to slice
	if pkg.Functions != nil {
		return pkg.Functions, nil
	}
	return nil, fmt.Errorf("package '%s' has no functions", pkg.Name)

}

// GetFunction ...
func (pkg *CXPackage) GetFunction(fnName string) (*CXFunction, error) {
	var found bool
	for _, fn := range pkg.Functions {
		if fn.Name == fnName {
			return fn, nil
		}
	}

	// now checking in imported packages
	if !found {
		for _, imp := range pkg.Imports {
			for _, fn := range imp.Functions {
				if fn.Name == fnName {
					return fn, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("function '%s' not found in package '%s' or its imports", fnName, pkg.Name)
}

// GetMethod ...
func (pkg *CXPackage) GetMethod(fnName string, receiverType string) (*CXFunction, error) {
	for _, fn := range pkg.Functions {
		if fn.Name == fnName && len(fn.Inputs) > 0 && fn.Inputs[0].CustomType != nil && fn.Inputs[0].CustomType.Name == receiverType {
			return fn, nil
		}
	}

	// Trying to find it in `Natives`.
	// Most likely a method from a core package.
	if opCode, found := OpCodes[pkg.Name+"."+fnName]; found {
		return Natives[opCode], nil
	}

	return nil, fmt.Errorf("method '%s' not found in package '%s'", fnName, pkg.Name)
}

// GetStruct ...
func (pkg *CXPackage) GetStruct(strctName string) (*CXStruct, error) {
	var foundStrct *CXStruct
	for _, strct := range pkg.Structs {
		if strct.Name == strctName {
			foundStrct = strct
			break
		}
	}

	if foundStrct == nil {
		//looking in imports
		for _, imp := range pkg.Imports {
			for _, strct := range imp.Structs {
				if strct.Name == strctName {
					foundStrct = strct
					break
				}
			}
		}
	}

	if foundStrct != nil {
		return foundStrct, nil
	}
	return nil, fmt.Errorf("struct '%s' not found in package '%s'", strctName, pkg.Name)

}

// GetGlobal ...
func (pkg *CXPackage) GetGlobal(defName string) (*CXArgument, error) {
	var foundDef *CXArgument
	for _, def := range pkg.Globals {
		if def.Name == defName {
			foundDef = def
			break
		}
	}

	if foundDef != nil {
		return foundDef, nil
	}
	return nil, fmt.Errorf("global '%s' not found in package '%s'", defName, pkg.Name)

}

// GetCurrentFunction ...
func (pkg *CXPackage) GetCurrentFunction() (*CXFunction, error) {
	if pkg.CurrentFunction != nil {
		return pkg.CurrentFunction, nil
	}

	return nil, errors.New("current function is nil")
}

// GetCurrentStruct ...
func (pkg *CXPackage) GetCurrentStruct() (*CXStruct, error) {
	if pkg.CurrentStruct != nil {
		return pkg.CurrentStruct, nil
	}

	return nil, errors.New("current struct is nil")
}

// ----------------------------------------------------------------
//                     `CXPackage` Member handling

// AddImport ...
func (pkg *CXPackage) AddImport(imp *CXPackage) *CXPackage {
	found := false
	for _, im := range pkg.Imports {
		if im.Name == imp.Name {
			found = true
			break
		}
	}
	if !found {
		pkg.Imports = append(pkg.Imports, imp)
	}

	return pkg
}

// RemoveImport ...
func (pkg *CXPackage) RemoveImport(impName string) {
	lenImps := len(pkg.Imports)
	for i, imp := range pkg.Imports {
		if imp.Name == impName {
			if i == lenImps-1 {
				pkg.Imports = pkg.Imports[:len(pkg.Imports)-1]
			} else {
				pkg.Imports = append(pkg.Imports[:i], pkg.Imports[i+1:]...)
			}
			break
		}
	}
}

// AddFunction ...
func (pkg *CXPackage) AddFunction(fn *CXFunction) *CXPackage {
	fn.Package = pkg

	found := false
	for i, f := range pkg.Functions {
		if f.Name == fn.Name {
			pkg.Functions[i].Name = fn.Name
			pkg.Functions[i].Inputs = fn.Inputs
			pkg.Functions[i].Outputs = fn.Outputs
			pkg.Functions[i].Expressions = fn.Expressions
			pkg.Functions[i].CurrentExpression = fn.CurrentExpression
			pkg.Functions[i].Package = fn.Package
			pkg.CurrentFunction = pkg.Functions[i]
			found = true
			break
		}
	}
	if found && !InREPL {
		println(CompilationError(fn.FileName, fn.FileLine), "function redeclaration")
	}
	if !found {
		pkg.Functions = append(pkg.Functions, fn)
		pkg.CurrentFunction = fn
	}

	return pkg
}

// RemoveFunction ...
func (pkg *CXPackage) RemoveFunction(fnName string) {
	lenFns := len(pkg.Functions)
	for i, fn := range pkg.Functions {
		if fn.Name == fnName {
			if i == lenFns-1 {
				pkg.Functions = pkg.Functions[:len(pkg.Functions)-1]
			} else {
				pkg.Functions = append(pkg.Functions[:i], pkg.Functions[i+1:]...)
			}
			break
		}
	}
}

// AddStruct ...
func (pkg *CXPackage) AddStruct(strct *CXStruct) *CXPackage {
	found := false
	for i, s := range pkg.Structs {
		if s.Name == strct.Name {
			pkg.Structs[i] = strct
			found = true
			break
		}
	}
	if !found {
		pkg.Structs = append(pkg.Structs, strct)
	}

	strct.Package = pkg
	pkg.CurrentStruct = strct

	return pkg
}

// RemoveStruct ...
func (pkg *CXPackage) RemoveStruct(strctName string) {
	lenStrcts := len(pkg.Structs)
	for i, strct := range pkg.Structs {
		if strct.Name == strctName {
			if i == lenStrcts-1 {
				pkg.Structs = pkg.Structs[:len(pkg.Structs)-1]
			} else {
				pkg.Structs = append(pkg.Structs[:i], pkg.Structs[i+1:]...)
			}
			break
		}
	}
}

// AddGlobal ...
func (pkg *CXPackage) AddGlobal(def *CXArgument) *CXPackage {
	def.Package = pkg
	found := false
	for i, df := range pkg.Globals {
		if df.Name == def.Name {
			pkg.Globals[i] = def
			found = true
			break
		}
	}
	if !found {
		pkg.Globals = append(pkg.Globals, def)
	}

	return pkg
}

// RemoveGlobal ...
func (pkg *CXPackage) RemoveGlobal(defName string) {
	lenGlobals := len(pkg.Globals)
	for i, def := range pkg.Globals {
		if def.Name == defName {
			if i == lenGlobals-1 {
				pkg.Globals = pkg.Globals[:len(pkg.Globals)-1]
			} else {
				pkg.Globals = append(pkg.Globals[:i], pkg.Globals[i+1:]...)
			}
			break
		}
	}
}

// ----------------------------------------------------------------
//                             `CXPackage` Selectors

// SelectFunction ...
func (pkg *CXPackage) SelectFunction(name string) (*CXFunction, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(cxt *CXProgram) {

	// 		if pkg, err := cxt.GetCurrentPackage(); err == nil {
	// 			pkg.SelectFunction(name)
	// 		}
	// 	},
	// }
	// saveProgramStep(prgrmStep, pkg.Context)

	var found *CXFunction
	for _, fn := range pkg.Functions {
		if fn.Name == name {
			pkg.CurrentFunction = fn
			found = fn
		}
	}

	if found == nil {
		return nil, fmt.Errorf("function '%s' does not exist", name)
	}

	return found, nil
}

// SelectStruct ...
func (pkg *CXPackage) SelectStruct(name string) (*CXStruct, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(cxt *CXProgram) {
	// 		if pkg, err := cxt.GetCurrentPackage(); err == nil {
	// 			pkg.SelectStruct(name)
	// 		}
	// 	},
	// }
	// saveProgramStep(prgrmStep, pkg.Context)

	var found *CXStruct
	for _, strct := range pkg.Structs {
		if strct.Name == name {
			pkg.CurrentStruct = strct
			found = strct
		}
	}

	if found == nil {
		return nil, errors.New("Desired structure does not exist")
	}

	return found, nil
}

// SelectExpression ...
func (pkg *CXPackage) SelectExpression(line int) (*CXExpression, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(cxt *CXProgram) {
	// 		if pkg, err := cxt.GetCurrentPackage(); err == nil {
	// 			pkg.SelectExpression(line)
	// 		}
	// 	},
	// }
	// saveProgramStep(prgrmStep, pkg.Context)
	fn, err := pkg.GetCurrentFunction()
	if err == nil {
		return fn.SelectExpression(line)
	}
	return nil, err
}

// MakeStruct ...
func MakeStruct(name string) *CXStruct {
	return &CXStruct{
		Name: name,
	}
}

// ----------------------------------------------------------------
//                             `CXStruct` Getters

// GetFields ...
func (strct *CXStruct) GetFields() ([]*CXArgument, error) {
	if strct.Fields != nil {
		return strct.Fields, nil
	}
	return nil, fmt.Errorf("structure '%s' has no fields", strct.Name)

}

// GetField ...
func (strct *CXStruct) GetField(name string) (*CXArgument, error) {
	for _, fld := range strct.Fields {
		if fld.Name == name {
			return fld, nil
		}
	}
	return nil, fmt.Errorf("field '%s' not found in struct '%s'", name, strct.Name)
}

// ----------------------------------------------------------------
//                     `CXStruct` Member handling

// AddField ...
func (strct *CXStruct) AddField(fld *CXArgument) *CXStruct {
	found := false
	for _, fl := range strct.Fields {
		if fl.Name == fld.Name {
			found = true
			break
		}
	}

	// FIXME: Shouldn't it be a compilation error if we define a new field
	// 	  with the same name as another field?
	if !found {
		numFlds := len(strct.Fields)
		strct.Fields = append(strct.Fields, fld)
		if numFlds != 0 {
			// Pre-compiling the offset of the field.
			// TODO: I think this would cause conflicts with affordances. For instance, affordances could remove a field and we'd need to re-calculate the offsets. We need to check this.
			lastFld := strct.Fields[numFlds-1]
			fld.Offset = lastFld.Offset + lastFld.TotalSize
		}
		strct.Size += GetSize(fld)
	} else {
		panic("duplicate field")
	}
	return strct
}

// RemoveField ...
func (strct *CXStruct) RemoveField(fldName string) {
	if len(strct.Fields) > 0 {
		lenFlds := len(strct.Fields)
		for i, fld := range strct.Fields {
			if fld.Name == fldName {
				if i == lenFlds-1 {
					strct.Fields = strct.Fields[:len(strct.Fields)-1]
				} else {
					strct.Fields = append(strct.Fields[:i], strct.Fields[i+1:]...)
				}
				break
			}
		}
	}
}

// MakeFunction creates an empty function.
//
// Later, parameters and contents can be added.
//
func MakeFunction(name string, fileName string, fileLine int) *CXFunction {
	return &CXFunction{
		Name:     name,
		FileName: fileName,
		FileLine: fileLine,
	}
}

// MakeNativeFunction creates a native function such as i32.add()
//
func MakeNativeFunction(opCode int, inputs []*CXArgument, outputs []*CXArgument) *CXFunction {
	fn := &CXFunction{
		IsNative: true,
		OpCode:   opCode,
        IntCode: -1,
		Version:1,
	}

	offset := 0
	for _, inp := range inputs {
		// for _, typCode := range inputs {
		// inp := MakeArgument("", "", -1).AddType(TypeNames[typCode])
		inp.Offset = offset
		offset += GetSize(inp)
		fn.Inputs = append(fn.Inputs, inp)
	}
	for _, out := range outputs {
		// for _, typCode := range outputs {
		// fn.Outputs = append(fn.Outputs, MakeArgument("", "", -1).AddType(TypeNames[typCode]))
		// out := MakeArgument("", "", -1).AddType(TypeNames[typCode])
		fn.Outputs = append(fn.Outputs, out)
		out.Offset = offset
		offset += GetSize(out)
	}

	return fn
}

// ----------------------------------------------------------------
//                             `CXFunction` Getters

// GetExpressions ...
func (fn *CXFunction) GetExpressions() ([]*CXExpression, error) {
	if fn.Expressions != nil {
		return fn.Expressions, nil
	}
	return nil, fmt.Errorf("function '%s' has no expressions", fn.Name)

}

// GetExpressionByLabel ...
func (fn *CXFunction) GetExpressionByLabel(lbl string) (*CXExpression, error) {
	if fn.Expressions != nil {
		for _, expr := range fn.Expressions {
			if expr.Label == lbl {
				return expr, nil
			}
		}

		return nil, fmt.Errorf("expression '%s' not found in function '%s'", lbl, fn.Name)
	}
	return nil, fmt.Errorf("function '%s' has no expressions", fn.Name)

}

// GetExpressionByLine ...
func (fn *CXFunction) GetExpressionByLine(line int) (*CXExpression, error) {
	if fn.Expressions != nil {
		if line <= len(fn.Expressions) {
			return fn.Expressions[line], nil
		}
		return nil, fmt.Errorf("expression line number '%d' exceeds number of expressions in function '%s'", line, fn.Name)

	}
	return nil, fmt.Errorf("function '%s' has no expressions", fn.Name)

}

// GetCurrentExpression ...
func (fn *CXFunction) GetCurrentExpression() (*CXExpression, error) {
	if fn.CurrentExpression != nil {
		return fn.CurrentExpression, nil
	} else if fn.Expressions != nil {
		return fn.Expressions[0], nil
	} else {
		return nil, errors.New("current expression is nil")
	}
}

// ----------------------------------------------------------------
//                     `CXFunction` Member handling

// AddInput ...
func (fn *CXFunction) AddInput(param *CXArgument) *CXFunction {
	found := false
	for _, inp := range fn.Inputs {
		if inp.Name == param.Name {
			found = true
			break
		}
	}
	if !found {
		fn.Inputs = append(fn.Inputs, param)
	}

	return fn
}

// RemoveInput ...
func (fn *CXFunction) RemoveInput(inpName string) {
	if len(fn.Inputs) > 0 {
		lenInps := len(fn.Inputs)
		for i, inp := range fn.Inputs {
			if inp.Name == inpName {
				if i == lenInps {
					fn.Inputs = fn.Inputs[:len(fn.Inputs)-1]
				} else {
					fn.Inputs = append(fn.Inputs[:i], fn.Inputs[i+1:]...)
				}
				break
			}
		}
	}
}

// AddOutput ...
func (fn *CXFunction) AddOutput(param *CXArgument) *CXFunction {
	found := false
	for _, out := range fn.Outputs {
		if out.Name == param.Name {
			found = true
			break
		}
	}
	if !found {
		fn.Outputs = append(fn.Outputs, param)
	}

	param.Package = fn.Package

	return fn
}

// RemoveOutput ...
func (fn *CXFunction) RemoveOutput(outName string) {
	if len(fn.Outputs) > 0 {
		lenOuts := len(fn.Outputs)
		for i, out := range fn.Outputs {
			if out.Name == outName {
				if i == lenOuts {
					fn.Outputs = fn.Outputs[:len(fn.Outputs)-1]
				} else {
					fn.Outputs = append(fn.Outputs[:i], fn.Outputs[i+1:]...)
				}
				break
			}
		}
	}
}

// AddExpression ...
func (fn *CXFunction) AddExpression(expr *CXExpression) *CXFunction {
	// expr.Program = fn.Program
	expr.Package = fn.Package
	expr.Function = fn
	fn.Expressions = append(fn.Expressions, expr)
	fn.CurrentExpression = expr
	fn.Length++
	return fn
}

// RemoveExpression ...
func (fn *CXFunction) RemoveExpression(line int) {
	if len(fn.Expressions) > 0 {
		lenExprs := len(fn.Expressions)
		if line >= lenExprs-1 || line < 0 {
			fn.Expressions = fn.Expressions[:len(fn.Expressions)-1]
		} else {
			fn.Expressions = append(fn.Expressions[:line], fn.Expressions[line+1:]...)
		}
		// for i, expr := range fn.Expressions {
		// 	expr.Index = i
		// }
	}
}

// ----------------------------------------------------------------
//                             `CXFunction` Selectors

// SelectExpression ...
func (fn *CXFunction) SelectExpression(line int) (*CXExpression, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(cxt *CXProgram) {
	// 		if mod, err := cxt.GetCurrentPackage(); err == nil {
	// 			if fn, err := mod.GetCurrentFunction(); err == nil {
	// 				fn.SelectExpression(line)
	// 			}
	// 		}
	// 	},
	// }
	// saveProgramStep(prgrmStep, fn.Context)
	if len(fn.Expressions) == 0 {
		return nil, errors.New("There are no expressions in this function")
	}

	if line >= len(fn.Expressions) {
		line = len(fn.Expressions) - 1
	}

	if line < 0 {
		line = 0
	}

	expr := fn.Expressions[line]
	fn.CurrentExpression = expr

	return expr, nil
}

// MakeExpression ...
func MakeExpression(op *CXFunction, fileName string, fileLine int) *CXExpression {
	return &CXExpression{
		Operator: op,
		FileLine: fileLine,
		FileName: fileName}
}

// ----------------------------------------------------------------
//                             `CXExpression` Getters

// GetInputs ...
func (expr *CXExpression) GetInputs() ([]*CXArgument, error) {
	if expr.Inputs != nil {
		return expr.Inputs, nil
	}
	return nil, errors.New("expression has no arguments")

}

// ----------------------------------------------------------------
//                     `CXExpression` Member handling

// AddInput ...
func (expr *CXExpression) AddInput(param *CXArgument) *CXExpression {
	// param.Package = expr.Package
	expr.Inputs = append(expr.Inputs, param)
	if param.Package == nil {
		param.Package = expr.Package
	}
	return expr
}

// RemoveInput ...
func (expr *CXExpression) RemoveInput() {
	if len(expr.Inputs) > 0 {
		expr.Inputs = expr.Inputs[:len(expr.Inputs)-1]
	}
}

// AddOutput ...
func (expr *CXExpression) AddOutput(param *CXArgument) *CXExpression {
	// param.Package = expr.Package
	expr.Outputs = append(expr.Outputs, param)
	if param.Package == nil {
		param.Package = expr.Package
	}
	return expr
}

// RemoveOutput ...
func (expr *CXExpression) RemoveOutput() {
	if len(expr.Outputs) > 0 {
		expr.Outputs = expr.Outputs[:len(expr.Outputs)-1]
	}
}

// AddLabel ...
func (expr *CXExpression) AddLabel(lbl string) *CXExpression {
	expr.Label = lbl
	return expr
}

/*
	FileName              string
- filename and line number
- can be moved to CX AST annotations (comments to be skipped or map)
	
	FileLine
*/

/*
Note: Dereference Levels, is possible unused

grep -rn "DereferenceLevels" .

./cxgo/actions/functions.go:328:			if fld.IsPointer && fld.DereferenceLevels == 0 {
./cxgo/actions/functions.go:329:				fld.DereferenceLevels++
./cxgo/actions/functions.go:333:		if arg.IsStruct && arg.IsPointer && len(arg.Fields) > 0 && arg.DereferenceLevels == 0 {
./cxgo/actions/functions.go:334:			arg.DereferenceLevels++
./cxgo/actions/functions.go:1132:					nameFld.DereferenceLevels = sym.DereferenceLevels
./cxgo/actions/functions.go:1150:						nameFld.DereferenceLevels++
./cxgo/actions/expressions.go:328:		exprOut.DereferenceLevels++
./CompilerDevelopment.md:70:* DereferenceLevels - How many dereference operations are performed to get this CXArgument?
./cx/serialize.go:149:	DereferenceLevels           int32
./cx/serialize.go:300:	s.Arguments[argOff].DereferenceLevels = int32(arg.DereferenceLevels)
./cx/serialize.go:1008:	arg.DereferenceLevels = int(sArg.DereferenceLevels)
./cx/cxargument.go:22:	DereferenceLevels     int
./cx/utilities.go:143:	if arg.DereferenceLevels > 0 {
./cx/utilities.go:144:		for c := 0; c < arg.DereferenceLevels; c++ {
*/

/*
Note: IndirectionLevels does not appear to be used at all

 grep -rn "IndirectionLevels" .
./cxgo/actions/functions.go:951:	sym.IndirectionLevels = arg.IndirectionLevels
./cxgo/actions/declarations.go:379:			declSpec.IndirectionLevels++
./cxgo/actions/declarations.go:383:			for c := declSpec.IndirectionLevels - 1; c > 0; c-- {
./cxgo/actions/declarations.go:384:				pointer.IndirectionLevels = c
./cxgo/actions/declarations.go:388:			declSpec.IndirectionLevels++
./CompilerDevelopment.md:69:* IndirectionLevels - how many discrete levels of indirection to this specific CXArgument?
Binary file ./bin/cx matches
./cx/serialize.go:148:	IndirectionLevels           int32
./cx/serialize.go:299:	s.Arguments[argOff].IndirectionLevels = int32(arg.IndirectionLevels)
./cx/serialize.go:1007:	arg.IndirectionLevels = int(sArg.IndirectionLevels)
./cx/cxargument.go:21:	IndirectionLevels     int
*/

/*
IsDereferenceFirst - is this both an array and a pointer, and if so, 
is the pointer first? Mutually exclusive with IsArrayFirst.

grep -rn "IsDereferenceFirst" .
./cxgo/actions/postfix.go:60:	if !elt.IsDereferenceFirst {
./cxgo/actions/expressions.go:331:			exprOut.IsDereferenceFirst = true
./CompilerDevelopment.md:76:* IsArrayFirst - is this both a pointer and an array, and if so, is the array first? Mutually exclusive with IsDereferenceFirst
./CompilerDevelopment.md:78:* IsDereferenceFirst - is this both an array and a pointer, and if so, is the pointer first? Mutually exclusive with IsArrayFirst.
Binary file ./bin/cx matches
./cx/serialize.go:161:	IsDereferenceFirst int32
./cx/serialize.go:314:	s.Arguments[argOff].IsDereferenceFirst = serializeBoolean(arg.IsDereferenceFirst)
./cx/serialize.go:1019:	arg.IsDereferenceFirst = dsBool(sArg.IsDereferenceFirst)
./cx/cxargument.go:32:	IsDereferenceFirst    bool // and then array
./cx/cxargument.go:43:IsDereferenceFirst - is this both an array and a pointer, and if so, 

*/


/*
All "Is" can be removed
- because there is a constants for type (int) for defining the types
- could look in definition, specifier
- but use int lookup
	IsSlice               bool
	IsArray               bool
	IsArrayFirst          bool // and then dereference
	IsPointer             bool
	IsReference           bool
	IsDereferenceFirst    bool // and then array
	IsStruct              bool
	IsRest                bool // pkg.var <- var is rest
	IsLocalDeclaration    bool
	IsShortDeclaration    bool
	IsInnerReference      bool // for example: &slice[0] or &struct.field

*/

/*

Note: PAssBy is not used too many place
Note: Low priority for deprecation
- isnt this same as "pointer"

grep -rn "PassBy" .
./cxgo/actions/misc.go:425:			arg.PassBy = PASSBY_REFERENCE
./cxgo/actions/functions.go:666:					out.PassBy = PASSBY_VALUE
./cxgo/actions/functions.go:678:		if elt.PassBy == PASSBY_REFERENCE &&
./cxgo/actions/functions.go:712:			out.PassBy = PASSBY_VALUE
./cxgo/actions/functions.go:723:				assignElt.PassBy = PASSBY_VALUE
./cxgo/actions/functions.go:915:			expr.Inputs[0].PassBy = PASSBY_REFERENCE
./cxgo/actions/functions.go:1153:					nameFld.PassBy = fld.PassBy
./cxgo/actions/functions.go:1157:						nameFld.PassBy = PASSBY_REFERENCE
./cxgo/actions/literals.go:219:				sym.PassBy = PASSBY_REFERENCE
./cxgo/actions/expressions.go:336:		baseOut.PassBy = PASSBY_REFERENCE
./cxgo/actions/assignment.go:57:		out.PassBy = PASSBY_REFERENCE
./cxgo/actions/assignment.go:208:		to[0].Outputs[0].PassBy = from[idx].Outputs[0].PassBy
./cxgo/actions/assignment.go:234:			to[0].Outputs[0].PassBy = from[idx].Operator.Outputs[0].PassBy
./cxgo/actions/assignment.go:244:			to[0].Outputs[0].PassBy = from[idx].Operator.Outputs[0].PassBy
./cxgo/actions/declarations.go:55:			glbl.PassBy = offExpr[0].Outputs[0].PassBy
./cxgo/actions/declarations.go:69:				declaration_specifiers.PassBy = glbl.PassBy
./cxgo/actions/declarations.go:85:				declaration_specifiers.PassBy = glbl.PassBy
./cxgo/actions/declarations.go:103:			declaration_specifiers.PassBy = glbl.PassBy
./cxgo/actions/declarations.go:324:			declarationSpecifiers.PassBy = initOut.PassBy
./cxgo/actions/declarations.go:417:		arg.PassBy = PASSBY_REFERENCE
./CompilerDevelopment.md:71:* PassBy - an int constant representing how the variable is passed - pass by value, or pass by reference.

./cx/op_http.go:50:	headerFld.PassBy = PASSBY_REFERENCE
./cx/op_http.go:75:	transferEncodingFld.PassBy = PASSBY_REFERENCE
./cx/serialize.go:168:	PassBy     int32
./cx/serialize.go:321:	s.Arguments[argOff].PassBy = int32(arg.PassBy)
./cx/serialize.go:1009:	arg.PassBy = int(sArg.PassBy)
./cx/execute.go:366:				if inp.PassBy == PASSBY_REFERENCE {
./cx/cxargument.go:23:	PassBy                int // pass by value or reference
./cx/op_misc.go:36:		switch elt.PassBy {
./cx/utilities.go:184:	if arg.PassBy == PASSBY_REFERENCE {
*/
// MakeArgument ...
func MakeArgument(name string, fileName string, fileLine int) *CXArgument {
	return &CXArgument{
		Name:     name,
		FileName: fileName,
		FileLine: fileLine}
}

// MakeField ...
func MakeField(name string, typ int, fileName string, fileLine int) *CXArgument {
	return &CXArgument{
		Name:     name,
		Type:     typ,
		FileName: fileName,
		FileLine: fileLine,
	}
}

// MakeGlobal ...
func MakeGlobal(name string, typ int, fileName string, fileLine int) *CXArgument {
	size := GetArgSize(typ)
	global := &CXArgument{
		Name:     name,
		Type:     typ,
		Size:     size,
		Offset:   HeapOffset,
		FileName: fileName,
		FileLine: fileLine,
	}
	HeapOffset += size
	return global
}

// ----------------------------------------------------------------
//                             `CXArgument` Getters

// ----------------------------------------------------------------
//                     `CXArgument` Member handling

// AddPackage assigns CX package `pkg` to CX argument `arg`.
func (arg *CXArgument) AddPackage(pkg *CXPackage) *CXArgument {
	// pkg, err := PROGRAM.GetPackage(pkgName)
	// if err != nil {
	// 	panic(err)
	// }
	arg.Package = pkg
	return arg
}

// AddType ...
func (arg *CXArgument) AddType(typ string) *CXArgument {
	if typCode, found := TypeCodes[typ]; found {
		arg.Type = typCode
		size := GetArgSize(typCode)
		arg.Size = size
		arg.TotalSize = size
		arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, DECL_BASIC)
	} else {
		arg.Type = TYPE_UNDEFINED
	}

	return arg
}

// AddInput adds input parameters to `arg` in case arg is of type `TYPE_FUNC`.
func (arg *CXArgument) AddInput(inp *CXArgument) *CXArgument {
	arg.Inputs = append(arg.Inputs, inp)
	if inp.Package == nil {
		inp.Package = arg.Package
	}
	return arg
}

// AddOutput adds output parameters to `arg` in case arg is of type `TYPE_FUNC`.
func (arg *CXArgument) AddOutput(out *CXArgument) *CXArgument {
	arg.Outputs = append(arg.Outputs, out)
	if out.Package == nil {
		out.Package = arg.Package
	}
	return arg
}
