// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2002-2006 Marcus Geelnard
// SPDX-FileCopyrightText: 2006-2019 Camilla Löwy <elmindreda@glfw.org>
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package glfw

import (
	"fmt"
	"image"
	"math"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/objc"

	"github.com/hajimehoshi/ebiten/v2/internal/cocoa"
	"github.com/hajimehoshi/ebiten/v2/internal/objcutil"
)

// NSPasteboardType strings.
var (
	nsPasteboardTypeString = cocoa.NSString_alloc().InitWithUTF8String("public.utf8-plain-text")
)

// NSDefaultRunLoopMode for event polling.
var nsDefaultRunLoopMode = cocoa.NSString_alloc().InitWithUTF8String("kCFRunLoopDefaultMode")

// Color space name for custom cursor creation.
var nsCalibratedRGBColorSpace = cocoa.NSString_alloc().InitWithUTF8String("NSCalibratedRGBColorSpace")

// Registered ObjC class references.
var (
	class_GLFWWindow         objc.Class
	class_GLFWWindowDelegate objc.Class
	class_GLFWContentView    objc.Class
)

// translateKey converts a macOS virtual key code to a GLFW key constant.
func translateKey(keyCode uint16) Key {
	if int(keyCode) >= len(_glfw.platformWindow.keycodes) {
		return KeyUnknown
	}
	return _glfw.platformWindow.keycodes[keyCode]
}

// translateFlags converts NSEvent modifier flags to GLFW modifier keys.
func translateFlags(flags uintptr) ModifierKey {
	var mods ModifierKey
	if flags&NSEventModifierFlagShift != 0 {
		mods |= ModShift
	}
	if flags&NSEventModifierFlagControl != 0 {
		mods |= ModControl
	}
	if flags&NSEventModifierFlagOption != 0 {
		mods |= ModAlt
	}
	if flags&NSEventModifierFlagCommand != 0 {
		mods |= ModSuper
	}
	if flags&NSEventModifierFlagCapsLock != 0 {
		mods |= ModCapsLock
	}
	return mods
}

// translateKeyToModifierFlag maps a GLFW key to its NSEvent modifier flag.
func translateKeyToModifierFlag(key Key) uintptr {
	switch key {
	case KeyLeftShift, KeyRightShift:
		return NSEventModifierFlagShift
	case KeyLeftControl, KeyRightControl:
		return NSEventModifierFlagControl
	case KeyLeftAlt, KeyRightAlt:
		return NSEventModifierFlagOption
	case KeyLeftSuper, KeyRightSuper:
		return NSEventModifierFlagCommand
	case KeyCapsLock:
		return NSEventModifierFlagCapsLock
	default:
		return 0
	}
}

// cursorInContentArea checks whether the mouse cursor is within a window's content area.
func cursorInContentArea(window *Window) bool {
	if window.platform.object == 0 {
		return false
	}

	pos := cocoa.ObjcSendNSPoint(window.platform.object, sel_mouseLocationOutsideOfEventStream)
	return cocoa.ObjcSendBoolPointRect(window.platform.view, sel_mouse_inRect, pos, cocoa.ObjcSendNSRect(window.platform.view, sel_frame))
}

// hideCursor hides the system cursor and optionally disables mouse/cursor association.
func hideCursor(window *Window) {
	if !_glfw.platformWindow.cursorHidden {
		cocoa.ObjcSend0(objc.ID(class_NSCursor), objc.RegisterName("hide"))
		_glfw.platformWindow.cursorHidden = true
	}
}

// showCursor shows the system cursor and re-enables mouse/cursor association.
func showCursor(window *Window) {
	if _glfw.platformWindow.cursorHidden {
		cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_unhide)
		_glfw.platformWindow.cursorHidden = false
	}
}

// updateCursorImage sets the appropriate cursor image based on cursor mode.
func updateCursorImage(window *Window) {
	if window.cursorMode == CursorNormal {
		showCursor(window)
		if window.cursor != nil && window.cursor.platform.object != 0 {
			cocoa.ObjcSend0(window.cursor.platform.object, sel_set)
		} else {
			cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_arrowCursor)), sel_set)
		}
	} else {
		hideCursor(window)
	}
}

// updateCursorMode applies cursor mode changes.
func updateCursorMode(window *Window) error {
	if window.cursorMode == CursorDisabled {
		_glfw.platformWindow.disabledCursorWindow = window
		_glfw.platformWindow.restoreCursorPosX, _glfw.platformWindow.restoreCursorPosY, _ = window.platformGetCursorPos()
		if err := window.centerCursorInContentArea(); err != nil {
			return err
		}
		cgAssociateMouseAndMouseCursorPosition(0)
	} else if _glfw.platformWindow.disabledCursorWindow == window {
		_glfw.platformWindow.disabledCursorWindow = nil
		if err := window.platformSetCursorPos(_glfw.platformWindow.restoreCursorPosX, _glfw.platformWindow.restoreCursorPosY); err != nil {
			return err
		}
		// NOTE: The matching CGAssociateMouseAndMouseCursorPosition call is
		//       made in platformSetCursorPos as part of a workaround
	}

	if cursorInContentArea(window) {
		updateCursorImage(window)
	}
	return nil
}

// windowForEvent finds the Window associated with a native NSWindow ID.
func windowForEvent(nsWindow objc.ID) *Window {
	for _, w := range _glfw.windows {
		if w.platform.object == nsWindow {
			return w
		}
	}
	return nil
}

// nsApp returns the shared NSApplication instance.
func nsApp() objc.ID {
	return objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
}

