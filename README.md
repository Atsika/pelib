# PElib

PElib (Portable Executable library) is a Go module for parsing PE, including extensive Windows types and structures definitions.

[![Static Badge](https://img.shields.io/badge/made_with-Go-007D9C)](https://go.dev/)
[![Go Report](https://goreportcard.com/badge/github.com/nicocha30/ligolo-ng)](https://goreportcard.com/report/github.com/atsika/pelib)
[![GitHub](https://img.shields.io/github/license/atsika/pelib)](https://choosealicense.com/licenses/mit/)

## Table of contents

- [PElib](#pelib)
  - [Table of contents](#table-of-contents)
  - [Features](#features)
    - [NewDLL](#newdll)
    - [NewProc](#newproc)
  - [Usage](#usage)
  - [Changelog](#changelog)
  - [Used in](#used-in)
  - [License](#license)

## Features

**PElib** exports multiple functions and structures related to PE parsing. The full list can be found here: [pkg.go.dev/github.com/atsika/pelib](https://pkg.go.dev/github.com/atsika/pelib) 

The most notable ones are:

- NewDLL
- NewProc

These functions are [generic functions](https://go.dev/doc/tutorial/generics#add_generic_function) used to resolve module (DLL) handles and function (proc) addresses. They use Go reflection to determine the type of passed parameters.

They can be resolved by hash ([API hashing](https://www.ired.team/offensive-security/defense-evasion/windows-api-hashing-in-malware)), a common technique used in malware development. The hashing algorithm used is [sdbm](http://www.cse.yorku.ca/~oz/hash.html#sdbm).

You can read more about those functions on my blog: [https://blog.atsika.ninja/posts/custom_getmodulehandle_getprocaddress/](https://blog.atsika.ninja/posts/custom_getmodulehandle_getprocaddress/)

### NewDLL

`NewDLL` accepts a _string_ (DLL name) or a _uint32_ (DLL 32-bit hash). It returns a `*windows.DLL`.
It retrieves a pointer to the PEB. Then, it parses the _InLoadOrderModuleList_ by comparing the DLL name or hash with the given parameter.
If the resolution fails, `nil` is returned.

### NewProc

`NewProc` takes a `*windows.DLL` as first parameters and either a _string_ (proc name), a _uint16_ (proc ordinal) or _uint32_ (proc 32-bit hash) as second parameter. It returns a `*windows.Proc`.
It retrieves a pointer to the EXPORT_DIRECTORY of a module. It then parses it using the provided parameter to retrieve the function address.
If the provided parameter is a string, then the resolution is done using binary search. Otherwise, a linear search is done.
If the resolution fails, `nil` is returned.

## Usage

1. Import the module in your project.

```go
import (
    "github.com/atsika/pelib"
)
```

2. Resolve modules and functions.

```go
kernel32 := pelib.NewDLL("kernel32.dll")
fmt.Printf("[string] kernel32 => %#2x\n", kernel32.Handle)

kernel32 = pelib.NewDLL(uint32(0x8f7ee672))
fmt.Printf("[hash]   kernel32 => %#2x\n", kernel32.Handle)

GetProcessHeap := pelib.NewProc(kernel32, "GetProcessHeap")
fmt.Printf("[string]  GetProcessHeap => %#2x\n", GetProcessHeap.Addr())

GetProcessHeap = pelib.NewProc(kernel32, uint16(0x2cc))
fmt.Printf("[ordinal] GetProcessHeap => %#2x\n", GetProcessHeap.Addr())

GetProcessHeap = pelib.NewProc(kernel32, uint32(0x4435efa5))
fmt.Printf("[hash]    GetProcessHeap => %#2x\n", GetProcessHeap.Addr())

hHeap, _ , _ := GetProcessHeap.Call()
```

3. Profit

> An example of a classic process injection technique using this module and API hashing can be found in the [example](/example) folder.

## Changelog

- 21/02/2024: Added TEB definition
- 31/08/2023: Added API sets V6 (Windows 10) resolution
- 21/08/2023: Export some useful functions (GetPEB, GetDosHeader, GetNtHeaders, GetDataDirectory,...)
- 01/08/2023: Initial release accompanying the blog post

## Used in

* [mkwinsyscall fork](https://github.com/atsika/mkwinsyscall)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
