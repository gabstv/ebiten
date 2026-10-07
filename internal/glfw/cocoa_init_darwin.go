// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2009-2019 Camilla Löwy <elmindreda@glfw.org>
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package glfw

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/hajimehoshi/ebiten/v2/internal/cocoa"
)

// cfString creates a CFStringRef from a Go string. The caller is responsible for releasing it.
func cfString(s string) uintptr {
	return cfStringCreateWithCString(0, s, kCFStringEncodingUTF8)
}

// createKeyTables builds the macOS virtual key code to GLFW key mapping tables.
func createKeyTables() {
	for i := range _glfw.platformWindow.keycodes {
		_glfw.platformWindow.keycodes[i] = -1
	}
	for i := range _glfw.platformWindow.scancodes {
		_glfw.platformWindow.scancodes[i] = -1
	}

	_glfw.platformWindow.keycodes[0x00] = KeyA
	_glfw.platformWindow.keycodes[0x01] = KeyS
	_glfw.platformWindow.keycodes[0x02] = KeyD
	_glfw.platformWindow.keycodes[0x03] = KeyF
	_glfw.platformWindow.keycodes[0x04] = KeyH
	_glfw.platformWindow.keycodes[0x05] = KeyG
	_glfw.platformWindow.keycodes[0x06] = KeyZ
	_glfw.platformWindow.keycodes[0x07] = KeyX
	_glfw.platformWindow.keycodes[0x08] = KeyC
	_glfw.platformWindow.keycodes[0x09] = KeyV
	_glfw.platformWindow.keycodes[0x0A] = KeyWorld1
	_glfw.platformWindow.keycodes[0x0B] = KeyB
	_glfw.platformWindow.keycodes[0x0C] = KeyQ
	_glfw.platformWindow.keycodes[0x0D] = KeyW
	_glfw.platformWindow.keycodes[0x0E] = KeyE
	_glfw.platformWindow.keycodes[0x0F] = KeyR
	_glfw.platformWindow.keycodes[0x10] = KeyY
	_glfw.platformWindow.keycodes[0x11] = KeyT
	_glfw.platformWindow.keycodes[0x12] = Key1
	_glfw.platformWindow.keycodes[0x13] = Key2
	_glfw.platformWindow.keycodes[0x14] = Key3
	_glfw.platformWindow.keycodes[0x15] = Key4
	_glfw.platformWindow.keycodes[0x16] = Key6
	_glfw.platformWindow.keycodes[0x17] = Key5
	_glfw.platformWindow.keycodes[0x18] = KeyEqual
	_glfw.platformWindow.keycodes[0x19] = Key9
	_glfw.platformWindow.keycodes[0x1A] = Key7
	_glfw.platformWindow.keycodes[0x1B] = KeyMinus
	_glfw.platformWindow.keycodes[0x1C] = Key8
	_glfw.platformWindow.keycodes[0x1D] = Key0
	_glfw.platformWindow.keycodes[0x1E] = KeyRightBracket
	_glfw.platformWindow.keycodes[0x1F] = KeyO
	_glfw.platformWindow.keycodes[0x20] = KeyU
	_glfw.platformWindow.keycodes[0x21] = KeyLeftBracket
	_glfw.platformWindow.keycodes[0x22] = KeyI
	_glfw.platformWindow.keycodes[0x23] = KeyP
	_glfw.platformWindow.keycodes[0x24] = KeyEnter
	_glfw.platformWindow.keycodes[0x25] = KeyL
	_glfw.platformWindow.keycodes[0x26] = KeyJ
	_glfw.platformWindow.keycodes[0x27] = KeyApostrophe
	_glfw.platformWindow.keycodes[0x28] = KeyK
	_glfw.platformWindow.keycodes[0x29] = KeySemicolon
	_glfw.platformWindow.keycodes[0x2A] = KeyBackslash
	_glfw.platformWindow.keycodes[0x2B] = KeyComma
	_glfw.platformWindow.keycodes[0x2C] = KeySlash
	_glfw.platformWindow.keycodes[0x2D] = KeyN
	_glfw.platformWindow.keycodes[0x2E] = KeyM
	_glfw.platformWindow.keycodes[0x2F] = KeyPeriod
	_glfw.platformWindow.keycodes[0x30] = KeyTab
	_glfw.platformWindow.keycodes[0x31] = KeySpace
	_glfw.platformWindow.keycodes[0x32] = KeyGraveAccent
	_glfw.platformWindow.keycodes[0x33] = KeyBackspace
	_glfw.platformWindow.keycodes[0x35] = KeyEscape
	_glfw.platformWindow.keycodes[0x37] = KeyLeftSuper
	_glfw.platformWindow.keycodes[0x38] = KeyLeftShift
	_glfw.platformWindow.keycodes[0x39] = KeyCapsLock
	_glfw.platformWindow.keycodes[0x3A] = KeyLeftAlt
	_glfw.platformWindow.keycodes[0x3B] = KeyLeftControl
	_glfw.platformWindow.keycodes[0x3C] = KeyRightShift
	_glfw.platformWindow.keycodes[0x3D] = KeyRightAlt
	_glfw.platformWindow.keycodes[0x36] = KeyRightSuper
	_glfw.platformWindow.keycodes[0x3E] = KeyRightControl
	_glfw.platformWindow.keycodes[0x40] = KeyF17
	_glfw.platformWindow.keycodes[0x41] = KeyKPDecimal
	_glfw.platformWindow.keycodes[0x43] = KeyKPMultiply
	_glfw.platformWindow.keycodes[0x45] = KeyKPAdd
	_glfw.platformWindow.keycodes[0x47] = KeyNumLock
	_glfw.platformWindow.keycodes[0x4B] = KeyKPDivide
	_glfw.platformWindow.keycodes[0x4C] = KeyKPEnter
	_glfw.platformWindow.keycodes[0x4E] = KeyKPSubtract
	_glfw.platformWindow.keycodes[0x4F] = KeyF18
	_glfw.platformWindow.keycodes[0x50] = KeyF19
	_glfw.platformWindow.keycodes[0x51] = KeyKPEqual
	_glfw.platformWindow.keycodes[0x52] = KeyKP0
	_glfw.platformWindow.keycodes[0x53] = KeyKP1
	_glfw.platformWindow.keycodes[0x54] = KeyKP2
	_glfw.platformWindow.keycodes[0x55] = KeyKP3
	_glfw.platformWindow.keycodes[0x56] = KeyKP4
	_glfw.platformWindow.keycodes[0x57] = KeyKP5
	_glfw.platformWindow.keycodes[0x58] = KeyKP6
	_glfw.platformWindow.keycodes[0x59] = KeyKP7
	_glfw.platformWindow.keycodes[0x5A] = KeyF20
	_glfw.platformWindow.keycodes[0x5B] = KeyKP8
	_glfw.platformWindow.keycodes[0x5C] = KeyKP9
	_glfw.platformWindow.keycodes[0x60] = KeyF5
	_glfw.platformWindow.keycodes[0x61] = KeyF6
	_glfw.platformWindow.keycodes[0x62] = KeyF7
	_glfw.platformWindow.keycodes[0x63] = KeyF3
	_glfw.platformWindow.keycodes[0x64] = KeyF8
	_glfw.platformWindow.keycodes[0x65] = KeyF9
	_glfw.platformWindow.keycodes[0x67] = KeyF11
	_glfw.platformWindow.keycodes[0x69] = KeyPrintScreen
	_glfw.platformWindow.keycodes[0x6A] = KeyF16
	_glfw.platformWindow.keycodes[0x6B] = KeyF14
	_glfw.platformWindow.keycodes[0x6D] = KeyF10
	_glfw.platformWindow.keycodes[0x6E] = KeyMenu
	_glfw.platformWindow.keycodes[0x6F] = KeyF12
	_glfw.platformWindow.keycodes[0x71] = KeyF15
	_glfw.platformWindow.keycodes[0x72] = KeyInsert
	_glfw.platformWindow.keycodes[0x73] = KeyHome
	_glfw.platformWindow.keycodes[0x74] = KeyPageUp
	_glfw.platformWindow.keycodes[0x75] = KeyDelete
	_glfw.platformWindow.keycodes[0x76] = KeyF4
	_glfw.platformWindow.keycodes[0x77] = KeyEnd
	_glfw.platformWindow.keycodes[0x78] = KeyF2
	_glfw.platformWindow.keycodes[0x79] = KeyPageDown
	_glfw.platformWindow.keycodes[0x7A] = KeyF1
	_glfw.platformWindow.keycodes[0x7B] = KeyLeft
	_glfw.platformWindow.keycodes[0x7C] = KeyRight
	_glfw.platformWindow.keycodes[0x7D] = KeyDown
	_glfw.platformWindow.keycodes[0x7E] = KeyUp

	for scancode := range 256 {
		if _glfw.platformWindow.keycodes[scancode] >= 0 {
			_glfw.platformWindow.scancodes[_glfw.platformWindow.keycodes[scancode]] = scancode
		}
	}
}