// registerGLFWClasses registers the GLFWWindow, GLFWWindowDelegate, and GLFWContentView
// ObjC classes. Called from platformInit.
func registerGLFWClasses() error {
	// GLFWWindow — NSWindow subclass.
	var err error
	class_GLFWWindow, err = objc.RegisterClass(
		"GLFWWindow",
		objc.GetClass("NSWindow"),
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("canBecomeKeyWindow"),
				Fn: func(_ objc.ID, _ objc.SEL) bool {
					return true
				},
			},
			{
				Cmd: objc.RegisterName("canBecomeMainWindow"),
				Fn: func(_ objc.ID, _ objc.SEL) bool {
					return true
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("glfw: failed to register GLFWWindow class: %w", err)
	}

	// GLFWWindowDelegate — NSObject subclass conforming to NSWindowDelegate.
	class_GLFWWindowDelegate, err = objc.RegisterClass(
		"GLFWWindowDelegate",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{objc.GetProtocol("NSWindowDelegate")},
		nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("windowShouldClose:"),
				Fn: func(self objc.ID, _ objc.SEL, sender objc.ID) bool {
					window := getGoWindow(self)
					if window == nil {
						return false
					}
					window.inputWindowCloseRequest()
					return false
				},
			},
			{
				Cmd: objc.RegisterName("windowDidResize:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}

					if window.context.source == NativeContextAPI {
						cocoa.ObjcSend0(window.context.platform.object, objc.RegisterName("update"))
					}

					if _glfw.platformWindow.disabledCursorWindow == window {
						_ = window.centerCursorInContentArea()
					}

					maximized := cocoa.ObjcSendBool(window.platform.object, sel_isZoomed)
					if window.platform.maximized != maximized {
						window.platform.maximized = maximized
						window.inputWindowMaximize(maximized)
					}

					updateWindowSize(window)
				},
			},
			{
				Cmd: objc.RegisterName("windowDidMove:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.platform.object == 0 {
						return
					}

					if window.context.source == NativeContextAPI {
						cocoa.ObjcSend0(window.context.platform.object, objc.RegisterName("update"))
					}

					if _glfw.platformWindow.disabledCursorWindow == window {
						_ = window.centerCursorInContentArea()
					}

					frame := cocoa.ObjcSendNSRect(window.platform.object, sel_frame)
					contentRect := cocoa.ObjcSendNSRectRect(window.platform.object, sel_contentRectForFrameRect, frame)
					xpos := int(contentRect.Origin.X)
					ypos := int(transformYNS(float32(contentRect.Origin.Y + contentRect.Size.Height - 1)))
					window.inputWindowPos(xpos, ypos)
				},
			},
			{
				Cmd: objc.RegisterName("windowDidMiniaturize:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.monitor != nil {
						_ = window.releaseMonitor()
					}
					window.inputWindowIconify(true)
				},
			},
			{
				Cmd: objc.RegisterName("windowDidDeminiaturize:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.monitor != nil {
						_ = window.acquireMonitor()
					}
					window.inputWindowIconify(false)
				},
			},
			{
				Cmd: objc.RegisterName("windowDidBecomeKey:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if _glfw.platformWindow.disabledCursorWindow == window {
						_ = window.centerCursorInContentArea()
					}
					window.inputWindowFocus(true)
					_ = updateCursorMode(window)
				},
			},
			{
				Cmd: objc.RegisterName("windowDidResignKey:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.monitor != nil && window.autoIconify {
						window.platformIconifyWindow()
					}
					window.inputWindowFocus(false)
				},
			},
			{
				Cmd: objc.RegisterName("windowDidChangeOcclusionState:"),
				Fn: func(self objc.ID, _ objc.SEL, notification objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.platform.object == 0 {
						return
					}
					state := cocoa.ObjcSend0(window.platform.object, sel_occlusionState)
					window.platform.occluded = (state & NSWindowOcclusionStateVisible) == 0
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("glfw: failed to register GLFWWindowDelegate class: %w", err)
	}

	// GLFWContentView — NSView subclass.
	class_GLFWContentView, err = objc.RegisterClass(
		"GLFWContentView",
		objc.GetClass("NSView"),
		[]*objc.Protocol{objc.GetProtocol("NSTextInputClient")},
		nil,
		[]objc.MethodDef{
			// Mouse button events.
			{
				Cmd: objc.RegisterName("mouseDown:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					window.inputMouseClick(MouseButton1, Press, translateFlags(flags))
				},
			},
			{
				Cmd: objc.RegisterName("mouseUp:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					window.inputMouseClick(MouseButton1, Release, translateFlags(flags))
				},
			},
			{
				Cmd: objc.RegisterName("rightMouseDown:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					window.inputMouseClick(MouseButton2, Press, translateFlags(flags))
				},
			},
			{
				Cmd: objc.RegisterName("rightMouseUp:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					window.inputMouseClick(MouseButton2, Release, translateFlags(flags))
				},
			},
			{
				Cmd: objc.RegisterName("otherMouseDown:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					button := MouseButton(cocoa.ObjcSend0(event, sel_buttonNumber))
					window.inputMouseClick(button, Press, translateFlags(flags))
				},
			},
			{
				Cmd: objc.RegisterName("otherMouseUp:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					button := MouseButton(cocoa.ObjcSend0(event, sel_buttonNumber))
					window.inputMouseClick(button, Release, translateFlags(flags))
				},
			},
			// Mouse move events.
			{
				Cmd: objc.RegisterName("mouseMoved:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					handleMouseMoved(window, event)
				},
			},
			{
				Cmd: objc.RegisterName("mouseDragged:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					handleMouseMoved(window, event)
				},
			},
			{
				Cmd: objc.RegisterName("rightMouseDragged:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					handleMouseMoved(window, event)
				},
			},
			{
				Cmd: objc.RegisterName("otherMouseDragged:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					handleMouseMoved(window, event)
				},
			},
			{
				Cmd: objc.RegisterName("mouseExited:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.cursorMode == CursorHidden {
						showCursor(window)
					}
					window.inputCursorEnter(false)
				},
			},
			{
				Cmd: objc.RegisterName("mouseEntered:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.cursorMode == CursorHidden {
						hideCursor(window)
					}
					window.inputCursorEnter(true)
				},
			},
			// Keyboard events.
			{
				Cmd: objc.RegisterName("keyDown:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					keyCode := uint16(cocoa.ObjcSend0(event, sel_keyCode))
					key := translateKey(keyCode)
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					mods := translateFlags(flags)

					window.inputKey(key, int(keyCode), Press, mods)

					// Interpret key events for text input.
					eventArray := objc.ID(cocoa.ObjcSend1(objc.ID(class_NSArray), sel_arrayWithObject, uintptr(event)))
					cocoa.ObjcSend1(self, sel_interpretKeyEvents, uintptr(eventArray))
				},
			},
			{
				Cmd: sel_performKeyEquivalent,
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) bool {
					window := getGoWindow(self)
					if window == nil {
						return objcutil.SendSuper[bool](self, class_GLFWContentView, sel_performKeyEquivalent, event)
					}
					keyCode := uint16(cocoa.ObjcSend0(event, sel_keyCode))
					key := translateKey(keyCode)
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					mods := translateFlags(flags)

					// Some key combinations are dispatched as key equivalents and
					// consumed before they reach keyDown:. Claim them here and emit
					// them as key presses. Mirrors performKeyEquivalent: from upstream
					// GLFW, which added it after the 3.3 series this backend was ported from.
					if mods&ModControl != 0 && (key == KeyTab || key == KeyEscape) {
						window.inputKey(key, int(keyCode), Press, mods)
						return true
					}
					if mods&ModSuper != 0 && key == KeyPeriod {
						window.inputKey(key, int(keyCode), Press, mods)
						return true
					}

					return objcutil.SendSuper[bool](self, class_GLFWContentView, sel_performKeyEquivalent, event)
				},
			},
			{
				Cmd: objc.RegisterName("keyUp:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					keyCode := uint16(cocoa.ObjcSend0(event, sel_keyCode))
					key := translateKey(keyCode)
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					mods := translateFlags(flags)

					window.inputKey(key, int(keyCode), Release, mods)
				},
			},
			{
				Cmd: objc.RegisterName("flagsChanged:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					keyCode := uint16(cocoa.ObjcSend0(event, sel_keyCode))
					key := translateKey(keyCode)
					flags := cocoa.ObjcSend0(event, sel_modifierFlags) & NSEventModifierFlagDeviceIndependentFlagsMask
					mods := translateFlags(flags)

					modFlag := translateKeyToModifierFlag(key)
					var action Action
					if modFlag != 0 && (flags&modFlag) != 0 {
						if window.keys[key] == Press {
							action = Release
						} else {
							action = Press
						}
					} else {
						action = Release
					}

					window.inputKey(key, int(keyCode), action, mods)
				},
			},
			// Scroll wheel.
			{
				Cmd: objc.RegisterName("scrollWheel:"),
				Fn: func(self objc.ID, _ objc.SEL, event objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					deltaX := cocoa.ObjcSendFloat64(event, sel_scrollingDeltaX)
					deltaY := cocoa.ObjcSendFloat64(event, sel_scrollingDeltaY)

					// AppKit's contract for scrollingDeltaX/Y: with precise deltas the values are in
					// points, which are device-independent pixels; otherwise they are to be multiplied
					// by a line height.
					scrollDeltaX, scrollDeltaY := deltaX, deltaY
					unit := ScrollUnitLine
					if cocoa.ObjcSendBool(event, sel_hasPreciseScrollingDeltas) {
						unit = ScrollUnitPixel
						deltaX *= 0.1
						deltaY *= 0.1
					}

					if deltaX != 0 || deltaY != 0 {
						window.inputScroll(deltaX, deltaY, scrollDeltaX, scrollDeltaY, unit)
					}
				},
			},
			// View lifecycle.
			{
				Cmd: objc.RegisterName("viewDidChangeBackingProperties"),
				Fn: func(self objc.ID, _ objc.SEL) {
					window := getGoWindow(self)
					if window == nil {
						return
					}

					contentRect := cocoa.ObjcSendNSRect(window.platform.view, sel_frame)
					fbRect := cocoa.ObjcSendNSRectRect(window.platform.view, sel_convertRectToBacking, contentRect)
					xscale := float32(fbRect.Size.Width / contentRect.Size.Width)
					yscale := float32(fbRect.Size.Height / contentRect.Size.Height)

					if xscale != window.platform.xscale || yscale != window.platform.yscale {
						if window.platform.retina && window.platform.layer != 0 {
							cocoa.ObjcSendFloat64Arg(window.platform.layer, objc.RegisterName("setContentsScale:"),
								cocoa.ObjcSendFloat64(window.platform.object, sel_backingScaleFactor))
						}

						window.platform.xscale = xscale
						window.platform.yscale = yscale
						window.inputWindowContentScale(xscale, yscale)
					}

					if int(fbRect.Size.Width) != window.platform.fbWidth ||
						int(fbRect.Size.Height) != window.platform.fbHeight {
						window.platform.fbWidth = int(fbRect.Size.Width)
						window.platform.fbHeight = int(fbRect.Size.Height)
						window.inputFramebufferSize(int(fbRect.Size.Width), int(fbRect.Size.Height))
					}
				},
			},
			{
				Cmd: objc.RegisterName("updateTrackingAreas"),
				Fn: func(self objc.ID, _ objc.SEL) {
					// Remove all existing tracking areas.
					areas := objc.ID(cocoa.ObjcSend0(self, sel_trackingAreas))
					areaCount := int(cocoa.ObjcSend0(areas, sel_count))
					for i := range areaCount {
						area := objc.ID(cocoa.ObjcSend1(areas, sel_objectAtIndex, uintptr(i)))
						cocoa.ObjcSend1(self, sel_removeTrackingArea, uintptr(area))
					}

					// Create new tracking area.
					bounds := cocoa.ObjcSendNSRect(self, sel_bounds)
					options := uintptr(NSTrackingMouseEnteredAndExited |
						NSTrackingActiveInKeyWindow |
						NSTrackingEnabledDuringMouseDrag |
						NSTrackingCursorUpdate |
						NSTrackingInVisibleRect |
						NSTrackingAssumeInside)

					trackingAreaAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSTrackingArea), sel_alloc))
					trackingArea := objc.ID(cocoa.ObjcSendRectIntIDInt(trackingAreaAlloc,
						sel_initWithRect_options_owner_userInfo,
						bounds, options, self, 0))
					cocoa.ObjcSend1(self, sel_addTrackingArea, uintptr(trackingArea))
					// Balance the alloc; the view now holds the only reference.
					cocoa.ObjcSend0(trackingArea, sel_release)

					// Call super.
					objcutil.SendSuper[struct{}](self, class_GLFWContentView, objc.RegisterName("updateTrackingAreas"))
				},
			},
			{
				Cmd: objc.RegisterName("dealloc"),
				Fn: func(self objc.ID, _ objc.SEL) {
					window := getGoWindow(self)
					if window != nil && window.platform.markedText != 0 {
						cocoa.ObjcSend0(window.platform.markedText, sel_release)
						window.platform.markedText = 0
					}
					delete(theGoWindows, self)
					objcutil.SendSuper[struct{}](self, class_GLFWContentView, objc.RegisterName("dealloc"))
				},
			},
			{
				Cmd: objc.RegisterName("canBecomeKeyView"),
				Fn: func(_ objc.ID, _ objc.SEL) bool {
					return true
				},
			},
			{
				Cmd: objc.RegisterName("isOpaque"),
				Fn: func(self objc.ID, _ objc.SEL) bool {
					window := getGoWindow(self)
					if window == nil {
						return false
					}
					return cocoa.ObjcSendBool(window.platform.object, objc.RegisterName("isOpaque"))
				},
			},
			{
				Cmd: objc.RegisterName("acceptsFirstResponder"),
				Fn: func(_ objc.ID, _ objc.SEL) bool {
					return true
				},
			},
			{
				Cmd: objc.RegisterName("wantsUpdateLayer"),
				Fn: func(_ objc.ID, _ objc.SEL) bool {
					return true
				},
			},
			{
				Cmd: objc.RegisterName("updateLayer"),
				Fn: func(self objc.ID, _ objc.SEL) {
					window := getGoWindow(self)
					if window == nil {
						return
					}

					if window.context.source == NativeContextAPI {
						cocoa.ObjcSend0(window.context.platform.object, objc.RegisterName("update"))
					}

					window.inputWindowDamage()
				},
			},
			{
				Cmd: objc.RegisterName("cursorUpdate:"),
				Fn: func(self objc.ID, _ objc.SEL, _ objc.ID) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					updateCursorImage(window)
				},
			},
			{
				Cmd: objc.RegisterName("acceptsFirstMouse:"),
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) bool {
					return true
				},
			},
			// NSTextInputClient methods.
			{
				Cmd: sel_hasMarkedText,
				Fn: func(self objc.ID, _ objc.SEL) bool {
					window := getGoWindow(self)
					if window == nil {
						return false
					}
					if window.platform.markedText != 0 {
						return cocoa.ObjcSend0(window.platform.markedText, sel_length) > 0
					}
					return false
				},
			},
			{
				Cmd: sel_markedRange,
				Fn: func(self objc.ID, _ objc.SEL) nsRange {
					window := getGoWindow(self)
					if window != nil && window.platform.markedText != 0 {
						length := cocoa.ObjcSend0(window.platform.markedText, sel_length)
						if length > 0 {
							return nsRange{Location: 0, Length: length - 1}
						}
					}
					return nsRange{Location: uintptr(math.MaxInt), Length: 0} // NSNotFound
				},
			},
			{
				Cmd: sel_selectedRange,
				Fn: func(_ objc.ID, _ objc.SEL) nsRange {
					return nsRange{Location: uintptr(math.MaxInt), Length: 0} // NSNotFound
				},
			},
			{
				Cmd: sel_setMarkedText_selectedRange_replacementRange,
				Fn: func(self objc.ID, _ objc.SEL, str objc.ID, _ nsRange, _ nsRange) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.platform.markedText != 0 {
						cocoa.ObjcSend0(window.platform.markedText, sel_release)
					}
					if cocoa.ObjcSendBool1(str, sel_isKindOfClass, uintptr(class_NSAttributedString)) {
						alloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMutableAttributedString), sel_alloc))
						window.platform.markedText = objc.ID(cocoa.ObjcSend1(alloc, sel_initWithAttributedString, uintptr(str)))
					} else {
						alloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMutableAttributedString), sel_alloc))
						window.platform.markedText = objc.ID(cocoa.ObjcSend1(alloc, sel_initWithString, uintptr(str)))
					}
				},
			},
			{
				Cmd: sel_unmarkText,
				Fn: func(self objc.ID, _ objc.SEL) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					if window.platform.markedText != 0 {
						ms := objc.ID(cocoa.ObjcSend0(window.platform.markedText, sel_mutableString))
						emptyStr := cocoa.NSString_alloc().InitWithUTF8String("")
						cocoa.ObjcSend1(ms, sel_setString, uintptr(emptyStr.ID))
						cocoa.ObjcSend0(emptyStr.ID, sel_release)
					}
				},
			},
			{
				Cmd: sel_validAttributesForMarkedText,
				Fn: func(_ objc.ID, _ objc.SEL) objc.ID {
					// Return an empty autoreleased NSArray (matching C's [NSArray array]).
					return objc.ID(cocoa.ObjcSend0(objc.ID(class_NSArray), objc.RegisterName("array")))
				},
			},
			{
				Cmd: sel_attributedSubstringForProposedRange_actualRange,
				Fn: func(_ objc.ID, _ objc.SEL, _ nsRange, _ uintptr) objc.ID {
					return 0 // nil
				},
			},
			{
				Cmd: sel_insertText_replacementRange,
				Fn: func(self objc.ID, _ objc.SEL, text objc.ID, _ nsRange) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					nsApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
					event := objc.ID(cocoa.ObjcSend0(nsApp, sel_currentEvent))
					flags := cocoa.ObjcSend0(event, sel_modifierFlags)
					mods := translateFlags(flags)
					plain := mods&ModSuper == 0

					// Get the string from the text object.
					// The text parameter can be either NSString or NSAttributedString.
					characters := text
					if cocoa.ObjcSendBool1(text, sel_isKindOfClass, uintptr(class_NSAttributedString)) {
						characters = objc.ID(cocoa.ObjcSend0(text, sel_string))
					}
					str := cocoa.NSString{ID: characters}
					s := str.String()
					for _, ch := range s {
						if ch >= 0xf700 && ch <= 0xf7ff {
							continue
						}
						window.inputChar(ch, mods, plain)
					}
				},
			},
			{
				Cmd: sel_characterIndexForPoint,
				Fn: func(_ objc.ID, _ objc.SEL, _ cocoa.NSPoint) uintptr {
					return 0
				},
			},
			{
				Cmd: sel_firstRectForCharacterRange_actualRange,
				Fn: func(self objc.ID, _ objc.SEL, _ nsRange, _ uintptr) cocoa.NSRect {
					window := getGoWindow(self)
					if window == nil {
						return cocoa.NSRect{}
					}
					frame := cocoa.ObjcSendNSRect(window.platform.view, sel_frame)
					return cocoa.NSRect{
						Origin: frame.Origin,
						Size:   cocoa.CGSize{Width: 0, Height: 0},
					}
				},
			},
			{
				Cmd: sel_doCommandBySelector,
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.SEL) {
					// Do nothing.
				},
			},
			// Drawing.
			{
				Cmd: objc.RegisterName("drawRect:"),
				Fn: func(self objc.ID, _ objc.SEL, _ cocoa.NSRect) {
					window := getGoWindow(self)
					if window == nil {
						return
					}
					window.inputWindowDamage()
				},
			},
			// Drag and drop.
			{
				Cmd: objc.RegisterName("draggingEntered:"),
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) uintptr {
					return NSDragOperationGeneric
				},
			},
			{
				Cmd: objc.RegisterName("performDragOperation:"),
				Fn: func(self objc.ID, _ objc.SEL, sender objc.ID) bool {
					window := getGoWindow(self)
					if window == nil {
						return false
					}

					// Update the cursor position to the drop location.
					contentRect := cocoa.ObjcSendNSRect(window.platform.view, sel_frame)
					pos := cocoa.ObjcSendNSPoint(sender, objc.RegisterName("draggingLocation"))
					window.inputCursorPos(pos.X, contentRect.Size.Height-pos.Y)

					pasteboard := objc.ID(cocoa.ObjcSend0(sender, sel_draggingPasteboard))
					urlClass := objc.ID(class_NSURL)
					classes := objc.ID(cocoa.ObjcSend1(objc.ID(class_NSArray), sel_arrayWithObject, uintptr(urlClass)))

					// Filter to file URLs only.
					nsYes := objc.ID(cocoa.ObjcSend1(objc.ID(objc.GetClass("NSNumber")), objc.RegisterName("numberWithBool:"), uintptr(1)))
					options := objc.ID(cocoa.ObjcSend2(objc.ID(objc.GetClass("NSDictionary")),
						objc.RegisterName("dictionaryWithObject:forKey:"),
						uintptr(nsYes), uintptr(nsPasteboardURLReadingFileURLsOnlyKey)))

					urls := objc.ID(cocoa.ObjcSend2(pasteboard, sel_readObjectsForClasses_options, uintptr(classes), uintptr(options)))
					var urlCount int
					if urls != 0 {
						urlCount = int(cocoa.ObjcSend0(urls, sel_count))
					}

					if urlCount > 0 {
						paths := make([]string, urlCount)
						for i := range urlCount {
							url := objc.ID(cocoa.ObjcSend1(urls, sel_objectAtIndex, uintptr(i)))
							// Use fileSystemRepresentation instead of path to handle
							// HFS+ Unicode normalization correctly.
							fsRep := cocoa.ObjcSend0(url, objc.RegisterName("fileSystemRepresentation"))
							if fsRep != 0 {
								paths[i] = goStringFromCString(fsRep)
							}
						}

						window.inputDrop(paths)
					}

					return true
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("glfw: failed to register GLFWContentView class: %w", err)
	}

	return nil
}

