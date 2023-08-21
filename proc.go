package pelib

import (
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	sizeofUint16 = unsafe.Sizeof(uint16(0))
	sizeofUint32 = unsafe.Sizeof(uint32(0))
)

// NewProc is the reimplementation of GetProcAddress using binary search (3x faster) with a linear search fallback.
// It implements search by name, by ordinal or by hash.
func NewProc[T ~string | ~uint16 | ~uint32](dll *windows.DLL, procedure T) *windows.Proc {

	if dll == nil || dll.Handle == 0 {
		return nil
	}

	var procName string
	var procOrdinal uint16
	var procHash uint32
	var procAddr uintptr

	switch reflect.TypeOf(procedure).Kind() {
	case reflect.String:
		procName = any(procedure).(string)
	case reflect.Uint16:
		procOrdinal = any(procedure).(uint16)
	case reflect.Uint32:
		procHash = any(procedure).(uint32)
	}

	proc := new(windows.Proc)
	proc.Dll = dll
	proc.Name = procName

	module := unsafe.Pointer(dll.Handle)

	dataDir := GetDataDirectory(module, IMAGE_DIRECTORY_ENTRY_EXPORT)
	exportDir := (*IMAGE_EXPORT_DIRECTORY)(unsafe.Add(module, dataDir.VirtualAddress))

	addrOfFunctions := unsafe.Add(module, exportDir.AddressOfFunctions)
	addrOfNames := unsafe.Add(module, exportDir.AddressOfNames)
	addrOfNameOrdinals := unsafe.Add(module, exportDir.AddressOfNameOrdinals)

	if procOrdinal != 0 {
		procOrdinal = procOrdinal - uint16(exportDir.Base)
		rva := *(*uint32)(unsafe.Add(addrOfFunctions, procOrdinal*uint16(sizeofUint32)))
		procAddr = uintptr(module) + uintptr(rva)
		goto Found
	}

	// binary search
	if procName != "" {
		left := uintptr(0)
		right := uintptr(exportDir.NumberOfNames - 1)

		for left != right {
			middle := left + ((right - left) >> 1)
			currentName := windows.BytePtrToString((*byte)(unsafe.Add(module, *(*uint32)(unsafe.Add(addrOfNames, middle*sizeofUint32)))))
			if currentName == procName {
				index := *(*uint16)(unsafe.Add(addrOfNameOrdinals, middle*sizeofUint16))
				procAddr = uintptr(module) + uintptr(*(*uint32)(unsafe.Add(addrOfFunctions, index*uint16(sizeofUint32))))
				goto Found
			} else if currentName < procName {
				left = middle + 1
			} else {
				right = middle - 1
			}
		}
	}

	// linear search
	if procAddr == 0 {
		for i := uintptr(0); i < uintptr(exportDir.NumberOfNames); i++ {
			currentName := windows.BytePtrToString((*byte)(unsafe.Add(module, *(*uint32)(unsafe.Add(addrOfNames, i*sizeofUint32)))))
			if currentName == procName || Hash(currentName) == procHash {
				index := *(*uint16)(unsafe.Add(addrOfNameOrdinals, i*sizeofUint16))
				procAddr = uintptr(module) + uintptr(*(*uint32)(unsafe.Add(addrOfFunctions, index*uint16(sizeofUint32))))
				goto Found
			}
		}
	}

	return nil

Found:
	// trick to set unexported addr field in windows.Proc structure
	addr := reflect.ValueOf(proc).Elem().FieldByName("addr")
	reflect.NewAt(addr.Type(), unsafe.Pointer(addr.UnsafeAddr())).Elem().Set(reflect.ValueOf(procAddr))

	return proc

}