// getAppName returns the application name from NSProcessInfo or the bundle.
func getAppName() string {
	// Try to figure out what the calling application is called.
	bundle := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSBundle), sel_mainBundle))
	if bundle != 0 {
		info := objc.ID(cocoa.ObjcSend0(bundle, sel_infoDictionary))
		if info != 0 {
			nameKeys := []string{
				"CFBundleDisplayName",
				"CFBundleName",
				"CFBundleExecutable",
			}
			for _, key := range nameKeys {
				nsKey := cocoa.NSString_alloc().InitWithUTF8String(key)
				name := objc.ID(cocoa.ObjcSend1(info, sel_objectForKey, uintptr(nsKey.ID)))
				cocoa.ObjcSend0(nsKey.ID, sel_release)
				if name != 0 && cocoa.ObjcSend1(name, sel_isKindOfClass, uintptr(objc.GetClass("NSString"))) != 0 {
					s := cocoa.NSString{ID: name}.String()
					if len(s) > 0 {
						return s
					}
				}
			}
		}
	}

	// Fall back to process name.
	pi := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSProcessInfo), sel_processInfo))
	name := cocoa.NSString{ID: objc.ID(cocoa.ObjcSend0(pi, sel_processName))}
	if s := name.String(); len(s) > 0 {
		return s
	}

	return "GLFW Application"
}

