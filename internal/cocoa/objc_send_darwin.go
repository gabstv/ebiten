// Copyright 2024 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cocoa

import (
	"math"
	"runtime"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// ObjcMsgSend is resolved in a variable initializer, not init(), because
// package-level variables are initialized before any init() runs.
var ObjcMsgSend = func() uintptr {
	lib, err := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	fn, err := purego.Dlsym(lib, "objc_msgSend")
	if err != nil {
		panic(err)
	}
	return fn
}()

func ObjcSend0(id objc.ID, sel objc.SEL) uintptr {
	r, _ := purego.SyscallN2(ObjcMsgSend, uintptr(id), uintptr(sel))
	return r
}

//go:uintptrescapes
func ObjcSend1(id objc.ID, sel objc.SEL, a0 uintptr) uintptr {
	r, _ := purego.SyscallN3(ObjcMsgSend, uintptr(id), uintptr(sel), a0)
	return r
}

//go:uintptrescapes
func ObjcSend2(id objc.ID, sel objc.SEL, a0, a1 uintptr) uintptr {
	r, _ := purego.SyscallN4(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1)
	return r
}

//go:uintptrescapes
func ObjcSend3(id objc.ID, sel objc.SEL, a0, a1, a2 uintptr) uintptr {
	r, _ := purego.SyscallN5(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1, a2)
	return r
}

//go:uintptrescapes
func ObjcSend4(id objc.ID, sel objc.SEL, a0, a1, a2, a3 uintptr) uintptr {
	r, _ := purego.SyscallN6(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1, a2, a3)
	return r
}

//go:uintptrescapes
func ObjcSend5(id objc.ID, sel objc.SEL, a0, a1, a2, a3, a4 uintptr) uintptr {
	r, _ := purego.SyscallN7(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1, a2, a3, a4)
	return r
}

func ObjcSendFloat64(id objc.ID, sel objc.SEL) float64 {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	_, f1, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return math.Float64frombits(uint64(f1))
}

func ObjcSendNSPoint(id objc.ID, sel objc.SEL) NSPoint {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	_, f1, f2, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return NSPoint{
		X: math.Float64frombits(uint64(f1)),
		Y: math.Float64frombits(uint64(f2)),
	}
}

func ObjcSendNSRect(id objc.ID, sel objc.SEL) NSRect {
	if !rectInFloatRegs {
		return objc.Send[NSRect](id, sel)
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	_, f1, f2, f3, f4 := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return rectFromFloats(f1, f2, f3, f4)
}

// rectInFloatRegs reports whether an NSRect (four float64s, 32 bytes) is passed and
// returned in float registers. That holds for arm64 (HFA in d0-d3) but not amd64,
// where a struct over 16 bytes goes through memory; there the NSRect helpers fall
// back to objc.Send, which is correct but allocates.
const rectInFloatRegs = runtime.GOARCH == "arm64"

func rectFromFloats(f1, f2, f3, f4 uintptr) NSRect {
	return NSRect{
		Origin: NSPoint{X: math.Float64frombits(uint64(f1)), Y: math.Float64frombits(uint64(f2))},
		Size:   NSSize{Width: math.Float64frombits(uint64(f3)), Height: math.Float64frombits(uint64(f4))},
	}
}

func rectToFloats(r NSRect) (f0, f1, f2, f3 uintptr) {
	return uintptr(math.Float64bits(r.Origin.X)),
		uintptr(math.Float64bits(r.Origin.Y)),
		uintptr(math.Float64bits(r.Size.Width)),
		uintptr(math.Float64bits(r.Size.Height))
}

// ObjcSendNSRectRect returns NSRect from a method that takes an NSRect arg.
func ObjcSendNSRectRect(id objc.ID, sel objc.SEL, r NSRect) NSRect {
	if !rectInFloatRegs {
		return objc.Send[NSRect](id, sel, r)
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0], floatArgs[1], floatArgs[2], floatArgs[3] = rectToFloats(r)
	_, f1, f2, f3, f4 := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return rectFromFloats(f1, f2, f3, f4)
}

// ObjcSendNSRectRectInt returns NSRect from a method with NSRect + uintptr args.
//
//go:uintptrescapes
func ObjcSendNSRectRectInt(id objc.ID, sel objc.SEL, r NSRect, a0 uintptr) NSRect {
	if !rectInFloatRegs {
		return objc.Send[NSRect](id, sel, r, a0)
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	intArgs[2] = a0
	floatArgs[0], floatArgs[1], floatArgs[2], floatArgs[3] = rectToFloats(r)
	_, f1, f2, f3, f4 := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return rectFromFloats(f1, f2, f3, f4)
}

// ObjcSendNSPointPoint returns NSPoint from a method that takes an NSPoint arg.
func ObjcSendNSPointPoint(id objc.ID, sel objc.SEL, p NSPoint) NSPoint {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(p.X))
	floatArgs[1] = uintptr(math.Float64bits(p.Y))
	_, f1, f2, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return NSPoint{
		X: math.Float64frombits(uint64(f1)),
		Y: math.Float64frombits(uint64(f2)),
	}
}

// ObjcSendBoolPointRect returns bool from a method with NSPoint + NSRect args.
func ObjcSendBoolPointRect(id objc.ID, sel objc.SEL, p NSPoint, r NSRect) bool {
	if !rectInFloatRegs {
		return objc.Send[bool](id, sel, p, r)
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(p.X))
	floatArgs[1] = uintptr(math.Float64bits(p.Y))
	floatArgs[2] = uintptr(math.Float64bits(r.Origin.X))
	floatArgs[3] = uintptr(math.Float64bits(r.Origin.Y))
	floatArgs[4] = uintptr(math.Float64bits(r.Size.Width))
	floatArgs[5] = uintptr(math.Float64bits(r.Size.Height))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1 != 0
}

// ObjcSendSize calls a method with an NSSize arg (HFA: 2 float64 in float regs).
func ObjcSendSize(id objc.ID, sel objc.SEL, s NSSize) {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(s.Width))
	floatArgs[1] = uintptr(math.Float64bits(s.Height))
	purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
}

// ObjcSendPoint calls a method with an NSPoint arg (HFA: 2 float64 in float regs).
func ObjcSendPoint(id objc.ID, sel objc.SEL, p NSPoint) {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(p.X))
	floatArgs[1] = uintptr(math.Float64bits(p.Y))
	purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
}

