// Copyright 2026 The Ebitengine Authors
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

// objcMsgSendStret is objc_msgSend_stret, which returns a struct in memory on amd64.
// It does not exist on arm64, where such structs are returned through x8 instead.
var objcMsgSendStret = func() uintptr {
	if runtime.GOARCH != "amd64" {
		return 0
	}
	lib, err := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	fn, err := purego.Dlsym(lib, "objc_msgSend_stret")
	if err != nil {
		panic(err)
	}
	return fn
}()

func ObjcSend0(id objc.ID, sel objc.SEL) uintptr {
	r, _, _ := purego.Syscall2(ObjcMsgSend, uintptr(id), uintptr(sel))
	return r
}

//go:uintptrescapes
func ObjcSend1(id objc.ID, sel objc.SEL, a0 uintptr) uintptr {
	r, _, _ := purego.Syscall3(ObjcMsgSend, uintptr(id), uintptr(sel), a0)
	return r
}

//go:uintptrescapes
func ObjcSend2(id objc.ID, sel objc.SEL, a0, a1 uintptr) uintptr {
	r, _, _ := purego.Syscall4(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1)
	return r
}

//go:uintptrescapes
func ObjcSend3(id objc.ID, sel objc.SEL, a0, a1, a2 uintptr) uintptr {
	r, _, _ := purego.Syscall5(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1, a2)
	return r
}

//go:uintptrescapes
func ObjcSend4(id objc.ID, sel objc.SEL, a0, a1, a2, a3 uintptr) uintptr {
	r, _, _ := purego.Syscall6(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1, a2, a3)
	return r
}

//go:uintptrescapes
func ObjcSend5(id objc.ID, sel objc.SEL, a0, a1, a2, a3, a4 uintptr) uintptr {
	r, _, _ := purego.Syscall7(ObjcMsgSend, uintptr(id), uintptr(sel), a0, a1, a2, a3, a4)
	return r
}

// The helpers below call objc_msgSend with float and struct arguments through purego.SyscallMixed.
//
// NSPoint and NSSize (two float64s) go in two float registers on both arm64 and amd64.
// NSRect (four float64s, 32 bytes) differs: on arm64 it is an HFA passed and returned in four
// float registers, while on amd64 it is passed in memory, in the first stack slots, and returned
// through objc_msgSend_stret with a hidden result pointer as the first integer argument.

// isAMD64 reports whether NSRect is passed in memory rather than in float registers.
const isAMD64 = runtime.GOARCH == "amd64"

// rectRecv is the index in the integer arguments of the receiver of a method that returns an
// NSRect: on amd64 the hidden result pointer comes first.
var rectRecv = boolInt(isAMD64)

// stackSlot is the index in the integer arguments of the first stack slot on amd64.
const stackSlot = 6

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type msgArgs struct {
	ints   [16]uintptr
	floats [8]uintptr
}

func (a *msgArgs) setFloat(i int, f float64) {
	a.floats[i] = uintptr(math.Float64bits(f))
}

// setRect places an NSRect argument: in floats[i:i+4] on arm64, or in the first stack slots on amd64.
func (a *msgArgs) setRect(i int, r NSRect) {
	words := [4]uintptr{
		uintptr(math.Float64bits(r.Origin.X)),
		uintptr(math.Float64bits(r.Origin.Y)),
		uintptr(math.Float64bits(r.Size.Width)),
		uintptr(math.Float64bits(r.Size.Height)),
	}
	if isAMD64 {
		copy(a.ints[stackSlot:], words[:])
		return
	}
	copy(a.floats[i:], words[:])
}

func (a *msgArgs) send() (r1, f1, f2 uintptr) {
	r1, f1, f2, _, _ = purego.SyscallMixed(ObjcMsgSend, &a.ints, &a.floats)
	return
}