// theGoWindows associates ObjC delegate and content-view instances with their Go Windows.
//
// theGoWindows must be accessed only from the main thread, like the rest of this package:
// the entries are written and removed while a window is created or destroyed, and read
// from the ObjC callbacks, which AppKit invokes on the thread that triggers them.
// Thus no synchronization is needed here.
var theGoWindows = map[objc.ID]*Window{}

// getGoWindow returns the Go Window associated with an ObjC instance, or nil if there is none.
//
// getGoWindow must be called from the main thread.
func getGoWindow(id objc.ID) *Window {
	return theGoWindows[id]
}

// setGoWindow associates an ObjC instance with a Go Window.
//
// setGoWindow must be called from the main thread.
func setGoWindow(id objc.ID, window *Window) {
	theGoWindows[id] = window
}

// handleMouseMoved processes mouse movement events.
func handleMouseMoved(window *Window, event objc.ID) {
	if window.cursorMode == CursorDisabled {
		dx := cocoa.ObjcSendFloat64(event, sel_deltaX)
		dy := cocoa.ObjcSendFloat64(event, sel_deltaY)

		dx -= window.platform.cursorWarpDeltaX
		dy -= window.platform.cursorWarpDeltaY

		window.inputCursorPos(
			window.virtualCursorPosX+dx,
			window.virtualCursorPosY+dy)
	} else {
		// Get the location in the content view.
		pos := cocoa.ObjcSendNSPoint(event, sel_locationInWindow)

		// Convert from Cocoa coordinates (origin at bottom-left) to GLFW coordinates (origin at top-left).
		contentRect := cocoa.ObjcSendNSRect(window.platform.view, sel_frame)
		pos.Y = contentRect.Size.Height - pos.Y

		window.inputCursorPos(pos.X, pos.Y)
	}

	window.platform.cursorWarpDeltaX = 0
	window.platform.cursorWarpDeltaY = 0
}

