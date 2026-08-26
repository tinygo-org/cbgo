# cbgo

cbgo implements Go bindings for [CoreBluetooth](https://developer.apple.com/documentation/corebluetooth?language=objc).

## Documentation

For documentation, see the [CoreBluetooth docs](https://developer.apple.com/documentation/corebluetooth?language=objc).

Examples are in the `examples` directory.

## Scope

cbgo aims to implement all functionality that is supported in macOS 10.13.

## Naming

Function and type names in cbgo are intended to match the corresponding CoreBluetooth functionality as closely as possible.  There are a few (consistent) deviations:

* All cbgo identifiers start with a capital letter to make them public.
* Named arguments in CoreBluetooth functions are eliminated.
* Properties are implemented as a pair of functions (`PropertyName` and `SetPropertyName`).

## L2CAP

Connection-oriented channels are supported in both roles.  A central opens a
channel to a peripheral and is handed the result through its peripheral
delegate; a peripheral publishes a PSM and is handed incoming channels through
its peripheral manager delegate.

Reads and writes on an open channel never block.  They return a count of zero
when the channel currently has no data to give or no space to take more, and
report the end of the stream once the far end goes away.  Register an event
handler on the channel to be told when either of those conditions changes,
rather than polling for it.

## Issues

There are definitely memory leaks.  ARC is not compatible with cgo, so objective C memory has to be managed manually.  I didn't see a set of consistent guidelines for object ownership in the CoreBluetooth documentation, so cbgo errs on the side of leaking.  Hopefully this is only an issue for very long running processes!  Any fixes here are much appreciated.
