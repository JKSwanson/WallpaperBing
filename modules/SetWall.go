package modules

import (
	"fmt"
	"syscall"
	"unsafe"
)

// UI Params for windows
const (
	spiSetdeskwallpaper = 0x0014

	uiParam = 0x0000

	pifUpdateINIFile = 0x01
	spifSendChange   = 0x02
)

// user32.dll and its proc
var (
	user32                = syscall.NewLazyDLL("user32.dll")
	systemParametersInfoW = user32.NewProc("SystemParametersInfoW")
)

// SetWallpaper is used to set wallpaper
func SetWallpaper(filename string) error {
	// we will pass the file name as UTF16 string pointer
	filenameUTF16Ptr, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return err
	}
	fmt.Println("Setting wallpaper...")
	// send message to windows to set wallpaper
	systemParametersInfoW.Call(
		uintptr(spiSetdeskwallpaper),              // Message
		uintptr(uiParam),                          // UI Param
		uintptr(unsafe.Pointer(filenameUTF16Ptr)), // User argument e.g. file name
		uintptr(pifUpdateINIFile|spifSendChange),  // we want to update the user profile and set this change into registry
	)

	fmt.Println("Your wallpaper is now set. Go check it!")

	return nil
}