// createMenuBar creates the standard macOS menu bar with app menu and window menu.
func createMenuBar() {
	appName := getAppName()

	menubar := objc.ID(cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMenu), sel_alloc)), sel_init))
	nsApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
	cocoa.ObjcSend1(nsApp, sel_setMainMenu, uintptr(menubar))

	// nsStr creates an NSString and schedules it for release.
	// This mirrors the behavior of @"..." literals in Objective-C which are
	// compile-time constants; here we use alloc/init and defer release.
	var nsStrings []cocoa.NSString
	nsStr := func(s string) objc.ID {
		str := cocoa.NSString_alloc().InitWithUTF8String(s)
		nsStrings = append(nsStrings, str)
		return str.ID
	}
	defer func() {
		for _, s := range nsStrings {
			cocoa.ObjcSend0(s.ID, sel_release)
		}
	}()

	// Create the application menu.
	appMenuItem := objc.ID(cocoa.ObjcSend3(menubar, sel_addItemWithTitle_action_keyEquivalent, uintptr(nsStr("")), uintptr(0), uintptr(nsStr(""))))
	appMenu := objc.ID(cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMenu), sel_alloc)), sel_init))
	cocoa.ObjcSend1(appMenuItem, sel_setSubmenu, uintptr(appMenu))

	// About <AppName>
	cocoa.ObjcSend3(appMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("About "+appName)),
		uintptr(sel_orderFrontStandardAboutPanel),
		uintptr(nsStr("")))

	cocoa.ObjcSend1(appMenu, sel_addItem, cocoa.ObjcSend0(objc.ID(class_NSMenuItem), sel_separatorItem))

	// Services submenu
	servicesMenu := objc.ID(cocoa.ObjcSend0(objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMenu), sel_alloc)), sel_init))
	cocoa.ObjcSend1(nsApp, sel_setServicesMenu, uintptr(servicesMenu))
	servicesMenuItem := objc.ID(cocoa.ObjcSend3(appMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Services")),
		uintptr(0),
		uintptr(nsStr(""))))
	cocoa.ObjcSend1(servicesMenuItem, sel_setSubmenu, uintptr(servicesMenu))
	cocoa.ObjcSend0(servicesMenu, sel_release)

	cocoa.ObjcSend1(appMenu, sel_addItem, cocoa.ObjcSend0(objc.ID(class_NSMenuItem), sel_separatorItem))

	// Hide <AppName>
	cocoa.ObjcSend3(appMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Hide "+appName)),
		uintptr(sel_hide),
		uintptr(nsStr("h")))

	// Hide Others
	hideOthersItem := objc.ID(cocoa.ObjcSend3(appMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Hide Others")),
		uintptr(sel_hideOtherApplications),
		uintptr(nsStr("h"))))
	// NSEventModifierFlagOption | NSEventModifierFlagCommand
	cocoa.ObjcSend1(hideOthersItem, sel_setKeyEquivalentModifierMask, uintptr(1<<19|1<<20))

	// Show All
	cocoa.ObjcSend3(appMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Show All")),
		uintptr(sel_unhideAllApplications),
		uintptr(nsStr("")))

	cocoa.ObjcSend1(appMenu, sel_addItem, cocoa.ObjcSend0(objc.ID(class_NSMenuItem), sel_separatorItem))

	// Quit <AppName>
	cocoa.ObjcSend3(appMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Quit "+appName)),
		uintptr(sel_terminate),
		uintptr(nsStr("q")))

	// Create the Window menu.
	windowMenuItem := objc.ID(cocoa.ObjcSend3(menubar, sel_addItemWithTitle_action_keyEquivalent, uintptr(nsStr("")), uintptr(0), uintptr(nsStr(""))))
	cocoa.ObjcSend0(menubar, sel_release)

	windowMenu := objc.ID(cocoa.ObjcSend1(objc.ID(cocoa.ObjcSend0(objc.ID(class_NSMenu), sel_alloc)),
		sel_initWithTitle, uintptr(nsStr("Window"))))
	cocoa.ObjcSend1(nsApp, sel_setWindowsMenu, uintptr(windowMenu))
	cocoa.ObjcSend1(windowMenuItem, sel_setSubmenu, uintptr(windowMenu))

	// Minimize
	cocoa.ObjcSend3(windowMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Minimize")),
		uintptr(sel_performMiniaturize),
		uintptr(nsStr("m")))

	// Zoom
	cocoa.ObjcSend3(windowMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Zoom")),
		uintptr(sel_performZoom),
		uintptr(nsStr("")))

	cocoa.ObjcSend1(windowMenu, sel_addItem, cocoa.ObjcSend0(objc.ID(class_NSMenuItem), sel_separatorItem))

	// Bring All to Front
	cocoa.ObjcSend3(windowMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Bring All to Front")),
		uintptr(sel_arrangeInFront),
		uintptr(nsStr("")))

	// Enter Full Screen
	cocoa.ObjcSend1(windowMenu, sel_addItem, cocoa.ObjcSend0(objc.ID(class_NSMenuItem), sel_separatorItem))
	fullScreenItem := objc.ID(cocoa.ObjcSend3(windowMenu, sel_addItemWithTitle_action_keyEquivalent,
		uintptr(nsStr("Enter Full Screen")),
		uintptr(sel_toggleFullScreen),
		uintptr(nsStr("f"))))
	// NSEventModifierFlagControl | NSEventModifierFlagCommand
	cocoa.ObjcSend1(fullScreenItem, sel_setKeyEquivalentModifierMask, uintptr(NSEventModifierFlagControl|NSEventModifierFlagCommand))

	// Prior to Snow Leopard, we need to use this oddly-named semi-private API
	// to get the application menu working properly.
	cocoa.ObjcSend2(nsApp, sel_performSelector_withObject,
		uintptr(sel_setAppleMenu), uintptr(appMenu))
}

