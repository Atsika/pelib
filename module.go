package pelib

import (
	"reflect"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// NewDLL parses PEB's InLoadOrderModuleList to retrieve a DLL handle.
// Pass an empty string or 0 to retrieve the current module.
func NewDLL[T ~string | ~uint32](module T) *windows.DLL {

	dll := new(windows.DLL)

	var modName string
	var modHash uint32

	switch reflect.TypeOf(module).Kind() {
	case reflect.String:
		modName = strings.ToLower(any(module).(string))
	case reflect.Uint32:
		modHash = any(module).(uint32)
	}

	peb := GetPEB()
	head := peb.Ldr.InLoadOrderModuleList

	// current module = first entry
	if modName == "" && modHash == 0 {
		entry := (*LDR_DATA_TABLE_ENTRY)(unsafe.Pointer(head.Flink))
		dll.Handle = windows.Handle(entry.DllBase)
		dll.Name = strings.ToLower(entry.BaseDllName.String())
		return dll
	}

	// search for module
	for next := head.Flink; *next != head; next = next.Flink {
		entry := (*LDR_DATA_TABLE_ENTRY)(unsafe.Pointer(next))
		currentName := strings.ToLower(entry.BaseDllName.String())
		if currentName == modName || hash(currentName) == modHash {
			dll.Handle = windows.Handle(entry.DllBase)
			dll.Name = currentName
			return dll
		}
	}

	// LoadLibrary fallback
	if modName != "" {
		kernel32 := NewDLL(uint32(0x8f7ee672))
		LoadLibrary := NewProc(kernel32, uint32(0xdf2bbc02))
		handle, _, _ := LoadLibrary.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(modName))))
		if handle != 0 {
			dll.Name = modName
			dll.Handle = windows.Handle(handle)
			return dll
		}
	}

	return nil

}