// updateWindowSize updates the cached window and framebuffer sizes, invoking callbacks as needed.
func updateWindowSize(window *Window) {
	if window.platform.object == 0 || window.platform.view == 0 {
		return
	}

	contentRect := cocoa.ObjcSendNSRect(window.platform.view, sel_frame)
	fbRect := cocoa.ObjcSendNSRectRect(window.platform.view, sel_convertRectToBacking, contentRect)

	fbWidth := int(fbRect.Size.Width)
	fbHeight := int(fbRect.Size.Height)

	if fbWidth != window.platform.fbWidth || fbHeight != window.platform.fbHeight {
		window.platform.fbWidth = fbWidth
		window.platform.fbHeight = fbHeight
		window.inputFramebufferSize(fbWidth, fbHeight)
	}

	width := int(contentRect.Size.Width)
	height := int(contentRect.Size.Height)

	if width != window.platform.width || height != window.platform.height {
		window.platform.width = width
		window.platform.height = height
		window.inputWindowSize(width, height)
	}
}

// createNativeWindow creates the actual NSWindow, delegate, and content view.
func createNativeWindow(window *Window, wndconfig *wndconfig, fbconfig_ *fbconfig) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// Create the delegate first (before the window, to avoid leaking the
	// window if delegate creation fails).
	delegateID := objc.ID(cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_GLFWWindowDelegate), sel_alloc)), sel_init))
	if delegateID == 0 {
		return fmt.Errorf("glfw: failed to create window delegate: %w", PlatformError)
	}
	setGoWindow(delegateID, window)
	window.platform.delegate = delegateID

	// Determine the content rect.
	var contentRect cocoa.NSRect
	if window.monitor != nil {
		mode, err := window.monitor.platformGetVideoMode()
		if err != nil {
			return err
		}
		xpos, ypos, _ := window.monitor.platformGetMonitorPos()
		contentRect = cocoa.NSRect{
			Origin: cocoa.NSPoint{X: float64(xpos), Y: float64(ypos)},
			Size:   cocoa.NSSize{Width: float64(mode.Width), Height: float64(mode.Height)},
		}
	} else {
		contentRect = cocoa.NSRect{
			Origin: cocoa.NSPoint{X: 0, Y: 0},
			Size:   cocoa.NSSize{Width: float64(wndconfig.width), Height: float64(wndconfig.height)},
		}
	}

	// Determine the style mask.
	styleMask := uintptr(NSWindowStyleMaskMiniaturizable)
	if window.monitor != nil || !wndconfig.decorated {
		styleMask |= NSWindowStyleMaskBorderless
	} else {
		styleMask |= NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
		if wndconfig.resizable {
			styleMask |= NSWindowStyleMaskResizable
		}
	}

	// Create the GLFWWindow instance.
	nsWindowAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_GLFWWindow), sel_alloc))
	nsWindow := objc.ID(cocoa.ObjcSendRectIntIntBool(nsWindowAlloc, sel_initWithContentRect_styleMask_backing_defer,
		contentRect, styleMask, uintptr(NSBackingStoreBuffered), false))
	if nsWindow == 0 {
		return fmt.Errorf("glfw: failed to create Cocoa window: %w", PlatformError)
	}

	window.platform.object = nsWindow

	if window.monitor != nil {
		cocoa.ObjcSend1(nsWindow, sel_setLevel, uintptr(NSMainMenuWindowLevel+1))
	} else {
		// Center the window on the screen.
		cocoa.ObjcSend0(nsWindow, objc.RegisterName("center"))
		cascadeIn := cocoa.NSPoint{
			X: _glfw.platformWindow.cascadePoint[0],
			Y: _glfw.platformWindow.cascadePoint[1],
		}
		cascadeOut := cocoa.ObjcSendNSPointPoint(nsWindow, objc.RegisterName("cascadeTopLeftFromPoint:"), cascadeIn)
		_glfw.platformWindow.cascadePoint[0] = cascadeOut.X
		_glfw.platformWindow.cascadePoint[1] = cascadeOut.Y

		if wndconfig.resizable {
			cocoa.ObjcSend1(nsWindow, sel_setCollectionBehavior, uintptr(_NSWindowCollectionBehaviorFullScreenPrimary|_NSWindowCollectionBehaviorManaged))
		} else {
			cocoa.ObjcSend1(nsWindow, sel_setCollectionBehavior, uintptr(_NSWindowCollectionBehaviorFullScreenNone))
		}

		if wndconfig.floating {
			cocoa.ObjcSend1(nsWindow, sel_setLevel, uintptr(NSFloatingWindowLevel))
		}

		if wndconfig.maximized {
			cocoa.ObjcSend1(nsWindow, sel_zoom, 0)
		}
	}

	if len(wndconfig.frameName) > 0 {
		name := cocoa.NSString_alloc().InitWithUTF8String(wndconfig.frameName)
		cocoa.ObjcSend1(nsWindow, sel_setFrameAutosaveName, uintptr(name.ID))
		cocoa.ObjcSend0(name.ID, sel_release)
	}

	// Create the content view.
	viewID := objc.ID(cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_GLFWContentView), sel_alloc)), sel_init))
	setGoWindow(viewID, window)
	window.platform.view = viewID
	window.platform.retina = wndconfig.retina
	window.platform.markedText = objc.ID(cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMutableAttributedString), sel_alloc)), objc.RegisterName("init")))

	// Handle transparent framebuffer.
	if fbconfig_.transparent {
		cocoa.ObjcSend1(nsWindow, sel_setOpaque, 0)
		cocoa.ObjcSend1(nsWindow, sel_setHasShadow, 0)
		cocoa.ObjcSend1(nsWindow, sel_setBackgroundColor, cocoa.ObjcSend0(objc.ID(class_NSColor), sel_clearColor))
	}

	cocoa.ObjcSend1(nsWindow, sel_setContentView, uintptr(viewID))
	cocoa.ObjcSend1(nsWindow, sel_makeFirstResponder, uintptr(viewID))
	titleStr := cocoa.NSString_alloc().InitWithUTF8String(wndconfig.title)
	cocoa.ObjcSend1(nsWindow, sel_setTitle, uintptr(titleStr.ID))
	cocoa.ObjcSend0(titleStr.ID, sel_release)
	cocoa.ObjcSend1(nsWindow, sel_setDelegate, uintptr(delegateID))
	cocoa.ObjcSend1(nsWindow, objc.RegisterName("setAcceptsMouseMovedEvents:"), uintptr(1))
	cocoa.ObjcSend1(nsWindow, sel_setRestorable, 0)

	// Disable window tabbing (macOS 10.12+).
	sel_setTabbingMode := objc.RegisterName("setTabbingMode:")
	if cocoa.ObjcSendBool1(nsWindow, objc.RegisterName("respondsToSelector:"), uintptr(sel_setTabbingMode)) {
		cocoa.ObjcSend1(nsWindow, sel_setTabbingMode, uintptr(2)) // NSWindowTabbingModeDisallowed = 2
	}

	cocoa.ObjcSend0(viewID, objc.RegisterName("updateTrackingAreas"))

	// Register for dragged types (URLs).
	typesArray := objc.ID(cocoa.ObjcSend1(objc.ID(class_NSArray), sel_arrayWithObject, uintptr(nsPasteboardTypeURL)))
	cocoa.ObjcSend1(viewID, sel_registerForDraggedTypes, uintptr(typesArray))

	// Update initial size cache.
	contentViewRect := cocoa.ObjcSendNSRect(viewID, sel_frame)
	window.platform.width = int(contentViewRect.Size.Width)
	window.platform.height = int(contentViewRect.Size.Height)

	fbRect := cocoa.ObjcSendNSRectRect(viewID, sel_convertRectToBacking, contentViewRect)
	window.platform.fbWidth = int(fbRect.Size.Width)
	window.platform.fbHeight = int(fbRect.Size.Height)

	return nil
}

