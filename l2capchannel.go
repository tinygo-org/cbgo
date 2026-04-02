package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import (
	"unsafe"
)

// L2CAPChannel: https://developer.apple.com/documentation/corebluetooth/cbl2capchannel
type L2CAPChannel struct {
	ptr unsafe.Pointer
}

// PSM returns the channel's Protocol/Service Multiplexer identifier.
// PSM: https://developer.apple.com/documentation/corebluetooth/cbl2capchannel/2880155-psm
func (ch L2CAPChannel) PSM() uint16 {
	return uint16(C.cb_l2cap_psm(ch.ptr))
}

// Read reads up to len(buf) bytes from the L2CAP channel's input stream.
// Returns the number of bytes read. Returns 0 if no bytes are currently
// available, or a negative value on error.
func (ch L2CAPChannel) Read(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	return int(C.cb_l2cap_read(ch.ptr, (*C.uint8_t)(unsafe.Pointer(&buf[0])), C.int(len(buf))))
}

// Write writes data to the L2CAP channel's output stream.
// Returns the number of bytes written. Returns 0 if the stream has no space
// available, or a negative value on error.
func (ch L2CAPChannel) Write(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return int(C.cb_l2cap_write(ch.ptr, (*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(len(data))))
}

// HasBytesAvailable returns true if the input stream has bytes available to read.
func (ch L2CAPChannel) HasBytesAvailable() bool {
	return bool(C.cb_l2cap_has_bytes_available(ch.ptr))
}

// HasSpaceAvailable returns true if the output stream has space available for writing.
func (ch L2CAPChannel) HasSpaceAvailable() bool {
	return bool(C.cb_l2cap_has_space_available(ch.ptr))
}

// Close closes the L2CAP channel's input and output streams.
func (ch L2CAPChannel) Close() {
	C.cb_l2cap_close(ch.ptr)
}
