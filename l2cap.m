#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#import "bt.h"

// l2cap.m: L2CAP Connection-Oriented Channel support (macOS 10.13+)

// Maps each of a channel's streams back to the channel that owns them, so a
// stream event can be reported to Go against the channel it belongs to.
// Entries are pointer-identity only; nothing here retains a stream.
static NSMapTable *l2cap_stream_map;
static NSLock *l2cap_stream_lock;
static dispatch_once_t l2cap_once;

static void l2cap_map_init(void) {
    dispatch_once(&l2cap_once, ^{
        l2cap_stream_map = [[NSMapTable alloc]
            initWithKeyOptions:NSPointerFunctionsOpaqueMemory | NSPointerFunctionsOpaquePersonality
                  valueOptions:NSPointerFunctionsOpaqueMemory | NSPointerFunctionsOpaquePersonality
                      capacity:4];
        l2cap_stream_lock = [[NSLock alloc] init];
    });
}

void *cb_l2cap_channel_for_stream(void *stream) {
    l2cap_map_init();

    [l2cap_stream_lock lock];
    void *ch = (void *)[l2cap_stream_map objectForKey:(id)stream];
    [l2cap_stream_lock unlock];

    return ch;
}

void cb_prph_open_l2cap_channel(void *prph, uint16_t psm) {
    CBPeripheral *p = (CBPeripheral *)prph;
    [p openL2CAPChannel:(CBL2CAPPSM)psm];
}

uint16_t cb_l2cap_psm(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return ch.PSM;
}

void *cb_l2cap_peer(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return ch.peer;
}

int cb_l2cap_read(void *channel, uint8_t *buf, int maxLen) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    NSInputStream *stream = ch.inputStream;
    if (![stream hasBytesAvailable]) {
        return 0;
    }
    return (int)[stream read:buf maxLength:maxLen];
}

int cb_l2cap_write(void *channel, const uint8_t *data, int len) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    NSOutputStream *stream = ch.outputStream;
    if (![stream hasSpaceAvailable]) {
        return 0;
    }
    return (int)[stream write:data maxLength:len];
}

struct bt_error cb_l2cap_input_stream_error(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    NSInputStream *stream = ch.inputStream;
    if (stream.streamStatus == NSStreamStatusError) {
        return nserror_to_bt_error(stream.streamError);
    }
    return (struct bt_error){0};
}

struct bt_error cb_l2cap_output_stream_error(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    NSOutputStream *stream = ch.outputStream;
    if (stream.streamStatus == NSStreamStatusError) {
        return nserror_to_bt_error(stream.streamError);
    }
    return (struct bt_error){0};
}

int cb_l2cap_input_stream_status(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return (int)ch.inputStream.streamStatus;
}

int cb_l2cap_output_stream_status(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return (int)ch.outputStream.streamStatus;
}

bool cb_l2cap_has_bytes_available(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return [ch.inputStream hasBytesAvailable];
}

bool cb_l2cap_has_space_available(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return [ch.outputStream hasSpaceAvailable];
}

// The streams are serviced on bt_queue, the same queue CoreBluetooth delivers
// its delegate callbacks on, so channel events are serialised with the rest of
// the library and no additional thread is involved.
void cb_l2cap_schedule_streams(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;

    l2cap_map_init();

    [l2cap_stream_lock lock];
    [l2cap_stream_map setObject:(id)ch forKey:(id)ch.inputStream];
    [l2cap_stream_map setObject:(id)ch forKey:(id)ch.outputStream];
    [l2cap_stream_lock unlock];

    ch.inputStream.delegate = bt_dlg;
    ch.outputStream.delegate = bt_dlg;

    CFReadStreamSetDispatchQueue((CFReadStreamRef)ch.inputStream, bt_queue);
    CFWriteStreamSetDispatchQueue((CFWriteStreamRef)ch.outputStream, bt_queue);

    [ch.inputStream open];
    [ch.outputStream open];
}

// The channel is retained when the delegate hands it to Go and is never
// released, matching how every other CoreBluetooth object crossing this
// boundary is owned. Closing only tears down the streams, so the pointer Go
// holds stays valid and every method remains safe to call afterwards.
void cb_l2cap_close(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;

    l2cap_map_init();

    [l2cap_stream_lock lock];
    [l2cap_stream_map removeObjectForKey:(id)ch.inputStream];
    [l2cap_stream_map removeObjectForKey:(id)ch.outputStream];
    [l2cap_stream_lock unlock];

    ch.inputStream.delegate = nil;
    ch.outputStream.delegate = nil;

    [ch.inputStream close];
    [ch.outputStream close];

    CFReadStreamSetDispatchQueue((CFReadStreamRef)ch.inputStream, NULL);
    CFWriteStreamSetDispatchQueue((CFWriteStreamRef)ch.outputStream, NULL);
}

void cb_pmgr_publish_l2cap_channel(void *pmgr, bool encryption) {
    CBPeripheralManager *pm = (CBPeripheralManager *)pmgr;
    [pm publishL2CAPChannelWithEncryption:encryption];
}

void cb_pmgr_unpublish_l2cap_channel(void *pmgr, uint16_t psm) {
    CBPeripheralManager *pm = (CBPeripheralManager *)pmgr;
    [pm unpublishL2CAPChannel:(CBL2CAPPSM)psm];
}
