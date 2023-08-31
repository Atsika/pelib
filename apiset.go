package pelib

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Resolve API Set V6
// Thanks to these references:
// https://lucasg.github.io/2017/10/15/Api-set-resolution/
// https://github.com/ajkhoury/ApiSet

// ResolveApiSet returns the name of the real function host.
func ResolveApiSet(apiSet string, parentName string) string {
	apiNamespace := GetApiSetNamespace()

	// api-ms-win-core-apiquery-l1-1-0.dll -> api-ms-win-core-apiquery-l1-1
	apiToResolve := apiSet[:strings.LastIndex(apiSet, "-")]

	entry := ApiSetSearchForApiSet(apiNamespace, apiToResolve)
	if entry == nil {
		return ""
	}

	hostLibEntry := new(API_SET_VALUE_ENTRY)

	if entry.ValueCount > 1 && parentName != "" {
		hostLibEntry = ApiSetSearchForApiSetHost(entry, parentName, apiNamespace)
	} else if entry.ValueCount > 0 {
		hostLibEntry = GetApiSetNamespaceValueEntry(apiNamespace, entry, 0)
	}

	name := GetApiSetValueEntryValue(apiNamespace, hostLibEntry)

	return name
}

// GetApiSetNamespace returns an API_SET_NAMESPACE structure from PEB
func GetApiSetNamespace() *API_SET_NAMESPACE {
	return GetPEB().ApiSetMap
}

func ApiSetSearchForApiSet(apiNamespace *API_SET_NAMESPACE, apiToResolve string) *API_SET_NAMESPACE_ENTRY {
	lower := strings.ToLower(apiToResolve)
	hashKey := uint32(0)

	for i := 0; i < len(lower); i++ {
		hashKey = hashKey*apiNamespace.HashFactor + uint32(lower[i])
	}

	// binary search
	low := uint32(0)
	middle := uint32(0)
	high := apiNamespace.Count - 1

	hashEntry := new(API_SET_HASH_ENTRY)

	for high >= low {
		middle = (high + low) >> 1
		hashEntry = GetApiSetHashEntry(apiNamespace, middle)

		if hashKey < hashEntry.Hash {
			high = middle - 1
		} else if hashKey > hashEntry.Hash {
			low = middle + 1
		} else {
			break
		}
	}

	// not found
	if high < low {
		return nil
	}

	foundEntry := GetApiSetNamespaceEntry(apiNamespace, hashEntry.Index)
	name := GetApiSetValueName(apiNamespace, foundEntry)

	// equivalent to truncate after last hyphen
	if strings.HasPrefix(lower, strings.ToLower(name)) {
		return nil
	}

	return foundEntry
}

func ApiSetSearchForApiSetHost(entry *API_SET_NAMESPACE_ENTRY, apiToResolve string, apiNamespace *API_SET_NAMESPACE) *API_SET_VALUE_ENTRY {

	foundEntry := GetApiSetNamespaceValueEntry(apiNamespace, entry, 0)

	high := entry.ValueCount - 1
	if high == 0 {
		return foundEntry
	}

	apiSetHostEntry := new(API_SET_VALUE_ENTRY)

	for low := uint32(1); low <= high; {
		middle := (low + high) >> 1
		apiSetHostEntry = GetApiSetNamespaceValueEntry(apiNamespace, entry, middle)

		switch name := GetApiSetValueEntryName(apiNamespace, apiSetHostEntry); {
		case apiToResolve == name:
			return GetApiSetNamespaceValueEntry(apiNamespace, entry, middle)
		case apiToResolve < name:
			high = middle - 1
		case apiToResolve > name:
			low = middle + 1
		}
	}

	return nil
}

func GetApiSetHashEntry(apiNamespace *API_SET_NAMESPACE, index uint32) *API_SET_HASH_ENTRY {
	return (*API_SET_HASH_ENTRY)(unsafe.Add(unsafe.Pointer(apiNamespace), apiNamespace.HashOffset+index*uint32(unsafe.Sizeof(API_SET_HASH_ENTRY{}))))
}

func GetApiSetNamespaceEntry(apiNamespace *API_SET_NAMESPACE, index uint32) *API_SET_NAMESPACE_ENTRY {
	return (*API_SET_NAMESPACE_ENTRY)(unsafe.Add(unsafe.Pointer(apiNamespace), apiNamespace.EntryOffset+index*uint32(unsafe.Sizeof(API_SET_NAMESPACE_ENTRY{}))))
}

func GetApiSetValueName(apiNamespace *API_SET_NAMESPACE, entry *API_SET_NAMESPACE_ENTRY) string {
	name := (*uint16)(unsafe.Add(unsafe.Pointer(apiNamespace), entry.NameOffset))
	return windows.UTF16PtrToString(name)
}

func GetApiSetNamespaceValueEntry(apiNamespace *API_SET_NAMESPACE, entry *API_SET_NAMESPACE_ENTRY, index uint32) *API_SET_VALUE_ENTRY {
	return (*API_SET_VALUE_ENTRY)(unsafe.Add(unsafe.Pointer(apiNamespace), entry.ValueOffset+index*uint32(unsafe.Sizeof(API_SET_VALUE_ENTRY{}))))
}

func GetApiSetValueEntryValue(apiNamespace *API_SET_NAMESPACE, entry *API_SET_VALUE_ENTRY) string {
	value := (*uint16)(unsafe.Add(unsafe.Pointer(apiNamespace), entry.ValueOffset))
	name := unsafe.Slice(value, entry.ValueLength/2)
	return windows.UTF16ToString(name)
}

func GetApiSetValueEntryName(apiNamespace *API_SET_NAMESPACE, entry *API_SET_VALUE_ENTRY) string {
	value := (*uint16)(unsafe.Add(unsafe.Pointer(apiNamespace), entry.NameOffset))
	name := unsafe.Slice(value, entry.NameLength/2)
	return windows.UTF16ToString(name)
}