// updateUnicodeDataNS updates the cached keyboard layout unicode data.
func updateUnicodeDataNS() error {
	// Release the previous input source before acquiring a new one.
	if _glfw.platformWindow.inputSource != 0 {
		cfRelease(_glfw.platformWindow.inputSource)
		_glfw.platformWindow.inputSource = 0
		_glfw.platformWindow.unicodeData = 0
	}

	_glfw.platformWindow.inputSource = _glfw.platformWindow.tis.CopyCurrentKeyboardLayoutInputSource()
	if _glfw.platformWindow.inputSource == 0 {
		return fmt.Errorf("glfw: failed to retrieve keyboard layout input source: %w", PlatformError)
	}

	_glfw.platformWindow.unicodeData = _glfw.platformWindow.tis.GetInputSourceProperty(
		_glfw.platformWindow.inputSource, _glfw.platformWindow.tis.kPropertyUnicodeKeyLayoutData)
	if _glfw.platformWindow.unicodeData == 0 {
		return fmt.Errorf("glfw: failed to retrieve keyboard layout Unicode data: %w", PlatformError)
	}

	return nil
}

// initializeTIS loads TIS (Text Input Source) symbols from the HIToolbox framework.
func initializeTIS() error {
	// When using Cgo, HIToolbox is loaded implicitly by linking against Cocoa.
	// With purego, we must load it explicitly so CFBundleGetBundleWithIdentifier can find it.
	_, err := purego.Dlopen("/System/Library/Frameworks/Carbon.framework/Versions/A/Frameworks/HIToolbox.framework/HIToolbox", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("glfw: failed to dlopen HIToolbox: %w", err)
	}

	bundleID := cfString("com.apple.HIToolbox")
	if bundleID == 0 {
		return fmt.Errorf("glfw: failed to create a CFString: %w", PlatformError)
	}
	defer cfRelease(bundleID)

	bundle := cfBundleGetBundleWithIdentifier(bundleID)
	if bundle == 0 {
		return fmt.Errorf("glfw: failed to load HIToolbox.framework: %w", PlatformError)
	}

	// Load TISCopyCurrentKeyboardLayoutInputSource
	fnName1 := cfString("TISCopyCurrentKeyboardLayoutInputSource")
	if fnName1 == 0 {
		return fmt.Errorf("glfw: failed to create a CFString: %w", PlatformError)
	}
	defer cfRelease(fnName1)
	ptr1 := cfBundleGetFunctionPointerForName(bundle, fnName1)
	if ptr1 == 0 {
		return fmt.Errorf("glfw: failed to load TIS API symbols: %w", PlatformError)
	}
	_glfw.platformWindow.tis.CopyCurrentKeyboardLayoutInputSource = func() uintptr {
		r, _, _ := purego.SyscallN(ptr1)
		return r
	}

	// Load TISGetInputSourceProperty
	fnName2 := cfString("TISGetInputSourceProperty")
	if fnName2 == 0 {
		return fmt.Errorf("glfw: failed to create a CFString: %w", PlatformError)
	}
	defer cfRelease(fnName2)
	ptr2 := cfBundleGetFunctionPointerForName(bundle, fnName2)
	if ptr2 == 0 {
		return fmt.Errorf("glfw: failed to load TIS API symbols: %w", PlatformError)
	}
	_glfw.platformWindow.tis.GetInputSourceProperty = func(inputSource uintptr, propertyKey uintptr) uintptr {
		r, _, _ := purego.SyscallN(ptr2, inputSource, propertyKey)
		return r
	}

	// Load LMGetKbdType
	fnName3 := cfString("LMGetKbdType")
	if fnName3 == 0 {
		return fmt.Errorf("glfw: failed to create a CFString: %w", PlatformError)
	}
	defer cfRelease(fnName3)
	ptr3 := cfBundleGetFunctionPointerForName(bundle, fnName3)
	if ptr3 == 0 {
		return fmt.Errorf("glfw: failed to load TIS API symbols: %w", PlatformError)
	}
	_glfw.platformWindow.tis.GetKbdType = func() uint8 {
		r, _, _ := purego.SyscallN(ptr3)
		return uint8(r)
	}

	// Load UCKeyTranslate from the Carbon umbrella framework.
	carbon, err2 := purego.Dlopen("/System/Library/Frameworks/Carbon.framework/Carbon", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err2 != nil {
		return fmt.Errorf("glfw: failed to dlopen Carbon: %w", err2)
	}
	ptr4, err2 := purego.Dlsym(carbon, "UCKeyTranslate")
	if err2 != nil {
		return fmt.Errorf("glfw: failed to load UCKeyTranslate: %w", err2)
	}
	_glfw.platformWindow.tis.UCKeyTranslate = func(keyLayoutPtr uintptr, virtualKeyCode uint16, keyAction uint16, modifierKeyState uint32, keyboardType uint32, keyTranslateOptions uint32, deadKeyState *uint32, maxStringLength int, actualStringLength *int, unicodeString *uint16) int32 {
		r, _, _ := purego.SyscallN(ptr4, keyLayoutPtr, uintptr(virtualKeyCode), uintptr(keyAction), uintptr(modifierKeyState), uintptr(keyboardType), uintptr(keyTranslateOptions), uintptr(unsafe.Pointer(deadKeyState)), uintptr(maxStringLength), uintptr(unsafe.Pointer(actualStringLength)), uintptr(unsafe.Pointer(unicodeString)))
		return int32(r)
	}

	// Load kTISPropertyUnicodeKeyLayoutData string constant
	dataName := cfString("kTISPropertyUnicodeKeyLayoutData")
	if dataName == 0 {
		return fmt.Errorf("glfw: failed to create a CFString: %w", PlatformError)
	}
	defer cfRelease(dataName)
	dataPtr := cfBundleGetDataPointerForName(bundle, dataName)
	if dataPtr == 0 {
		return fmt.Errorf("glfw: failed to load TIS API symbols: %w", PlatformError)
	}
	_glfw.platformWindow.tis.kPropertyUnicodeKeyLayoutData = *(*uintptr)(unsafe.Pointer(dataPtr))

	return nil
}

// GLFWHelper and GLFWApplicationDelegate class references.
var (
	class_GLFWHelper              objc.Class
	class_GLFWApplicationDelegate objc.Class
)

// checkWindowServerAccess returns an error if the process is denied access to the window server.
func checkWindowServerAccess() error {
	// Without access, the application never finishes launching and NSApplication's run blocks forever.
	// The service name is undocumented, so only a definite denial is an error.
	// Any other failure, such as an unknown service, is left to AppKit.
	var port uint32
	switch bootstrapLookUp(*bootstrapPort, "com.apple.windowserver.active", &port) {
	case _KERN_SUCCESS:
		machPortDeallocate(*machTaskSelf, port)
	case _BOOTSTRAP_NOT_PRIVILEGED:
		return fmt.Errorf("glfw: access to the window server is denied: %w", PlatformError)
	}
	return nil
}

// platformInit performs the full macOS platform initialization.
func platformInit() error {
	if err := checkWindowServerAccess(); err != nil {
		return err
	}

	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// Register GLFWHelper class — an NSObject subclass with a method to handle
	// keyboard input source change notifications.
	helper, err := objc.RegisterClass(
		"GLFWHelper",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: sel_selectedKeyboardInputSourceChanged,
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) {
					_ = updateUnicodeDataNS()
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("glfw: failed to register GLFWHelper class: %w", err)
	}
	class_GLFWHelper = helper

	// Register GLFWWindow, GLFWWindowDelegate, and GLFWContentView classes.
	if err := registerGLFWClasses(); err != nil {
		return err
	}

	// Register GLFWApplicationDelegate class — implements NSApplicationDelegate.
	// The NSApplicationDelegate protocol is intentionally NOT declared: on
	// macOS 14 doing so makes AppKit crash inside nextEventMatchingMask:
	// shortly after the app activates (issue #3451). AppKit still calls our
	// delegate methods because it falls back to respondsToSelector:.
	delegate, err := objc.RegisterClass(
		"GLFWApplicationDelegate",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{
				Cmd: sel_applicationShouldTerminate,
				Fn: func(_ objc.ID, _ objc.SEL, sender objc.ID) uintptr {
					// Post close events to all windows.
					for _, window := range _glfw.windows {
						window.inputWindowCloseRequest()
					}
					return _NSApplicationTerminateCancel
				},
			},
			{
				Cmd: sel_applicationDidChangeScreenParameters,
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) {
					for _, window := range _glfw.windows {
						if window.context.client != NoAPI {
							cocoa.ObjcSend0(window.context.platform.object, sel_update)
						}
					}
					_ = pollMonitorsNS()
				},
			},
			{
				Cmd: sel_applicationWillFinishLaunching,
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) {
					// In the C original, this first tries to load MainMenu.nib from
					// the bundle, and only falls back to createMenuBar() if no nib exists.
					bundle := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSBundle), sel_mainBundle))
					mainMenuNib := cocoa.NSString_alloc().InitWithUTF8String("MainMenu")
					nibType := cocoa.NSString_alloc().InitWithUTF8String("nib")
					nibPath := cocoa.ObjcSend2(bundle, sel_pathForResource_ofType, uintptr(mainMenuNib.ID), uintptr(nibType.ID))
					cocoa.ObjcSend0(nibType.ID, sel_release)
					if nibPath != 0 {
						cocoa.ObjcSend3(bundle, sel_loadNibNamed_owner_topLevelObjects,
							uintptr(mainMenuNib.ID),
							cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication),
							uintptr(unsafe.Pointer(&_glfw.platformWindow.nibObjects)))
					} else {
						createMenuBar()
					}
					cocoa.ObjcSend0(mainMenuNib.ID, sel_release)
				},
			},
			{
				Cmd: sel_applicationDidFinishLaunching,
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) {
					nsApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
					postEmptyEvent()
					// In case we are unbundled, make us a proper UI application.
					// The C code gates this on _glfw.hints.init.ns.menubar which
					// defaults to true. Since Ebitengine always wants a menubar,
					// this is called unconditionally.
					cocoa.ObjcSend1(nsApp, sel_setActivationPolicy, _NSApplicationActivationPolicyRegular)
					cocoa.ObjcSend1(nsApp, sel_stop, 0)
				},
			},
			{
				Cmd: sel_applicationDidHide,
				Fn: func(_ objc.ID, _ objc.SEL, _ objc.ID) {
					for _, monitor := range _glfw.monitors {
						monitor.restoreVideoModeNS()
					}
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("glfw: failed to register GLFWApplicationDelegate class: %w", err)
	}
	class_GLFWApplicationDelegate = delegate

	// Create the shared NSApplication instance.
	nsApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))

	// Create and set the application delegate.
	_glfw.platformWindow.delegate = objc.ID(cocoa.ObjcSend0(
		objc.ID(cocoa.ObjcSend0(objc.ID(class_GLFWApplicationDelegate), sel_alloc)), sel_init))
	cocoa.ObjcSend1(nsApp, sel_setDelegate, uintptr(_glfw.platformWindow.delegate))

	// Create GLFWHelper instance and register for keyboard input source change notifications.
	_glfw.platformWindow.helper = objc.ID(cocoa.ObjcSend0(
		objc.ID(cocoa.ObjcSend0(objc.ID(class_GLFWHelper), sel_alloc)), sel_init))

	notificationCenter := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSNotificationCenter), sel_defaultCenter))
	nsTextInputContextKeyboardSelectionDidChangeNotification := cocoa.NSString_alloc().InitWithUTF8String(
		"NSTextInputContextKeyboardSelectionDidChangeNotification")
	cocoa.ObjcSend4(notificationCenter, sel_addObserver_selector_name_object,
		uintptr(_glfw.platformWindow.helper),
		uintptr(sel_selectedKeyboardInputSourceChanged),
		uintptr(nsTextInputContextKeyboardSelectionDidChangeNotification.ID),
		0)
	cocoa.ObjcSend0(nsTextInputContextKeyboardSelectionDidChangeNotification.ID, sel_release)

	// Add a local monitor for keyUp events to work around Cocoa swallowing
	// key-up events when the menu bar is active.
	keyUpBlock := objc.NewBlock(func(_ objc.Block, event objc.ID) objc.ID {
		if cocoa.ObjcSend0(event, sel_modifierFlags)&NSEventModifierFlagCommand != 0 {
			app := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
			keyWin := objc.ID(cocoa.ObjcSend0(app, sel_keyWindow))
			cocoa.ObjcSend1(keyWin, sel_sendEvent, uintptr(event))
		}
		return event
	})
	defer keyUpBlock.Release()
	_glfw.platformWindow.keyUpMonitor = objc.ID(cocoa.ObjcSend2(objc.ID(class_NSEvent),
		sel_addLocalMonitorForEventsMatchingMask_handler,
		_NSEventMaskKeyUp,
		uintptr(keyUpBlock)))

	// Create a CGEventSource for synthesized events.
	_glfw.platformWindow.eventSource = cgEventSourceCreate(_kCGEventSourceStateHIDSystemState)
	if _glfw.platformWindow.eventSource == 0 {
		return fmt.Errorf("glfw: failed to create CGEventSource: %w", PlatformError)
	}

	// Set the suppression interval to zero so that synthesized events
	// are not suppressed after a warp.
	cgEventSourceSetLocalEventsSuppressionInterval(_glfw.platformWindow.eventSource, 0.0)

	// Initialize TIS (Text Input Source) framework bindings.
	if err := initializeTIS(); err != nil {
		return err
	}

	// Build key code translation tables.
	createKeyTables()

	// Cache the current keyboard layout unicode data.
	if err := updateUnicodeDataNS(); err != nil {
		return err
	}

	// Initialize the high-resolution timer.
	initTimerNS()

	// Detect and register connected monitors.
	if err := pollMonitorsNS(); err != nil {
		return err
	}

	// Run the application to process initial events, but only if it hasn't
	// already finished launching. The delegate's applicationDidFinishLaunching:
	// calls stop: and posts an empty event, so this returns quickly.
	currentApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSRunningApplication), sel_currentApplication))
	if cocoa.ObjcSend0(currentApp, sel_isFinishedLaunching) == 0 {
		cocoa.ObjcSend0(nsApp, sel_run)
	}

	// Initialize NSGL (OpenGL context support).
	if err := initNSGL(); err != nil {
		return err
	}

	return nil
}