// --- Platform window functions ---

func (w *Window) platformCreateWindow(wndconfig *wndconfig, ctxconfig *ctxconfig, fbconfig_ *fbconfig) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if err := createNativeWindow(w, wndconfig, fbconfig_); err != nil {
		return err
	}

	if ctxconfig.client != NoAPI {
		if ctxconfig.source == NativeContextAPI {
			if err := initNSGL(); err != nil {
				return err
			}
			if err := w.createContextNSGL(ctxconfig, fbconfig_); err != nil {
				return err
			}
		}
		if err := w.refreshContextAttribs(ctxconfig); err != nil {
			return err
		}
	}

	if wndconfig.mousePassthrough {
		if err := w.platformSetWindowMousePassthrough(true); err != nil {
			return err
		}
	}

	if w.monitor != nil {
		w.platformShowWindow()
		if err := w.platformFocusWindow(); err != nil {
			return err
		}
		if err := w.acquireMonitor(); err != nil {
			return err
		}
		if wndconfig.centerCursor {
			if err := w.centerCursorInContentArea(); err != nil {
				return err
			}
		}
	} else {
		if wndconfig.visible {
			w.platformShowWindow()
			if wndconfig.focused {
				if err := w.platformFocusWindow(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (w *Window) platformDestroyWindow() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if _glfw.platformWindow.disabledCursorWindow == w {
		_glfw.platformWindow.disabledCursorWindow = nil
	}

	if w.platform.object != 0 {
		cocoa.ObjcSend1(w.platform.object, sel_orderOut, 0)
	}

	if w.monitor != nil {
		if err := w.releaseMonitor(); err != nil {
			return err
		}
	}

	if w.context.destroy != nil {
		if err := w.context.destroy(w); err != nil {
			return err
		}
	}

	if w.platform.delegate != 0 {
		cocoa.ObjcSend1(w.platform.object, sel_setDelegate, 0)
		delete(theGoWindows, w.platform.delegate)
		cocoa.ObjcSend0(w.platform.delegate, sel_release)
		w.platform.delegate = 0
	}

	if w.platform.view != 0 {
		// The view's theGoWindows entry is removed in the dealloc callback, which fires
		// only after the NSWindow releases its content view.
		cocoa.ObjcSend0(w.platform.view, sel_release)
		w.platform.view = 0
	}

	if w.platform.object != 0 {
		cocoa.ObjcSend0(w.platform.object, objc.RegisterName("close"))
		w.platform.object = 0
	}

	// HACK: Allow Cocoa to catch up before returning
	if err := platformPollEvents(); err != nil {
		return err
	}

	return nil
}

func (w *Window) platformSetWindowTitle(title string) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	s := cocoa.NSString_alloc().InitWithUTF8String(title)
	cocoa.ObjcSend1(w.platform.object, sel_setTitle, uintptr(s.ID))
	// HACK: Set the miniwindow title explicitly as setTitle: doesn't update it
	//       if the window lacks NSWindowStyleMaskTitled
	cocoa.ObjcSend1(w.platform.object, sel_setMiniwindowTitle, uintptr(s.ID))
	cocoa.ObjcSend0(s.ID, sel_release)
	return nil
}

func (w *Window) platformSetWindowIcon(images []*Image) error {
	// macOS does not support per-window icons. The dock icon is set at the application level.
	return nil
}

func (w *Window) platformGetWindowPos() (xpos, ypos int, err error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	frame := cocoa.ObjcSendNSRect(w.platform.object, sel_frame)
	contentRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_contentRectForFrameRect, frame)

	xpos = int(contentRect.Origin.X)
	ypos = int(transformYNS(float32(contentRect.Origin.Y + contentRect.Size.Height - 1)))
	return xpos, ypos, nil
}

func (w *Window) platformSetWindowPos(xpos, ypos int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	viewFrame := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	dummyRect := cocoa.NSRect{
		Origin: cocoa.NSPoint{
			X: float64(xpos),
			Y: float64(transformYNS(float32(float64(ypos) + viewFrame.Size.Height - 1))),
		},
		Size: cocoa.NSSize{Width: 0, Height: 0},
	}

	frameRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_frameRectForContentRect, dummyRect)
	cocoa.ObjcSendPoint(w.platform.object, sel_setFrameOrigin, frameRect.Origin)
	return nil
}

func (w *Window) platformGetWindowSize() (width, height int, err error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	contentRect := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	return int(contentRect.Size.Width), int(contentRect.Size.Height), nil
}

func (w *Window) platformSetWindowSize(width, height int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if w.monitor != nil {
		if w.monitor.window == w {
			if err := w.acquireMonitor(); err != nil {
				return err
			}
		}
		return nil
	}

	contentRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_contentRectForFrameRect,
		cocoa.ObjcSendNSRect(w.platform.object, sel_frame))
	contentRect.Origin.Y += contentRect.Size.Height - float64(height)
	contentRect.Size = cocoa.NSSize{Width: float64(width), Height: float64(height)}
	frameRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_frameRectForContentRect, contentRect)
	cocoa.ObjcSendRectBool(w.platform.object, objc.RegisterName("setFrame:display:"), frameRect, true)
	return nil
}

func (w *Window) platformSetWindowSizeLimits(minwidth, minheight, maxwidth, maxheight int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if minwidth == DontCare || minheight == DontCare {
		cocoa.ObjcSendSize(w.platform.object, sel_setContentMinSize, cocoa.NSSize{Width: 0, Height: 0})
	} else {
		cocoa.ObjcSendSize(w.platform.object, sel_setContentMinSize, cocoa.NSSize{Width: float64(minwidth), Height: float64(minheight)})
	}

	if maxwidth == DontCare || maxheight == DontCare {
		cocoa.ObjcSendSize(w.platform.object, sel_setContentMaxSize, cocoa.NSSize{Width: math.MaxFloat64, Height: math.MaxFloat64})
	} else {
		cocoa.ObjcSendSize(w.platform.object, sel_setContentMaxSize, cocoa.NSSize{Width: float64(maxwidth), Height: float64(maxheight)})
	}

	return nil
}

func (w *Window) platformSetWindowAspectRatio(numer, denom int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if numer == DontCare || denom == DontCare {
		cocoa.ObjcSendSize(w.platform.object, sel_setResizeIncrements, cocoa.NSSize{Width: 1, Height: 1})
	} else {
		cocoa.ObjcSendSize(w.platform.object, sel_setContentAspectRatio, cocoa.NSSize{Width: float64(numer), Height: float64(denom)})
	}
	return nil
}

func (w *Window) platformGetFramebufferSize() (width, height int, err error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	contentRect := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	fbRect := cocoa.ObjcSendNSRectRect(w.platform.view, sel_convertRectToBacking, contentRect)
	return int(fbRect.Size.Width), int(fbRect.Size.Height), nil
}

func (w *Window) platformGetWindowFrameSize() (left, top, right, bottom int, err error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	contentRect := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	frameRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_frameRectForContentRect, contentRect)

	left = int(contentRect.Origin.X - frameRect.Origin.X)
	top = int((frameRect.Origin.Y + frameRect.Size.Height) - (contentRect.Origin.Y + contentRect.Size.Height))
	right = int((frameRect.Origin.X + frameRect.Size.Width) - (contentRect.Origin.X + contentRect.Size.Width))
	bottom = int(contentRect.Origin.Y - frameRect.Origin.Y)
	return left, top, right, bottom, nil
}

func (w *Window) platformGetWindowContentScale() (xscale, yscale float32, err error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	points := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	pixels := cocoa.ObjcSendNSRectRect(w.platform.view, sel_convertRectToBacking, points)

	return float32(pixels.Size.Width / points.Size.Width), float32(pixels.Size.Height / points.Size.Height), nil
}

func (w *Window) platformIconifyWindow() {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	cocoa.ObjcSend1(w.platform.object, sel_miniaturize, 0)
}

func (w *Window) platformRestoreWindow() {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if cocoa.ObjcSendBool(w.platform.object, objc.RegisterName("isMiniaturized")) {
		cocoa.ObjcSend1(w.platform.object, sel_deminiaturize, 0)
	} else if cocoa.ObjcSendBool(w.platform.object, sel_isZoomed) {
		cocoa.ObjcSend1(w.platform.object, sel_zoom, 0)
	}
}