// sendRect calls a method that returns an NSRect. The receiver must be at ints[rectRecv].
func (a *msgArgs) sendRect() NSRect {
	var w [4]uintptr
	if isAMD64 {
		w = purego.SyscallMixedStret(objcMsgSendStret, &a.ints, &a.floats)
	} else {
		_, w[0], w[1], w[2], w[3] = purego.SyscallMixed(ObjcMsgSend, &a.ints, &a.floats)
	}
	return NSRect{
		Origin: NSPoint{X: math.Float64frombits(uint64(w[0])), Y: math.Float64frombits(uint64(w[1]))},
		Size:   NSSize{Width: math.Float64frombits(uint64(w[2])), Height: math.Float64frombits(uint64(w[3]))},
	}
}

func pointFromWords(f1, f2 uintptr) NSPoint {
	return NSPoint{X: math.Float64frombits(uint64(f1)), Y: math.Float64frombits(uint64(f2))}
}

// ObjcSendBool calls a method that returns BOOL. Only the low byte of the result register is defined.
func ObjcSendBool(id objc.ID, sel objc.SEL) bool {
	return ObjcSend0(id, sel)&0xff != 0
}

// ObjcSendBool1 calls a method with one integer argument that returns BOOL.
//
//go:uintptrescapes
func ObjcSendBool1(id objc.ID, sel objc.SEL, a0 uintptr) bool {
	return ObjcSend1(id, sel, a0)&0xff != 0
}

func ObjcSendFloat64(id objc.ID, sel objc.SEL) float64 {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	_, f1, _ := a.send()
	return math.Float64frombits(uint64(f1))
}

func ObjcSendNSPoint(id objc.ID, sel objc.SEL) NSPoint {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	_, f1, f2 := a.send()
	return pointFromWords(f1, f2)
}

func ObjcSendNSRect(id objc.ID, sel objc.SEL) NSRect {
	var a msgArgs
	a.ints[rectRecv], a.ints[rectRecv+1] = uintptr(id), uintptr(sel)
	return a.sendRect()
}

// ObjcSendNSRectRect returns NSRect from a method that takes an NSRect arg.
func ObjcSendNSRectRect(id objc.ID, sel objc.SEL, r NSRect) NSRect {
	var a msgArgs
	a.ints[rectRecv], a.ints[rectRecv+1] = uintptr(id), uintptr(sel)
	a.setRect(0, r)
	return a.sendRect()
}

// ObjcSendNSRectRectInt returns NSRect from a method with NSRect + uintptr args.
//
//go:uintptrescapes
func ObjcSendNSRectRectInt(id objc.ID, sel objc.SEL, r NSRect, a0 uintptr) NSRect {
	var a msgArgs
	a.ints[rectRecv], a.ints[rectRecv+1], a.ints[rectRecv+2] = uintptr(id), uintptr(sel), a0
	a.setRect(0, r)
	return a.sendRect()
}

// ObjcSendNSPointPoint returns NSPoint from a method that takes an NSPoint arg.
func ObjcSendNSPointPoint(id objc.ID, sel objc.SEL, p NSPoint) NSPoint {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, p.X)
	a.setFloat(1, p.Y)
	_, f1, f2 := a.send()
	return pointFromWords(f1, f2)
}

// ObjcSendBoolPointRect returns bool from a method with NSPoint + NSRect args.
func ObjcSendBoolPointRect(id objc.ID, sel objc.SEL, p NSPoint, r NSRect) bool {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, p.X)
	a.setFloat(1, p.Y)
	a.setRect(2, r)
	r1, _, _ := a.send()
	return r1&0xff != 0
}

// ObjcSendSize calls a method with an NSSize arg.
func ObjcSendSize(id objc.ID, sel objc.SEL, s NSSize) {
	ObjcSendSizeRet(id, sel, s)
}

// ObjcSendPoint calls a method with an NSPoint arg.
func ObjcSendPoint(id objc.ID, sel objc.SEL, p NSPoint) {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, p.X)
	a.setFloat(1, p.Y)
	a.send()
}