// postEmptyEvent posts a no-op application-defined event to wake the run loop.
func postEmptyEvent() {
	nsApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
	// NSApplicationDefined = 15
	event := objc.Send[objc.ID](objc.ID(class_NSEvent), sel_otherEventWithType_location_modifierFlags_timestamp_windowNumber_context_subtype_data1_data2,
		uintptr(15),               // NSApplicationDefined
		cocoa.CGPoint{X: 0, Y: 0}, // location (NSPoint)
		uintptr(0),                // modifierFlags
		float64(0),                // timestamp
		uintptr(0),                // windowNumber
		uintptr(0),                // context (nil)
		uintptr(0),                // subtype
		uintptr(0),                // data1
		uintptr(0),                // data2
	)
	cocoa.ObjcSend2(nsApp, sel_postEvent_atStart, uintptr(event), cocoa.BoolToUintptr(true))
}

// platformTerminate cleans up macOS platform resources.
func platformTerminate() error {
	pool := cocoa.NSAutoreleasePool_new()
	defer pool.Release()

	// Release TIS input source if held.
	if _glfw.platformWindow.inputSource != 0 {
		cfRelease(_glfw.platformWindow.inputSource)
		_glfw.platformWindow.inputSource = 0
		_glfw.platformWindow.unicodeData = 0
	}

	// Release the CGEventSource.
	if _glfw.platformWindow.eventSource != 0 {
		cfRelease(_glfw.platformWindow.eventSource)
		_glfw.platformWindow.eventSource = 0
	}

	// Release the application delegate.
	if _glfw.platformWindow.delegate != 0 {
		nsApp := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSApplication), sel_sharedApplication))
		cocoa.ObjcSend1(nsApp, sel_setDelegate, 0)
		cocoa.ObjcSend0(_glfw.platformWindow.delegate, sel_release)
		_glfw.platformWindow.delegate = 0
	}

	// Release the helper and remove notification observers.
	if _glfw.platformWindow.helper != 0 {
		notificationCenter := objc.ID(cocoa.ObjcSend0(objc.ID(class_NSNotificationCenter), sel_defaultCenter))
		cocoa.ObjcSend1(notificationCenter, sel_removeObserver, uintptr(_glfw.platformWindow.helper))
		cocoa.ObjcSend0(_glfw.platformWindow.helper, sel_release)
		_glfw.platformWindow.helper = 0
	}

	// Remove the global keyUp monitor.
	if _glfw.platformWindow.keyUpMonitor != 0 {
		cocoa.ObjcSend1(objc.ID(class_NSEvent), sel_removeMonitor, uintptr(_glfw.platformWindow.keyUpMonitor))
		_glfw.platformWindow.keyUpMonitor = 0
	}

	// Terminate NSGL (OpenGL context support).
	terminateNSGL()

	return nil
}
