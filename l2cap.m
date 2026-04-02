#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#import "bt.h"

// l2cap.m: L2CAP Connection-Oriented Channel support (macOS 10.13+)

void cb_prph_open_l2cap_channel(void *prph, uint16_t psm) {
    CBPeripheral *p = (CBPeripheral *)prph;
    [p openL2CAPChannel:(CBL2CAPPSM)psm];
}

uint16_t cb_l2cap_psm(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return ch.PSM;
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

bool cb_l2cap_has_bytes_available(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return [ch.inputStream hasBytesAvailable];
}

bool cb_l2cap_has_space_available(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    return [ch.outputStream hasSpaceAvailable];
}

void cb_l2cap_close(void *channel) {
    CBL2CAPChannel *ch = (CBL2CAPChannel *)channel;
    [ch.inputStream close];
    [ch.outputStream close];
}

void cb_pmgr_publish_l2cap_channel(void *pmgr, bool encryption) {
    CBPeripheralManager *pm = (CBPeripheralManager *)pmgr;
    [pm publishL2CAPChannelWithEncryption:encryption];
}

void cb_pmgr_unpublish_l2cap_channel(void *pmgr, uint16_t psm) {
    CBPeripheralManager *pm = (CBPeripheralManager *)pmgr;
    [pm unpublishL2CAPChannel:(CBL2CAPPSM)psm];
}