func (w *Window) platformMaximizeWindow() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if !cocoa.ObjcSendBool(w.platform.object, sel_isZoomed) {
		cocoa.ObjcSend1(w.platform.object, sel_zoom, 0)
	}
	return nil
}

func (w *Window) platformMaximizeSupported() bool {
	return true
}

func (w *Window) platformIconifySupported() bool {
	return true
}

func (w *Window) platformRestoreSupported() bool {
	return true
}

func (w *Window) platformShowWindow() {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	cocoa.ObjcSend1(w.platform.object, sel_orderFront, 0)
}

func (w *Window) platformHideWindow() {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	cocoa.ObjcSend1(w.platform.object, sel_orderOut, 0)
}

func (w *Window) platformRequestWindowAttention() {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// NSInformationalRequest = 10
	cocoa.ObjcSend1(nsApp(), sel_requestUserAttention, uintptr(10))
}

func (w *Window) platformFocusWindow() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	cocoa.ObjcSend1(nsApp(), sel_activateIgnoringOtherApps, uintptr(1))
	cocoa.ObjcSend1(w.platform.object, sel_makeKeyAndOrderFront, 0)
	return nil
}

func (w *Window) platformSetWindowMonitor(monitor *Monitor, xpos, ypos, width, height, refreshRate int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if w.monitor == monitor {
		if monitor != nil {
			if monitor.window == w {
				if err := w.acquireMonitor(); err != nil {
					return err
				}
			}
		} else {
			contentRect := cocoa.NSRect{
				Origin: cocoa.NSPoint{
					X: float64(xpos),
					Y: float64(transformYNS(float32(ypos + height - 1))),
				},
				Size: cocoa.NSSize{Width: float64(width), Height: float64(height)},
			}
			styleMask := cocoa.ObjcSend0(w.platform.object, sel_styleMask)
			frameRect := cocoa.ObjcSendNSRectRectInt(w.platform.object, sel_frameRectForContentRect_styleMask, contentRect, styleMask)
			cocoa.ObjcSendRectBool(w.platform.object, objc.RegisterName("setFrame:display:"), frameRect, true)
		}
		return nil
	}

	if w.monitor != nil {
		if err := w.releaseMonitor(); err != nil {
			return err
		}
	}

	w.inputWindowMonitor(monitor)

	// HACK: Allow the state cached in Cocoa to catch up to reality
	if err := platformPollEvents(); err != nil {
		return err
	}

	styleMask := cocoa.ObjcSend0(w.platform.object, sel_styleMask)

	if w.monitor != nil {
		styleMask &^= NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskResizable
		styleMask |= NSWindowStyleMaskBorderless
	} else {
		if w.decorated {
			styleMask &^= NSWindowStyleMaskBorderless
			styleMask |= NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
		}
		if w.resizable {
			styleMask |= NSWindowStyleMaskResizable
		} else {
			styleMask &^= NSWindowStyleMaskResizable
		}
	}

	cocoa.ObjcSend1(w.platform.object, sel_setStyleMask, styleMask)
	// HACK: Changing the style mask can cause the first responder to be cleared
	cocoa.ObjcSend1(w.platform.object, sel_makeFirstResponder, uintptr(w.platform.view))

	if w.monitor != nil {
		cocoa.ObjcSend1(w.platform.object, sel_setLevel, uintptr(NSMainMenuWindowLevel+1))
		cocoa.ObjcSend1(w.platform.object, sel_setHasShadow, 0)

		if err := w.acquireMonitor(); err != nil {
			return err
		}
	} else {
		contentRect := cocoa.NSRect{
			Origin: cocoa.NSPoint{
				X: float64(xpos),
				Y: float64(transformYNS(float32(ypos + height - 1))),
			},
			Size: cocoa.NSSize{Width: float64(width), Height: float64(height)},
		}
		frameRect := cocoa.ObjcSendNSRectRectInt(w.platform.object, sel_frameRectForContentRect_styleMask, contentRect, styleMask)
		cocoa.ObjcSendRectBool(w.platform.object, objc.RegisterName("setFrame:display:"), frameRect, true)

		if w.numer != DontCare && w.denom != DontCare {
			cocoa.ObjcSendSize(w.platform.object, sel_setContentAspectRatio, cocoa.NSSize{Width: float64(w.numer), Height: float64(w.denom)})
		}

		if w.minwidth != DontCare && w.minheight != DontCare {
			cocoa.ObjcSendSize(w.platform.object, sel_setContentMinSize, cocoa.NSSize{Width: float64(w.minwidth), Height: float64(w.minheight)})
		}

		if w.maxwidth != DontCare && w.maxheight != DontCare {
			cocoa.ObjcSendSize(w.platform.object, sel_setContentMaxSize, cocoa.NSSize{Width: float64(w.maxwidth), Height: float64(w.maxheight)})
		}

		if w.floating {
			cocoa.ObjcSend1(w.platform.object, sel_setLevel, uintptr(NSFloatingWindowLevel))
		} else {
			cocoa.ObjcSend1(w.platform.object, sel_setLevel, uintptr(NSNormalWindowLevel))
		}

		if w.resizable {
			cocoa.ObjcSend1(w.platform.object, sel_setCollectionBehavior, uintptr(_NSWindowCollectionBehaviorFullScreenPrimary|_NSWindowCollectionBehaviorManaged))
		} else {
			cocoa.ObjcSend1(w.platform.object, sel_setCollectionBehavior, uintptr(_NSWindowCollectionBehaviorFullScreenNone))
		}

		cocoa.ObjcSend1(w.platform.object, sel_setHasShadow, uintptr(1))
		// HACK: Clearing NSWindowStyleMaskTitled resets and disables the window
		//       title property but the miniwindow title property is unaffected
		miniTitle := cocoa.ObjcSend0(w.platform.object, sel_miniwindowTitle)
		cocoa.ObjcSend1(w.platform.object, sel_setTitle, miniTitle)
	}

	return nil
}

func (w *Window) platformWindowFocused() bool {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	return cocoa.ObjcSendBool(w.platform.object, sel_isKeyWindow)
}

func (w *Window) platformWindowIconified() bool {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	return cocoa.ObjcSendBool(w.platform.object, sel_isMiniaturized)
}

func (w *Window) platformWindowVisible() bool {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	return cocoa.ObjcSendBool(w.platform.object, sel_isVisible)
}

func (w *Window) platformWindowMaximized() bool {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if w.resizable {
		return cocoa.ObjcSendBool(w.platform.object, sel_isZoomed)
	}
	return false
}

func (w *Window) platformWindowHovered() (bool, error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	pos := cocoa.ObjcSendNSPoint(objc.ID(class_NSEvent), objc.RegisterName("mouseLocation"))

	// Check if this window is the topmost window at the cursor position.
	topWindowNumber := cocoa.ObjcSendIntPointInt(objc.ID(class_NSWindow), objc.RegisterName("windowNumberAtPoint:belowWindowWithWindowNumber:"), pos, uintptr(0))
	if topWindowNumber != cocoa.ObjcSend0(w.platform.object, sel_windowNumber) {
		return false, nil
	}

	viewFrame := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	screenRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_convertRectToScreen, viewFrame)
	// Match NSMouseInRect(point, rect, NO) behavior for non-flipped coordinates:
	// x >= origin.x && x < maxX && y > origin.y && y <= maxY
	return pos.X >= screenRect.Origin.X &&
		pos.X < screenRect.Origin.X+screenRect.Size.Width &&
		pos.Y > screenRect.Origin.Y &&
		pos.Y <= screenRect.Origin.Y+screenRect.Size.Height, nil
}

func (w *Window) platformFramebufferTransparent() bool {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	return !cocoa.ObjcSendBool(w.platform.object, objc.RegisterName("isOpaque")) &&
		!cocoa.ObjcSendBool(w.platform.view, objc.RegisterName("isOpaque"))
}

func (w *Window) platformSetWindowResizable(enabled bool) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	mask := cocoa.ObjcSend0(w.platform.object, objc.RegisterName("styleMask"))
	if enabled {
		mask |= NSWindowStyleMaskResizable
	} else {
		mask &^= NSWindowStyleMaskResizable
	}
	cocoa.ObjcSend1(w.platform.object, objc.RegisterName("setStyleMask:"), mask)
	if enabled {
		cocoa.ObjcSend1(w.platform.object, sel_setCollectionBehavior, uintptr(_NSWindowCollectionBehaviorFullScreenPrimary|_NSWindowCollectionBehaviorManaged))
	} else {
		cocoa.ObjcSend1(w.platform.object, sel_setCollectionBehavior, uintptr(_NSWindowCollectionBehaviorFullScreenNone))
	}
	return nil
}

func (w *Window) platformSetWindowDecorated(enabled bool) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	mask := cocoa.ObjcSend0(w.platform.object, sel_styleMask)
	if enabled {
		mask |= NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
		mask &^= NSWindowStyleMaskBorderless
	} else {
		mask |= NSWindowStyleMaskBorderless
		mask &^= (NSWindowStyleMaskTitled | NSWindowStyleMaskClosable)
	}
	cocoa.ObjcSend1(w.platform.object, sel_setStyleMask, mask)
	cocoa.ObjcSend1(w.platform.object, sel_makeFirstResponder, uintptr(w.platform.view))
	return nil
}

