// Package Enchant provides simplified bindings to Enchant spell checking library.
package enchant

/*
#cgo LDFLAGS: -lenchant
#include <string.h>
#include <stdlib.h>
#include <sys/types.h>
#include "enchant/enchant.h"

static char* getString(char ** c, int i) {
    return c[i];
}
*/
import "C"

import (
	"errors"
	"reflect"
	"unsafe"
)

// Enchant is a type that encapsulates Enchant internals
type Enchant struct {
	broker *C.EnchantBroker
}

// Dict encapsulates dictionaries
type Dict struct {
	dict   *C.EnchantDict
}

// NewEnchant creates a new Enchant instance for access
// to the rest of the Enchant API.
//
// The returned value is an Enchant struct.
//
// Example usage:
//
// 		enchant, err := enchant.NewEnchant()
// 		if err != nil {
// 			panic("Enchant error: " + err.Error())
// 		}
// 		defer enchant.Free()
//      fmt.Println(enchant.DictExists("zh"))
//
// Because the Enchant package is a binding to Enchant C library, memory
// allocated by the NewEnchant() call has to be disposed explicitly.
// This is why the above example contains a deferred call to Free().
func NewEnchant() (e Enchant, err error) {
	broker := C.enchant_broker_init()
	e = Enchant{broker}
	// we don't return errors at the moment, but we might in the future
	return e, nil
}

// Free frees the Enchant broker and dictionary, and needs
// to be called when use of the library is no longer needed
// to prevent memory leaks.
func (e *Enchant) Free() {
	C.enchant_broker_free(e.broker)
}

// FreeDict frees a dictionary.
func (e *Enchant) FreeDict(d *Dict) {
	C.enchant_broker_free_dict(e.broker, d.dict)
}


// DictExists wraps enchant_broker_dict_exists.
// It takes a language code name, such as "en_GB", as string
// argument, and it returns whether or not such a dictionary
// is installed on the system.
func (e *Enchant) DictExists(name string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	exists := C.enchant_broker_dict_exists(e.broker, cName)
	return exists > 0
}

// LoadDict wraps enchant_broker_request_dict, and adds
// the loaded dictionary to the Enchant instance.
// It takes a language code name, such as "en_GB", as string
// argument, and it returns a EnchantDict representation
// of this dictionary.
func (e *Enchant) LoadDict(name string) (Dict,error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	dict := C.enchant_broker_request_dict(e.broker, cName)
	if dict == nil {
		return Dict{}, errors.New("Cannot load dictionary")
	}
	return Dict{dict}, nil
}

// Check whether a given word is in the currently loaded dictionary.
// This wraps enchant_dict_check.
// It returns a boolean value: true if the word is in the dictionary,
// false otherwise.
func (d Dict) Check(word string) bool {
	if len(word) == 0 {
		return true
	}

	cWord := (*C.char)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&word)).Data))
	size := C.ssize_t(uintptr(len(word)))

	return C.enchant_dict_check(d.dict, cWord, size) == 0
}

// Suggest words based on the given word.
// This is a wrapper for enchant_dict_suggest.
// It returns a slice of suggestion strings.
func (d Dict) Suggest(word string) (suggestions []string) {
	if len(word) == 0 {
		return suggestions
	}

	cWord := C.CString(word)
	defer C.free(unsafe.Pointer(cWord))

	size := uintptr(len(word))
	s := (*C.ssize_t)(unsafe.Pointer(&size))

	var n int
	nSugg := uintptr(n)
	ns := (*C.size_t)(unsafe.Pointer(&nSugg))

	// get the suggestions; ns will be modified to store the
	// number of suggestions returned
	response := C.enchant_dict_suggest(d.dict, cWord, *s, ns)

	for i := 0; i < int(*ns); i++ {
		ci := C.int(i)
		suggestions = append(suggestions, C.GoString(C.getString(response, ci)))
	}

	return suggestions
}
