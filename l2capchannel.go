package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import (
	"errors"
	"io"
	"unsafe"
)

// ErrL2CAPRead is returned by L2CAPChannel.Read when the input stream fails
// but reports no underlying error.
var ErrL2CAPRead = errors.New("cbgo: l2cap input stream read failed")

// ErrL2CAPWrite is returned by L2CAPChannel.Write when the output stream fails
// but reports no underlying error.
var ErrL2CAPWrite = errors.New("cbgo: l2cap output stream write failed")

// StreamStatus describes the state of one of an L2CAP channel's streams.
// StreamStatus: https://developer.apple.com/documentation/foundation/stream/status
type StreamStatus int

const (
	StreamStatusNotOpen StreamStatus = 0
	StreamStatusOpening StreamStatus = 1
	StreamStatusOpen    StreamStatus = 2
	StreamStatusReading StreamStatus = 3
	StreamStatusWriting StreamStatus = 4
	StreamStatusAtEnd   StreamStatus = 5
	StreamStatusClosed  StreamStatus = 6
	StreamStatusError   StreamStatus = 7
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

// InputStreamStatus returns the current state of the channel's input stream.
func (ch L2CAPChannel) InputStreamStatus() StreamStatus {
	return StreamStatus(C.cb_l2cap_input_stream_status(ch.ptr))
}

// OutputStreamStatus returns the current state of the channel's output stream.
func (ch L2CAPChannel) OutputStreamStatus() StreamStatus {
	return StreamStatus(C.cb_l2cap_output_stream_status(ch.ptr))
}

// Read reads up to len(buf) bytes from the L2CAP channel's input stream.
// It returns the number of bytes read, and 0 with a nil error if no bytes are
// currently available. It returns io.EOF once the remote end has closed the
// stream. On failure it returns 0 and a non-nil error.
func (ch L2CAPChannel) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	n := int(C.cb_l2cap_read(ch.ptr, (*C.uint8_t)(unsafe.Pointer(&buf[0])), C.int(len(buf))))
	if n < 0 {
		e := C.cb_l2cap_input_stream_error(ch.ptr)
		if err := btErrorToNSError(&e); err != nil {
			return 0, err
		}
		return 0, ErrL2CAPRead
	}
	if n == 0 && ch.InputStreamStatus() == StreamStatusAtEnd {
		return 0, io.EOF
	}
	return n, nil
}

// Write writes data to the L2CAP channel's output stream.
// It returns the number of bytes written, and 0 with a nil error if the stream
// currently has no space available. It returns io.ErrClosedPipe once the
// stream can no longer accept data. On failure it returns 0 and a non-nil
// error.
func (ch L2CAPChannel) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	n := int(C.cb_l2cap_write(ch.ptr, (*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(len(data))))
	if n < 0 {
		e := C.cb_l2cap_output_stream_error(ch.ptr)
		if err := btErrorToNSError(&e); err != nil {
			return 0, err
		}
		return 0, ErrL2CAPWrite
	}
	if n == 0 {
		switch ch.OutputStreamStatus() {
		case StreamStatusAtEnd, StreamStatusClosed:
			return 0, io.ErrClosedPipe
		}
	}
	return n, nil
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