func (w *Window) platformSetWindowFloating(enabled bool) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if enabled {
		cocoa.ObjcSend1(w.platform.object, sel_setLevel, uintptr(NSFloatingWindowLevel))
	} else {
		cocoa.ObjcSend1(w.platform.object, sel_setLevel, uintptr(NSNormalWindowLevel))
	}
	return nil
}

func (w *Window) platformSetWindowMousePassthrough(enabled bool) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	cocoa.ObjcSend1(w.platform.object, sel_setIgnoresMouseEvents, cocoa.BoolToUintptr(enabled))
	return nil
}

func (w *Window) platformGetWindowOpacity() (float32, error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	return float32(cocoa.ObjcSendFloat64(w.platform.object, sel_alphaValue)), nil
}

func (w *Window) platformSetWindowOpacity(opacity float32) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	cocoa.ObjcSendFloat64Arg(w.platform.object, sel_setAlphaValue, float64(opacity))
	return nil
}

func (w *Window) platformSetRawMouseMotion(enabled bool) error {
	// Raw mouse motion is not supported on macOS.
	return nil
}

// --- Event polling ---

var (
	class_NSDate      = objc.GetClass("NSDate")
	sel_distantPast   = objc.RegisterName("distantPast")
	sel_distantFuture = objc.RegisterName("distantFuture")
)

func platformPollEvents() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	distantPast := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSDate), sel_distantPast))
	for {
		event := objc.ID(cocoa.ObjcSend4(nsApp(), sel_nextEventMatchingMask_untilDate_inMode_dequeue,
			uintptr(NSEventMaskAny),
			uintptr(distantPast),
			uintptr(nsDefaultRunLoopMode.ID),
			uintptr(1)))
		if event == 0 {
			break
		}
		cocoa.ObjcSend1(nsApp(), sel_sendEvent, uintptr(event))
	}
	return nil
}

func platformWaitEvents() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// Wait for an event with no timeout (distantFuture).
	distantFuture := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSDate), sel_distantFuture))
	event := objc.ID(cocoa.ObjcSend4(nsApp(), sel_nextEventMatchingMask_untilDate_inMode_dequeue,
		uintptr(NSEventMaskAny),
		uintptr(distantFuture),
		uintptr(nsDefaultRunLoopMode.ID),
		uintptr(1)))
	if event != 0 {
		cocoa.ObjcSend1(nsApp(), sel_sendEvent, uintptr(event))
	}

	if err := platformPollEvents(); err != nil {
		return err
	}
	return nil
}

func platformWaitEventsTimeout(timeout float64) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// Create an NSDate for the timeout.
	date := objc.ID(cocoa.ObjcSendFloat64ArgRet(objc.ID(objc.GetClass("NSDate")), objc.RegisterName("dateWithTimeIntervalSinceNow:"), timeout))
	event := objc.ID(cocoa.ObjcSend4(nsApp(), sel_nextEventMatchingMask_untilDate_inMode_dequeue,
		uintptr(NSEventMaskAny),
		uintptr(date),
		uintptr(nsDefaultRunLoopMode.ID),
		uintptr(1)))
	if event != 0 {
		cocoa.ObjcSend1(nsApp(), sel_sendEvent, uintptr(event))
	}

	if err := platformPollEvents(); err != nil {
		return err
	}
	return nil
}

func platformPostEmptyEvent() error {
	// Unlike most of the platform functions, this can be called from any goroutine.
	// An autorelease pool belongs to the OS thread that created it, so the goroutine must not
	// migrate to another thread between creating and releasing the pool.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	postEmptyEvent()
	return nil
}

// --- Cursor functions ---

func (w *Window) platformGetCursorPos() (xpos, ypos float64, err error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	contentRect := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	// NOTE: The returned location uses base 0,1 not 0,0
	pos := cocoa.ObjcSendNSPoint(w.platform.object, sel_mouseLocationOutsideOfEventStream)

	xpos = pos.X
	ypos = contentRect.Size.Height - pos.Y

	return xpos, ypos, nil
}

func (w *Window) platformSetCursorPos(xpos, ypos float64) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	updateCursorImage(w)

	contentRect := cocoa.ObjcSendNSRect(w.platform.view, sel_frame)
	// NOTE: The returned location uses base 0,1 not 0,0
	pos := cocoa.ObjcSendNSPoint(w.platform.object, sel_mouseLocationOutsideOfEventStream)

	w.platform.cursorWarpDeltaX += xpos - pos.X
	w.platform.cursorWarpDeltaY += ypos - contentRect.Size.Height + pos.Y

	if w.monitor != nil {
		cgDisplayMoveCursorToPoint(w.monitor.platform.displayID, cocoa.CGPoint{X: xpos, Y: ypos})
	} else {
		localRect := cocoa.NSRect{
			Origin: cocoa.NSPoint{X: xpos, Y: contentRect.Size.Height - ypos - 1},
			Size:   cocoa.NSSize{Width: 0, Height: 0},
		}
		globalRect := cocoa.ObjcSendNSRectRect(w.platform.object, sel_convertRectToScreen, localRect)
		globalPoint := globalRect.Origin

		cgWarpMouseCursorPosition(cocoa.CGPoint{
			X: globalPoint.X,
			Y: float64(transformYNS(float32(globalPoint.Y))),
		})
	}

	// HACK: Calling this right after setting the cursor position prevents macOS
	//       from freezing the cursor for a fraction of a second afterwards.
	if w.cursorMode != CursorDisabled {
		cgAssociateMouseAndMouseCursorPosition(1)
	}

	return nil
}

func (w *Window) platformSetCursorMode(mode int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if w.platformWindowFocused() {
		if err := updateCursorMode(w); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cursor) platformCreateCursor(img *image.NRGBA, xhot, yhot int) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()

	repAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSBitmapImageRep), sel_alloc))
	rep := repAlloc.Send(sel_initWithBitmapDataPlanes_pixelsWide_pixelsHigh_bitsPerSample_samplesPerPixel_hasAlpha_isPlanar_colorSpaceName_bitmapFormat_bytesPerRow_bitsPerPixel,
		uintptr(0),                   // planes (NULL = allocate)
		uintptr(w),                   // pixelsWide
		uintptr(h),                   // pixelsHigh
		uintptr(8),                   // bitsPerSample
		uintptr(4),                   // samplesPerPixel
		true,                         // hasAlpha
		false,                        // isPlanar
		nsCalibratedRGBColorSpace.ID, // colorSpaceName
		uintptr(1<<0),                // bitmapFormat: NSBitmapFormatAlphaNonpremultiplied = 1 << 0
		uintptr(w*4),                 // bytesPerRow
		uintptr(32),                  // bitsPerPixel
	)
	if rep == 0 {
		return fmt.Errorf("glfw: failed to create NSBitmapImageRep: %w", PlatformError)
	}

	// Copy pixel data into the bitmap row by row to honor the image's stride
	// and non-zero origin.
	bitmapData := cocoa.ObjcSend0(rep, sel_bitmapData)
	if bitmapData != 0 {
		dst := unsafe.Slice((*byte)(unsafe.Pointer(bitmapData)), w*h*4)
		for y := range h {
			src := img.PixOffset(b.Min.X, b.Min.Y+y)
			copy(dst[y*w*4:(y+1)*w*4], img.Pix[src:src+w*4])
		}
	}

	nativeAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSImage), sel_alloc))
	native := objc.ID(cocoa.ObjcSendSizeRet(nativeAlloc, sel_initWithSize,
		cocoa.CGSize{Width: float64(w), Height: float64(h)}))
	cocoa.ObjcSend1(native, sel_addRepresentation, uintptr(rep))

	cursorAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_alloc))
	cursor := objc.ID(cocoa.ObjcSendIDPoint(cursorAlloc,
		sel_initWithImage_hotSpot,
		native,
		cocoa.NSPoint{X: float64(xhot), Y: float64(yhot)}))

	cocoa.ObjcSend0(native, sel_release)
	cocoa.ObjcSend0(rep, sel_release)

	if cursor == 0 {
		return fmt.Errorf("glfw: failed to create custom cursor: %w", PlatformError)
	}

	c.platform.object = cursor
	return nil
}