// ObjcSendFloat64Arg calls a method with a float64 arg.
func ObjcSendFloat64Arg(id objc.ID, sel objc.SEL, f float64) {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(f))
	purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
}

// ObjcSendFloat64ArgRet calls a method with a float64 arg and returns uintptr.
func ObjcSendFloat64ArgRet(id objc.ID, sel objc.SEL, f float64) uintptr {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(f))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendRectBool calls a method with NSRect + bool args (e.g. setFrame:display:).
func ObjcSendRectBool(id objc.ID, sel objc.SEL, r NSRect, b bool) {
	if !rectInFloatRegs {
		id.Send(sel, r, b)
		return
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	if b {
		intArgs[2] = 1
	}
	floatArgs[0], floatArgs[1], floatArgs[2], floatArgs[3] = rectToFloats(r)
	purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
}

// ObjcSendRectIntIntBool calls a method with NSRect + uintptr + uintptr + bool args.
//
//go:uintptrescapes
func ObjcSendRectIntIntBool(id objc.ID, sel objc.SEL, r NSRect, a0, a1 uintptr, b bool) uintptr {
	if !rectInFloatRegs {
		return uintptr(id.Send(sel, r, a0, a1, b))
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	intArgs[2] = a0
	intArgs[3] = a1
	if b {
		intArgs[4] = 1
	}
	floatArgs[0], floatArgs[1], floatArgs[2], floatArgs[3] = rectToFloats(r)
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendIntPointInt returns uintptr from a method with NSPoint + uintptr args.
//
//go:uintptrescapes
func ObjcSendIntPointInt(id objc.ID, sel objc.SEL, p NSPoint, a0 uintptr) uintptr {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	intArgs[2] = a0
	floatArgs[0] = uintptr(math.Float64bits(p.X))
	floatArgs[1] = uintptr(math.Float64bits(p.Y))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendRectIntIDInt calls a method with NSRect + uintptr + ID + uintptr args.
//
//go:uintptrescapes
func ObjcSendRectIntIDInt(id objc.ID, sel objc.SEL, r NSRect, a0 uintptr, a1 objc.ID, a2 uintptr) uintptr {
	if !rectInFloatRegs {
		return uintptr(id.Send(sel, r, a0, a1, a2))
	}
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	intArgs[2] = a0
	intArgs[3] = uintptr(a1)
	intArgs[4] = a2
	floatArgs[0], floatArgs[1], floatArgs[2], floatArgs[3] = rectToFloats(r)
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendFloat4 calls a method with 4 float64 args (e.g. colorWithSRGBRed:green:blue:alpha:).
func ObjcSendFloat4(id objc.ID, sel objc.SEL, f0, f1a, f2a, f3a float64) uintptr {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(f0))
	floatArgs[1] = uintptr(math.Float64bits(f1a))
	floatArgs[2] = uintptr(math.Float64bits(f2a))
	floatArgs[3] = uintptr(math.Float64bits(f3a))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendPointSizeFloat calls a method with NSPoint + NSSize + float64 args.
func ObjcSendPointSizeFloat(id objc.ID, sel objc.SEL, p NSPoint, s NSSize, f float64) uintptr {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(p.X))
	floatArgs[1] = uintptr(math.Float64bits(p.Y))
	floatArgs[2] = uintptr(math.Float64bits(s.Width))
	floatArgs[3] = uintptr(math.Float64bits(s.Height))
	floatArgs[4] = uintptr(math.Float64bits(f))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendSizeRet calls a method with an NSSize arg and returns uintptr (e.g. initWithSize:).
func ObjcSendSizeRet(id objc.ID, sel objc.SEL, s NSSize) uintptr {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	floatArgs[0] = uintptr(math.Float64bits(s.Width))
	floatArgs[1] = uintptr(math.Float64bits(s.Height))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

// ObjcSendIDPoint calls a method with an objc.ID + NSPoint args and returns uintptr (e.g. initWithImage:hotSpot:).
func ObjcSendIDPoint(id objc.ID, sel objc.SEL, a0 objc.ID, p NSPoint) uintptr {
	var intArgs, floatArgs [8]uintptr
	intArgs[0] = uintptr(id)
	intArgs[1] = uintptr(sel)
	intArgs[2] = uintptr(a0)
	floatArgs[0] = uintptr(math.Float64bits(p.X))
	floatArgs[1] = uintptr(math.Float64bits(p.Y))
	r1, _, _, _, _ := purego.SyscallNMixed(ObjcMsgSend, &intArgs, &floatArgs)
	return r1
}

func BoolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