// ObjcSendFloat64Arg calls a method with a float64 arg.
func ObjcSendFloat64Arg(id objc.ID, sel objc.SEL, f float64) {
	ObjcSendFloat64ArgRet(id, sel, f)
}

// ObjcSendFloat64ArgRet calls a method with a float64 arg and returns uintptr.
func ObjcSendFloat64ArgRet(id objc.ID, sel objc.SEL, f float64) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, f)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendRectBool calls a method with NSRect + bool args (e.g. setFrame:display:).
func ObjcSendRectBool(id objc.ID, sel objc.SEL, r NSRect, b bool) {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel), BoolToUintptr(b)}}
	a.setRect(0, r)
	a.send()
}

// ObjcSendRectIntIntBool calls a method with NSRect + uintptr + uintptr + bool args.
//
//go:uintptrescapes
func ObjcSendRectIntIntBool(id objc.ID, sel objc.SEL, r NSRect, a0, a1 uintptr, b bool) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel), a0, a1, BoolToUintptr(b)}}
	a.setRect(0, r)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendIntPointInt returns uintptr from a method with NSPoint + uintptr args.
//
//go:uintptrescapes
func ObjcSendIntPointInt(id objc.ID, sel objc.SEL, p NSPoint, a0 uintptr) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel), a0}}
	a.setFloat(0, p.X)
	a.setFloat(1, p.Y)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendRectIntIDInt calls a method with NSRect + uintptr + ID + uintptr args.
//
//go:uintptrescapes
func ObjcSendRectIntIDInt(id objc.ID, sel objc.SEL, r NSRect, a0 uintptr, a1 objc.ID, a2 uintptr) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel), a0, uintptr(a1), a2}}
	a.setRect(0, r)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendFloat4 calls a method with 4 float64 args (e.g. colorWithSRGBRed:green:blue:alpha:).
func ObjcSendFloat4(id objc.ID, sel objc.SEL, f0, f1, f2, f3 float64) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, f0)
	a.setFloat(1, f1)
	a.setFloat(2, f2)
	a.setFloat(3, f3)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendStack calls a method whose only argument is a struct passed in memory on amd64
// (e.g. MTLViewport, MTLScissorRect, MTLClearColor), given as the struct's words.
func ObjcSendStack(id objc.ID, sel objc.SEL, words ...uintptr) {
	if !isAMD64 {
		panic("cocoa: ObjcSendStack is only for amd64")
	}
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	if copy(a.ints[stackSlot:], words) != len(words) {
		panic("cocoa: too many words for ObjcSendStack")
	}
	a.send()
}

// ObjcSendPointSizeFloat calls a method with NSPoint + NSSize + float64 args.
func ObjcSendPointSizeFloat(id objc.ID, sel objc.SEL, p NSPoint, s NSSize, f float64) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, p.X)
	a.setFloat(1, p.Y)
	a.setFloat(2, s.Width)
	a.setFloat(3, s.Height)
	a.setFloat(4, f)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendSizeRet calls a method with an NSSize arg and returns uintptr (e.g. initWithSize:).
func ObjcSendSizeRet(id objc.ID, sel objc.SEL, s NSSize) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel)}}
	a.setFloat(0, s.Width)
	a.setFloat(1, s.Height)
	r1, _, _ := a.send()
	return r1
}

// ObjcSendIDPoint calls a method with an objc.ID + NSPoint args and returns uintptr (e.g. initWithImage:hotSpot:).
func ObjcSendIDPoint(id objc.ID, sel objc.SEL, a0 objc.ID, p NSPoint) uintptr {
	a := msgArgs{ints: [16]uintptr{uintptr(id), uintptr(sel), uintptr(a0)}}
	a.setFloat(0, p.X)
	a.setFloat(1, p.Y)
	r1, _, _ := a.send()
	return r1
}

func BoolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