func (c *Cursor) platformCreateStandardCursor(shape StandardCursor) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// Try private selectors for resize cursors first.
	var cursorSelector objc.SEL
	switch shape {
	case HResizeCursor:
		cursorSelector = objc.RegisterName("_windowResizeEastWestCursor")
	case VResizeCursor:
		cursorSelector = objc.RegisterName("_windowResizeNorthSouthCursor")
	case ResizeNWSECursor:
		cursorSelector = objc.RegisterName("_windowResizeNorthWestSouthEastCursor")
	case ResizeNESWCursor:
		cursorSelector = objc.RegisterName("_windowResizeNorthEastSouthWestCursor")
	}

	var cursor objc.ID
	if cursorSelector != 0 && cocoa.ObjcSendBool1(objc.ID(class_NSCursor), sel_respondsToSelector, uintptr(cursorSelector)) {
		id := objc.ID(cocoa.ObjcSend1(objc.ID(class_NSCursor), sel_performSelector, uintptr(cursorSelector)))
		if id != 0 && cocoa.ObjcSendBool1(id, sel_isKindOfClass, uintptr(class_NSCursor)) {
			cursor = id
		}
	}

	if cursor == 0 {
		switch shape {
		case ArrowCursor:
			cursor = objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_arrowCursor))
		case IBeamCursor:
			cursor = objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_IBeamCursor))
		case CrosshairCursor:
			cursor = objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_crosshairCursor))
		case HandCursor:
			cursor = objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_pointingHandCursor))
		case ResizeAllCursor:
			// Use the OS's resource: https://stackoverflow.com/a/21786835/5435443
			cursorName := cocoa.NSString_alloc().InitWithUTF8String("move")
			basePath := cocoa.NSString_alloc().InitWithUTF8String("/System/Library/Frameworks/ApplicationServices.framework/Versions/A/Frameworks/HIServices.framework/Versions/A/Resources/cursors")
			cursorPath := cocoa.NSString{ID: objc.ID(cocoa.ObjcSend1(basePath.ID, sel_stringByAppendingPathComponent, uintptr(cursorName.ID)))}
			cocoa.ObjcSend0(cursorName.ID, sel_release)
			cocoa.ObjcSend0(basePath.ID, sel_release)
			cursorPDF := cocoa.NSString_alloc().InitWithUTF8String("cursor.pdf")
			imagePath := cocoa.NSString{ID: objc.ID(cocoa.ObjcSend1(cursorPath.ID, sel_stringByAppendingPathComponent, uintptr(cursorPDF.ID)))}
			cocoa.ObjcSend0(cursorPDF.ID, sel_release)
			infoPlist := cocoa.NSString_alloc().InitWithUTF8String("info.plist")
			infoPath := cocoa.NSString{ID: objc.ID(cocoa.ObjcSend1(cursorPath.ID, sel_stringByAppendingPathComponent, uintptr(infoPlist.ID)))}
			cocoa.ObjcSend0(infoPlist.ID, sel_release)
			imageAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSImage), sel_alloc))
			image := objc.ID(cocoa.ObjcSend1(imageAlloc, sel_initByReferencingFile, uintptr(imagePath.ID)))
			info := objc.ID(cocoa.ObjcSend1(objc.ID(class_NSDictionary), sel_dictionaryWithContentsOfFile, uintptr(infoPath.ID)))
			if image != 0 && info != 0 {
				hotxKey := cocoa.NSString_alloc().InitWithUTF8String("hotx")
				hotx := cocoa.ObjcSendFloat64(objc.ID(cocoa.ObjcSend1(info, sel_valueForKey, uintptr(hotxKey.ID))), sel_doubleValue)
				cocoa.ObjcSend0(hotxKey.ID, sel_release)
				hotyKey := cocoa.NSString_alloc().InitWithUTF8String("hoty")
				hoty := cocoa.ObjcSendFloat64(objc.ID(cocoa.ObjcSend1(info, sel_valueForKey, uintptr(hotyKey.ID))), sel_doubleValue)
				cocoa.ObjcSend0(hotyKey.ID, sel_release)
				// alloc/init returns an owned object. Autorelease it so that the retain
				// below leaves exactly one owned reference.
				cursorAlloc := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_alloc))
				cursor = objc.ID(cocoa.ObjcSendIDPoint(cursorAlloc, sel_initWithImage_hotSpot, image, cocoa.NSPoint{X: hotx, Y: hoty}))
				cocoa.ObjcSend0(cursor, sel_autorelease)
			}
			if image != 0 {
				cocoa.ObjcSend0(image, sel_release)
			}
		case NotAllowedCursor:
			cursor = objc.ID(cocoa.ObjcSend0(objc.ID(class_NSCursor), sel_operationNotAllowedCursor))
		}
	}

	if cursor == 0 {
		return fmt.Errorf("glfw: failed to create standard cursor: %w", PlatformError)
	}

	cocoa.ObjcSend0(cursor, sel_retain)
	c.platform.object = cursor
	return nil
}

func (c *Cursor) platformDestroyCursor() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if c.platform.object != 0 {
		cocoa.ObjcSend0(c.platform.object, sel_release)
		c.platform.object = 0
	}
	return nil
}

func (w *Window) platformSetCursor(cursor *Cursor) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	if cursorInContentArea(w) {
		updateCursorImage(w)
	}
	return nil
}

// --- Clipboard ---

func platformSetClipboardString(str string) error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	pasteboard := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSPasteboard), sel_generalPasteboard))
	types := objc.ID(cocoa.ObjcSend1(objc.ID(class_NSArray), sel_arrayWithObject, uintptr(nsPasteboardTypeString.ID)))
	cocoa.ObjcSend2(pasteboard, sel_declareTypes_owner, uintptr(types), 0)
	clipStr := cocoa.NSString_alloc().InitWithUTF8String(str)
	cocoa.ObjcSend2(pasteboard, sel_setString_forType,
		uintptr(clipStr.ID),
		uintptr(nsPasteboardTypeString.ID))
	cocoa.ObjcSend0(clipStr.ID, sel_release)
	return nil
}

func platformGetClipboardString() (string, error) {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	pasteboard := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSPasteboard), sel_generalPasteboard))

	types := objc.ID(cocoa.ObjcSend0(pasteboard, sel_types))
	if !cocoa.ObjcSendBool1(types, sel_containsObject, uintptr(nsPasteboardTypeString.ID)) {
		return "", fmt.Errorf("glfw: failed to retrieve string from pasteboard: %w", FormatUnavailable)
	}

	strID := objc.ID(cocoa.ObjcSend1(pasteboard, sel_stringForType, uintptr(nsPasteboardTypeString.ID)))
	if strID == 0 {
		return "", fmt.Errorf("glfw: failed to retrieve object from pasteboard: %w", PlatformError)
	}

	str := cocoa.NSString{ID: strID}
	return str.String(), nil
}

// --- Key functions ---

func platformGetScancodeName(scancode int) (string, error) {
	if scancode < 0 || scancode > 0xff {
		return "", fmt.Errorf("glfw: invalid scancode %d: %w", scancode, InvalidValue)
	}
	key := _glfw.platformWindow.keycodes[scancode]
	if key == KeyUnknown {
		return "", nil
	}

	unicodeData := _glfw.platformWindow.unicodeData
	if unicodeData == 0 {
		return "", nil
	}

	const (
		kUCKeyActionDisplay          = 3
		kUCKeyTranslateNoDeadKeysBit = 0
	)

	var deadKeyState uint32
	var characters [4]uint16
	var characterCount int

	layoutPtr := cfDataGetBytePtr(unicodeData)
	if layoutPtr == 0 {
		return "", nil
	}

	if _glfw.platformWindow.tis.UCKeyTranslate(layoutPtr,
		uint16(scancode),
		kUCKeyActionDisplay,
		0,
		uint32(_glfw.platformWindow.tis.GetKbdType()),
		kUCKeyTranslateNoDeadKeysBit,
		&deadKeyState,
		len(characters),
		&characterCount,
		&characters[0]) != 0 {
		return "", nil
	}

	if characterCount == 0 {
		return "", nil
	}

	str := cfStringCreateWithCharacters(0, &characters[0], characterCount)
	if str == 0 {
		return "", nil
	}
	defer cfRelease(str)

	length := cfStringGetLength(str)
	size := cfStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8)
	if size < 0 {
		return "", nil
	}
	buf := make([]byte, size+1)
	if !cfStringGetCString(str, &buf[0], len(buf), kCFStringEncodingUTF8) {
		return "", nil
	}

	// Find the null terminator.
	name := cStringToGoString(buf)
	_glfw.platformWindow.keynames[key] = name
	return _glfw.platformWindow.keynames[key], nil
}

func platformGetKeyScancode(key Key) int {
	return _glfw.platformWindow.scancodes[key]
}

func platformRawMouseMotionSupported() bool {
	return false
}

// --- Monitor acquire/release ---

func (w *Window) acquireMonitor() error {
	if w.monitor == nil {
		return nil
	}
	vm := &w.videoMode
	if err := w.monitor.setVideoModeNS(vm); err != nil {
		return err
	}
	bounds := cgDisplayBounds(w.monitor.platform.displayID)
	frame := cocoa.NSRect{
		Origin: cocoa.NSPoint{
			X: bounds.X,
			Y: float64(transformYNS(float32(bounds.Y + bounds.Height - 1))),
		},
		Size: cocoa.NSSize{Width: bounds.Width, Height: bounds.Height},
	}
	cocoa.ObjcSendRectBool(w.platform.object, objc.RegisterName("setFrame:display:"), frame, true)

	w.monitor.inputMonitorWindow(w)
	return nil
}

func (w *Window) releaseMonitor() error {
	if w.monitor == nil {
		return nil
	}
	if w.monitor.window != w {
		return nil
	}
	w.monitor.inputMonitorWindow(nil)
	w.monitor.restoreVideoModeNS()
	return nil
}

// goStringFromCString converts a null-terminated C string pointer to a Go string.
func goStringFromCString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	p := (*byte)(unsafe.Pointer(ptr))
	var n int
	for {
		if *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) == 0 {
			break
		}
		n++
	}
	return string(unsafe.Slice(p, n))
}
